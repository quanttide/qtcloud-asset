#!/usr/bin/env python3
"""
归档产品日志

将已处理的产品日志从 journal 子模块移动到 archive 子模块，
保持工作区整洁。

用法:
    python examples/backup_product_journal.py                    # 归档 product 标识下所有产品
    python examples/backup_product_journal.py product qtadmin    # 归档指定产品
"""

import shutil
import sys
from pathlib import Path

JOURNAL_BASE = Path("docs/journal")
ARCHIVE_BASE = Path("docs/archive/journal")


def get_products(slug: str) -> list[str]:
    """获取指定标识下的所有产品目录"""
    journal_dir = JOURNAL_BASE / slug
    if not journal_dir.exists():
        print(f"错误: 找不到 {journal_dir}")
        sys.exit(1)
    return [d.name for d in sorted(journal_dir.iterdir()) if d.is_dir()]


def backup_product(slug: str, product: str) -> bool:
    """归档单个产品的所有日志文件"""
    src_dir = JOURNAL_BASE / slug / product
    dst_dir = ARCHIVE_BASE / slug / product

    if not src_dir.exists():
        print(f"  跳过: {product} (源目录不存在)")
        return False

    md_files = list(src_dir.glob("*.md"))
    if not md_files:
        print(f"  跳过: {product} (无 .md 文件)")
        return False

    # 创建目标目录
    dst_dir.mkdir(parents=True, exist_ok=True)

    # 移动文件
    moved = []
    for f in sorted(md_files):
        dst = dst_dir / f.name
        shutil.move(str(f), str(dst))
        moved.append(f.name)

    # 清理空目录
    if not any(src_dir.iterdir()):
        src_dir.rmdir()
        print(f"  已删除空目录: {src_dir}")

    print(f"  已归档 {product}: {', '.join(moved)}")
    return True


def main():
    slug = (
        sys.argv[1]
        if len(sys.argv) > 1 and not sys.argv[1].startswith("--")
        else "product"
    )
    target_product = sys.argv[2] if len(sys.argv) > 2 else None

    products = get_products(slug)
    if target_product:
        if target_product not in products:
            print(f"错误: 找不到产品 {target_product}")
            sys.exit(1)
        products = [target_product]

    print(f"待归档产品: {', '.join(products)}")

    count = 0
    for product in products:
        if backup_product(slug, product):
            count += 1

    print(f"\n完成! 共归档 {count} 个产品")
    print("\n提示: 请在 journal 和 archive 子模块中分别提交更改")


if __name__ == "__main__":
    main()
