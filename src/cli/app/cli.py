#!/usr/bin/env python3
"""CLI 命令定义"""

from __future__ import annotations

from pathlib import Path

import typer

__version__ = "0.1.0"
__stage__ = "alpha"
__stage_description__ = "本机验证"

from .file_operator import archive_product
from .workflow import resolve_workflow, print_workflow_summary
from .scanner import scan_local_directory, scan_output_to_dict
from .validator import validate_directory
from .render import (
    print_success,
    print_error,
    print_warning,
    print_info,
    print_header,
    print_asset_table,
    print_workflow_summary as _print_workflow_summary,
    print_validate_report,
)

app = typer.Typer(
    help="量潮数字资产云 CLI — 面向非技术用户的资产管理工具",
    no_args_is_help=True,
)


@app.command()
def run(
    skill: str = typer.Option(
        "archive-journal",
        "-s",
        "--skill",
        help="技能名称（默认 archive-journal）",
    ),
    input_dir: Path = typer.Option(
        Path.cwd(),
        "-i",
        "--input",
        help="输入目录（默认当前目录）",
    ),
    output_dir: Path = typer.Option(
        Path.cwd() / "output",
        "-o",
        "--output",
        help="输出目录（默认 ./output）",
    ),
    pattern: str = typer.Option(
        None,
        "-p",
        "--pattern",
        help="文件匹配模式（默认从契约读取）",
    ),
    dry_run: bool = typer.Option(False, "-n", "--dry-run", help="预览模式"),
    verbose: bool = typer.Option(False, "-v", "--verbose", help="详细输出"),
) -> None:
    """🚀 一键执行资产管理工作流（归档）"""
    print_header("资产管理工作流")

    try:
        workflow = resolve_workflow(
            skill_name=skill,
            input_dir=input_dir,
            output_dir=output_dir,
            pattern=pattern,
        )
    except (FileNotFoundError, KeyError) as e:
        print_error(f"无法启动工作流: {e}")
        raise typer.Exit(1)

    if verbose:
        _print_workflow_summary(
            workflow.name,
            workflow.pattern,
            len(workflow.tasks),
            dry_run=dry_run,
        )

    success_products = []
    failed_products = []

    for task in workflow.tasks:
        result = archive_product(
            task.src_dir,
            task.dst_dir,
            pattern=workflow.pattern,
            dry_run=dry_run,
        )

        if result.ok:
            if dry_run:
                typer.echo(f"  [预览] {task.product}: {', '.join(result.moved)}")
            else:
                label = f"归档 {len(result.moved)} 个文件"
                if result.source_removed:
                    label += "，已清理空源目录"
                print_success(f"{task.product}: {label}")
            success_products.append(task.product)
        else:
            print_error(f"{task.product}: {result.error}")
            failed_products.append(task.product)

    total = len(workflow.tasks)
    if failed_products:
        print_warning(f"完成: {len(success_products)}/{total} 成功，{len(failed_products)} 失败")
        raise typer.Exit(1)
    else:
        print_success(f"完成: {len(success_products)}/{total} 全部成功")
        raise typer.Exit(0)


@app.command()
def scan(
    path: Path = typer.Option(
        Path.cwd(),
        "-i",
        "--input",
        help="要扫描的目录（默认当前目录）",
    ),
    verbose: bool = typer.Option(False, "-v", "--verbose", help="详细输出"),
) -> None:
    """🔍 扫描目录，列出所有资产"""
    print_header("资产扫描")

    print_info(f"扫描目录: {path}")

    try:
        output = scan_local_directory(path)
    except FileNotFoundError as e:
        print_error(str(e))
        raise typer.Exit(1)
    except NotADirectoryError as e:
        print_error(str(e))
        raise typer.Exit(1)

    assets_dict = scan_output_to_dict(output)

    print_success(f"扫描完成: {output.total_dirs} 个资产目录，{output.total_files} 个文件")

    if verbose:
        print_asset_table(assets_dict["assets"])

    raise typer.Exit(0)


@app.command()
def validate(
    path: Path = typer.Option(
        Path.cwd(),
        "-i",
        "--input",
        help="要验证的目录（默认当前目录）",
    ),
    contract_path: Path = typer.Option(
        None,
        "-c",
        "--contract",
        help="契约文件路径（默认自动查找）",
    ),
    verbose: bool = typer.Option(False, "-v", "--verbose", help="详细输出"),
) -> None:
    """✅ 验证资产是否符合契约要求"""
    print_header("契约验证")

    print_info(f"验证目录: {path}")

    try:
        report = validate_directory(path, contract_path)
    except FileNotFoundError as e:
        print_error(f"验证失败: {e}")
        raise typer.Exit(1)
    except ValueError as e:
        print_error(f"验证配置错误: {e}")
        raise typer.Exit(1)

    print_validate_report(
        report.passed_assets,
        report.failed_assets,
        report.total_assets,
    )

    if verbose:
        for result in report.results:
            status = "✅" if result.passed else "❌"
            print(f"  {status} {result.name}")
            if not result.passed:
                for detail in result.details:
                    if detail["status"] == "fail":
                        print(f"      - {detail['rule']}")

    if report.failed_assets > 0:
        raise typer.Exit(1)
    else:
        raise typer.Exit(0)


@app.command()
def config(
    action: str = typer.Option(
        "show",
        "-a",
        "--action",
        help="操作: show（显示契约）/list（列出资产）",
    ),
    contract_path: Path = typer.Option(
        None,
        "-c",
        "--contract",
        help="契约文件路径（默认自动查找）",
    ),
) -> None:
    """⚙️ 查看契约配置"""
    from .contract import Contract

    print_header("契约配置")

    try:
        if contract_path:
            contract = Contract(root=contract_path.parent)
        else:
            contract = Contract()

        if action == "show":
            print_info(f"契约位置: {contract.root / '.quanttide/asset/contract.yaml'}")
            print_info(f"资产数: {len(contract.config.assets)}")
            print_info(f"技能数: {len(contract.config.skills)}")

            print("\n资产列表:")
            for name, asset in contract.config.assets.items():
                print(f"  - {name}: {asset.type}")

            print("\n技能列表:")
            for name, skill in contract.config.skills.items():
                print(f"  - {name}: v{skill.version}")

        elif action == "list":
            print_info("资产配置:")
            for name, asset in contract.config.assets.items():
                print(f"  {name}")

    except FileNotFoundError as e:
        print_error(f"契约文件不存在: {e}")
        raise typer.Exit(1)


@app.command()
def version(
) -> None:
    """🏷️ 显示版本和预发布阶段"""
    stage_icons = {
        "alpha": "🔵",
        "beta": "🟡",
        "rc": "🟠",
        "stable": "🟢",
    }
    icon = stage_icons.get(__stage__, "⚪")

    print_header("版本信息")
    print_info(f"qtcloud-asset-cli v{__version__}")
    print(f"\n  {icon} 阶段: {__stage__.upper()}")
    print(f"  📝 {__stage_description__}")

    raise typer.Exit(0)


if __name__ == "__main__":
    app()
