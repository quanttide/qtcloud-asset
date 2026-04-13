#!/usr/bin/env python3
"""契约层测试"""

from __future__ import annotations

import pytest
import yaml

from app.contract import Asset, Contract, Skill, Transform


class TestContract:
    def test_find_root(self, tmp_path):
        root = tmp_path / "project"
        (root / ".quanttide" / "asset").mkdir(parents=True)
        (root / ".quanttide" / "asset" / "contract.yaml").write_text(
            "skills:\n  test:\n    title: Test\n"
        )
        assert Contract.find_root(root) == root

    def test_find_root_from_subdir(self, tmp_path):
        root = tmp_path / "project"
        (root / ".quanttide" / "asset").mkdir(parents=True)
        (root / ".quanttide" / "asset" / "contract.yaml").write_text(
            "skills:\n  test:\n    title: Test\n"
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
            "skills:\n  archive:\n    title: Archive\n    transform:\n      pattern: '*.md'\n"
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
                {
                    "skills": {
                        "archive": {
                            "title": "Archive",
                            "transform": {"pattern": "*.md"},
                        }
                    }
                }
            )
        )
        contract = Contract(root)
        skill = contract.get_skill("archive")
        assert skill.title == "Archive"
        assert skill.transform.pattern == "*.md"

    def test_raises_on_missing_skill(self, tmp_path):
        root = tmp_path / "project"
        (root / ".quanttide" / "asset").mkdir(parents=True)
        (root / ".quanttide" / "asset" / "contract.yaml").write_text(
            yaml.dump({"skills": {"existing": {"title": "Existing"}}})
        )
        contract = Contract(root)
        with pytest.raises(KeyError) as exc_info:
            contract.get_skill("missing")
        assert "missing" in str(exc_info.value)

    def test_get_asset(self, tmp_path):
        root = tmp_path / "project"
        (root / ".quanttide" / "asset").mkdir(parents=True)
        (root / ".quanttide" / "asset" / "contract.yaml").write_text(
            yaml.dump(
                {
                    "assets": {
                        "docs": {
                            "title": "Docs",
                            "type": "docs",
                            "category": "doc",
                            "path": "docs/",
                        }
                    }
                }
            )
        )
        contract = Contract(root)
        asset = contract.get_asset("docs")
        assert asset.title == "Docs"

    def test_raises_on_missing_asset(self, tmp_path):
        root = tmp_path / "project"
        (root / ".quanttide" / "asset").mkdir(parents=True)
        (root / ".quanttide" / "asset" / "contract.yaml").write_text(
            yaml.dump(
                {
                    "assets": {
                        "existing": {
                            "title": "Existing",
                            "type": "docs",
                            "category": "doc",
                            "path": "e/",
                        }
                    }
                }
            )
        )
        contract = Contract(root)
        with pytest.raises(KeyError) as exc_info:
            contract.get_asset("missing")
        assert "missing" in str(exc_info.value)


class TestModels:
    def test_asset_model(self):
        asset = Asset(
            title="Docs", type="docs", category="doc", path="docs/", description="Test"
        )
        assert asset.title == "Docs"
        assert asset.description == "Test"

    def test_skill_model(self):
        skill = Skill(
            title="Archive", description="Test", transform=Transform(pattern="*.txt")
        )
        assert skill.title == "Archive"
        assert skill.transform.pattern == "*.txt"

    def test_skill_default_transform(self):
        skill = Skill(title="Test")
        assert skill.transform.pattern == "*.md"
