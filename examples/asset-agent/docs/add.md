# 资产版本因果链可视化 — 核心设计文档

状态：已确认
日期：2026-05-16
领域：数字资产管理 × 智能体协作

---

1. 概述

1.1 项目定位

通过版本因果链可视化，回答三个问题：当前仓库各资产处于什么版本、版本之间谁驱动了谁、协作过程经历了多少轮迭代才收敛。

1.2 数据模型

```
graphData
├── graphNodes[]
│   └── { id, label, group, parent }
│       其中 group ∈ { Dir, File, Agent, Version }
└── graphEdges[]
    └── { id, from, to, type, label, events[] }
        其中 type ∈ { write, trigger, feedback }
```

节点类型：

| 类型 | 含义 | 出现在 |
|------|------|--------|
| Dir | 资产目录 | 结构页树 |
| File | 资产文件 | 结构页树 |
| Agent | 智能体角色 | 结构页树（badge） |
| Version | 资产版本 | 演化页图 |

边类型：

| 类型 | 含义 | 方向 |
|------|------|------|
| write | 智能体产出文件 | Agent → File |
| trigger | 上游通知下游 | 版本 → 版本 |
| feedback | 下游反馈上游 | 版本 → 版本 |

2. 核心架构

2.1 资产目录结构

```
资产仓库/
├── docs/drd/          ← drd-agent
│   └── 支付设计.md
├── src/frontend/     ← code-frontend-agent
│   └── payment.js
└── src/backend/      ← code-backend-agent
    └── payment.go
```

每个智能体挂载在一个目录下，职责边界由目录路径定义。

2.2 版本追踪

每个智能体独立维护版本计数器。版本在以下情况递增：

- 收到 trigger 事件 → 按源版本对齐
- 收到 feedback 事件 → 当前版本 +1
- 自迭代（内部修改） → 当前版本 +1，不产生 trigger 或 feedback

版本节点 ID 格式：`{agent-id}-v{number}`，如 `drd-agent-v2`。

2.3 迭代边生成逻辑

跨版本连线的依据是事件。一次协作产生两种边：

- trigger：从 source 的某版本指向 target 的某版本（设计驱动开发）
- feedback：从 source 的某版本指向 target 的下一个版本（实现反馈设计）

同列内部不存在 trigger 或 feedback 边。自迭代不产生跨列连线。

2.4 收敛

收敛没有显式标记，而是隐式状态：当所有关联路径上没有未处理的 feedback 时，该资产版本视为收敛。演化图中表现为最后一版只有 trigger 边，没有 feedback 边。

3. 数据流

```
app.py（内嵌 AGENTS 定义 + traces）
  │ 扫描 sample/ 目录
  │ 解析智能体挂载关系
  │ 生成 trigger/feedback 边
  │ 从 traces 生成版本节点和迭代边
  ▼
data.js → window.graphData
  ├── index.html（结构页）
  │   读取 Dir + File + Agent 节点
  │   渲染目录树 + 角色 badge
  └── evolution.html（演化页）
      读取 Version 节点 + 迭代边
      渲染三列版本网格
```

4. 当前技术选型

| 层 | 选型 | 理由 |
|:---|:-----|:-----|
| 前端渲染 | vanilla HTML + vis-network | 零构建、双击即用 |
| 数据格式 | data.js（静态 JSON + 全局变量） | 支持 file:// 协议 |
| 数据生成 | app.py（内嵌 agent 定义） | 无需外部配置，可复现 |
| 布局计算 | 代码内 BFS 求层次 + 显式坐标 | 避免 vis 自动布局不稳定 |

5. 下一步

- 自迭代（同列内版本递增无 trigger）当前未实现，traces 中需要增加 self-version-bump 事件类型
- 收敛标准目前是隐式的，未在数据中显式标记
- 版本号目前由 app.py 按 traces 顺序分配，未接入真实 Git tag
