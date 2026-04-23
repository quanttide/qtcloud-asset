#!/usr/bin/env python3
"""基于声明式策略验证资产"""

import json
import yaml
import os
from pathlib import Path


def load_contract(contract_path="contract.yaml"):
    """读取契约文件"""
    if not os.path.exists(contract_path):
        print(f"❌ 契约文件不存在: {contract_path}")
        return {}
    with open(contract_path, encoding="utf-8") as f:
        return yaml.safe_load(f)


def load_catalog(base_dir="data/feishu_wiki"):
    """读取目录元数据"""
    catalog_path = os.path.join(base_dir, "metadata.json")
    if not os.path.exists(catalog_path):
        print(f"❌ 目录元数据不存在: {catalog_path}")
        return {}
    with open(catalog_path, encoding="utf-8") as f:
        return json.load(f)


def match_selector(space_name, selector):
    """检查 space 是否匹配 selector"""
    if selector == "**":
        return True
    if selector.endswith("/**"):
        prefix = selector[:-3]
        return space_name.startswith(prefix)
    return space_name == selector


def get_space_categories(space):
    """获取 space 的所有一级分类名称"""
    nodes = space.get("nodes", [])
    categories = set()

    for node in nodes:
        parent = node.get("parent_token", "")
        # parent_token 为空的是根节点
        # 一级分类的 parent 是根节点
        # 这里简单处理：直接收集所有分类节点的 title
        if node.get("has_child", False):
            categories.add(node.get("title", ""))

    return categories


def validate_space(space, policies):
    """验证单个 space，返回结果列表"""
    space_name = space.get("name", "")
    node_titles = [n.get("title", "") for n in space.get("nodes", [])]
    results = []

    for policy in policies:
        selector = policy["selector"]
        mode = policy["mode"]

        if match_selector(space_name, selector):
            if mode == "ATOMIC":
                # 原子模式：必须包含 required_categories
                required = policy.get("required_categories", [])
                for cat in required:
                    if cat in node_titles:
                        results.append({
                            "rule": f"必须包含分类 '{cat}'",
                            "status": "pass"
                        })
                    else:
                        results.append({
                            "rule": f"必须包含分类 '{cat}'",
                            "status": "fail"
                        })

            elif mode == "SCOPED":
                # 范围模式：space 存在即可
                results.append({
                    "rule": "space 存在",
                    "status": "pass"
                })

            # 首位命中
            break

    return results


def main():
    print("=== 资产验证报告 ===\n")

    contract = load_contract()
    catalog = load_catalog()

    policies = contract.get("validation", {}).get("policies", [])
    if not policies:
        print("❌ 契约文件中未定义验证策略")
        return

    print(f"契约版本: {contract.get('spec_version', 'unknown')}")
    print(f"验证策略数: {len(policies)}\n")

    all_spaces = catalog.get("spaces", [])
    print(f"发现 Space 数: {len(all_spaces)}\n")

    print("-" * 50)

    total_pass = 0
    total_fail = 0

    for space in all_spaces:
        space_name = space.get("name", "")
        results = validate_space(space, policies)

        if not results:
            continue

        print(f"\n📁 {space_name}")

        for r in results:
            if r["status"] == "pass":
                total_pass += 1
                print(f"  ✅ {r['rule']}")
            else:
                total_fail += 1
                print(f"  ❌ {r['rule']}")

    print("\n" + "-" * 50)
    print(f"\n=== 统计 ===")
    print(f"通过: {total_pass}")
    print(f"失败: {total_fail}")
    print(f"总计: {total_pass + total_fail}")

    if total_fail == 0:
        print("\n✅ 所有验证通过！")
    else:
        print(f"\n⚠️  有 {total_fail} 项验证失败")


if __name__ == "__main__":
    main()
