# 数字资产工作流 DSL

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

## 健康检查

- 对每个适配器提供 healthcheck 接口：检查平台 API 可达性、关键字段/能力是否存在
- 引擎定期执行：success 更新 last_healthy_at，failure 触发告警 + 标记 degraded/failed
- 契约版本与平台版本绑定：声明 platform.min_version / max_version
- 健康检查失败后，契约状态变为 suspended，通过人工恢复或显式降级版本才能继续
