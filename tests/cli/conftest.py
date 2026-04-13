"""测试夹具 — 统一引用 assets/fixtures 作为唯一事实源"""

from __future__ import annotations

from pathlib import Path

import pytest


@pytest.fixture
def fixtures_root() -> Path:
    """返回 assets/fixtures 目录"""
    return Path(__file__).parents[3] / "assets" / "fixtures"


@pytest.fixture
def journal_dir(fixtures_root: Path) -> Path:
    """返回 fixtures/docs/journal 目录"""
    return fixtures_root / "docs" / "journal"


@pytest.fixture
def archive_dir(fixtures_root: Path) -> Path:
    """返回 fixtures/docs/archive 目录"""
    return fixtures_root / "docs" / "archive"


@pytest.fixture
def contract_file(fixtures_root: Path) -> Path:
    """返回 fixtures/.quanttide/asset/contract.yaml"""
    return fixtures_root / ".quanttide" / "asset" / "contract.yaml"
