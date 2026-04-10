# 操作层 — `file_operator.py`

## 职责

执行实际的文件系统操作：移动文件、失败回滚、清理空目录。

**不依赖任何配置源（YAML、CLI 参数等），只接收具体路径。**

## 数据结构

### `FileResult`

单个文件的操作结果：

```python
@dataclass
class FileResult:
    name: str       # 文件名
    success: bool   # 是否成功
    reason: str     # 失败/跳过原因，成功时为空
```

### `ArchiveResult`

单个产品目录的操作结果：

```python
@dataclass
class ArchiveResult:
    product: str
    total: int                  # 匹配文件总数
    moved: list[str]            # 已移动的文件
    skipped: list[str]          # 跳过的文件（已存在）
    failed: list[str]           # 失败的文件
    source_removed: bool        # 源目录是否已被删除
    error: str | None           # 错误信息

    @property
    def ok(self) -> bool:       # True = 无错误且无失败
        return self.error is None and not self.failed
```

## 公开接口

### `archive_product()`

```python
archive_product(
    src_dir: Path,
    dst_dir: Path,
    *,
    pattern: str = "*.md",
    dry_run: bool = False,
) -> ArchiveResult
```

流程：
1. 检查源目录是否存在
2. 按 `pattern` 收集文件
3. **dry_run** → 填充 `moved` 列表，直接返回
4. 创建目标目录（`mkdir -p`）
5. 逐文件移动（`shutil.copy2` + `unlink`），已存在则跳过
6. **有失败** → 调用 `_rollback` 回滚已移动的文件
7. **全部成功** → 源目录为空则 `rmdir`，标记 `source_removed`
8. 返回 `ArchiveResult`

## 内部函数

| 函数 | 说明 |
|------|------|
| `_move_file(src, dst)` | 复制文件（保留元数据）+ 删除源文件 |
| `_rollback(src_dir, dst_dir, moved)` | 尽力回滚：将已移动的文件拷回源目录 |

## 设计原则

- 不 import `yaml`、`typer`、`planner`
- 每个函数接收具体 `Path`，返回结构化结果
- 失败自动回滚，保证原子性
- 可独立单元测试（传入临时目录）
