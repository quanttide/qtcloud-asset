# 质量保证 — CLI 实现差距审计

> 2026-04-09 对比文档设想的 BRD / PRD / ADD 与当前 CLI 实际实现，记录以下已发现问题。

---

## Q001: CLI 只有一个 `archive` 命令，覆盖面过窄

| 维度 | 说明 |
|------|------|
| **现状** | CLI 仅实现 `archive` 一个命令（journal → archive 文件移动） |
| **期望** | BRD 描述了多个业务场景——跨平台迁移、AI 提炼路线图、成熟度判断、审批流程等 |
| **影响** | 产品定位为"数字资产云"，但用户可见功能仅一个归档操作 |
| **来源** | `docs/prd/workflow.md`、`docs/brd/index.md` |

## Q002: 契约是"哑"的路径映射，缺少状态机和 ISDL

| 维度 | 说明 |
|------|------|
| **现状** | `contracts.yaml` 仅包含 `name`、`version`、`paths.journal`、`paths.archive` 四个字段 |
| **期望** | PRD 定义契约包含状态机（draft → pending_review → active → suspended → archived）、成熟度评分、治理语言记录、CID 内容寻址、`forked_from` 血缘 |
| **影响** | 契约无法表达生命周期、审批流程、成熟度触发等核心概念 |
| **来源** | `docs/prd/contract.md`、`docs/add/workflow.md` (ISDL Lite v1.0) |

## Q003: 没有 `engine.py`，架构偏离 ADD 演进路径

| 维度 | 说明 |
|------|------|
| **现状** | `planner.py` 和 `file_operator.py` 直接耦合，没有独立的执行引擎 |
| **期望** | ADD 定义演进路径 Step 2 为"抽取通用引擎 `engine.py`"，负责 读取契约 → 遍历文件 → 执行 transform → 写入 → 记录日志 |
| **影响** | 无法支持链式 transform、事件记录、可插拔的转换动作 |
| **来源** | `docs/add/index.md` |

## Q004: 没有适配器抽象

| 维度 | 说明 |
|------|------|
| **现状** | 所有操作硬编码本地文件系统（`shutil.copy2` + `unlink`） |
| **期望** | ADD 规划 `LocalFSAdapter`、`FeishuAdapter`、`GithubAdapter` 等插件化适配器 |
| **影响** | 无法实现跨平台转换（飞书 → GitHub）这一核心业务场景 |
| **来源** | `docs/add/index.md`、`docs/prd/workflow.md` |

## Q005: 缺少 `roadmap` 命令（AI 提炼产品路线图场景）

| 维度 | 说明 |
|------|------|
| **现状** | CLI 无此命令，`examples/roadmap/generate_product_roadmap.py` 脚本独立存在未被集成 |
| **期望** | PRD 工作流明确包含"产品路线图生成场景"，BRD 描述"日志提炼为结构化路线图"的需求 |
| **影响** | 已验证的 AI 转换能力未产品化 |
| **来源** | `docs/prd/workflow.md`、`docs/brd/index.md` |

## Q006: `src/provider` 为空壳，测试 import 路径不一致

| 维度 | 说明 |
|------|------|
| **现状** | `src/provider/app/` 下仅存空文件；测试引用 `src.provider.app.services.planner`，但实际代码在 `src.cli.app.planner` |
| **期望** | 测试应能正常运行 |
| **影响** | `pytest` 在 CLI 模块下运行会因 import 错误直接失败 |
| **来源** | `src/cli/tests/test_planner.py`、`src/cli/tests/test_file_operator.py` |

## Q007: 缺少触发器机制

| 维度 | 说明 |
|------|------|
| **现状** | 只能手动执行 `cli archive` |
| **期望** | ISDL Lite 规划了 manual（V1.0）、cron（V2.0）、event（V2.0）三种触发方式；PRD 提到成熟度阈值自动触发迁移 |
| **影响** | 无法实现 BRD 中"系统提示创始人何时该迁移"的核心痛点 |
| **来源** | `docs/prd/workflow.md`、`docs/add/workflow.md` |

## Q008: 缺少事件溯源 / 执行日志

| 维度 | 说明 |
|------|------|
| **现状** | `archive_product()` 返回 `ArchiveResult` 仅包含当次统计信息，不持久化、不可追溯 |
| **期望** | ADD 定义"每个执行步骤记录结构化事件"；PRD 定义执行日志包含 `correlation_id`、`step_id`、`action_id`、`params_hash`、`result`、`timestamp`、`actor_id` |
| **影响** | 无法满足 BRD"治理过程可追溯可审计"的成功标准 |
| **来源** | `docs/add/index.md`、`docs/add/contract.md` |
