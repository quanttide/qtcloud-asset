from __future__ import annotations

from pathlib import Path

import pytest
from src.cli.app.cli import app
from typer.testing import CliRunner

runner = CliRunner()


class TestArchiveCommand:
    def test_help(self):
        result = runner.invoke(app, ["--help"])
        assert result.exit_code == 0

    def test_dry_run(self, journal_dir: Path):
        if not journal_dir.exists():
            pytest.skip("Fixtures journal not found")

        result = runner.invoke(
            app,
            [
                "run",
                "-s",
                "archive-journal",
                "-i",
                str(journal_dir),
                "-o",
                "/tmp/archive-out",
                "-n",
            ],
        )
        assert result.exit_code == 0

    def test_unknown_skill(self, tmp_path: Path):
        result = runner.invoke(
            app,
            [
                "run",
                "-s",
                "nonexistent",
                "-i",
                str(tmp_path),
                "-o",
                str(tmp_path / "out"),
            ],
        )
        assert result.exit_code != 0

    def test_missing_input_dir(self, tmp_path: Path):
        result = runner.invoke(
            app,
            [
                "run",
                "-s",
                "archive-journal",
                "-i",
                str(tmp_path / "missing"),
                "-o",
                str(tmp_path),
            ],
        )
        assert result.exit_code != 0
