#!/usr/bin/env python3
"""
使用本地 Ollama 模型从产品日志生成工作蓝图

读取 docs/journal/<slug>/<product>/ 下的日志文件，
调用本地 Ollama 模型生成结构化的产品蓝图，
参考 qtcloud-think 蓝图的格式。

用法:
    python examples/generate_blueprint_ollama.py                    # 处理 product 标识
    python examples/generate_blueprint_ollama.py product qtadmin    # 处理指定产品

配置:
    contracts.yaml - 契约配置
"""

import sys
from pathlib import Path

import requests
import yaml

CONTRACTS_FILE = Path(__file__).parent / "contracts.yaml"
OLLAMA_URL = "http://localhost:11434"
MODEL = "qwen2.5-coder:3b"

SYSTEM_PROMPT = """你是一个产品蓝图分析师。请根据产品日志内容，生成结构化的产品工作蓝图。

严格遵循以下要求：
1. 不使用粗体、表格、emoji 等装饰
2. 使用纯列表和编号结构
3. 内容简洁，去除冗余描述
4. 使用中文输出

输出格式：

# 产品名称蓝图

## 产品定位

- 核心定位描述
- 关键价值点

## 核心设计

### 原子定义

1. 核心概念定义
2. 关键规则

### 功能设计

- 功能1：描述
- 功能2：描述

## 数据处理流程

- 数据流转描述
- 关键处理节点

## 关键洞察

### 主题1

- 洞察点

### 主题2

- 洞察点

## 架构原则

- 原则1
- 原则2

## 验证方向

- 验证点1
- 验证点2

## 困惑与挑战

1. 挑战1
2. 挑战2"""


def load_contract(name: str) -> dict:
    """从 contracts.yaml 加载契约配置"""
    if not CONTRACTS_FILE.exists():
        print(f"错误: 找不到契约配置文件 {CONTRACTS_FILE}")
        sys.exit(1)

    with open(CONTRACTS_FILE, encoding="utf-8") as f:
        data = yaml.safe_load(f)

    contracts = data.get("contracts", {})
    if name not in contracts:
        print(f"错误: 找不到契约 {name}")
        print(f"可用契约: {', '.join(contracts.keys())}")
        sys.exit(1)

    return contracts[name]


def get_paths(contract: dict) -> dict:
    """从契约配置获取路径"""
    return contract.get("paths", {})


def call_ollama(prompt: str, system: str, model: str = MODEL) -> str:
    """调用本地 Ollama 模型"""
    url = f"{OLLAMA_URL}/api/generate"
    payload = {
        "model": model,
        "prompt": prompt,
        "system": system,
        "stream": False,
    }
    resp = requests.post(url, json=payload, timeout=300)
    resp.raise_for_status()
    return resp.json().get("response", "")


def load_product_journal(journal_base: Path, slug: str, product: str) -> str:
    """加载指定产品的所有日志内容"""
    journal_dir = journal_base / slug / product
    if not journal_dir.exists():
        print(f"错误: 找不到 {journal_dir}")
        sys.exit(1)

    contents = []
    for f in sorted(journal_dir.glob("*.md")):
        content = f.read_text(encoding="utf-8")
        contents.append(f"# {f.stem}\n\n{content}")
    return "\n\n---\n\n".join(contents)


def generate_blueprint(slug: str, product: str) -> str:
    """调用 Ollama 生成产品蓝图"""
    journal_content = load_product_journal(slug, product)
    prompt = f"以下是产品日志内容，请生成工作蓝图：\n\n{journal_content}"
    print(f"  正在调用 Ollama ({MODEL}) 生成 {product} 蓝图...")
    return call_ollama(prompt, SYSTEM_PROMPT)


def main():
    contract_name = (
        sys.argv[1]
        if len(sys.argv) > 1 and not sys.argv[1].startswith("--")
        else "feishu_to_github"
    )
    slug = (
        sys.argv[2]
        if len(sys.argv) > 2 and not sys.argv[2].startswith("--")
        else "product"
    )
    target_product = sys.argv[3] if len(sys.argv) > 3 else None

    contract = load_contract(contract_name)
    paths = get_paths(contract)

    journal_base = Path(paths.get("journal", "docs/journal"))
    roadmap_base = Path(paths.get("roadmap", "docs/roadmap"))

    journal_dir = journal_base / slug
    if not journal_dir.exists():
        print(f"错误: 找不到 {journal_dir}")
        sys.exit(1)

    products = [d.name for d in sorted(journal_dir.iterdir()) if d.is_dir()]
    if target_product:
        if target_product not in products:
            print(f"错误: 找不到产品 {target_product}")
            sys.exit(1)
        products = [target_product]

    output_dir = roadmap_base / slug
    output_dir.mkdir(parents=True, exist_ok=True)

    for product in products:
        blueprint = generate_blueprint(slug, product)
        if not blueprint.strip():
            print(f"  警告: {product} 生成结果为空，跳过")
            continue

        output = output_dir / f"{product}.md"
        output.write_text(blueprint, encoding="utf-8")
        print(f"  输出: {output}")

    print(f"\n完成!")


if __name__ == "__main__":
    main()
