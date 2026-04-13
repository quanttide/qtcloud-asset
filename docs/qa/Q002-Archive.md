# Q002: 归档逻辑

**[ 需求来源 ]**：`PRD` 定义数字资产从 journal（草稿区）到 archive（归档区）的生命周期流转。
**[ 核心契约 ]**：`examples/archive/sample/contracts.yaml`

---

## 1. 现实需求：确保"移动"而非"复制"
* **验收点**：执行归档后，源目录文件被删除，目标目录出现相同文件。
* **验证资产**：`examples/archive/sample/journal/` → `examples/archive/sample/archive/`
* **状态**：✅ *Passed*

## 2. 现实需求：处理"重名冲突"
* **验收点**：若 archive 已存在同名文件，跳过不覆盖。
* **验证资产**：`src/cli/tests/test_file_operator.py::test_skip_existing`
* **状态**：✅ *Passed*

## 3. 现实需求：空源目录自动清理
* **验收点**：所有文件移动成功后，空源目录被删除。
* **验证资产**：`src/cli/tests/test_file_operator.py::test_clean_empty_dir`
* **状态**：✅ *Passed*

## 4. 现实需求：无匹配文件时跳过
* **验收点**：journal 目录存在但无匹配文件时，不报错，标记为 skipped。
* **验证资产**：`src/cli/tests/test_file_operator.py::test_no_matching_files`
* **状态**：✅ *Passed*

## 5. 现实需求：目标目录不存在时自动创建
* **验收点**：archive 目录不存在时，自动创建（含父目录）。
* **验证资产**：`src/cli/app/file_operator.py`（mkdir parents=True）
* **状态**：✅ *Passed*
