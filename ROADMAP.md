# 数字资产云路线图

## Phase 2: DSL 语法与引擎原型（当前阶段）

### 目标

把现实需求翻译成 DSL 语法，对齐 `examples` 和 `tests/fixtures`。

### 待办

1. 提取现有脚本需求
   - 分析 `backup_product_journal.py` 的输入/输出/触发条件
   - 分析 `generate_product_roadmap.py` 的转换逻辑
2. 编写 DSL 示例
   - 将上述需求转换为 `tests/fixtures/*.yaml` 文件
   - 验证 ISDL 结构（assets, transforms, triggers）的表达能力
3. 实现最小解析器
   - 读取 YAML 契约
   - 校验 Schema 合法性
   - 模拟执行转换步骤
