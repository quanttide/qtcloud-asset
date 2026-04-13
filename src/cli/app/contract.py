#!/usr/bin/env python3
"""契约层 — 识别契约目录、加载契约配置。"""

from __future__ import annotations

from pathlib import Path
from typing import Any

import yaml


def find_contract_dir(start: Path | None = None) -> Path:
    """向上查找包含 .quanttide/asset/contract.yaml 的目录。

    Args:
        start: 起始目录，默认为当前工作目录

    Returns:
        契约目录路径

    Raises:
        FileNotFoundError: 未找到契约目录
    """
    current = start or Path.cwd()
    for parent in [current, *current.parents]:
        contract = parent / ".quanttide" / "asset" / "contract.yaml"
        if contract.exists():
            return parent
    raise FileNotFoundError(
        f"未找到契约目录，请确保存在 .quanttide/asset/contract.yaml"
    )


def load_contract(root: Path | None = None) -> dict[str, Any]:
    """加载契约配置。

    Args:
        root: 项目根目录，默认自动查找

    Returns:
        契约字典
    """
    root = root or find_contract_dir()
    path = root / ".quanttide" / "asset" / "contract.yaml"
    with open(path, encoding="utf-8") as f:
        data = yaml.safe_load(f)
    return data or {}


def get_skill(contract: dict, name: str) -> dict:
    """从契约中获取指定技能配置。

    Args:
        contract: 契约字典
        name: 技能名称

    Returns:
        技能配置

    Raises:
        KeyError: 技能不存在
    """
    skills = contract.get("skills", {})
    if name not in skills:
        available = ", ".join(skills.keys()) or "(空)"
        raise KeyError(f"找不到技能 '{name}'，可用: {available}")
    return skills[name]


def get_asset(contract: dict, name: str) -> dict:
    """从契约中获取指定资产配置。

    Args:
        contract: 契约字典
        name: 资产名称

    Returns:
        资产配置

    Raises:
        KeyError: 资产不存在
    """
    assets = contract.get("assets", {})
    if name not in assets:
        available = ", ".join(assets.keys()) or "(空)"
        raise KeyError(f"找不到资产 '{name}'，可用: {available}")
    return assets[name]
