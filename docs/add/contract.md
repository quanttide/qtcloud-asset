# 数字资产契约技术架构

## 三大工程公理

契约采用单向下发模式，结构化契约是唯一的 Single Source of Truth，内联注释纯粹给人看，引擎解析时忽略。废弃双向校验，引擎绝不尝试从自然语言反向生成契约结构，UI 渲染时可提取注释作为步骤标题但不参与校验逻辑。

引擎禁止自动合并，回归 PR 提案模型。引擎是只读校验器加状态推进器，不具备语义合并能力。合并即创建新契约，由人工生成新 CID，forked_from 只记录血缘。只要 forked_from 非空，契约进入 pending_review 状态，必须走完提案审查批准流程。

消除静默失败，采用显式暂停与死信队列。契约只有明确的运行态 active 和阻断态 suspended，无降级运行态。失败后唯一选项是显式暂停，推入死信队列 DLQ，触发人工介入。降级方案必须是另一个显式的契约版本，由人工批准后替换，降级路径本身也被完全审计。

## 契约生命周期状态机

契约生命周期定义见 PRD 文档。状态迁移规则为 draft 到 pending_review 提交审查，pending_review 到 active 审批通过，active 到 suspended 执行失败或人工暂停，suspended 到 active 人工恢复或显式降级版本替换，active 或 suspended 到 archived 契约废弃或完成。

## 契约即证据

每个契约实例有内容寻址 CID、version、snapshot。生命周期事件包含 created_at、activated_at、deprecated_at。执行日志包含 correlation_id、step_id、action_id、params_hash、result、timestamp、actor_id。证据导出接口给定契约 ID 和时间区间，导出事件流和对应的快照哈希链。