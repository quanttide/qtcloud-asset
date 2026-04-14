#!/usr/bin env python3
"""CLI 层测试"""

from __future__ import annotations

import os
import tempfile
from pathlib import Path

import pytest
from typer.testing import CliRunner

from app.cli import app


@pytest.fixture
def temp_project():
    """创建临时项目结构"""
    with tempfile.TemporaryDirectory() as tmpdir:
        tmp = Path(tmpdir)
        root = tmp / "project"
        root.mkdir()
        contract_dir = root / ".quanttide" / "asset"
        contract_dir.mkdir(parents=True)
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
    """创建带有文件的临时项目"""
    (temp_project["input"] / "product1").mkdir()
    (temp_project["input"] / "product1" / "note.md").write_text("# Test 1")
    (temp_project["input"] / "product2").mkdir()
    (temp_project["input"] / "product2" / "note.md").write_text("# Test 2")
    return temp_project


class TestCLI:
    @pytest.fixture
    def runner(self):
        return CliRunner()

    def test_run_command_help(self, runner):
        result = runner.invoke(app, ["run", "--help"])
        assert result.exit_code == 0
        assert "--input" in result.stdout
        assert "--skill" in result.stdout
        assert "--output" in result.stdout

    def test_missing_input_raises_error(self, runner, temp_project):
        result = runner.invoke(
            app,
            [
                "-s",
                "archive",
                "-i",
                str(temp_project["input"]),
                "-o",
                str(temp_project["output"]),
            ],
        )
        assert result.exit_code == 1

    def test_missing_contract_raises_error(self, runner, populated_project):
        result = runner.invoke(
            app,
            [
                "-s",
                "archive",
                "-i",
                str(populated_project["input"]),
                "-o",
                str(populated_project["output"]),
            ],
        )
        assert result.exit_code == 1
        assert (
            "错误" in result.stdout
            or "错误" in result.stderr
            or "契约" in result.stdout
            or "契约" in result.stderr
        )

    def test_successful_archive(self, runner, populated_project, monkeypatch):
        contract_file = populated_project["contract_dir"] / "contract.yaml"
        contract_file.write_text(
            "skills:\n  archive:\n    version: '1.0'\n    params:\n      pattern: '*.md'\n"
        )
        monkeypatch.chdir(populated_project["root"])

        result = runner.invoke(
            app,
            [
                "-s",
                "archive",
                "-i",
                str(populated_project["input"]),
                "-o",
                str(populated_project["output"]),
            ],
        )

        assert result.exit_code == 0
        assert "完成" in result.stdout or "OK" in result.stdout
        assert (populated_project["output"] / "product1" / "note.md").exists()

    def test_dry_run_mode(self, runner, populated_project, monkeypatch):
        contract_file = populated_project["contract_dir"] / "contract.yaml"
        contract_file.write_text(
            "skills:\n  archive:\n    version: '1.0'\n    params:\n      pattern: '*.md'\n"
        )
        monkeypatch.chdir(populated_project["root"])

        result = runner.invoke(
            app,
            [
                "-s",
                "archive",
                "-i",
                str(populated_project["input"]),
                "-o",
                str(populated_project["output"]),
                "-n",
            ],
        )

        assert result.exit_code == 0
        assert "预览" in result.stdout or "Preview" in result.stdout
        assert not (populated_project["output"] / "product1").exists()

    def test_verbose_mode(self, runner, populated_project, monkeypatch):
        contract_file = populated_project["contract_dir"] / "contract.yaml"
        contract_file.write_text(
            "skills:\n  archive:\n    version: '1.0'\n    params:\n      pattern: '*.md'\n"
        )
        monkeypatch.chdir(populated_project["root"])

        result = runner.invoke(
            app,
            [
                "-s",
                "archive",
                "-i",
                str(populated_project["input"]),
                "-o",
                str(populated_project["output"]),
                "-v",
            ],
        )

        assert result.exit_code == 0

    def test_pattern_override(self, runner, populated_project, monkeypatch):
        contract_file = populated_project["contract_dir"] / "contract.yaml"
        contract_file.write_text(
            "skills:\n  archive:\n    version: '1.0'\n    params:\n      pattern: '*.txt'\n"
        )
        monkeypatch.chdir(populated_project["root"])

        result = runner.invoke(
            app,
            [
                "-s",
                "archive",
                "-i",
                str(populated_project["input"]),
                "-o",
                str(populated_project["output"]),
                "-p",
                "*.md",
            ],
        )

        assert result.exit_code == 0


class TestCLIExitCodes:
    @pytest.fixture
    def runner(self):
        return CliRunner()

    @pytest.fixture
    def ready_project(self, temp_project, monkeypatch):
        contract_file = temp_project["contract_dir"] / "contract.yaml"
        contract_file.write_text(
            "skills:\n  archive:\n    version: '1.0'\n    params:\n      pattern: '*.md'\n"
        )
        (temp_project["input"] / "product1").mkdir()
        (temp_project["input"] / "product1" / "note.md").write_text("# Test")
        monkeypatch.chdir(temp_project["root"])
        return temp_project

    def test_exit_0_on_success(self, runner, ready_project):
        result = runner.invoke(
            app,
            [
                "-s",
                "archive",
                "-i",
                str(ready_project["input"]),
                "-o",
                str(ready_project["output"]),
            ],
        )
        assert result.exit_code == 0

    def test_exit_1_on_failure(self, runner, tmp_path):
        old_cwd = os.getcwd()
        root = tmp_path / "project"
        root.mkdir()
        contract_dir = root / ".quanttide" / "asset"
        contract_dir.mkdir(parents=True)
        (contract_dir / "contract.yaml").write_text(
            "skills:\n  archive:\n    version: '1.0'\n    params:\n      pattern: '*.md'\n"
        )
        input_dir = root / "input"
        input_dir.mkdir()
        (input_dir / "product1").mkdir()
        (input_dir / "product1" / "note.md").write_text("# Test")
        output_dir = root / "output"
        output_dir.mkdir()
        os.chmod(output_dir, 0o444)
        try:
            os.chdir(root)
            result = runner.invoke(
                app,
                [
                    "-s",
                    "archive",
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
