#!/usr/bin/env python3
"""契约层 — 识别契约目录、加载契约配置。"""

from __future__ import annotations

from pathlib import Path
from typing import Any

import yaml


class Contract:
    """数字资产契约。

    自动向上查找 .quanttide/asset/contract.yaml 并加载配置。
    """

    CONTRACT_PATH = Path(".quanttide") / "asset" / "contract.yaml"

    def __init__(self, root: Path | None = None) -> None:
        self.root = root or self.find_root()
        self.path = self.root / self.CONTRACT_PATH
        self.data = self._load()

    @staticmethod
    def find_root(start: Path | None = None) -> Path:
        """向上查找包含契约文件的目录。

        Args:
            start: 起始目录，默认为当前工作目录

        Returns:
            契约根目录

        Raises:
            FileNotFoundError: 未找到契约文件
        """
        current = start or Path.cwd()
        for parent in [current, *current.parents]:
            if (parent / Contract.CONTRACT_PATH).exists():
                return parent
        raise FileNotFoundError(f"未找到契约文件 {Contract.CONTRACT_PATH}")

    def _load(self) -> dict[str, Any]:
        """加载契约 YAML 文件。"""
        if not self.path.exists():
            raise FileNotFoundError(f"契约文件不存在: {self.path}")
        with open(self.path, encoding="utf-8") as f:
            data = yaml.safe_load(f)
        return data or {}

    @property
    def assets(self) -> dict[str, Any]:
        """返回所有资产定义。"""
        return self.data.get("assets", {})

    @property
    def skills(self) -> dict[str, Any]:
        """返回所有技能定义。"""
        return self.data.get("skills", {})

    def get_skill(self, name: str) -> dict[str, Any]:
        """获取指定技能配置。

        Args:
            name: 技能名称

        Returns:
            技能配置字典

        Raises:
            KeyError: 技能不存在
        """
        if name not in self.skills:
            available = ", ".join(self.skills.keys()) or "(空)"
            raise KeyError(f"找不到技能 '{name}'，可用: {available}")
        return self.skills[name]

    def get_asset(self, name: str) -> dict[str, Any]:
        """获取指定资产配置。

        Args:
            name: 资产名称

        Returns:
            资产配置字典

        Raises:
            KeyError: 资产不存在
        """
        if name not in self.assets:
            available = ", ".join(self.assets.keys()) or "(空)"
            raise KeyError(f"找不到资产 '{name}'，可用: {available}")
        return self.assets[name]
