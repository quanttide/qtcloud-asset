#!/usr/bin/env python3
"""量潮数字资产云 CLI — 入口文件"""

from __future__ import annotations

import sys
from pathlib import Path

import typer

from .app.cli import app, run as run_cmd


def main() -> None:
    """入口：如果没有指定命令，默认执行 run"""
    # 去掉模块名，只保留用户参数
    args = sys.argv[1:]

    # 如果没有任何参数，默认执行 run
    if not args:
        # 创建一个最小化的 Typer 来调用 run
        standalone = typer.Typer(rich_help_panel=None)
        standalone.command()(run_cmd)
        standalone()
    else:
        # 有参数，正常走命令解析
        app()


if __name__ == "__main__":
    main()
