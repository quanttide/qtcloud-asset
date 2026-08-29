# 阶段七文件夹分享上线记录

> 验证日期：2026-08-28
>
> 状态：已完成
> 正式入口：`https://asset.cloud.quanttide.com`
> API 入口：`https://api.quanttide.com/qtcloud-asset`

## Provider 和 RDS

- 现实需求：文件夹分享记录可以持久化，并由生产 Provider 提供只读访问。
- 判定标准：PostgreSQL `folder_shares` 迁移成功；生产函数通过 VPC 访问 RDS；迁移开关在启动后移除；Provider 健康检查正常。
- 存证：RDS 数据库为 `qtcloud_asset`，分享表迁移使用 `folder-shares-postgres-v1`，生产函数配置为 PostgreSQL RDS 分享存储和 `qtcloud-asset-studio` 白名单。直连 `/health` 返回 200，未知分享返回 404；本次生产函数 codeChecksum 为 `16728507634091352455`。
- 状态：✅ 符合

## API 网关路由

- 现实需求：Studio 通过正式 API 网关访问分享能力，不暴露函数计算内部地址。
- 判定标准：分享创建、分享列表、分享读取、对象列表、对象链接、ZIP 下载和撤销七条路由均在 `RELEASE` 阶段部署，并映射到对应 Provider 路径。
- 存证：
  - `POST /shares`：`5f57f022b01044f79758a48ff80bd22a`
  - `GET /shares`：`f89c29695f6449d2896ca96742b3081d`
  - `GET /shares/{token}`：`7079455ebdba446d95cd7d9f29b58411`
  - `GET /shares/{token}/objects`：`a43fc641d01d4874a20df1fc9d63b40e`
  - `GET /shares/{token}/object-url`：`fcef0ad3b74844fb8ed0736c0c58beca`
  - `GET /shares/{token}/download`：`76b65d990076487484724d8a7c7c5470`，返回类型为 `BINARY`，后端超时为 300 秒
  - `DELETE /shares/{token}`：`e85e6e97c8e84fc1a1a7f08aa0ffe9c7`
  - 七条 API 均为 `RELEASE/DEPLOYED`，并映射到函数计算后端；下载 API 的最新 `RELEASE` 生效版本为 `20260828215305552`。
- 状态：✅ 符合

## 认证和 CORS

- 现实需求：分享写操作仍受登录态保护，公开读取路径可以被分享页调用。
- 判定标准：未登录列表、创建和撤销返回 401；未知 token 读取返回 404；正式 Studio 来源的 DELETE 预检返回 200，并返回精确来源、必要方法和请求头。
- 存证：网关线上验收结果为 `GET /shares` 401、`POST /shares` 401、`DELETE /shares/{token}` 401、`GET /shares/not-a-real-token` 404、`GET /shares/not-a-real-token/download` 404，下载接口返回 Provider 的 `{"error":"Not Found","message":"share not found"}`，未再返回网关 `I404NF`。下载接口预检返回 `Access-Control-Allow-Origin: https://asset.cloud.quanttide.com`、`Access-Control-Allow-Methods: GET,POST,DELETE,OPTIONS`、`Access-Control-Allow-Headers: Content-Type,Authorization,X-CSRF-Token` 和 `Access-Control-Allow-Credentials: true`。分享 CORS 插件 `qtcloud_asset_share_cors_allowlist` 已绑定七条 API。
- 状态：✅ 符合

## Studio 发布

- 现实需求：正式 Studio 可以进入分享页面，并调用分享 API。
- 判定标准：线上构建包含分享客户端和公开分享页面；正式入口及分享深链接可以返回 Flutter 应用入口；构建产物不包含内部函数地址、凭证标记或临时文件。
- 存证：线上 `main.dart.js` 返回 200，ETag 为 `B1B65D2DCCA78EDD75778DA883EDD2F1`，与当前构建一致；分享链接使用 `https://asset.cloud.quanttide.com/#/share/{token}`，其请求部分返回应用入口内容，Hash 片段由浏览器交给 Flutter 解析，不会被 OSS 当成对象路径。Flutter 测试 56 个用例全部通过，`flutter analyze` 无问题，静态构建安全扫描发现 0 个禁止标记和 0 个禁止产物。
- 状态：✅ 符合

## 回滚和未覆盖项

- 现实需求：发布异常时可以恢复上一版 Provider。
- 判定标准：上一版和当前版发布包均保留，并有 SHA-256 校验值。
- 存证：上一版回滚包为 `src/provider/dist/provider-20260828-182840-share-download-all/qtcloud-asset-provider-go-linux-amd64-share-download-all-20260828.zip`，SHA-256 为 `46E0F66BB4987386A05BCB3DC4BA3746F927D556A8F6E980925B4D3D226BC33C`；当前包为 `src/provider/dist/provider-20260828-183648-share-hash-route/qtcloud-asset-provider-go-linux-amd64-share-hash-route-20260828-183649.zip`，SHA-256 为 `E7E781F1D11C6686E86441D0B6A872EE74CD3AB40E18BBB809E157864F995E7D`。
- 状态：✅ 符合

- 现实需求：完整验证管理员创建分享、公开浏览文件夹和撤销分享的真实生命周期。
- 判定标准：使用已授权管理员会话创建一个真实测试分享，访问对象列表和对象链接，再撤销并确认访问失效。
- 存证：本次执行环境没有认证验收脚本所需的登录环境变量，因此未猜测或绕过管理员凭证；已完成健康检查、未知 token、未登录写操作和 CORS 验收。真实生命周期验收应在具备管理员会话后补做。
- 状态：⏸️ 待实现
