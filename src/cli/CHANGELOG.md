# CHANGELOG

## [Unreleased]

## [cli/v0.0.1] - 2026-04-17

### Added

- 新增 CLI 模块：三层架构（入口层 cli.py、配置层 workflow.py、操作层 file_operator.py）
- 新增契约系统：Pydantic 模型定义，自动识别契约目录
- 新增集成测试和单元测试：CLI 模块测试覆盖率达 99%

### Changed

- 重构 CLI 为 --input/--contract/--output 模式
- 重构契约系统：使用 Contract 类和 ContractSchema

### Fixed

- 修复 AI 契约：添加强制执行声明和触发条件
