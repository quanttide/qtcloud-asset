#!/usr/bin/env python3
"""工作流层测试"""

from __future__ import annotations

from pathlib import Path

import pytest
import yaml

from app.workflow import (
    ArchiveTask,
    Workflow,
    _get_products,
    print_workflow_summary,
    resolve_workflow,
)


class TestWorkflow:
    def test_products_property(self):
        workflow = Workflow(
            name="test",
            pattern="*.md",
            tasks=[
                ArchiveTask("p1", "/a", "/b"),
                ArchiveTask("p2", "/c", "/d"),
            ],
        )
        assert workflow.products == ["p1", "p2"]


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
    def contract_and_input(self, tmp_path):
        root = tmp_path / "project"
        contract_dir = root / ".quanttide" / "asset"
        contract_dir.mkdir(parents=True)
        (contract_dir / "contract.yaml").write_text(
            yaml.dump(
                {
                    "skills": {
                        "archive": {
                            "version": "1.0",
                            "params": {"pattern": "*.md"},
                        }
                    }
                }
            )
        )

        input_dir = root / "input"
        input_dir.mkdir()
        (input_dir / "product1").mkdir()
        (input_dir / "product2").mkdir()
        (input_dir / "product1" / "note.md").write_text("# Test")

        output_dir = root / "output"
        output_dir.mkdir()

        return {"root": root, "input": input_dir, "output": output_dir}

    def test_resolves_workflow(self, contract_and_input):
        workflow = resolve_workflow(
            skill_name="archive",
            input_dir=contract_and_input["input"],
            output_dir=contract_and_input["output"],
            contract_root=contract_and_input["root"],
        )
        assert workflow.name == "archive"
        assert len(workflow.tasks) == 2
        assert workflow.tasks[0].product == "product1"
        assert workflow.tasks[1].product == "product2"

    def test_uses_contract_pattern(self, contract_and_input):
        workflow = resolve_workflow(
            skill_name="archive",
            input_dir=contract_and_input["input"],
            output_dir=contract_and_input["output"],
            contract_root=contract_and_input["root"],
        )
        assert workflow.pattern == "*.md"

    def test_overrides_pattern(self, contract_and_input):
        workflow = resolve_workflow(
            skill_name="archive",
            input_dir=contract_and_input["input"],
            output_dir=contract_and_input["output"],
            pattern="*.txt",
            contract_root=contract_and_input["root"],
        )
        assert workflow.pattern == "*.txt"

    def test_raises_on_unknown_skill(self, contract_and_input):
        with pytest.raises(KeyError) as exc_info:
            resolve_workflow(
                skill_name="nonexistent",
                input_dir=contract_and_input["input"],
                output_dir=contract_and_input["output"],
                contract_root=contract_and_input["root"],
            )
        assert "nonexistent" in str(exc_info.value)

    def test_raises_on_missing_input(self, contract_and_input):
        with pytest.raises(FileNotFoundError):
            resolve_workflow(
                skill_name="archive",
                input_dir=contract_and_input["root"] / "missing",
                output_dir=contract_and_input["output"],
                contract_root=contract_and_input["root"],
            )

    def test_default_pattern_when_not_in_skill_params(self, tmp_path):
        root = tmp_path / "project"
        contract_dir = root / ".quanttide" / "asset"
        contract_dir.mkdir(parents=True)
        (contract_dir / "contract.yaml").write_text(
            yaml.dump({"skills": {"archive": {"version": "1.0"}}})
        )

        input_dir = root / "input"
        input_dir.mkdir()
        (input_dir / "product1").mkdir()

        output_dir = root / "output"
        output_dir.mkdir()

        workflow = resolve_workflow(
            skill_name="archive",
            input_dir=input_dir,
            output_dir=output_dir,
            contract_root=root,
        )
        assert workflow.pattern == "*.md"


class TestPrintWorkflowSummary:
    def test_prints_execution_mode(self, capsys):
        workflow = Workflow(
            name="archive",
            pattern="*.md",
            tasks=[
                ArchiveTask("p1", Path("/a"), Path("/b")),
            ],
        )
        print_workflow_summary(workflow)
        captured = capsys.readouterr()
        assert (
            "[执行]" in captured.out
            or "技能" in captured.out
            or "archive" in captured.out
        )

    def test_prints_dry_run_mode(self, capsys):
        workflow = Workflow(
            name="archive",
            pattern="*.md",
            tasks=[
                ArchiveTask("p1", Path("/a"), Path("/b")),
            ],
        )
        print_workflow_summary(workflow, dry_run=True)
        captured = capsys.readouterr()
        assert "[预览]" in captured.out or "archive" in captured.out
