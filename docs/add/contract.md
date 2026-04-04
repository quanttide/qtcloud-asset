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

## 契约生命周期状态机

- draft：契约定义阶段，可自由修改
- pending_review：契约分叉或修改后，需经审批
- active：契约生效，开始执行
- suspended：契约因错误或人工干预被阻断
- archived：契约废弃或完成，保留历史记录

状态迁移：
- draft -> pending_review：提交审查
- pending_review -> active：审批通过
- active -> suspended：执行失败或人工暂停
- suspended -> active：人工恢复或显式降级版本替换
- active/suspended -> archived：契约废弃或完成

## 契约即证据

- 每个契约实例有内容寻址 CID、version、snapshot
- 生命周期事件：created_at、activated_at、deprecated_at
- 执行日志包含：correlation_id、step_id、action_id、params_hash、result、timestamp、actor_id
- 证据导出接口：给定契约 ID + 时间区间，导出事件流 + 对应的快照哈希链
