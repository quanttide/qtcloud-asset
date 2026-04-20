# 操作层 — `file_operator.py`

## 职责

提供原子文件操作（移动、复制、删除），失败自动回滚。

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

## 公开接口

### `move_file()`

```python
move_file(src: Path, dst: Path) -> FileResult
```

复制文件（保留元数据）并删除源文件。目标已存在时跳过。

### `copy_file()`

```python
copy_file(src: Path, dst: Path) -> FileResult
```

复制文件（保留元数据），保留源文件。

### `delete_file()`

```python
delete_file(path: Path) -> FileResult
```

删除文件或目录。不存在时跳过。

### `rollback()`

```python
rollback(src_dir: Path, dst_dir: Path, moved: list[str]) -> list[FileResult]
```

尽力回滚：将已移动的文件拷回源目录。

## 设计原则

- 不 import `yaml`、`typer`、`workflow`
- 每个函数接收具体 `Path`，返回结构化结果
- 失败自动回滚，保证原子性
- 可独立单元测试（传入临时目录）
