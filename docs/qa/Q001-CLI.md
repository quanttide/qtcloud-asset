# Q001: CLI 工具链

**[ 需求来源 ]**：`BRD` 提到创始人需要简单的命令行工具管理知识库归档。
**[ 核心契约 ]**：`src/cli/contracts.yaml`

---

## 1. 现实需求：一个命令就能归档
* **验收点**：`qtcloud-asset archive` 执行后，journal 文件移动到 archive，源目录清空。
* **验证资产**：`src/cli/tests/test_planner.py`、`src/cli/tests/test_file_operator.py`
* **状态**：✅ *Passed*

## 2. 现实需求：执行前先预览
* **验收点**：`--dry-run` 只显示预览，不移动文件。
* **验证资产**：`src/cli/tests/test_file_operator.py::test_dry_run`
* **状态**：✅ *Passed*

## 3. 现实需求：契约不存在时报错
* **验收点**：指定不存在的契约名称时，CLI 输出错误信息并退出。
* **验证资产**：`tests/cli/test_archive.py`
* **状态**：✅ *Passed*

## 4. 现实需求：journal 目录不存在时报错
* **验收点**：journal 目录不存在时，CLI 输出错误信息并退出。
* **验证资产**：`tests/cli/test_archive.py`
* **状态**：✅ *Passed*

## 5. 现实需求：失败时自动回滚
* **验收点**：移动文件失败时，已移动的文件退回源目录。
* **验证资产**：`src/cli/tests/test_file_operator.py::test_rollback_on_failure`
* **状态**：✅ *Passed*
