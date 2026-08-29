# CHANGELOG

## [Unreleased]

### Changed
- 本地登录和管理员邀请从邮箱主键改为账号主键，保留旧 `email` 请求体和 `LOCAL_AUTH_EMAIL` 作为兼容 fallback
- 明确对象链接仅支持公开桶；私密桶和 `quanttide-terraform-state` 不生成访问链接

## [0.1.1] - 2026-08-27

### Added
- CLI 新增 `oss` 子命令（list / ls / url），复用 Provider 只读接口管理 OSS
- Provider 列表接口支持排序（`sort`/`order`）与对象分页（`limit`/`marker`）
- Provider 对象列表支持 `prefix` 过滤（OSS 原生）
- CLI `oss list` / `oss ls` 支持 `--sort` / `--order` / `--limit` / `--prefix` 参数
- Provider 新增账号门禁，覆盖 `/auth/login`、`/auth/callback`、`/auth/me` 和 `/auth/logout`
- Provider 新增 `viewer` / `admin` 两级角色，保护桶列表、对象列表、对象链接和管理接口
- Provider 新增本地账号密码登录模式，使用 PBKDF2-SHA256 哈希和服务端 `HttpOnly` 会话 Cookie
- Provider 新增管理员用户管理接口，支持邀请、禁用、改角色和撤销会话
- Provider 新增结构化审计日志 `qtcloud_asset_audit`，便于函数计算 stdout 进入 SLS 后持久查询
- Studio 新增登录态初始化、账号密码登录页、当前用户展示、退出登录和管理员入口
- QA 新增阶段五线上验收记录、阶段六回滚准备记录和 `qtcloud-asset` 旧桶对象清单

### Changed
- Studio 桶列表排序调整为创建时间和桶名两个独立开关，支持四种组合状态
- Studio 对象列表补齐续页逻辑，继续请求 Provider 返回的 continuation marker
- CLI `oss` 子命令对齐认证后 Provider 的 401、403、404 和 429 错误语义
- `README.md`、`ROADMAP.md`、`docs/prd/*` 和 `docs/qa/*` 对齐新域名、账号门禁和验收状态

### Fixed
- 修复对象访问链接经 API 网关时 `name`、`key`、`expires` 参数透传不完整的问题
- 修复线上 API 网关 CORS 仍返回 `*` 的问题，改为正式入口和兼容入口精确来源
- 修复管理员私密桶对象元数据访问缺少最小只读 RAM 权限的问题

### Security
- `/buckets`、`/buckets/{name}/objects` 和 `/buckets/{name}/object-url` 默认需要登录
- `viewer` 默认隐藏 `-private` 桶和 `quanttide-terraform-state`，私密桶对象能力返回 403
- 管理写接口校验 `Origin`，未登记来源返回 403
- 登录接口增加限流，避免本地账号密码模式被暴力尝试
- 旧域名 `asset.quanttide.com` 保留为兼容入口，本轮不执行旧入口下线、跳转或删除

## [0.1.0] - 2026-08-21

### Added
- Provider 采用 Go 技术栈，`OssAdapter` 作为 `SourceAdapter` 首个实现
- 只读端点 `/buckets`：列出 OSS 桶（含 region、存储类型、创建时间）
- 只读端点 `/buckets/{name}/objects`：列出桶内对象
- 端点 `/buckets/{name}/object-url`：生成对象访问链接，私密桶走签名 URL、公开桶走永久直链
- Studio 桶列表页：按用途分类筛选、搜索、创建时间排序
- Studio 文件浏览页：目录层级下钻、搜索、日期/大小排序
- Studio 复制访问链接：有效期自选（1 天 / 7 天 / 30 天）

## [0.0.1] - 2026-04-17

### Added
- CLI 模块：三层架构（入口层 cli.py、配置层 workflow.py、操作层 file_operator.py）
- 契约系统：Pydantic 模型定义，自动识别契约目录
- Studio 客户端：Flutter Web 应用，数字资产可视化管理界面
- Provider 服务端：FastAPI 服务，对接阿里云函数计算和 OSS
- Docker 部署配置：Studio 和 Provider 的容器化部署
- 完整文档体系：BRD、PRD、IXD、ADD、QA、用户文档
- AI 执行审核契约：强制执行声明和触发条件
- 原子技能编排需求文档：商业需求和产品需求
- 集成测试和单元测试：CLI 模块测试覆盖率达 99%
- 阿里云基础设施代码：函数计算、OSS、VPC 等定义

### Changed
- 重构 CLI 为 --input/--contract/--output 模式
- 重构契约系统：使用 Contract 类和 ContractSchema
- 重命名 .quanttide/ai 为 .quanttide/agent

### Fixed
- 修复 AI 契约：添加强制执行声明和触发条件
