# CLI 状态报告

> 更新日期：2026-08-16
> 位置：`src/cli/`
> 技术栈：Rust (cargo)
> 最新版本：v0.1.0-alpha.2 (2026-07-30)

## 版本历史

| 版本 | 日期 | 内容 |
|------|------|------|
| v0.1.0-alpha.2 | 2026-07-30 | Rust 全量重写（替换 Python）：run/scan/validate/config/version 五命令，56 测试 |
| v0.1.0-alpha.1 | 2026-04-28 | CLI 重构，支持 alpha 预发布阶段 |
| v0.0.1 | 2026-04-17 | 初始版本（Python），基础归档工作流 |

## 当前状态

- 五命令已实现：`run`（归档工作流）、`scan`（资产扫描）、`validate`（声明式策略验证）、`config`（契约配置）、`version`
- 56 个自动化测试（31 单元 + 25 集成）
- 支持 `--json` 输出、dry-run、回滚、空目录清理
- 参考实现 `examples/asset-cli/`（自 qtadmin 迁移）：archive/quality/status 三命令

## 规划进度

见根 `ROADMAP.md` 目标 1（数字资产契约）与目标 2（质量控制文档）：

- 契约解析器、Skill 执行引擎、契约验证与 diff 均未开始
- CLI 与 Studio 契约驱动的衔接（目标 3）未开始

## 注意事项

- `app/`、`main.py`、`__init__.py` 为 Python 遗留文件，待移除
