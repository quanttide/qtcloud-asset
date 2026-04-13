from __future__ import annotations

from pathlib import Path

import typer

from .file_operator import archive_product
from .workflow import print_workflow_summary, resolve_workflow

app = typer.Typer(help="量潮数字资产云 CLI")


@app.command()
def run(
    input: Path = typer.Option(
        ..., "-i", "--input", help="数据源目录", exists=True, file_okay=False
    ),
    skill: str = typer.Option(..., "-s", "--skill", help="技能名称"),
    output: Path = typer.Option(
        ..., "-o", "--output", help="输出目标目录", file_okay=False
    ),
    pattern: str = typer.Option(
        None, "-p", "--pattern", help="文件匹配模式（覆盖契约配置）"
    ),
    dry_run: bool = typer.Option(False, "-n", "--dry-run", help="预览模式"),
    verbose: bool = typer.Option(False, "-v", "--verbose", help="详细输出"),
) -> None:
    """数据转换：输入 → 契约(转换) → 输出"""
    try:
        workflow = resolve_workflow(
            skill_name=skill,
            input_dir=input,
            output_dir=output,
            pattern=pattern,
        )
    except (FileNotFoundError, KeyError) as e:
        typer.secho(f"错误: {e}", fg=typer.colors.RED, err=True)
        raise typer.Exit(1)

    if verbose:
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
                typer.echo(f"  [OK] {task.product}: {label}")
            success.append(task.product)
        else:
            typer.echo(f"  [FAIL] {task.product}: {result.error}", err=True)
            skipped.append(task.product)

    total = len(workflow.tasks)
    suffix = f"，{len(skipped)} 跳过" if skipped else ""
    typer.echo(f"\n完成: {len(success)}/{total} 成功{suffix}")

    raise typer.Exit(0 if not skipped else 1)


if __name__ == "__main__":
    app()
