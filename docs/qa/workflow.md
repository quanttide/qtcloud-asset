# 治理工作流质量保证

将每一项 QA 拆解为「准则 — 判定 — 证据」。

## 归档业务完整性

准则：一个命令完成整个归档周期，不留下烂摊子。

### 命令可用性判定

- 现实需求：一个命令就能归档
- 判定标准：`qtcloud-asset` 执行后，journal 文件移动到 archive，源目录清空
- 存证：`tests/cli/test_archive.py::test_actual_archive`
- 状态：✅ 符合

### 预览模式判定

- 现实需求：执行前先预览，不误操作
- 判定标准：`--dry-run` / `-n` 只显示预览，不移动文件
- 存证：`tests/cli/test_archive.py::test_dry_run_archives_nothing`
- 状态：✅ 符合

## 异常处理完备性

准则：边界条件全部覆盖，不崩溃。

### 契约不存在判定

- 现实需求：技能不存在时报错
- 判定标准：指定不存在的技能名称时，CLI 输出错误信息并退出
- 存证：`tests/cli/test_archive.py::test_unknown_skill`
- 状态：✅ 符合

### 输入目录不存在判定

- 现实需求：输入目录不存在时报错
- 判定标准：输入目录不存在时，CLI 输出错误信息并退出
- 存证：`tests/cli/test_archive.py::test_missing_input_dir`
- 状态：✅ 符合

### 无匹配文件判定

- 现实需求：目录存在但无匹配文件时，返回 skipped 标记
- 判定标准：执行归档后返回 skipped 信息，不报错
- 存证：`src/cli/tests/test_file_operator.py::test_no_matching_files`
- 状态：✅ 符合

## 数据安全性

准则：归档过程中出问题能回退，不丢文件。

### 重名冲突判定

- 现实需求：处理重名冲突，不覆盖已有文件
- 判定标准：若 archive 已存在同名文件，跳过不覆盖
- 存证：`src/cli/tests/test_file_operator.py::TestArchiveProduct::test_skips_existing_files`
- 状态：✅ 符合

### 失败回滚判定

- 现实需求：失败时自动回滚，不丢文件
- 判定标准：移动文件失败时，已移动的文件退回源目录
- 存证：`src/cli/tests/test_file_operator.py::TestRollback::test_rollback_restores_files`
- 状态：✅ 符合

### 空目录清理判定

- 现实需求：归档后清理空源目录
- 判定标准：所有文件移动成功后，空源目录被删除
- 存证：`src/cli/tests/test_file_operator.py::TestArchiveProduct::test_removes_empty_src_dir`
- 状态：✅ 符合
