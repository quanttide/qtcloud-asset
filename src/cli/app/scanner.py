#!/usr/bin/env python3
"""扫描器 — 扫描本地目录，返回资产清单"""

from __future__ import annotations

import os
from dataclasses import dataclass, field
from pathlib import Path
from typing import Any


@dataclass
class AssetInfo:
    """单个资产的信息"""
    name: str
    path: str
    asset_type: str = "directory"
    size: int = 0
    files_count: int = 0
    categories: list[str] = field(default_factory=list)


@dataclass
class ScanOutput:
    """扫描结果"""
    root_path: str
    total_dirs: int = 0
    total_files: int = 0
    assets: list[AssetInfo] = field(default_factory=list)
    synced_at: str = ""


def _get_dir_stats(path: Path) -> tuple[int, int]:
    """获取目录的文件数和总大小"""
    files_count = 0
    total_size = 0
    for _, _, files in os.walk(path):
        files_count += len(files)
        for f in files:
            total_size += os.path.getsize(os.path.join(path, f)) if os.path.exists(os.path.join(path, f)) else 0
    return files_count, total_size


def _guess_asset_type(path: Path) -> str:
    """根据目录名推测资产类型"""
    name_lower = path.name.lower()
    if "议事档案" in name_lower or "archive" in name_lower:
        return "议事档案"
    if "journal" in name_lower:
        return "日志"
    if any(kw in name_lower for kw in ["prd", "brd", "ixd", "add", "qa"]):
        return "文档"
    if any(kw in name_lower for kw in ["docs", "doc"]):
        return "文档"
    return "通用"


def scan_local_directory(path: Path) -> ScanOutput:
    """扫描本地目录，返回资产清单

    Args:
        path: 要扫描的目录路径

    Returns:
        ScanOutput 扫描结果
    """
    if not path.exists():
        raise FileNotFoundError(f"目录不存在: {path}")

    if not path.is_dir():
        raise NotADirectoryError(f"不是目录: {path}")

    assets: list[AssetInfo] = []
    total_dirs = 0
    total_files = 0

    for entry in sorted(path.iterdir()):
        if not entry.is_dir():
            continue

        total_dirs += 1
        files_count, size = _get_dir_stats(entry)
        total_files += files_count

        # 获取子目录作为分类
        categories = [d.name for d in entry.iterdir() if d.is_dir()]

        assets.append(AssetInfo(
            name=entry.name,
            path=str(entry),
            asset_type=_guess_asset_type(entry),
            size=size,
            files_count=files_count,
            categories=categories,
        ))

    return ScanOutput(
        root_path=str(path),
        total_dirs=total_dirs,
        total_files=total_files,
        assets=assets,
        synced_at="",  # 本地扫描不涉及同步时间
    )


def scan_output_to_dict(output: ScanOutput) -> dict[str, Any]:
    """将 ScanOutput 转换为可序列化的字典"""
    return {
        "root_path": output.root_path,
        "total_dirs": output.total_dirs,
        "total_files": output.total_files,
        "assets": [
            {
                "name": a.name,
                "path": a.path,
                "type": a.asset_type,
                "size": a.size,
                "files_count": a.files_count,
                "categories": a.categories,
            }
            for a in output.assets
        ],
        "synced_at": output.synced_at,
    }
