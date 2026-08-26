# 阶段五线上验收记录

> 验证日期：2026-08-26
>
> 状态：已完成
> 正式入口：`https://asset.cloud.quanttide.com`

## 已验证

### 正式入口可访问

- 现实需求：用户可以通过正式域名打开 Studio。
- 判定标准：页面返回 HTTP 200，标题为「量潮资产云」。
- 存证：`curl -I -L https://asset.cloud.quanttide.com`；浏览器标题为「量潮资产云」。
- 状态：✅ 符合

### Provider 基础链路

- 现实需求：Studio 可以访问 Provider。
- 判定标准：`/health` 返回 HTTP 200；账号门禁发布后，未登录 `/buckets` 返回 HTTP 401。
- 存证：`curl` 请求 `https://api.quanttide.com/qtcloud-asset/health` 返回 200；`Origin: https://asset.cloud.quanttide.com` 请求 `/buckets` 返回 401 `authentication required`。
- 状态：✅ 符合

### 私密桶对象保护

- 现实需求：私密桶不暴露对象内容和访问链接。
- 判定标准：未登录私密桶对象列表返回 HTTP 401；登录后权限不足应返回 HTTP 403；Studio 不显示无权限对象浏览和复制链接入口。
- 存证：账号门禁发布后，`/buckets/qtadmin-private/objects` 未登录返回 401 `authentication required`；临时签名 `viewer` 会话访问 `/buckets/qtadmin-private/objects?limit=1` 返回 403 `bucket object listing is disabled`，访问 `/admin/users` 返回 403 `admin role required`；`viewer` 桶列表返回 44 个桶且不包含 `-private` 桶和 `quanttide-terraform-state`。
- 状态：✅ 未登录、`viewer` 权限不足和 `admin` 私密桶对象列表路径均符合

## 仓库侧修复

### 排序状态

- 现实需求：创建时间排序和桶名排序可以独立关闭、开启和切换方向。
- 判定标准：支持两项都关闭、单独开启、同时开启；排序条件变化后回到第一页。
- 存证：`src/studio/lib/main.dart` 的 `BucketSortMode` 和 `BucketListView`；`src/studio/test/widget_test.dart` 的四态排序与分页测试。
- 状态：✅ 符合

### 对象续页

- 现实需求：大桶浏览不能静默遗漏超过单页上限的对象。
- 判定标准：Studio 沿 Provider 返回的 `next_marker` 拉取所有续页。
- 存证：`fetchAllObjectPages` 及其 Flutter 测试。
- 状态：✅ 符合

### 对象 URL 编码

- 现实需求：对象名包含空格或保留字符时，复制的公开链接仍可用。
- 判定标准：对象 key 按 URL 路径编码，保留目录分隔符。
- 存证：`src/provider/internal/repository/oss_sort_test.go` 的 URL 编码测试。
- 状态：✅ 符合

### 线上发布刷新

- 现实需求：仓库修复后的 Studio bundle 已重新上线。
- 判定标准：正式域名 `main.dart.js` 与最新构建一致。
- 存证：`https://asset.cloud.quanttide.com/main.dart.js` 返回 `ETag: "3CB211BCEC3A1AF0E2910383AD6BF2D0"`，`Last-Modified: Wed, 26 Aug 2026 07:42:32 GMT`；`index.html` 仍为 `F79DEB9E8A3CB1B40066B0B881C415BC`。
- 状态：✅ 符合

### Studio 覆盖率

- 现实需求：Studio 质量门禁达到契约要求。
- 判定标准：`flutter test --coverage` 的覆盖率不低于 80%。
- 存证：`flutter test --coverage`，当前结果 `LH:838 / LF:931`，约 `90.01%`。
- 状态：✅ 符合

### 网关 CORS 来源收口

- 现实需求：API 网关和 Provider 精确来源白名单保持一致，不再向任意来源放行业务请求。
- 判定标准：正式入口和兼容入口精确回显来源；未登记来源返回 403；OPTIONS 预检允许必要方法和请求头。
- 存证：已创建 CORS 插件 `qtcloud_asset_cors_allowlist`（`abd87bdc91ea48758074152310a16a10`）并绑定到既有 7 个 `qtcloud-asset-*` API 和新增 9 个 `/auth/*`、`/admin/users*` API；`https://asset.cloud.quanttide.com` 和 `https://asset.quanttide.com` 回显对应 `Access-Control-Allow-Origin`，`https://evil.example.com` 返回 403，OPTIONS 预检返回 200。
- 状态：✅ 符合

### 账号门禁线上发布

- 现实需求：敏感 API 不再匿名暴露，Studio 登录态路由不再在 API 网关层 404。
- 判定标准：`/auth/me`、`/admin/users`、`/buckets` 未登录返回 401；`POST /auth/login` 本地账号密码登录返回 200 并下发 `HttpOnly; Secure; SameSite=None` Cookie；登录后 `/auth/me`、`/buckets`、`/admin/users` 和公开桶对象列表返回 200；`/auth/logout` 返回 204 后旧 Cookie 失效；管理员 `PATCH` 路由不被 FC 触发器拦截。
- 存证：Provider 账号门禁包最初以 `provider/qtcloud-asset-provider-go-linux-amd64-auth-logout-unixmode-20260826-173715.zip` 发布，函数 codeChecksum 为 `2120679429903001663`；后续审计日志增强包 `provider/qtcloud-asset-provider-go-linux-amd64-audit-json-unixmode-20260826-194042.zip` 已发布，当前函数 codeChecksum 为 `12722312677886720452`。CloudAPI 已为 `/auth/*`、`/admin/users*`、`/buckets*` 相关路由补充 `Cookie` 和 `Origin` Header 转发映射；FC HTTP 触发器 methods 已包含 `PATCH`。线上回归结果：未登录 `/auth/me` 返回 401；本地管理员账号登录返回 200；登录后 `/auth/me`、`/buckets`、`/admin/users` 和公开桶对象列表返回 200；`/auth/logout` 返回 204；退出后旧 Cookie 再访问 `/auth/me` 返回 401；未登记来源返回 403。
- 状态：✅ 网关、后端门禁、本地账号登录链路、`viewer` 权限不足路径、`admin` 私密桶对象列表和审计日志持久化均符合

### 私密桶对象列表权限

- 现实需求：`admin` 可以查看私密桶对象元数据，`viewer` 仍不能访问私密桶对象级能力。
- 判定标准：`admin` 访问 `-private` 桶和 `quanttide-terraform-state` 对象列表返回 HTTP 200；`viewer` 访问同一路径返回 HTTP 403。
- 存证：已创建并挂载自定义策略 `QtcloudAssetProviderPrivateBucketRead` 到 `qtcloud-asset-provider-role`，只授予 `qtadmin-private`、`qtclass-private`、`qtcloud-private`、`qtconsult-private`、`qtdata-private`、`qtrecruit-private` 和 `quanttide-terraform-state` 的 `oss:ListObjects` 与 `oss:GetObject`，不包含写入、删除、ACL 或权限变更。线上回归结果：`admin` 访问 `/buckets/qtadmin-private/objects?limit=1`、`/buckets/qtcloud-private/objects?limit=1`、`/buckets/quanttide-terraform-state/objects?limit=1` 均返回 200；此前已验证 `viewer` 访问私密桶对象列表返回 403。
- 状态：✅ 符合

### 对象访问链接网关参数映射

- 现实需求：Studio 的「复制链接」可以经 API 网关生成对象访问链接。
- 判定标准：未登录请求返回 HTTP 401；管理员登录后，请求 `key` 和 `expires` 参数可透传到 Provider 并返回 HTTP 200。
- 存证：已修改并发布 CloudAPI `qtcloud-asset-object-url-v2`（`b871f9ef05084a3d8e6d3c2eb9b15add`），补齐 `name` Path 参数和 `key`、`expires` Query 参数映射，RELEASE 生效版本为 `20260826185323691`。线上回归结果：未登录访问 `/buckets/qtcloud-asset-studio/object-url?key=index.html&expires=600` 返回 401；本地管理员账号登录后同一路径返回 200，响应包含 `bucket=qtcloud-asset-studio`、`key=index.html` 和公开 OSS URL。
- 状态：✅ 符合

### 真实浏览器控制台验收

- 现实需求：正式 Studio 在浏览器中无关键资源 404、CORS 错误或证书错误。
- 判定标准：未登录页面只触发预期 `/auth/me` 401；管理员会话页面可正常请求 `/auth/me`、`/health` 和 `/buckets`；无运行时异常。
- 存证：干净 Chrome headless 会话访问 `https://asset.cloud.quanttide.com`，页面标题为「量潮资产云」；未登录 API 请求仅 `/qtcloud-asset/auth/me` 返回 401；临时签名 `admin` 会话下 `/auth/me`、`/health`、`/buckets` 均返回 200；未发现关键资源 404、CORS 或证书错误。控制台仅有 Flutter 模板的 `apple-mobile-web-app-capable` 过期提示，以及未登录 401 的预期资源日志。
- 状态：✅ 符合

### 登录后浏览器路径验收

- 现实需求：正式 Studio 登录后可完成桶列表、搜索、排序、对象浏览和退出登录主路径。
- 判定标准：真实浏览器使用本地管理员账号登录成功；页面显示当前用户和桶列表；分类和分页可点击；可搜索 `qtcloud-asset-studio`；排序按钮可切换；可进入公开桶对象列表并看到 `index.html` / `assets` / `main.dart.js` 等对象；退出登录后返回登录面板；除预期 `/auth/me` 401 外无 HTTP 4xx/5xx、关键资源 404、CORS 或证书错误。
- 存证：Chrome headless 执行 `npx --yes playwright test --config tmp-playwright/playwright.config.js`，环境变量注入 `QTCLOUD_ASSET_LOGIN_EMAIL`、`QTCLOUD_ASSET_LOGIN_PASSWORD` 和 `PLAYWRIGHT_CHROMIUM_EXECUTABLE_PATH`，结果 `1 passed (11.5s)`。脚本 `tmp-playwright/phase5-ui.spec.js` 覆盖登录、桶列表、分类、分页、桶名搜索、排序按钮、公开桶 `qtcloud-asset-studio` 对象浏览和退出登录。
- 状态：✅ 符合

### 审计日志持久化

- 现实需求：线上敏感行为可以跨函数重启持久查询，至少覆盖登录、登出、拒绝访问、对象访问和管理动作。
- 判定标准：Provider 输出结构化审计 JSON；FC `logConfig` 绑定 SLS；线上触发拒绝访问后可在 SLS 查询到 `qtcloud_asset_audit` 事件。
- 存证：Provider 当前 codeChecksum 为 `12722312677886720452`；函数 `logConfig.project` 为 `serverless-cn-hangzhou-964290a1-c4dc-5463-ab76-144fc28dd906`，`logConfig.logstore` 为 `default-logs`；线上 `/health` 返回 200，未登录 `/buckets` 返回 401；SLS 查询 `auth_failed` 返回结构化消息 `{"event":"qtcloud_asset_audit","action":"auth_failed","target":"/buckets","result":"denied"...}`。发布中曾出现一次 Windows zip 权限位不被 FC 识别导致 `CAFilePermission`，已立即回滚到上一版可用包，并用 Unix zip 权限位 `0100755` 重新发布成功。
- 状态：✅ 符合

## 后续增强项

### 平台 SSO 入口

- 现状：`POST /qtcloud-asset/auth/login` 本地账号密码登录已可用；`GET /qtcloud-asset/auth/login` 仍是平台 SSO 占位入口，未接入飞书/Lark 或平台 SSO。
- 影响：本地账号内测链路和阶段五线上验收已完成；真实身份源接入作为后续增强，不阻断当前验收。
- 后续：单独规划平台 SSO/真实身份源接入，确认身份提供方、回调地址、用户映射和回滚策略后再执行。

### 旧入口 DNS

- 现状：`asset.quanttide.com` 保持兼容入口，旧 DNS 记录未删除。
- 后续：确认兼容周期和跳转策略后，再单独审批删除或调整。
