#!/usr/bin/env python3
"""契约层 — 基于 Pydantic 的契约模型。"""

from __future__ import annotations

from pathlib import Path
from typing import Any

import yaml
from pydantic import BaseModel, ConfigDict, Field, model_validator


class SkillConfig(BaseModel):
    """技能配置。"""

    model_config = ConfigDict(frozen=True)

    version: str = "1.0"
    entrypoint: str = ""
    params: dict[str, Any] = Field(default_factory=dict)


class AssetConfig(BaseModel):
    """资产配置。"""

    model_config = ConfigDict(frozen=True)

    type: str
    provider: str = ""
    metadata: dict[str, Any] = Field(default_factory=dict)


class ContractSchema(BaseModel):
    """契约配置的强类型定义。"""

    model_config = ConfigDict(frozen=True)

    assets: dict[str, AssetConfig] = Field(default_factory=dict)
    skills: dict[str, SkillConfig] = Field(default_factory=dict)

    @model_validator(mode="after")
    def validate_logic(self) -> ContractSchema:
        # 跨字段校验逻辑
        return self


class Contract:
    """数字资产契约。

    自动向上查找 .quanttide/asset/contract.yaml 并加载为只读 Pydantic 模型。
    """

    CONTRACT_PATH = Path(".quanttide") / "asset" / "contract.yaml"

    def __init__(self, root: Path | None = None) -> None:
        self.root = root or self.find_root()
        self._config = self._load()

    @property
    def config(self) -> ContractSchema:
        """契约配置（只读）。"""
        return self._config

    @staticmethod
    def find_root(start: Path | None = None) -> Path:
        """向上查找包含契约文件的目录。"""
        current = start or Path.cwd()
        for parent in [current, *current.parents]:
            if (parent / Contract.CONTRACT_PATH).exists():
                return parent
        raise FileNotFoundError(f"未找到契约文件 {Contract.CONTRACT_PATH}")

    def _load(self) -> ContractSchema:
        """加载契约 YAML 文件并解析为 Pydantic 模型。"""
        path = self.root / self.CONTRACT_PATH
        if not path.exists():
            raise FileNotFoundError(f"契约文件不存在: {path}")
        with open(path, encoding="utf-8") as f:
            raw: dict[str, Any] = yaml.safe_load(f) or {}
        return ContractSchema.model_validate(raw)

    def get_skill(self, name: str) -> SkillConfig:
        """获取指定技能配置。"""
        if name not in self.config.skills:
            available = ", ".join(self.config.skills.keys()) or "(空)"
            raise KeyError(f"找不到技能 '{name}'，可用: {available}")
        return self.config.skills[name]

    def get_asset(self, name: str) -> AssetConfig:
        """获取指定资产配置。"""
        if name not in self.config.assets:
            available = ", ".join(self.config.assets.keys()) or "(空)"
            raise KeyError(f"找不到资产 '{name}'，可用: {available}")
        return self.config.assets[name]
