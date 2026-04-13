# 架构设计与维护指南

本文档总结量潮数字资产云（qtcloud-asset）的架构设计风格与维护习惯，供开发者和 AI 参考。

## 设计风格

### 1. 契约即接口 (Contract as Interface)

契约是模块与外界的唯一约定，而非随意的配置文件。

- **强类型定义**：使用 Pydantic Schema（`ContractSchema`, `AssetConfig`, `SkillConfig`）定义契约，拒绝魔法字符串和隐式解析。
- **不可变性**：所有 Schema 设置 `frozen=True`。配置在运行时是静态只读的，消除了“配置何时被修改”的状态追踪复杂度。
- **自动校验**：契约加载即验证（`model_validate`），失败立即报错，绝不静默降级。

### 2. 分层通信 (Layered Communication)

架构严格分为四层，层间通过**类型**通信，严禁跨层调用或隐式依赖。

| 层级 | 职责 | 禁区 |
|:---|:---|:---|
| **入口层 (CLI)** | 解析参数，编排流程，展示结果 | 不操作文件系统，不包含业务逻辑 |
| **契约层 (Contract)** | 自动向上查找 `.quanttide/asset/contract.yaml` 并解析 | 不依赖硬编码路径，不感知上层 UI |
| **工作流层 (Workflow)** | 将契约转化为具体任务列表 (`ArchiveTask`) | 纯数据转换，不执行 IO 操作 |
| **操作层 (Operator)** | 接收具体 `Path`，执行文件移动、回滚 | 不依赖 YAML、不感知 CLI 参数 |

### 3. 唯一事实源 (Single Source of Truth)

- **测试数据**：集中在 `assets/fixtures/`。`src/cli/tests` 和 `tests/cli` 通过 `conftest.py` 的 fixture 引用该目录。**严禁在多处复制同一份样例数据。**
- **契约查找**：`Contract.find_root()` 从当前目录向上遍历，直到找到 `.quanttide/asset/contract.yaml`。CLI 可在项目任意子目录运行。

### 4. QA 即交付证明 (QA as Proof)

质量文档不记录测试过程，只回答一个问题：**我们证明了什么？**

- **结构**：「准则 → 判定 → 证据」。
- **准则**：一句话总结底线（如“契约必须是地图而非账本”）。
- **判定**：具体的通过条件。
- **证据**：指向具体的测试代码或运行结果。

## 维护习惯

### 命名规范

- **文件**：`snake_case.py`（如 `file_operator.py`）。
- **类/Schema**：`PascalCase`（如 `ContractSchema`, `ArchiveTask`）。
- **测试用例**：`test_动词_场景`（如 `test_raises_on_unknown_skill`），描述行为而非实现。
- **意图优先**：名称反映业务意图，而非技术实现（如 `workflow.py` 优于 `planner.py`）。

### 重构原则

- **小步提交**：每个提交只做一件事（例：“重命名 planner 为 workflow"与“增加 Contract 类”分两个提交）。
- **删除优于隐藏**：废弃代码直接删除，不留 `# TODO: delete later`。
- **结构跟随功能**：当代码职责发生变化时，敢于重构目录结构（如将 `contracts.yaml` 移入 `assets/fixtures/`）。

### 文档同步

- **架构文档 (`docs/add/index.md`)**：记录 **Why**（设计原则、分层理由、核外挂载思想）。
- **代码注释**：模块头部 Docstring 记录 **How**（输入输出、依赖关系）。
- **契约示例**：`assets/fixtures/.quanttide/asset/contract.yaml` 是最小的、可运行的架构示例。

## 给 AI 的提示 (AI Hints)

1.  **修改代码时**：先确认功能属于哪一层。**严禁跨层修改**。
2.  **新增配置时**：优先在 `ContractSchema` 中定义类型，再读取 YAML。不要写 `config.get('key')`。
3.  **测试失败时**：检查 `assets/fixtures/` 是否符合契约 Schema 定义，而不是去 mock 内部路径查找逻辑。
4.  **寻找数据时**：永远使用 `conftest.py` 提供的 fixture，不要拼接相对路径。
