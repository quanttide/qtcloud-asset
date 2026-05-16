# 贡献指南

## 环境要求

- Python 3.8+
- 浏览器（双击 HTML 文件打开即可）

无其他依赖。HTML 页面零配置运行，无需构建步骤。

## 快速运行

```bash
# 生成图数据（使用 sample/ 目录）
python app.py sample/ --output data.js --seed-traces

# 打开视图
open index.html      # 资产结构
open evolution.html  # 版本演化
```

## 文件约定

| 文件 | 职责 | 修改触发条件 |
|------|------|-------------|
| `app.py` | 图数据生成 | 数据模型变更、新增智能体、目录结构变化 |
| `data.js` | 静态数据（由 app.py 生成） | 重新运行 app.py 后自动覆盖 |
| `index.html` | 资产结构视图 | 树渲染逻辑、UI 样式变更 |
| `evolution.html` | 版本演化视图 | 网格布局、边样式、交互逻辑变更 |
| `brd.md` | 业务需求 | 业务场景或用户角色变化 |
| `drd.md` | 数据需求 | 数据模型或业务假设变化 |
| `ixd.md` | 交互设计 | 页面布局或交互方式变化 |
| `add.md` | 架构设计 | 数据流或技术选型变化 |
| `qa.md` | 质量验证 | 验证项或通过标准变化 |
| `user.md` | 用户指南 | 用户可见的功能变更 |
| `ROADMAP.md` | 演进路线 | 评审后更新待办优先级 |
| `AGENTS.md` | 设计原理 | 设计原则或结论变化 |
| `vis-network.min.js` | 第三方库 | 版本升级 |

## 扩展指南

### 新增智能体

1. 在 `app.py` 的 `AGENTS` 列表中添加定义（id、path、outputs、watch）
2. 在 `evolution.html` 的 `colMap` 和 `colLabel` 中添加对应列
3. 在 `drd.md` 的智能体表格中增加一行
4. 在 `sample/` 中添加对应的目录结构
5. 运行 `python app.py sample/ --output data.js --seed-traces` 验证

### 修改数据模型

1. 更新 `drd.md` 中的 schema 定义
2. 更新 `add.md` 中的核心架构描述
3. 修改 `app.py` 中的节点/边生成逻辑
4. 更新 `index.html` 或 `evolution.html` 的渲染逻辑
5. 更新 `qa.md` 中的验证项

### 添加新视图

1. 先在 `ixd.md` 中说明视图用途和布局
2. 确认不需新增视图时遵循减法优先原则
3. 新增 HTML 文件，在 `ixd.md` 更新数据流图
4. 在 `qa.md` 中添加对应的验证项

## 开发原则

- **减法优先**：能用一个视图解决的问题不用两个
- **数据驱动**：`data.js` 是唯一数据源，两个页面各自消费
- **零配置**：HTML 页面不依赖构建工具，双击即用
- **契约对齐**：修改代码前先更新对应文档
