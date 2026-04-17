# CHANGELOG

## [Unreleased]

## [v0.0.1] - 2026-04-17

### Added

- 新增 CLI 模块：三层架构（入口层 cli.py、配置层 workflow.py、操作层 file_operator.py）
- 新增契约系统：Pydantic 模型定义，自动识别契约目录
- 新增 Studio 客户端：Flutter Web 应用，数字资产可视化管理界面
- 新增 Provider 服务端：FastAPI 服务，对接阿里云函数计算和 OSS
- 新增 Docker 部署配置：Studio 和 Provider 的容器化部署
- 新增完整文档体系：BRD、PRD、IXD、ADD、QA、用户文档
- 新增 AI 执行审核契约：强制执行声明和触发条件
- 新增原子技能编排需求文档：商业需求和产品需求
- 新增集成测试和单元测试：CLI 模块测试覆盖率达 99%
- 新增阿里云基础设施代码：函数计算、OSS、VPC 等定义

### Changed

- 重构 CLI 为 --input/--contract/--output 模式
- 重构契约系统：使用 Contract 类和 ContractSchema
- 重命名 .quanttide/ai 为 .quanttide/agent

### Fixed

- 修复 AI 契约：添加强制执行声明和触发条件
