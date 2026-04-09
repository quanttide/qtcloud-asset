#!/usr/bin/env python3
"""
操作层 — 纯文件系统操作，不依赖任何配置源

职责：
- 移动文件
- 失败回滚
- 清理空目录
- 预览模式

设计原则：
- 每个函数接收具体路径，返回结构化结果
- 不 import yaml / argparse / contracts
- 可独立单元测试
"""

from __future__ import annotations

import os
import shutil
from dataclasses import dataclass, field
from pathlib import Path


# ── 数据结构 ──────────────────────────────────────────


@dataclass
class FileResult:
    """单个文件的归档结果"""
    name: str
    success: bool
    reason: str = ""  # 成功时为空，失败/跳过时填写原因


@dataclass
class ArchiveResult:
    """单个产品目录的归档结果"""
    product: str
    total: int = 0
    moved: list[str] = field(default_factory=list)
    skipped: list[str] = field(default_factory=list)
    failed: list[str] = field(default_factory=list)
    source_removed: bool = False
    error: str | None = None

    @property
    def ok(self) -> bool:
        return self.error is None and len(self.failed) == 0


# ── 核心操作 ──────────────────────────────────────────


def _move_file(src: Path, dst: Path) -> None:
    """移动单个文件，保留元数据"""
    shutil.copy2(src, dst)
    os.unlink(src)


def _rollback(src_dir: Path, dst_dir: Path, moved: list[str]) -> list[str]:
    """将已移动的文件退回源目录（尽力回滚）"""
    rolled_back: list[str] = []
    for name in moved:
        dst = dst_dir / name
        src = src_dir / name
        if dst.exists() and not src.exists():
            try:
                shutil.copy2(dst, src)
                os.unlink(dst)
                rolled_back.append(name)
            except OSError:
                pass
    return rolled_back


def archive_product(
    src_dir: Path,
    dst_dir: Path,
    *,
    pattern: str = "*.md",
    dry_run: bool = False,
) -> ArchiveResult:
    """归档单个产品目录

    Args:
        src_dir:   源目录（journal 中的产品目录）
        dst_dir:   目标目录（archive 中的产品目录）
        pattern:   文件匹配模式
        dry_run:   预览模式，不实际移动

    Returns:
        ArchiveResult 结构化结果
    """
    product = src_dir.name
    result = ArchiveResult(product=product)

    # ── 检查源目录 ──
    if not src_dir.exists():
        result.error = f"源目录不存在: {src_dir}"
        return result

    # ── 收集文件 ──
    files = sorted(src_dir.glob(pattern))
    result.total = len(files)

    if not files:
        result.skipped = [f"(无匹配 {pattern} 文件)"]
        return result

    # ── 预览模式 ──
    if dry_run:
        result.moved = [f.name for f in files]
        return result

    # ── 创建目标目录 ──
    try:
        dst_dir.mkdir(parents=True, exist_ok=True)
    except OSError as e:
        result.error = f"无法创建目标目录: {e}"
        return result

    # ── 逐文件移动 ──
    for f in files:
        dst = dst_dir / f.name
        if dst.exists():
            result.skipped.append(f.name)
            continue
        try:
            _move_file(f, dst)
            result.moved.append(f.name)
        except OSError as e:
            result.failed.append(f.name)
            result.error = f"移动失败: {f.name} — {e}"

    # ── 失败时回滚 ──
    if result.failed:
        rolled = _rollback(src_dir, dst_dir, result.moved)
        result.moved.clear()

    # ── 清理空源目录 ──
    if not result.failed and not result.error:
        try:
            if not any(src_dir.iterdir()):
