from __future__ import annotations

import pytest
import yaml

from src.provider.app.services.planner import (
    ArchiveTask,
    Workflow,
    _get_contract,
    _get_products,
    _load_yaml,
    resolve_workflow,
)


class TestWorkflow:
    def test_products_property(self):
        workflow = Workflow(
            name="test",
            slug="test",
            pattern="*.md",
            tasks=[
                ArchiveTask("p1", "/a", "/b"),
                ArchiveTask("p2", "/c", "/d"),
            ],
        )
        assert workflow.products == ["p1", "p2"]


class TestLoadYaml:
    def test_loads_yaml_file(self, tmp_path):
        yaml_file = tmp_path / "test.yaml"
        yaml_file.write_text("key: value\nlist:\n  - 1\n  - 2")
        data = _load_yaml(yaml_file)
        assert data == {"key": "value", "list": [1, 2]}

    def test_raises_on_missing_file(self, tmp_path):
        with pytest.raises(FileNotFoundError):
            _load_yaml(tmp_path / "nonexistent.yaml")

    def test_returns_empty_dict_for_empty_file(self, tmp_path):
        yaml_file = tmp_path / "empty.yaml"
        yaml_file.write_text("")
        data = _load_yaml(yaml_file)
        assert data == {}


class TestGetContract:
    def test_returns_contract(self):
        data = {"contracts": {"test": {"name": "Test"}}}
        contract = _get_contract(data, "test")
        assert contract == {"name": "Test"}

    def test_raises_on_missing_contract(self):
        data = {"contracts": {"existing": {}}}
        with pytest.raises(KeyError) as exc_info:
            _get_contract(data, "missing")
        assert "missing" in str(exc_info.value)
        assert "existing" in str(exc_info.value)


class TestGetProducts:
    def test_returns_sorted_subdirs(self, tmp_path):
        journal = tmp_path / "journal"
        journal.mkdir()
        (journal / "zebra").mkdir()
        (journal / "apple").mkdir()
        (journal / "banana").mkdir()
        products = _get_products(journal)
        assert products == ["apple", "banana", "zebra"]

    def test_raises_on_missing_dir(self, tmp_path):
        with pytest.raises(FileNotFoundError):
            _get_products(tmp_path / "nonexistent")


class TestResolveWorkflow:
    @pytest.fixture
    def contracts_and_journal(self, tmp_path):
        contracts_file = tmp_path / "contracts.yaml"
        contracts = {
            "contracts": {
                "test_contract": {
                    "name": "测试契约",
                    "paths": {
                        "journal": str(tmp_path / "journal"),
                        "archive": str(tmp_path / "archive"),
                    },
                }
            }
        }
        contracts_file.write_text(yaml.dump(contracts))

        journal = tmp_path / "journal"
        slug_dir = journal / "product"
        product_dir = slug_dir / "qtcloud-asset"
        product_dir.mkdir(parents=True)
        (product_dir / "note.md").write_text("# Test")

        return {"contracts_file": contracts_file, "journal": journal, "slug_dir": slug_dir}

    def test_resolves_workflow(self, contracts_and_journal):
        workflow = resolve_workflow(
            "test_contract",
            "product",
            contracts_file=contracts_and_journal["contracts_file"],
        )
        assert workflow.name == "test_contract"
        assert workflow.slug == "product"
        assert len(workflow.tasks) == 1
        assert workflow.tasks[0].product == "qtcloud-asset"

    def test_filters_by_product(self, contracts_and_journal):
        (contracts_and_journal["slug_dir"] / "other").mkdir()
        workflow = resolve_workflow(
            "test_contract",
            "product",
            product="qtcloud-asset",
            contracts_file=contracts_and_journal["contracts_file"],
        )
        assert len(workflow.tasks) == 1
        assert workflow.tasks[0].product == "qtcloud-asset"

    def test_raises_on_unknown_product(self, contracts_and_journal):
        with pytest.raises(KeyError) as exc_info:
            resolve_workflow(
                "test_contract",
                "product",
                product="nonexistent",
                contracts_file=contracts_and_journal["contracts_file"],
            )
        assert "nonexistent" in str(exc_info.value)
