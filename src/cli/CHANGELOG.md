# Changelog

## [0.1.0-alpha.2] - 2026-07-30

### 变更

- Rust 全量重写（替换 Python）
- 新增 `validate` 子命令（ATOMIC / SCOPED 声明式策略验证）
- 新增 JSON 输出支持（`--json` 标志）
- 56 个自动化测试（31 单元 + 25 集成）

### 命令

- `run` — 归档工作流（dry-run / 回滚 / 空目录清理）
- `scan` — 目录资产扫描
- `validate` — 声明式策略验证
- `config` — 契约配置查看
- `version` — 版本信息

## [0.1.0-alpha.1] - 2026-04-28

### 变更

- CLI 重构，支持 alpha 预发布阶段
- `version` 命令支持预发布阶段标识

## [0.0.1] - 2026-04-17

### 变更

- 初始版本（Python）
- 基础归档工作流
- 契约驱动配置
