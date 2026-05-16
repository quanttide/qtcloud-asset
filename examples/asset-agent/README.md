# asset-agent — 资产智能体协作壳

三位智能体（DRD、前端、后端）通过文件变化和事件驱动协作的可视化原型。

## 已实现

### 双页视图

| 页面 | 用途 | 打开方式 |
|------|------|----------|
| `src/index.html` | 资产结构树 + 角色 badge，回答「谁负责什么、上下游是谁」 | 双击 |
| `src/evolution.html` | 版本演化网格，回答「迭代了多少轮、卡在哪、怎么收敛的」 | 双击 |

### 数据生成

```bash
python src/app.py sample/ --output src/data.js --seed-traces
```

扫描真实目录结构，生成图数据。`--seed-traces` 生成 3 轮迭代的版本节点和 trigger/feedback 边。

### 3 智能体模型

- **drd-agent** — 产出设计文档
- **code-frontend-agent** — 实现前端页面
- **code-backend-agent** — 实现后端接口

三种协作边：`write`（产出）、`trigger`（设计驱动开发）、`feedback`（实现反馈设计）。各角色独立版本线，迭代收敛于无反馈状态。

### 文档体系

| 文档 | 内容 |
|------|------|
| `docs/brd.md` | 3 个业务场景（版本不可追溯、迭代不可见、收敛不明确） |
| `docs/drd.md` | 图数据模型和 6 项业务假设 |
| `docs/ixd.md` | 双页设计原则和布局 |
| `docs/add.md` | 核心数据模型和版本追踪逻辑 |
| `docs/qa.md` | 4 组验证项和通过标准 |
| `docs/user.md` | 最终用户指南 |

### 示例仓库

`sample/` 目录包含一个模拟资产仓库（`sample/docs/drd/`、`sample/src/backend/`、`sample/src/frontend/`），可直接作为 `src/app.py` 的输入。

## 快速开始

```bash
# 1. 生成图数据
python src/app.py sample/ --output src/data.js --seed-traces

# 2. 打开视图
open src/index.html   # 或 src/evolution.html
```

## 设计原则

- 职责可见：目录树 + 角色 badge 表达分工
- 过程可诊：版本网格暴露瓶颈和反馈闭环
- 减法优先：能用一个视图解决的问题不用两个
- 不假设版本同步：每个角色独立迭代
