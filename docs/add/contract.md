# 数字资产契约技术架构

## 三大工程公理

### 公理一：单向下发 + 内联注释为唯一真相源

- 结构化契约是唯一的 Single Source of Truth
- 废弃双向校验：引擎绝不尝试从自然语言反向生成契约结构
- 内联注释（annotation）纯粹给人看，引擎解析时忽略
- UI 渲染时可提取注释作为步骤标题，但不参与任何校验逻辑

### 公理二：禁止引擎级自动合并，回归 PR/提案模型

- 引擎是只读校验器 + 状态推进器，不具备语义合并能力
- 合并即创建新契约：由人工生成新 CID，forked_from 只记录血缘
- 只要 forked_from 非空，契约进入 pending_review 状态，必须走完提案-审查-批准流程

### 公理三：消除静默失败，显式暂停与死信队列

- 契约只有明确的运行态（active）和阻断态（suspended），无降级运行态
- 失败后唯一选项：显式暂停，推入死信队列（DLQ），触发人工介入
- 降级方案必须是另一个显式的契约版本，由人工批准后替换，降级路径本身也被完全审计

## ISDL 核心结构

契约遵循 ISDL（统一数字资产描述语言），分为四个核心块：

### assets（资产声明）

- 每个资产声明 platform（适配器标识）、asset_type（平台内资源类型）
- schema 定义 required_fields 和 constraints（状态必须满足的谓词）

### transforms（转换规则）

- id、input_asset_ref、output_asset_ref
- action 引用外部适配器实现的转换逻辑
- params 定义 action 所需的参数
- condition 定义执行条件（满足时才执行）
- error_handling：max_retries、retry_delay、on_failure（唯一值 suspend）、dlq_binding、escalation

### triggers（触发器）

- event_type、filter（asset_type/label/状态变更等）、cron、manual
- 每个触发器绑定 transform 或 approval_gate
- 可选并发控制（max_in_flight、strategy）

### lifecycle（生命周期与治理）

- 状态机：draft -> pending_review -> active -> suspended -> archived
- 血缘：forked_from 记录来源 CID，证明血缘不代表引擎做了合并计算
- 审批：required_approvals、approver_roles
- 部署绑定：platform credential_ref、settings 注入
- 健康检查：enabled、interval、checks（adapter_health）、on_unhealthy（suspend）

## 适配器层设计

- 分层设计：core（状态机 + 转换步骤 + 事件端口）与 adapter（飞书/GitHub 等）分离
- 每个适配器声明支持的 asset_types、action_types、必需/可选 params、支持的 event_filters
- 引擎在部署时做能力协商：合约所需能力 与 平台支持能力的交集非空才允许绑定
- 合约只引用资源 ID，不直接写平台细节

## 事件溯源

- 物理上分两流：contract_events（定义变更）与 execution_events（每次运行）
- 逻辑上共享同一基础设施：统一的 event_id、timestamp、source、correlation_id、version
- 执行事件包含：contract_id、contract_version、input_hash、output_hash、step_id、action_id、result
- 支持统一回放与时间旅行调试

## 契约即证据

- 每个契约实例有内容寻址 CID、version、snapshot
- 生命周期事件：created_at、activated_at、deprecated_at
- 执行日志包含：correlation_id、step_id、action_id、params_hash、result、timestamp、actor_id
- 证据导出接口：给定契约 ID + 时间区间，导出事件流 + 对应的快照哈希链

## 健康检查

- 对每个适配器提供 healthcheck 接口：检查平台 API 可达性、关键字段/能力是否存在
- 引擎定期执行：success 更新 last_healthy_at，failure 触发告警 + 标记 degraded/failed
- 契约版本与平台版本绑定：声明 platform.min_version / max_version
- 健康检查失败后，契约状态变为 suspended，通过人工恢复或显式降级版本才能继续
