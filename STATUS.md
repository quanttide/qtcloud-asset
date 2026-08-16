# qtcloud-asset 状态报告

> 更新日期：2026-08-16
> 仓库：quanttide/qtcloud-asset
> 最新 commit：6ad2429（领先 origin/main 1 提交，未推送）
> 最新版本：cli/v0.1.0-alpha.2 (2026-07-30)

## 版本历史

| 版本 | 日期 | 内容 |
|------|------|------|
| cli/v0.1.0-alpha.2 | 2026-07-30 | CLI Rust 全量重写：run/scan/validate/config/version 五命令，56 测试 |
| cli/v0.1.0-alpha.1 | 2026-04-28 | CLI 重构，支持 alpha 预发布阶段 |
| v0.0.1 | 2026-04-17 | 首版：Python CLI + Flutter Studio + FastAPI Provider |

## 组件进度

| 组件 | 技术 | 版本 | 状态 |
|------|------|------|------|
| CLI (`src/cli`) | Rust (cargo) | v0.1.0-alpha.2 | 五命令已实现，56 测试通过 |
| Provider (`src/provider`) | Go 1.22 | — | Go 重写完成，API 骨架（/health、/config），Service/Repository 层待实现 |
| Studio (`src/studio`) | Flutter | — | 骨架，`asset_contract_screen.dart` 待按契约驱动重构 |

## ROADMAP 进度

| 目标 | 文件 | 阶段 |
|------|------|------|
| 当前阶段：文档体系对齐 (v0.1.x) | `docs/` | 仅 docs/index.md 完稿，BRD/PRD/IXD/ADD/QA 待对齐 |
| 目标 1：数字资产契约 | `ROADMAP.md` | 未开始（统一 Schema、解析器、执行引擎、验证、diff） |
| 目标 2：质量控制文档 | `ROADMAP.md` | 未开始（指标体系、准入/准出、自动化检查、质量报告） |
| 目标 3：Studio 资产浏览 | `ROADMAP.md` | 刚规划（资产契约落地、目录引擎、通用页面、只读浏览） |

## 文档覆盖

| 目录 | 状态 |
|------|------|
| docs/brd/ | 有（graph、harness、pricing、skill-composition） |
| docs/prd/ | 有（graph、harness、pricing、skill-composition、workflow） |
| docs/ixd/ | 仅索引页 |
| docs/add/ | 缺失 |
| docs/qa/ | 有（架构决策记录 + 质量承诺范式） |
| docs/user/ | 有（安装、快速开始、配置、契约、归档） |
