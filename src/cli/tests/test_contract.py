#!/usr/bin/env python3
"""契约层测试"""

from __future__ import annotations

import pytest
import yaml

from app.contract import AssetConfig, Contract, SkillConfig


class TestContract:
    def test_find_root(self, tmp_path):
        root = tmp_path / "project"
        (root / ".quanttide" / "asset").mkdir(parents=True)
        (root / ".quanttide" / "asset" / "contract.yaml").write_text(
            "skills:\n  test:\n    version: '1.0'\n"
        )
        assert Contract.find_root(root) == root

    def test_find_root_from_subdir(self, tmp_path):
        root = tmp_path / "project"
        (root / ".quanttide" / "asset").mkdir(parents=True)
        (root / ".quanttide" / "asset" / "contract.yaml").write_text(
            "skills:\n  test:\n    version: '1.0'\n"
        )
        subdir = root / "src" / "cli"
        subdir.mkdir(parents=True)
        assert Contract.find_root(subdir) == root

    def test_raises_when_not_found(self, tmp_path):
        with pytest.raises(FileNotFoundError):
            Contract.find_root(tmp_path / "no_contract")

    def test_loads_contract(self, tmp_path):
        root = tmp_path / "project"
        (root / ".quanttide" / "asset").mkdir(parents=True)
        (root / ".quanttide" / "asset" / "contract.yaml").write_text(
            "skills:\n  archive:\n    version: '1.0'\n    entrypoint: archive\n"
        )
        contract = Contract(root)
        assert "archive" in contract.config.skills

    def test_raises_on_missing_file(self, tmp_path):
        with pytest.raises(FileNotFoundError):
            Contract(tmp_path / "nonexistent")

    def test_get_skill(self, tmp_path):
        root = tmp_path / "project"
        (root / ".quanttide" / "asset").mkdir(parents=True)
        (root / ".quanttide" / "asset" / "contract.yaml").write_text(
            yaml.dump(
                {"skills": {"archive": {"version": "1.0", "entrypoint": "archive"}}}
            )
        )
        contract = Contract(root)
        skill = contract.get_skill("archive")
        assert skill.version == "1.0"
        assert skill.entrypoint == "archive"

    def test_raises_on_missing_skill(self, tmp_path):
        root = tmp_path / "project"
        (root / ".quanttide" / "asset").mkdir(parents=True)
        (root / ".quanttide" / "asset" / "contract.yaml").write_text(
            yaml.dump({"skills": {"existing": {"version": "1.0"}}})
        )
        contract = Contract(root)
        with pytest.raises(KeyError) as exc_info:
            contract.get_skill("missing")
        assert "missing" in str(exc_info.value)

    def test_get_asset(self, tmp_path):
        root = tmp_path / "project"
        (root / ".quanttide" / "asset").mkdir(parents=True)
        (root / ".quanttide" / "asset" / "contract.yaml").write_text(
            yaml.dump({"assets": {"docs": {"type": "docs", "provider": "local"}}})
        )
        contract = Contract(root)
        asset = contract.get_asset("docs")
        assert asset.type == "docs"

    def test_raises_on_missing_asset(self, tmp_path):
        root = tmp_path / "project"
        (root / ".quanttide" / "asset").mkdir(parents=True)
        (root / ".quanttide" / "asset" / "contract.yaml").write_text(
            yaml.dump({"assets": {"existing": {"type": "docs", "provider": "local"}}})
        )
        contract = Contract(root)
        with pytest.raises(KeyError) as exc_info:
            contract.get_asset("missing")
        assert "missing" in str(exc_info.value)


class TestModels:
    def test_asset_config(self):
        asset = AssetConfig(type="docs", provider="local", metadata={"key": "value"})
        assert asset.type == "docs"
        assert asset.provider == "local"
        assert asset.metadata == {"key": "value"}

    def test_skill_config(self):
        skill = SkillConfig(version="2.0", entrypoint="main", params={"verbose": True})
        assert skill.version == "2.0"
        assert skill.entrypoint == "main"
        assert skill.params == {"verbose": True}

    def test_skill_default_values(self):
        skill = SkillConfig()
        assert skill.version == "1.0"
        assert skill.entrypoint == ""
        assert skill.params == {}

    def test_frozen_models(self):
        skill = SkillConfig()
        with pytest.raises(Exception):  # pydantic raises ValidationError or similar
            skill.version = "2.0"
