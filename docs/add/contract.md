# 数字资产契约技术架构

## 当前实现（V1.0）

`contracts.yaml` 是最简化的契约结构，仅包含路径映射：

```yaml
contracts:
  journal_backup:
    name: 产品日志归档
    version: 1
    paths:
      journal: docs/journal      # 源路径
      archive: docs/archive/journal  # 目标路径
```

字段说明：

| 字段 | 类型 | 说明 |
|------|------|------|
| `contracts.<name>` | 映射 | 契约名称，CLI 通过名称引用 |
| `name` | 字符串 | 人类可读的契约描述 |
| `version` | 整数 | 契约版本号 |
| `paths.journal` | 路径 | journal 基础目录 |
| `paths.archive` | 路径 | archive 目标目录 |

运行时通过 `slug` 参数拼接完整路径：`{journal}/{slug}/{product}` → `{archive}/{slug}/{product}`。

## 设计原则（三大工程公理）

> 以下公理指导 V2.0 契约引擎设计，V1.0 尚未实现。

**单向下发** — 结构化契约是唯一的 Single Source of Truth，内联注释纯粹给人看，引擎解析时忽略。废弃双向校验，引擎绝不尝试从自然语言反向生成契约结构。

**禁止自动合并** — 回归 PR 提案模型。引擎是只读校验器加状态推进器，不具备语义合并能力。合并即创建新契约，由人工生成新 CID，`forked_from` 只记录血缘。只要 `forked_from` 非空，契约进入 `pending_review` 状态，必须走完提案审查批准流程。

**消除静默失败** — 契约只有明确的运行态 active 和阻断态 suspended，无降级运行态。失败后唯一选项是显式暂停，推入死信队列 DLQ，触发人工介入。降级方案必须是另一个显式的契约版本，由人工批准后替换，降级路径本身也被完全审计。

## V2.0 规划：契约生命周期状态机

> 当前 V1.0 无状态机概念，契约即路径配置，直接执行。

```
draft ──提交──→ pending_review ──审批──→ active ──失败/暂停──→ suspended
  │                                                        │
  │                                                        └─人工恢复/降级──→ active
  │
  └────────────────────────────── 废弃/完成 ──────────────→ archived
```

状态迁移：
- `draft` → `pending_review`：提交审查
- `pending_review` → `active`：审批通过
- `active` → `suspended`：执行失败或人工暂停
- `suspended` → `active`：人工恢复或显式降级版本替换
- `active` / `suspended` → `archived`：契约废弃或完成

## V2.0 规划：契约即证据

每个契约实例包含：

- **身份**：内容寻址 CID、`version`、`snapshot`
- **生命周期事件**：`created_at`、`activated_at`、`deprecated_at`
- **执行日志**：`correlation_id`、`step_id`、`action_id`、`params_hash`、`result`、`timestamp`、`actor_id`
- **证据导出**：给定契约 ID 和时间区间，导出事件流和对应的快照哈希链
