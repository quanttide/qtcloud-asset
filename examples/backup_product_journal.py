#!/usr/bin/env python3
"""
归档产品日志

将已处理的产品日志从 journal 子模块移动到 archive 子模块，
保持工作区整洁。

用法:
    python examples/backup_product_journal.py                    # 归档 product 标识下所有产品
    python examples/backup_product_journal.py product qtadmin    # 归档指定产品

配置:
    contracts.yaml - 契约配置
"""

import shutil
import sys
from pathlib import Path

import yaml

CONTRACTS_FILE = Path(__file__).parent.parent / "contracts.yaml"


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


def get_products(journal_dir: Path) -> list[str]:
    """获取指定标识下的所有产品目录"""
    if not journal_dir.exists():
        print(f"错误: 找不到 {journal_dir}")
        sys.exit(1)
    return [d.name for d in sorted(journal_dir.iterdir()) if d.is_dir()]


def backup_product(src_dir: Path, dst_dir: Path) -> bool:
    """归档单个产品的所有日志文件"""
    if not src_dir.exists():
        print(f"  跳过: {src_dir.name} (源目录不存在)")
        return False

    md_files = list(src_dir.glob("*.md"))
    if not md_files:
        print(f"  跳过: {src_dir.name} (无 .md 文件)")
        return False

    dst_dir.mkdir(parents=True, exist_ok=True)

    moved = []
    for f in sorted(md_files):
        dst = dst_dir / f.name
        shutil.move(str(f), str(dst))
        moved.append(f.name)

    if not any(src_dir.iterdir()):
        src_dir.rmdir()
        print(f"  已删除空目录: {src_dir}")

    print(f"  已归档 {src_dir.name}: {', '.join(moved)}")
    return True


def main():
    contract_name = (
        sys.argv[1]
        if len(sys.argv) > 1 and not sys.argv[1].startswith("--")
        else "journal_backup"
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
    archive_base = Path(paths.get("archive", "docs/archive/journal"))

    journal_dir = journal_base / slug
    products = get_products(journal_dir)

    if target_product:
        if target_product not in products:
            print(f"错误: 找不到产品 {target_product}")
            sys.exit(1)
        products = [target_product]

    print(f"待归档产品: {', '.join(products)}")

    count = 0
    for product in products:
        src_dir = journal_base / slug / product
        dst_dir = archive_base / slug / product
        if backup_product(src_dir, dst_dir):
            count += 1

    print(f"\n完成! 共归档 {count} 个产品")
    print("\n提示: 请在 journal 和 archive 子模块中分别提交更改")


if __name__ == "__main__":
    main()
