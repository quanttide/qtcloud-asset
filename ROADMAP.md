# 数字资产云路线图

## Phase 2: DSL 语法与引擎原型（当前阶段）

### 目标

把现实需求翻译成 DSL 语法，对齐 `examples`、`tests/fixtures` 和文档设计，提炼通用代码结构，为 CLI 做准备。

### 待办

1. 提取需求：分析现有脚本的输入/输出/触发条件/转换逻辑
2. 编写 DSL：将需求转换为 `tests/fixtures/*.yaml`，验证 ISDL 结构表达能力
3. 实现解析器：读取 YAML 契约，校验 Schema，模拟执行转换
4. 重构代码：对齐 `docs/prd` 和 `docs/add` 设计，提炼通用结构，为 CLI 做准备
