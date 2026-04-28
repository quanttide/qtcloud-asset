#!/usr/bin/env python3
"""验证器 — 基于声明式策略验证资产是否符合契约"""

from __future__ import annotations

import os
from dataclasses import dataclass, field
from pathlib import Path
from typing import Any

import yaml


@dataclass
class ValidationRule:
    """单个验证规则"""
    selector: str
    mode: str
    required_categories: list[str] = field(default_factory=list)


@dataclass
class ValidationResult:
    """单个资产的验证结果"""
    name: str
    passed: bool
    rules_passed: int = 0
    rules_failed: int = 0
    details: list[dict[str, Any]] = field(default_factory=list)


@dataclass
class ValidationReport:
    """整体验证报告"""
    total_assets: int = 0
    passed_assets: int = 0
    failed_assets: int = 0
    results: list[ValidationResult] = field(default_factory=list)


def load_contract(contract_path: str | Path) -> dict[str, Any]:
    """读取契约文件"""
    path = Path(contract_path)
    if not path.exists():
        raise FileNotFoundError(f"契约文件不存在: {contract_path}")

    with open(path, encoding="utf-8") as f:
        return yaml.safe_load(f) or {}


def _match_selector(name: str, selector: str) -> bool:
    """检查名称是否匹配 selector"""
    if selector == "**":
        return True
    if selector.endswith("/**"):
        prefix = selector[:-3]
        return name.startswith(prefix)
    return name == selector


def _load_assets_from_dir(base_dir: Path) -> list[dict[str, Any]]:
    """从目录加载资产信息（基于 scanner.py 格式）"""
    assets = []
    if not base_dir.exists():
        return assets

    for entry in sorted(base_dir.iterdir()):
        if not entry.is_dir():
            continue
        categories = [d.name for d in entry.iterdir() if d.is_dir()]
        assets.append({
            "name": entry.name,
            "path": str(entry),
            "categories": categories,
        })
    return assets


def validate_asset(
    asset_name: str,
    asset_categories: list[str],
    policies: list[ValidationRule],
) -> ValidationResult:
    """验证单个资产是否符合策略"""
    result = ValidationResult(name=asset_name, passed=True)

    for policy in policies:
        if _match_selector(asset_name, policy.selector):
            if policy.mode == "ATOMIC":
                # 原子模式：必须包含所有 required_categories
                for cat in policy.required_categories:
                    if cat in asset_categories:
                        result.details.append({
                            "rule": f"必须包含分类 '{cat}'",
                            "status": "pass",
                        })
                        result.rules_passed += 1
                    else:
                        result.details.append({
                            "rule": f"必须包含分类 '{cat}'",
                            "status": "fail",
                        })
                        result.rules_failed += 1
                        result.passed = False

            elif policy.mode == "SCOPED":
                # 范围模式：资产存在即可
                result.details.append({
                    "rule": "资产存在",
                    "status": "pass",
                })
                result.rules_passed += 1

            # 首位命中
            break

    if result.rules_failed > 0:
        result.passed = False

    return result


def validate_directory(
    directory: Path,
    contract_path: Path | None = None,
) -> ValidationReport:
    """验证目录中的资产是否符合契约要求

    Args:
        directory: 要验证的目录
        contract_path: 契约文件路径，默认查找 .quanttide/asset/contract.yaml

    Returns:
        ValidationReport 验证报告
    """
    # 查找契约文件
    if contract_path is None:
        root = directory
        while root.parent != root:
            contract_path = root / ".quanttide" / "asset" / "contract.yaml"
            if contract_path.exists():
                break
            root = root.parent
        else:
            raise FileNotFoundError("未找到契约文件")

    contract = load_contract(contract_path)
    policies_raw = contract.get("validation", {}).get("policies", [])

    # 解析策略
    policies: list[ValidationRule] = []
    for p in policies_raw:
        policies.append(ValidationRule(
            selector=p.get("selector", "**"),
            mode=p.get("mode", "SCOPED"),
            required_categories=p.get("required_categories", []),
        ))

    if not policies:
        raise ValueError("契约文件中未定义验证策略")

    # 加载资产
    assets = _load_assets_from_dir(directory)

    # 验证每个资产
    report = ValidationReport(total_assets=len(assets))
    for asset in assets:
        result = validate_asset(
            asset["name"],
            asset.get("categories", []),
            policies,
        )
        report.results.append(result)
        if result.passed:
            report.passed_assets += 1
        else:
            report.failed_assets += 1

    return report
