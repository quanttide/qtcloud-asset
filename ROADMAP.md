# ROADMAP

> 格式：Keep a Changelog + checkbox 任务清单。

## [0.1.0] — 可视化对象存储 MVP

### Added
- [x] Provider 采用 Go 技术栈，`OssAdapter` 作为 `SourceAdapter` 首个实现
- [x] 只读端点 `/buckets`：列出 OSS 桶（含 region、存储类型、创建时间）
- [x] 只读端点 `/buckets/{name}/objects`：列出桶内对象
- [x] 端点 `/buckets/{name}/object-url`：生成对象访问链接，私密桶走签名 URL、公开桶走永久直链
- [x] Studio 桶列表页：按用途分类筛选、搜索、创建时间排序
- [x] Studio 文件浏览页：目录层级下钻、搜索、日期/大小排序
- [x] Studio 复制访问链接：有效期自选（1 天 / 7 天 / 30 天）
- [x] CLI 新增 `oss` 子命令（list / ls / url），复用 Provider 只读接口

### Changed
- [x] 清理 `src/provider/README.md` 遗留的 git 冲突标记
- [x] 修订 `code/contract.yaml`：Provider 技术栈统一为 Go，补充 `aliyun-oss-go-sdk`
- [x] `oss-integration.md` 完成阶段一基线口径对齐

## [0.1.1] — 账号与门禁

### Added
- [ ] 身份源选型与接入：飞书/Lark SSO 优先，必要时启用邀请码/邮箱验证码
- [ ] 登录态与会话：`/auth/login`、`/auth/callback`、`/auth/logout`、`/auth/me`
- [ ] 角色体系：`viewer` / `operator` / `admin`
- [ ] 私密桶仅授权人员可见，公开桶继续只读
- [ ] 访问审计：用户、IP、bucket、action、result
- [ ] 管理员邀请 / 禁用用户
- [ ] Studio 登录态展示与退出登录
- [ ] 认证态 smoke test 与灰度发布

### Security
- [ ] `/buckets`、`/objects`、`/object-url` 全部加鉴权
- [ ] CORS 收口到已登记来源，必要时启用 credentials
- [ ] 会话 cookie、安全头、CSRF 保护
- [ ] 登录与敏感接口限流
- [ ] AK/SK 凭证迁移到阿里云 KMS

## [0.2.0] — 补齐与加固

### Added
- [ ] 桶/对象列表自动刷新
- [ ] `ListObjects` 分页，突破单次 1000 个对象上限
- [ ] 文件上传、下载、删除等写操作

### Security
- [ ] 确认私密桶链接默认有效期与档位（是否含「永久」）

### Changed
- [x] 在 `.quanttide/asset/contract.yaml` 登记 35 个 OSS 桶为资产条目

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
