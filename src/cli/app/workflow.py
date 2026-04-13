#!/usr/bin/env python3
"""配置层 — 将契约配置解析为可执行的工作流。"""

from __future__ import annotations

from dataclasses import dataclass, field
from pathlib import Path

from .contract import Contract


@dataclass
class ArchiveTask:
    """单个归档任务"""

    product: str
    src_dir: Path
    dst_dir: Path


@dataclass
class Workflow:
    """归档工作流"""

    name: str
    pattern: str
    tasks: list[ArchiveTask] = field(default_factory=list)

    @property
    def products(self) -> list[str]:
        return [t.product for t in self.tasks]


def _get_products(directory: Path) -> list[str]:
    if not directory.exists():
        raise FileNotFoundError(f"目录不存在: {directory}")
    return sorted(d.name for d in directory.iterdir() if d.is_dir())


def resolve_workflow(
    skill_name: str,
    input_dir: Path,
    output_dir: Path,
    pattern: str | None = None,
    contract_root: Path | None = None,
) -> Workflow:
    """解析契约，生成工作流。

    Args:
        skill_name: 技能名称
        input_dir: 源目录
        output_dir: 目标目录
        pattern: 文件匹配模式，优先使用技能配置中的 pattern
        contract_root: 契约根目录，默认自动查找
    """
    root = contract_root or Contract.find_root()
    contract = Contract(root)
    skill = contract.get_skill(skill_name)

    # 优先使用技能配置中的 pattern
    effective_pattern = pattern or skill.transform.pattern

    products = _get_products(input_dir)

    tasks = [
        ArchiveTask(
            product=p,
            src_dir=input_dir / p,
            dst_dir=output_dir / p,
        )
        for p in products
    ]

    return Workflow(
        name=skill_name,
        pattern=effective_pattern,
        tasks=tasks,
    )


def print_workflow_summary(workflow: Workflow, *, dry_run: bool = False) -> None:
    """打印工作流摘要"""
    mode = "预览" if dry_run else "执行"
    print(f"[{mode}] 技能: {workflow.name}")
    print(f"[{mode}] 产品 ({len(workflow.tasks)}): {', '.join(workflow.products)}")
    print(f"[{mode}] 模式: {workflow.pattern}")
