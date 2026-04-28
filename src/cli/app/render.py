#!/usr/bin/env python3
"""渲染层 — 基于 rich 的彩色输出"""

from __future__ import annotations

import sys
from dataclasses import dataclass
from typing import Any

# Windows UTF-8 处理
if sys.platform == "win32":
    try:
        sys.stdout.reconfigure(encoding="utf-8", errors="replace")
        sys.stderr.reconfigure(encoding="utf-8", errors="replace")
    except Exception:
        pass

try:
    from rich.console import Console
    from rich.table import Table
    from rich.panel import Panel
    from rich.progress import Progress, SpinnerColumn, TextColumn
    _RICH_AVAILABLE = True
except ImportError:
    _RICH_AVAILABLE = False
    Console = None
    Table = None
    Panel = None
    Progress = None

console = Console() if _RICH_AVAILABLE else None


@dataclass
class ScanResult:
    """扫描结果"""
    space_name: str
    total_dirs: int
    total_files: int
    assets: list[dict[str, Any]]
    synced_at: str = ""


def _emoji(text: str) -> str:
    """在 Windows GBK 环境下移除 emoji 避免编码错误"""
    if sys.platform != "win32":
        return text
    emap = {
        "✅": "[OK]", "❌": "[FAIL]", "⚠️ ": "[WARN] ",
        "ℹ️ ": "[INFO] ", "🚀": "[RUN]", "🔍": "[SCAN]",
        "⚙️": "[CONFIG]", "✅": "[OK]",
    }
    for k, v in emap.items():
        text = text.replace(k, v)
    return text


def print_success(message: str) -> None:
    """打印成功信息"""
    msg = _emoji(f"✅ {message}")
    if _RICH_AVAILABLE:
        console.print(f"[green]{msg}[/green]")
    else:
        print(msg)


def print_error(message: str) -> None:
    """打印错误信息"""
    msg = _emoji(f"❌ {message}")
    if _RICH_AVAILABLE:
        console.print(f"[red]{msg}[/red]")
    else:
        print(msg)


def print_warning(message: str) -> None:
    """打印警告信息"""
    msg = _emoji(f"⚠️  {message}")
    if _RICH_AVAILABLE:
        console.print(f"[yellow]{msg}[/yellow]")
    else:
        print(msg)


def print_info(message: str) -> None:
    """打印普通信息"""
    msg = _emoji(f"ℹ️  {message}")
    if _RICH_AVAILABLE:
        console.print(f"[blue]{msg}[/blue]")
    else:
        print(msg)


def print_header(message: str) -> None:
    """打印标题"""
    if _RICH_AVAILABLE:
        console.print(f"\n[bold cyan]{message}[/bold cyan]")
    else:
        print(f"\n=== {message} ===")


def print_asset_table(assets: list[dict[str, Any]]) -> None:
    """打印资产表格"""
    if not assets:
        print_warning("未发现任何资产")
        return

    if _RICH_AVAILABLE:
        table = Table(title="资产清单")
        table.add_column("名称", style="cyan")
        table.add_column("类型", style="magenta")
        table.add_column("路径", style="dim")

        for asset in assets:
            table.add_row(
                asset.get("name", "-"),
                asset.get("type", "-"),
                asset.get("path", "-"),
            )
        console.print(table)
    else:
        print("\n资产清单:")
        for asset in assets:
            print(f"  - {asset.get('name', '-')} ({asset.get('type', '-')})")


def print_workflow_summary(name: str, pattern: str, tasks_count: int, dry_run: bool = False) -> None:
    """打印工作流摘要"""
    mode = "预览" if dry_run else "执行"
    if _RICH_AVAILABLE:
        console.print(Panel(
            f"[{mode}] 技能: {name} | 模式: {pattern} | 产品数: {tasks_count}",
            title="工作流摘要"
        ))
    else:
        print(f"\n[{mode}] 技能: {name}")
        print(f"[{mode}] 模式: {pattern}")
        print(f"[{mode}] 产品数: {tasks_count}")


def print_validate_report(passed: int, failed: int, total: int) -> None:
    """打印验证报告"""
    if _RICH_AVAILABLE:
        if failed == 0:
            console.print(Panel(
                f"[green]所有验证通过！({passed}/{total})[/green]",
                title="验证报告"
            ))
        else:
            console.print(Panel(
                f"[red]有 {failed} 项验证失败 ({passed}/{total} 通过)[/red]",
                title="验证报告"
            ))
    else:
        if failed == 0:
            print(f"\n✅ 所有验证通过！({passed}/{total})")
        else:
            print(f"\n⚠️  有 {failed} 项验证失败 ({passed}/{total} 通过)")


def create_progress() -> Progress | None:
    """创建进度条"""
    if _RICH_AVAILABLE:
        return Progress(
            SpinnerColumn(),
            TextColumn("[progress.description]{task.description}"),
            console=console,
        )
    return None
