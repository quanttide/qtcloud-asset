#!/usr/bin/env python3
"""契约层测试"""

from __future__ import annotations

import pytest
import yaml

from app.contract import Contract


class TestContract:
    def test_find_root(self, tmp_path):
        root = tmp_path / "project"
        contract_dir = root / ".quanttide" / "asset"
        contract_dir.mkdir(parents=True)
        (contract_dir / "contract.yaml").write_text(
            "skills:\n  test:\n    title: Test\n"
        )

        result = Contract.find_root(root)
        assert result == root

    def test_find_root_from_subdir(self, tmp_path):
        root = tmp_path / "project"
        contract_dir = root / ".quanttide" / "asset"
        contract_dir.mkdir(parents=True)
        (contract_dir / "contract.yaml").write_text(
            "skills:\n  test:\n    title: Test\n"
        )

        subdir = root / "src" / "cli"
        subdir.mkdir(parents=True)

        result = Contract.find_root(subdir)
        assert result == root

    def test_raises_when_not_found(self, tmp_path):
        with pytest.raises(FileNotFoundError):
            Contract.find_root(tmp_path / "no_contract")

    def test_loads_contract(self, tmp_path):
        root = tmp_path / "project"
        contract_dir = root / ".quanttide" / "asset"
        contract_dir.mkdir(parents=True)
        contract_file = contract_dir / "contract.yaml"
        contract_file.write_text(
            "skills:\n  archive:\n    title: Archive\n    transform:\n      pattern: '*.md'\n"
        )

        contract = Contract(root)
        assert "archive" in contract.skills
        assert "skills" not in contract.assets

    def test_returns_empty_dict_for_empty_file(self, tmp_path):
        root = tmp_path / "project"
        contract_dir = root / ".quanttide" / "asset"
        contract_dir.mkdir(parents=True)
        (contract_dir / "contract.yaml").write_text("")

        contract = Contract(root)
        assert contract.assets == {}
        assert contract.skills == {}

    def test_raises_on_missing_file(self, tmp_path):
        with pytest.raises(FileNotFoundError):
            Contract(tmp_path / "nonexistent")

    def test_get_skill(self, tmp_path):
        root = tmp_path / "project"
        contract_dir = root / ".quanttide" / "asset"
        contract_dir.mkdir(parents=True)
        (contract_dir / "contract.yaml").write_text(
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
        assert skill["title"] == "Archive"

    def test_raises_on_missing_skill(self, tmp_path):
        root = tmp_path / "project"
        contract_dir = root / ".quanttide" / "asset"
        contract_dir.mkdir(parents=True)
        (contract_dir / "contract.yaml").write_text(
            yaml.dump({"skills": {"existing": {}}})
        )

        contract = Contract(root)
        with pytest.raises(KeyError) as exc_info:
            contract.get_skill("missing")
        assert "missing" in str(exc_info.value)
        assert "existing" in str(exc_info.value)

    def test_get_asset(self, tmp_path):
        root = tmp_path / "project"
        contract_dir = root / ".quanttide" / "asset"
        contract_dir.mkdir(parents=True)
        (contract_dir / "contract.yaml").write_text(
            yaml.dump(
                {"assets": {"docs": {"title": "Docs", "type": "docs", "path": "docs/"}}}
            )
        )

        contract = Contract(root)
        asset = contract.get_asset("docs")
        assert asset["title"] == "Docs"

    def test_raises_on_missing_asset(self, tmp_path):
        root = tmp_path / "project"
        contract_dir = root / ".quanttide" / "asset"
        contract_dir.mkdir(parents=True)
        (contract_dir / "contract.yaml").write_text(
            yaml.dump({"assets": {"existing": {}}})
        )

        contract = Contract(root)
        with pytest.raises(KeyError) as exc_info:
            contract.get_asset("missing")
        assert "missing" in str(exc_info.value)
        assert "existing" in str(exc_info.value)
