#!/usr/bin/env python3
"""归档产品日志：从 journal 移动到 archive"""

import shutil

from .config import ARCHIVE, JOURNAL


def archive_product(product: str, slug: str = "product") -> dict:
    src = JOURNAL / slug / product
    dst = ARCHIVE / slug / product
    if not src.exists():
        return {"ok": False, "error": f"源目录不存在: {src}"}

    files = list(src.glob("*.md"))
    if not files:
        return {"ok": True, "moved": [], "skipped": ["无 *.md 文件"]}

    dst.mkdir(parents=True, exist_ok=True)
    moved, skipped = [], []

    for f in files:
        target = dst / f.name
        if target.exists():
            skipped.append(f.name)
        else:
            shutil.copy2(f, target)
            f.unlink()
            moved.append(f.name)

    if not any(src.iterdir()):
        src.rmdir()

    return {"ok": True, "moved": moved, "skipped": skipped}


if __name__ == "__main__":
    import sys

    product = sys.argv[1] if len(sys.argv) > 1 else "qtcloud-asset"
    result = archive_product(product)
    print(f"[{'OK' if result['ok'] else 'FAIL'}] {product}")
    print(f"  移动: {result['moved']}")
    print(f"  跳过: {result['skipped']}")
