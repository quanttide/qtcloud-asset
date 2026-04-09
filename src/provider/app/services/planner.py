#!/usr/bin/env python3
"""
配置层 — 将契约配置解析为可执行的工作流

职责：
- 加载 contracts.yaml
- 解析路径模板
- 生成具体的归档任务列表

设计原则：
- 不执行任何文件系统写操作
- 不 import archiver_ops（通过 dataclass 解耦）
- 可独立测试（传入 mock data）
"""

from __future__ import annotations

from dataclasses import dataclass, field
from pathlib import Path

import yaml


CONTRACTS_FILE = Path(__file__).parent.parent / "contracts.yaml"


# ── 数据结构（层间契约）──────────────────────────────


@dataclass
class ArchiveTask:
    """单个归档任务"""
    product: str
    src_dir: Path
    dst_dir: Path


@dataclass
class Workflow:
    """归档工作流 — 配置层的唯一输出"""
    name: str
    slug: str
    pattern: str
    tasks: list[ArchiveTask] = field(default_factory=list)

    @property
    def products(self) -> list[str]:
        return [t.product for t in self.tasks]


# ── 配置加载 ──────────────────────────────────────────


def _load_yaml(path: Path) -> dict:
    """加载 YAML 文件"""
    if not path.exists():
        raise FileNotFoundError(f"契约配置文件不存在: {path}")
    with open(path, encoding="utf-8") as f:
        data = yaml.safe_load(f)
    return data or {}


def _get_contract(data: dict, name: str) -> dict:
    """从 YAML 数据中提取指定契约"""
    contracts = data.get("contracts", {})
    if name not in contracts:
        available = ", ".join(contracts.keys()) or "(空)"
        raise KeyError(f"找不到契约 \'{name}\'，可用: {available}")
    return contracts[name]


def _get_products(journal_dir: Path) -> list[str]:
    """扫描 journal 目录下的产品子目录"""
    if not journal_dir.exists():
        raise FileNotFoundError(f"目录不存在: {journal_dir}")
    return sorted(d.name for d in journal_dir.iterdir() if d.is_dir())


# ── 公开接口 ──────────────────────────────────────────


def resolve_workflow(
    contract_name: str,
    slug: str,
    *,
    product: str | None = None,
    pattern: str = "*.md",
    contracts_file: Path | None = None,
) -> Workflow:
    """解析契约配置，生成归档工作流

    Args:
        contract_name:   contracts.yaml 中的契约名称
        slug:            产品标识（如 product / qtadmin）
        product:         指定单个产品（None 表示全部）
        pattern:         文件匹配模式
        contracts_file:  自定义配置文件路径（测试用）

    Returns:
        Workflow 对象，包含所有待执行的 ArchiveTask

    Raises:
        FileNotFoundError: 配置文件或目录不存在
        KeyError:          契约不存在
    """
    file = contracts_file or CONTRACTS_FILE
    data = _load_yaml(file)
    contract = _get_contract(data, contract_name)
    paths = contract.get("paths", {})

    journal_base = Path(paths.get("journal", "docs/journal"))
    archive_base = Path(paths.get("archive", "docs/archive/journal"))

    journal_dir = journal_base / slug
    products = _get_products(journal_dir)

    if product:
        if product not in products:
            raise KeyError(
                f"找不到产品 \'{product}\'，可用: {', '.join(products)}"
            )
        products = [product]

    tasks = [
        ArchiveTask(
            product=p,
            src_dir=journal_base / slug / p,
            dst_dir=archive_base / slug / p,
        )
        for p in products
    ]

    return Workflow(
        name=contract_name,
        slug=slug,
        pattern=pattern,
        tasks=tasks,
    )


def print_workflow_summary(workflow: Workflow, *, dry_run: bool = False) -> None:
    """打印工作流摘要（纯展示，不执行操作）"""
    mode = "预览" if dry_run else "执行"
    print(f"[{mode}] 契约: {workflow.name}  标识: {workflow.slug}")
    print(f"[{mode}] 产品 ({len(workflow.tasks)}): {', '.join(workflow.products)}")
    print(f"[{mode}] 模式: {workflow.pattern}")
