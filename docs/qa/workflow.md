# 治理工作流

需求来源：`BRD` 提到创始人需要简单的命令行工具管理知识库归档
核心契约：`src/cli/contracts.yaml`

## 一个命令就能归档

- 验收点：`qtcloud-asset archive` 执行后，journal 文件移动到 archive，源目录清空
- 验证资产：`src/cli/tests/test_planner.py`、`src/cli/tests/test_file_operator.py`
- 状态：✅ 通过

## 执行前先预览

- 验收点：`--dry-run` 只显示预览，不移动文件
- 验证资产：`src/cli/tests/test_file_operator.py::test_dry_run`
- 状态：✅ 通过

## 确保「移动」而非「复制」

- 验收点：执行归档后，源目录文件被删除，目标目录出现相同文件
- 验证资产：`examples/archive/sample/journal/` → `examples/archive/sample/archive/`
- 状态：✅ 通过

## 处理重名冲突

- 验收点：若 archive 已存在同名文件，跳过不覆盖
- 验证资产：`src/cli/tests/test_file_operator.py::test_skip_existing`
- 状态：✅ 通过

## 空源目录自动清理

- 验收点：所有文件移动成功后，空源目录被删除
- 验证资产：`src/cli/tests/test_file_operator.py::test_clean_empty_dir`
- 状态：✅ 通过

## 无匹配文件时跳过

- 验收点：journal 目录存在但无匹配文件时，不报错，标记为 skipped
- 验证资产：`src/cli/tests/test_file_operator.py::test_no_matching_files`
- 状态：✅ 通过

## 目标目录不存在时自动创建

- 验收点：archive 目录不存在时，自动创建（含父目录）
- 验证资产：`src/cli/app/file_operator.py`（`mkdir parents=True`）
- 状态：✅ 通过

## 契约不存在时报错

- 验收点：指定不存在的契约名称时，CLI 输出错误信息并退出
- 验证资产：`tests/cli/test_archive.py`
- 状态：✅ 通过

## journal 目录不存在时报错

- 验收点：journal 目录不存在时，CLI 输出错误信息并退出
- 验证资产：`tests/cli/test_archive.py`
- 状态：✅ 通过

## 失败时自动回滚

- 验收点：移动文件失败时，已移动的文件退回源目录
- 验证资产：`src/cli/tests/test_file_operator.py::test_rollback_on_failure`
- 状态：✅ 通过
