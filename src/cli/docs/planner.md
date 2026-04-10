# 配置层 — `planner.py`

## 职责

将 `contracts.yaml` 契约配置解析为可执行的 `Workflow` 对象。

**不执行任何文件系统写操作，是纯数据转换层。**

## 数据结构

### `ArchiveTask`

单个产品的归档任务：

```python
@dataclass
class ArchiveTask:
    product: str      # 产品名称
    src_dir: Path     # 源目录
    dst_dir: Path     # 目标目录
```

### `Workflow`

归档工作流：

```python
@dataclass
class Workflow:
    name: str                      # 契约名称
    pattern: str                   # 文件匹配模式
    tasks: list[ArchiveTask]       # 任务列表
```

## 公开接口

### `resolve_workflow_simple()`

```python
resolve_workflow_simple(
    contract_name: str,
    input_dir: Path,
    output_dir: Path,
    pattern: str = "*.md",
    contracts_file: Path | None = None,
) -> Workflow
```

流程：
1. 加载 `contracts.yaml`
2. 扫描 `input_dir` 下的产品子目录
3. 为每个产品生成 `ArchiveTask`
4. 返回 `Workflow`

### `print_workflow_summary()`

打印工作流摘要。

## 内部函数

| 函数 | 说明 |
|------|------|
| `_load_yaml(path)` | 加载 YAML 文件 |
| `_get_contract(data, name)` | 提取指定契约 |
| `_get_products(directory)` | 扫描产品子目录 |

## 设计原则

- 不 import `file_operator`
- 无副作用，可独立单元测试
