from __future__ import annotations

import typer

from src.cli.app.services.planner import resolve_workflow, print_workflow_summary
from src.cli.app.repositories.file_operator import archive_product

app = typer.Typer(help="数字资产云 CLI")


@app.command()
def archive(
    contract: str = typer.Argument("journal_backup", help="契约名称"),
    slug: str = typer.Argument("product", help="产品标识"),
    product: str | None = typer.Option(None, "-p", "--product", help="指定产品"),
    pattern: str = typer.Option("*.md", "--pattern", help="文件匹配模式"),
    dry_run: bool = typer.Option(False, "-n", "--dry-run", help="仅预览"),
) -> None:
    """归档产品日志：从 journal 移动到 archive"""
    try:
        workflow = resolve_workflow(
            contract_name=contract,
            slug=slug,
            product=product,
            pattern=pattern,
        )
    except (FileNotFoundError, KeyError) as e:
        typer.secho(f"错误: {e}", fg=typer.colors.RED, err=True)
        raise typer.Exit(1)

    print_workflow_summary(workflow, dry_run=dry_run)

    success, skipped = [], []

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
                    label += "，已删除空源目录"
                typer.secho(f"  ✓ {task.product}: {label}", fg=typer.colors.GREEN)
            success.append(task.product)
        else:
            typer.secho(f"  ✗ {task.product}: {result.error}", fg=typer.colors.RED, err=True)
            skipped.append(task.product)

    total = len(workflow.tasks)
    suffix = f"，{len(skipped)} 跳过" if skipped else ""
    typer.echo(f"\n完成: {len(success)}/{total} 成功{suffix}")

    if not dry_run:
        typer.echo("提示: 请在 journal 和 archive 子模块中分别提交更改")

    raise typer.Exit(0 if not skipped else 1)


if __name__ == "__main__":
    app()
