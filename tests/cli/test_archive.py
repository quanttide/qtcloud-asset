#!/usr/bin env python3
"""集成测试 — CLI 端到端测试"""

from __future__ import annotations

import os
import shutil
import tempfile
from pathlib import Path

import pytest
from typer.testing import CliRunner

from app.cli import app


@pytest.fixture
def runner():
    return CliRunner()


@pytest.fixture
def temp_project():
    with tempfile.TemporaryDirectory() as tmpdir:
        tmp = Path(tmpdir)
        root = tmp / "project"
        root.mkdir()
        contract_dir = root / ".quanttide" / "asset"
        contract_dir.mkdir(parents=True)
        (contract_dir / "contract.yaml").write_text(
            """skills:
  archive-journal:
    version: "1.0"
    params:
      pattern: "*.md"
  archive:
    version: "1.0"
    params:
      pattern: "*.md"
"""
        )
        input_dir = root / "input"
        input_dir.mkdir()
        output_dir = root / "output"
        output_dir.mkdir()
        yield {
            "root": root,
            "input": input_dir,
            "output": output_dir,
            "contract_dir": contract_dir,
        }


@pytest.fixture
def populated_project(temp_project):
    input_dir = temp_project["input"]
    for product in ["product1", "product2", "product3"]:
        product_dir = input_dir / product
        product_dir.mkdir()
        (product_dir / "note.md").write_text(f"# {product}")
        (product_dir / "draft.txt").write_text("draft")
    return temp_project


class TestArchiveHelp:
    def test_help(self):
        runner = CliRunner()
        result = runner.invoke(app, ["--help"])
        assert result.exit_code == 0
        assert "--input" in result.stdout
        assert "--skill" in result.stdout
        assert "--output" in result.stdout

    def test_no_args(self):
        runner = CliRunner()
        result = runner.invoke(app, [])
        assert result.exit_code != 0


class TestArchiveWorkflow:
    def test_dry_run_archives_nothing(self, populated_project, monkeypatch):
        runner = CliRunner()
        monkeypatch.chdir(populated_project["root"])
        result = runner.invoke(
            app,
            [
                "-s",
                "archive-journal",
                "-i",
                str(populated_project["input"]),
                "-o",
                str(populated_project["output"]),
                "-n",
            ],
        )
        assert result.exit_code == 0
        assert "预览" in result.stdout
        assert not (populated_project["output"] / "product1").exists()

    def test_actual_archive(self, populated_project, monkeypatch):
        runner = CliRunner()
        monkeypatch.chdir(populated_project["root"])
        result = runner.invoke(
            app,
            [
                "-s",
                "archive-journal",
                "-i",
                str(populated_project["input"]),
                "-o",
                str(populated_project["output"]),
            ],
        )
        assert result.exit_code == 0
        assert "完成" in result.stdout or "3/3" in result.stdout
        assert (populated_project["output"] / "product1" / "note.md").exists()
        assert (populated_project["output"] / "product2" / "note.md").exists()
        assert (populated_project["output"] / "product3" / "note.md").exists()

    def test_verbose_mode(self, populated_project, monkeypatch):
        runner = CliRunner()
        monkeypatch.chdir(populated_project["root"])
        result = runner.invoke(
            app,
            [
                "-s",
                "archive-journal",
                "-i",
                str(populated_project["input"]),
                "-o",
                str(populated_project["output"]),
                "-v",
            ],
        )
        assert result.exit_code == 0

    def test_pattern_override(self, populated_project, monkeypatch):
        runner = CliRunner()
        monkeypatch.chdir(populated_project["root"])
        result = runner.invoke(
            app,
            [
                "-s",
                "archive-journal",
                "-i",
                str(populated_project["input"]),
                "-o",
                str(populated_project["output"]),
                "-p",
                "*.txt",
            ],
        )
        assert result.exit_code == 0


class TestArchiveErrors:
    def test_unknown_skill(self, temp_project, monkeypatch):
        runner = CliRunner()
        monkeypatch.chdir(temp_project["root"])
        result = runner.invoke(
            app,
            [
                "-s",
                "nonexistent",
                "-i",
                str(temp_project["input"]),
                "-o",
                str(temp_project["output"]),
            ],
        )
        assert result.exit_code == 1
        assert "错误" in result.stderr or "找不到" in result.stderr

    def test_missing_input_dir(self, temp_project, monkeypatch):
        runner = CliRunner()
        monkeypatch.chdir(temp_project["root"])
        result = runner.invoke(
            app,
            [
                "-s",
                "archive-journal",
                "-i",
                str(temp_project["input"] / "missing"),
                "-o",
                str(temp_project["output"]),
            ],
        )
        assert result.exit_code in (1, 2)

    def test_output_permission_error(self, temp_project, monkeypatch):
        runner = CliRunner()
        input_dir = temp_project["input"]
        (input_dir / "product1").mkdir()
        (input_dir / "product1" / "note.md").write_text("# Test")
        output_dir = temp_project["output"]
        old_cwd = os.getcwd()
        os.chmod(output_dir, 0o444)
        try:
            os.chdir(temp_project["root"])
            result = runner.invoke(
                app,
                [
                    "-s",
                    "archive-journal",
                    "-i",
                    str(input_dir),
                    "-o",
                    str(output_dir),
                ],
            )
            assert result.exit_code == 1
        finally:
            os.chmod(output_dir, 0o644)
            os.chdir(old_cwd)


class TestArchiveExitCodes:
    def test_exit_0_on_success(self, populated_project, monkeypatch):
        runner = CliRunner()
        monkeypatch.chdir(populated_project["root"])
        result = runner.invoke(
            app,
            [
                "-s",
                "archive-journal",
                "-i",
                str(populated_project["input"]),
                "-o",
                str(populated_project["output"]),
            ],
        )
        assert result.exit_code == 0

    def test_exit_1_on_failure(self, temp_project, monkeypatch):
        runner = CliRunner()
        old_cwd = os.getcwd()
        (temp_project["input"] / "product1").mkdir()
        (temp_project["input"] / "product1" / "note.md").write_text("# Test")
        os.chmod(temp_project["output"], 0o444)
        try:
            os.chdir(temp_project["root"])
            result = runner.invoke(
                app,
                [
                    "-s",
                    "archive-journal",
                    "-i",
                    str(temp_project["input"]),
                    "-o",
                    str(temp_project["output"]),
                ],
            )
            assert result.exit_code == 1
        finally:
            os.chmod(temp_project["output"], 0o644)
            os.chdir(old_cwd)
