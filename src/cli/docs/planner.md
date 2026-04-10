# 配置层 — `planner.py`

## 职责

将 `contracts.yaml` 契约配置和文件系统扫描结果解析为可执行的 `Workflow` 对象。

**不执行任何文件系统写操作，是纯数据转换层。**

## 数据结构

### `ArchiveTask`

单个产品的归档任务：

```python
@dataclass
class ArchiveTask:
    product: str      # 产品名称
    src_dir: Path     # journal 中的源路径
    dst_dir: Path     # archive 中的目标路径
```

### `Workflow`

归档工作流，配置层的唯一输出：

```python
@dataclass
class Workflow:
    name: str                         # 契约名称
    slug: str                         # 产品标识
    pattern: str                      # 文件匹配模式
    tasks: list[ArchiveTask]          # 任务列表
```

属性 `products` → 提取所有产品名称的便捷方法。

## 公开接口

### `resolve_workflow()`

```python
resolve_workflow(
    contract_name: str,
    slug: str,
    *,
    product: str | None = None,
    pattern: str = "*.md",
    contracts_file: Path | None = None,
) -> Workflow
```

流程：
1. 加载 `contracts.yaml`（或自定义路径）
2. 提取指定契约，获取 `journal` / `archive` 基础路径
3. 扫描 `journal/{slug}/` 下的产品子目录
4. 可选：过滤到单个产品
5. 为每个产品生成 `ArchiveTask`
6. 返回 `Workflow`

异常：
- `FileNotFoundError` — 配置文件或目录不存在
- `KeyError` — 契约名称或产品名称不存在（错误信息附带可用选项）

### `print_workflow_summary()`

```python
print_workflow_summary(workflow: Workflow, *, dry_run: bool = False) -> None
```

打印 `[预览/执行] 契约 / 标识 / 产品列表 / 模式` 摘要，纯展示用途。

## 内部函数

| 函数 | 说明 |
|------|------|
| `_load_yaml(path)` | 加载 YAML 文件，返回 dict |
| `_get_contract(data, name)` | 从 parsed YAML 中提取指定契约 |
| `_get_products(journal_dir)` | 扫描目录下所有子目录（产品），返回排序后的列表 |

## 设计原则

- 不 import `file_operator`（通过 dataclass 解耦）
- 无副作用，可独立单元测试
- `CONTRACTS_FILE` 默认指向同级 `contracts.yaml`，测试可通过 `contracts_file` 参数覆盖
