#!/usr/bin/env python3
"""
归档产品日志

将已处理的产品日志从 journal 子模块移动到 archive 子模块，
保持工作区整洁。

用法:
    python examples/backup_product_journal.py                          # 归档 product 标识下所有产品
    python examples/backup_product_journal.py journal_backup product    # 指定契约和标识
    python examples/backup_product_journal.py -p qtadmin               # 归档指定产品
    python examples/backup_product_journal.py --dry-run                # 仅预览

配置:
    contracts.yaml - 契约配置
"""

import argparse
import os
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


def get_products(journal_dir: Path) -> list[str]:
    """获取指定标识下的所有产品目录"""
    if not journal_dir.exists():
        print(f"错误: 找不到 {journal_dir}")
        sys.exit(1)
    return [d.name for d in sorted(journal_dir.iterdir()) if d.is_dir()]


def backup_product(src_dir: Path, dst_dir: Path, *, dry_run: bool = False) -> bool:
    """归档单个产品的所有日志文件

    Returns:
        True 如果执行了归档操作，False 如果跳过
    """
    if not src_dir.exists():
        print(f"  跳过: {src_dir.name} (源目录不存在)")
        return False

    md_files = list(src_dir.glob("*.md"))
    if not md_files:
        print(f"  跳过: {src_dir.name} (无 .md 文件)")
        return False

    if dry_run:
        print(f"  [预览] {src_dir.name}: {', '.join(f.name for f in md_files)}")
        return True

    dst_dir.mkdir(parents=True, exist_ok=True)

    moved = []
    try:
        for f in sorted(md_files):
            dst = dst_dir / f.name
            if dst.exists():
                print(f"    跳过: {f.name} (目标已存在)")
                continue
            shutil.copy2(f, dst)
            os.unlink(f)
            moved.append(f.name)
    except OSError as e:
        print(f"  错误: 移动文件失败 - {e}，正在回滚...")
        for name in moved:
            src_back = src_dir / name
            dst_back = dst_dir / name
            if dst_back.exists() and not src_back.exists():
                shutil.move(dst_back, src_back)
        return False

    # 清理空源目录
    try:
        if not any(src_dir.iterdir()):
            src_dir.rmdir()
            print(f"  已删除空目录: {src_dir}")
    except OSError:
        pass

    print(f"  已归档 {src_dir.name}: {', '.join(moved)}")
    return True


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="归档产品日志：从 journal 子模块移动到 archive 子模块"
    )
    parser.add_argument("contract", nargs="?", default="journal_backup",
                        help="契约名称 (默认: journal_backup)")
    parser.add_argument("slug", nargs="?", default="product",
                        help="标识 (默认: product)")
    parser.add_argument("-p", "--product", default=None,
                        help="指定产品名称（不指定则归档所有产品）")
    parser.add_argument("-n", "--dry-run", action="store_true",
                        help="仅预览，不实际移动文件")
    return parser.parse_args()


def main():
    args = parse_args()

    contract = load_contract(args.contract)
    paths = contract.get("paths", {})

    journal_base = Path(paths.get("journal", "docs/journal"))
    archive_base = Path(paths.get("archive", "docs/archive/journal"))

    journal_dir = journal_base / args.slug
    products = get_products(journal_dir)

    if args.product:
        if args.product not in products:
            print(f"错误: 找不到产品 {args.product}")
            print(f"可用产品: {', '.join(products)}")
            sys.exit(1)
        products = [args.product]

    action = "预览归档" if args.dry_run else "待归档"
    print(f"{action}产品: {', '.join(products)}")

    count = 0
    for product in products:
        src_dir = journal_base / args.slug / product
        dst_dir = archive_base / args.slug / product
        if backup_product(src_dir, dst_dir, dry_run=args.dry_run):
            count += 1

    print(f"\n完成! 共归档 {count} 个产品")
    if not args.dry_run:
        print("\n提示: 请在 journal 和 archive 子模块中分别提交更改")


if __name__ == "__main__":
    main()


