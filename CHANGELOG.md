# CHANGELOG

## [Unreleased]

### Added
- CLI 新增 `oss` 子命令（list / ls / url），复用 Provider 只读接口管理 OSS
- Provider 列表接口支持排序（`sort`/`order`）与对象分页（`limit`/`marker`）
- Provider 对象列表支持 `prefix` 过滤（OSS 原生）
- CLI `oss list` / `oss ls` 支持 `--sort` / `--order` / `--limit` / `--prefix` 参数

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
