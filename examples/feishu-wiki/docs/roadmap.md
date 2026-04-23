# 路线图

仅供参考。

在不破坏现有简洁性的前提下，用最小改动引入契约和验证能力。

## 极简演化路线（两步走）

### 第一步：增加一个独立的验证脚本（不修改 catalog.py）

**目标**：在现有目录快照生成后，能用另一个脚本对照契约做检查。

#### 1.1 契约文件（最简单格式）
```yaml
# contract.yaml
spaces:
  - id: "7123456789"          # 必须存在的知识库ID
    name: "核心知识库"
    required_nodes:
      - "BRD"                 # 标题中包含此关键字的节点
```

#### 1.2 验证脚本 `validate.py`
```python
#!/usr/bin/env python3
"""对照契约检查已下载的目录快照"""

import json
import yaml
from config import BASE_DIR

def load_contract():
    with open("contract.yaml") as f:
        return yaml.safe_load(f)

def load_catalog_metadata():
    with open(f"{BASE_DIR}/metadata.json") as f:
        return json.load(f)

def validate():
    contract = load_contract()
    catalog = load_catalog_metadata()

    catalog_space_ids = {s["space_id"] for s in catalog["spaces"]}

    for required_space in contract["spaces"]:
        space_id = required_space["id"]
        if space_id not in catalog_space_ids:
            print(f"❌ 缺失知识库: {required_space['name']} ({space_id})")
        else:
            print(f"✅ 知识库存在: {required_space['name']}")

if __name__ == "__main__":
    validate()
```

**复杂度**：新增一个 30 行的独立脚本，不动 `catalog.py`。

---

### 第二步：将发现范围配置化（可选）

**目标**：让 `catalog.py` 支持只同步契约中声明的知识库，而非全量拉取。

#### 2.1 修改 `catalog.py` 的 `main()`，增加过滤逻辑
```python
# 在 main() 中，获取 spaces 后
contract_spaces = load_contract().get("spaces", [])
target_space_ids = {s["id"] for s in contract_spaces} if contract_spaces else None

for space in spaces:
    if target_space_ids and space["space_id"] not in target_space_ids:
        print(f"⏭️  跳过未在契约中声明的知识库: {space['name']}")
        continue
    # 原有同步逻辑...
```

**复杂度**：约 5 行代码改动。

---

## 为什么这样就够了

| 治理需求 | 实现方式 | 改动量 |
|---------|---------|--------|
| 知道应该有什么 | `contract.yaml` 文件 | 新增一个文件 |
| 验证实际有什么 | `validate.py` 读取目录快照对比 | 新增一个脚本 |
| 按需同步 | `catalog.py` 读取契约过滤 | 5 行代码 |

**核心原则**：
- **不拆模块**：保持单文件脚本的便利性。
- **不引入抽象接口**：YAML + JSON 就是数据接口。
- **行为分离但工具独立**：发现/注册留在 `catalog.py`，验证留在 `validate.py`，各司其职。

---

## 最终文件结构

```
.
├── catalog.py          # 原有的下载脚本（微调过滤逻辑）
├── validate.py         # 新增的验证脚本
├── contract.yaml       # 新增的契约文件
├── config.py           # 不变
└── data/               # 目录快照存储（不变）
```
