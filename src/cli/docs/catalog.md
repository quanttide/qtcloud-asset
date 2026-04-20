# 资产目录层 — `catalog.py`

## 职责

存储和管理通过发现步骤（Discovery）生成的资产清单数据。

**作为发现结果的数据载体，连接发现层与后续处理层。**

## 数据结构

### `AssetEntry`

单个资产的元数据：

```python
@dataclass
class AssetEntry:
    id: str           # 资产唯一标识
    name: str         # 资产名称
    type: str         # 资产类型（doc, code, image 等）
    source: str       # 来源路径或链接
    discovered_at: str # 发现时间
    metadata: dict    # 扩展属性
```

### `Catalog`

资产清单集合：

```python
@dataclass
class Catalog:
    project_id: str               # 项目标识
    assets: list[AssetEntry]      # 资产列表
    generated_at: str             # 生成时间

    def find(self, asset_id: str) -> AssetEntry | None: ...
    def filter(self, type: str) -> list[AssetEntry]: ...
```

## 公开接口

### `load_catalog()`

```python
load_catalog(path: Path) -> Catalog
```

从本地文件加载资产目录。

### `save_catalog()`

```python
save_catalog(catalog: Catalog, path: Path) -> None
```

将资产目录序列化到本地文件。

### `create_catalog()`

```python
create_catalog(project_id: str) -> Catalog
```

创建空的资产目录实例。

## 设计原则

- 纯数据结构，不包含业务逻辑
- 支持序列化/反序列化（JSON/YAML）
- 可独立单元测试
