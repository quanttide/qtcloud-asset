# 架构设计文档契约

## 模块定位

CLI 三层架构、契约技术规范、ISDL DSL 设计

## 资产定义

| 资产 | 类型 | 分类 | 路径 | 描述 |
|------|------|------|------|------|
| 架构设计文档 | docs | add | docs/add | CLI 三层架构、契约技术规范、ISDL DSL 设计 |

## 资产详情

### 架构设计文档 (add)

- **类型**: docs
- **分类**: add
- **路径**: `docs/add`
- **描述**: CLI 三层架构、契约技术规范、ISDL DSL 设计

## 设计原则

### 三大工程公理

1. **单向下发** — 结构化契约是唯一的 Single Source of Truth，内联注释纯粹给人看，引擎解析时忽略。废弃双向校验。

2. **禁止自动合并** — 回归 PR 提案模型。引擎是只读校验器加状态推进器，不具备语义合并能力。合并即创建新契约。

3. **消除静默失败** — 契约只有明确的运行态 active 和阻断态 suspended，无降级运行态。

## 契约状态机

```
draft ──提交──→ pending_review ──审批──→ active ──失败/暂停──→ suspended
  │                                                        │
  │                                                        └─人工恢复/降级──→ active
  │
  └────────────────────────────── 废弃/完成 ──────────────→ archived
```

## 契约即证据

每个契约实例包含：
- **身份**: 内容寻址 CID、`version`、`snapshot`
- **生命周期事件**: `created_at`、`activated_at`、`deprecated_at`
- **执行日志**: `correlation_id`、`step_id`、`action_id`、`result`
