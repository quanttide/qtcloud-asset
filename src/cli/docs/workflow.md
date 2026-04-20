# 工作流 — `workflow.py` 和 `file_operator.py`

## 职责

将契约配置解析为操作计划，并执行原子文件操作。

**planner 是纯数据转换层，file_operator 是具体执行层。**

## 数据结构

### `Operation`

单个操作定义：

```python
@dataclass
class Operation:
    action: str       # 操作类型：move / copy / delete
    src: Path         # 源路径
    dst: Path | None  # 目标路径（delete 时为空）
```

### `Plan`

操作计划：

```python
@dataclass
class Plan:
    name: str                    # 计划名称
    operations: list[Operation]  # 操作列表
```

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

### `resolve_plan()`

```python
resolve_plan(
    contract_path: Path,
    context: dict | None = None,
) -> Plan
```

流程：
1. 加载 `contract.yaml`
2. 解析 `skills` 配置中的操作定义
3. 根据上下文参数解析路径变量
4. 生成 `Plan`

### `move_file()` / `copy_file()` / `delete_file()`

```python
move_file(src: Path, dst: Path) -> FileResult
copy_file(src: Path, dst: Path) -> FileResult
delete_file(path: Path) -> FileResult
```

原子文件操作。`move_file` 复制并删除源，`copy_file` 仅复制，`delete_file` 删除文件或目录。

### `rollback()`

```python
rollback(src_dir: Path, dst_dir: Path, moved: list[str]) -> list[FileResult]
```

尽力回滚：将已移动的文件拷回源目录。

## 设计原则

- planner 不执行写操作，纯数据转换
- file_operator 不依赖配置源，只接收具体路径
- 失败自动回滚，保证原子性
- 可独立单元测试
