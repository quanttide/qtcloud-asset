# ROADMAP

> 格式：Keep a Changelog + checkbox 任务清单。

## [0.1.0] — 可视化对象存储 MVP

### Added
- [x] Provider 采用 Go 技术栈，`OssAdapter` 作为 `SourceAdapter` 首个实现
- [x] 只读端点 `/buckets`：列出 OSS 桶（含 region、存储类型、创建时间）
- [x] 只读端点 `/buckets/{name}/objects`：列出桶内对象
- [x] 端点 `/buckets/{name}/object-url`：仅生成公开桶永久直链，私密桶拒绝对象链接请求
- [x] Studio 桶列表页：按用途分类筛选、搜索、创建时间排序
- [x] Studio 文件浏览页：目录层级下钻、搜索、日期/大小排序
- [x] Studio 复制访问链接：公开桶直接复制永久直链，私密桶不提供链接分享
- [x] CLI 新增 `oss` 子命令（list / ls / url），复用 Provider 只读接口

### Changed
- [x] 清理 `src/provider/README.md` 遗留的 git 冲突标记
- [x] 修订 `code/contract.yaml`：Provider 技术栈统一为 Go，补充 `aliyun-oss-go-sdk`
- [x] `oss-integration.md` 完成阶段一基线口径对齐

## [0.1.1] — 账号与门禁

### Added
- [x] 身份源选型：飞书/Lark SSO 优先，必要时启用邀请码/邮箱验证码
- [x] 登录态与会话：`/auth/login`、`/auth/callback`、`/auth/logout`、`/auth/me`
- [x] 本地账号密码登录：内测使用 `AUTH_MODE=local`、PBKDF2-SHA256 密码哈希和服务端会话
- [x] 角色体系：`viewer` / `admin`
- [x] 私密桶仅授权人员可见，公开桶继续只读
- [x] 私密桶对象只展示元数据，不生成访问链接
- [x] 访问审计：用户、IP、bucket、action、result
- [x] 审计日志持久化：Provider 输出结构化 `qtcloud_asset_audit`，函数计算 stdout 接入 SLS
- [x] 管理员邀请 / 禁用用户
- [x] Studio 登录态展示与退出登录
- [x] 认证态 smoke test 与线上浏览器主路径验收

### Security
- [x] `/buckets`、`/objects`、`/object-url` 全部加鉴权
- [x] Provider CORS 收口到已登记来源，必要时启用 credentials
- [x] 平台 API 网关 CORS 收口到已登记来源
- [x] 会话 cookie、安全头、管理写接口 Origin 校验
- [x] 登录与敏感接口限流

## [0.1.2] — 身份源与发布收口

### Added
- [ ] 接入平台 SSO 或真实身份源，替换默认 SSO 占位实现
- [ ] 将线上登录后浏览器验收脚本固化为可重复执行的 smoke test
- [ ] 补齐 GitHub Flow 发布记录：PR、CI、release audit、tag 和 GitHub Release
- [ ] PR 检查确认不包含 `plan.md`、`plans.md` 或 `plan-a.md`

### Changed
- [x] 本地登录和管理员邀请从邮箱主键改为账号主键，保留旧邮箱配置兼容 fallback
- [ ] 决定 `qtcloud-asset/provider/` 是否迁到专用 Provider 制品桶或制品仓库
- [ ] 明确旧入口 `asset.quanttide.com` 的保留、跳转或下线策略
- [ ] 同步 `CHANGELOG.md`、`README.md`、`src/provider/README.md` 和 `docs/prd/oss-integration.md`

### Security
- [ ] 明确本地账号密码登录的内测退场条件和管理员账号轮换流程
- [x] 确认私密桶和 `quanttide-terraform-state` 不生成对象访问链接
- [ ] 明确 API 网关、DNS、CDN、证书、RAM 权限和 OSS 删除操作的单独授权清单
- [ ] 将长期 OSS AK/SK 使用路径迁移到 RAM 角色、临时凭证或 KMS 管控链路

## [0.2.0] — 补齐与加固

### Added
- [ ] 桶/对象列表自动刷新
- [ ] `ListObjects` 分页，突破单次 1000 个对象上限
- [ ] 文件上传、下载、删除等写操作

### Changed
- [x] 在 `.quanttide/asset/contract.yaml` 登记 35 个 OSS 桶为资产条目

### Security
- [x] 私密桶不提供链接分享；未来如需访问，另行设计授权与审计机制

## [0.3.0] — 全景与多源

### Added
- [ ] 桶/对象注册为契约资产，建立跨平台关联关系
- [ ] 接入 GitHub 适配器
- [ ] 接入飞书适配器
- [ ] 资产全景总览：按平台 + 用途分组的统一视图

## [0.0.1] — 已发布

### Added
- [x] CLI 模块：run / scan / validate / config / version
- [x] 契约系统：Pydantic 模型，自动识别契约目录
- [x] Studio 客户端与 Provider 服务端骨架
- [x] 文档体系：BRD、PRD、IXD、QA、用户文档
- [x] 阿里云基础设施代码（函数计算、OSS、VPC）
