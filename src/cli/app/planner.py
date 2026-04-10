#!/usr/bin/env python3
"""
配置层 — 将契约配置解析为可执行的工作流
"""

from __future__ import annotations

from dataclasses import dataclass, field
from pathlib import Path

import yaml


CONTRACTS_FILE = Path(__file__).parent.parent / "contracts.yaml"


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


def _load_yaml(path: Path) -> dict:
    if not path.exists():
        raise FileNotFoundError(f"契约配置文件不存在: {path}")
    with open(path, encoding="utf-8") as f:
        data = yaml.safe_load(f)
    return data or {}


def _get_contract(data: dict, name: str) -> dict:
    contracts = data.get("contracts", {})
    if name not in contracts:
        available = ", ".join(contracts.keys()) or "(空)"
        raise KeyError(f"找不到契约 '{name}'，可用: {available}")
    return contracts[name]


def _get_products(directory: Path) -> list[str]:
    if not directory.exists():
        raise FileNotFoundError(f"目录不存在: {directory}")
    return sorted(d.name for d in directory.iterdir() if d.is_dir())


def resolve_workflow_simple(
    contract_name: str,
    input_dir: Path,
    output_dir: Path,
    pattern: str = "*.md",
    contracts_file: Path | None = None,
) -> Workflow:
    """解析契约，生成简单工作流

    Args:
        contract_name: 契约名称
        input_dir: 源目录
        output_dir: 目标目录
        pattern: 文件匹配模式
        contracts_file: 自定义配置文件路径
    """
    file = contracts_file or CONTRACTS_FILE
    data = _load_yaml(file)
    contract = _get_contract(data, contract_name)

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
        name=contract_name,
        pattern=pattern,
        tasks=tasks,
    )


def print_workflow_summary(workflow: Workflow, *, dry_run: bool = False) -> None:
    """打印工作流摘要"""
    mode = "预览" if dry_run else "执行"
    print(f"[{mode}] 契约: {workflow.name}")
    print(f"[{mode}] 产品 ({len(workflow.tasks)}): {', '.join(workflow.products)}")
    print(f"[{mode}] 模式: {workflow.pattern}")
