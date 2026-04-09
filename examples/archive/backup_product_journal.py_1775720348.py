#!/usr/bin/env python3
"""
归档产品日志 — CLI 入口

将已处理的产品日志从 journal 子模块移动到 archive 子模块，
保持工作区整洁。

用法:
    python backup_product_journal.py                            # 归档 product 标识下所有产品
    python backup_product_journal.py journal_backup product      # 指定契约和标识
    python backup_product_journal.py -p qtadmin                 # 归档指定产品
    python backup_product_journal.py --dry-run                  # 仅预览

配置:
    contracts.yaml - 契约配置
"""

from __future__ import annotations

import argparse
import sys

from archiver_config import resolve_workflow, print_workflow_summary
from archiver_ops import archive_product, cleanup_empty_dirs


def parse_args(argv: list[str] | None = None) -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="归档产品日志：从 journal 移动到 archive",
    )
    parser.add_argument(
        "contract", nargs="?", default="journal_backup",
        help="契约名称 (默认: journal_backup)",
    )
    parser.add_argument(
        "slug", nargs="?", default="product",
        help="产品标识 (默认: product)",
    )
    parser.add_argument(
        "-p", "--product", default=None,
        help="指定产品（不指定则归档全部）",
    )
    parser.add_argument(
        "--pattern", default="*.md",
        help="文件匹配模式 (默认: *.md)",
    )
    parser.add_argument(
        "-n", "--dry-run", action="store_true",
        help="仅预览，不实际移动",
    )
    return parser.parse_args(argv)


def main(argv: list[str] | None = None) -> int:
    args = parse_args(argv)

    # ── 配置层：解析工作流 ──
    try:
        workflow = resolve_workflow(
            contract_name=args.contract,
            slug=args.slug,
            product=args.product,
            pattern=args.pattern,
        )
    except (FileNotFoundError, KeyError) as e:
        print(f"错误: {e}", file=sys.stderr)
        return 1

    print_workflow_summary(workflow, dry_run=args.dry_run)

    # ── 操作层：执行归档 ──
    success, skipped = [], []
    empty_src_dirs: list = []

    for task in workflow.tasks:
        result = archive_product(
            task.src_dir,
            task.dst_dir,
            pattern=workflow.pattern,
            dry_run=args.dry_run,
        )

        if result.ok:
            if args.dry_run:
                print(f"  [预览] {task.product}: {', '.join(result.moved)}")
            else:
                label = f"归档 {len(result.moved)} 个文件"
                if result.source_removed:
                    label += "，已删除空源目录"
                print(f"  ✓ {task.product}: {label}")
            success.append(task.product)
            if result.source_removed:
                empty_src_dirs.append(task.src_dir.parent)
        else:
            print(f"  ✗ {task.product}: {result.error}", file=sys.stderr)
            skipped.append(task.product)

    # ── 清理空父目录 ──
    if not args.dry_run and empty_src_dirs:
        removed = cleanup_empty_dirs(empty_src_dirs)
        if removed:
            names = ", ".join(str(d) for d in removed)
            print(f"\n清理空目录: {names}")

    # ── 汇总 ──
    total = len(workflow.tasks)
    suffix = f"，{len(skipped)} 跳过" if skipped else ""
    print(f"\n完成: {len(success)}/{total} 成功{suffix}")

    if not args.dry_run:
        print("提示: 请在 journal 和 archive 子模块中分别提交更改")

    return 0 if not skipped else 1


if __name__ == "__main__":
    sys.exit(main())
