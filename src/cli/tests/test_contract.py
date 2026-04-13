#!/usr/bin/env python3
"""契约层测试"""

from __future__ import annotations

import pytest
import yaml
from app.contract import find_contract_dir, get_asset, get_skill, load_contract


class TestFindContractDir:
    def test_finds_contract_in_current_dir(self, tmp_path):
        root = tmp_path / "project"
        contract_dir = root / ".quanttide" / "asset"
        contract_dir.mkdir(parents=True)
        (contract_dir / "contract.yaml").write_text(
            "skills:\n  test:\n    title: Test\n"
        )

        result = find_contract_dir(root)
        assert result == root

    def test_finds_contract_in_parent_dir(self, tmp_path):
        root = tmp_path / "project"
        contract_dir = root / ".quanttide" / "asset"
        contract_dir.mkdir(parents=True)
        (contract_dir / "contract.yaml").write_text(
            "skills:\n  test:\n    title: Test\n"
        )

        subdir = root / "src" / "cli"
        subdir.mkdir(parents=True)

        result = find_contract_dir(subdir)
        assert result == root

    def test_raises_when_not_found(self, tmp_path):
        with pytest.raises(FileNotFoundError):
            find_contract_dir(tmp_path / "no_contract")


class TestLoadContract:
    def test_loads_contract(self, tmp_path):
        root = tmp_path / "project"
        contract_dir = root / ".quanttide" / "asset"
        contract_dir.mkdir(parents=True)
        contract_file = contract_dir / "contract.yaml"
        contract_file.write_text(
            "skills:\n  archive:\n    title: Archive\n    transform:\n      pattern: '*.md'\n"
        )

        data = load_contract(root)
        assert "skills" in data
        assert "archive" in data["skills"]

    def test_returns_empty_dict_for_empty_file(self, tmp_path):
        root = tmp_path / "project"
        contract_dir = root / ".quanttide" / "asset"
        contract_dir.mkdir(parents=True)
        (contract_dir / "contract.yaml").write_text("")

        data = load_contract(root)
        assert data == {}


class TestGetSkill:
    def test_returns_skill(self):
        contract = {
            "skills": {
                "archive": {"title": "Archive", "transform": {"pattern": "*.md"}},
            }
        }
        skill = get_skill(contract, "archive")
        assert skill["title"] == "Archive"

    def test_raises_on_missing_skill(self):
        contract = {"skills": {"existing": {}}}
        with pytest.raises(KeyError) as exc_info:
            get_skill(contract, "missing")
        assert "missing" in str(exc_info.value)
        assert "existing" in str(exc_info.value)


class TestGetAsset:
    def test_returns_asset(self):
        contract = {
            "assets": {
                "docs": {"title": "Docs", "type": "docs", "path": "docs/"},
            }
        }
        asset = get_asset(contract, "docs")
        assert asset["title"] == "Docs"

    def test_raises_on_missing_asset(self):
        contract = {"assets": {"existing": {}}}
        with pytest.raises(KeyError) as exc_info:
            get_asset(contract, "missing")
        assert "missing" in str(exc_info.value)
        assert "existing" in str(exc_info.value)
