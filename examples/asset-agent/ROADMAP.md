# asset-agent 原型路线图

## 当前状态

原型已通过双页分离验证（结构页职责可见 + 演化页过程诊断），五层文档体系（BRD→DRD→IXD→ADD→QA）覆盖完整。

文件已按 `docs/`（文档）和 `src/`（代码）重组。

## 已修复

- **P0** — `app.py` 默认输出改为 `data.js`（原为 `data.json`）
- **P2** — `ixd.md` 数据流图更新，删除已不存在的 `agent_rules.yaml` 引用

## 待办问题

### P1 — 收敛标记硬编码

`src/evolution.html` 图例 `✓ v3 收敛` 不随数据变化，实际收敛版本与显示不一致。

- 修复：动态计算收敛状态（各列底部版本无 outgoing feedback 边）

### P3 — Agent ID 硬编码

`src/app.py`、`src/evolution.html` 均硬编码三个 agent ID，增减 agent 需修改多处。

- 修复：统一从数据源读取 agent 列表

### P4 — 版本计数器语义

`--seed-traces` 版本分配逻辑不完全遵循独立版本线设计。

- 修复：每个 agent 维护独立计数器，trigger 只对齐版本不覆写

### P5 — 已知待办（AGENTS.md 已记录）

- 自迭代（智能体内部版本提升无触发）未实现
- 收敛标准未显式化为数据
- 版本号未接入真实 Git tag
- 事件应从边属性中拆出为独立结构
