from __future__ import annotations

from pathlib import Path

import pytest
from typer.testing import CliRunner

from src.cli.app.cli import app

runner = CliRunner()


@pytest.fixture(autouse=True)
def setup_archive_dirs(tmp_path: Path):
    archive = tmp_path / "archive"
    archive.mkdir()
    yield {"archive": archive, "tmp": tmp_path}


class TestArchiveCommand:
    def test_archive_help(self):
        result = runner.invoke(app, ["archive", "--help"])
        assert result.exit_code == 0
        assert "归档" in result.output

    def test_archive_dry_run_with_real_files(self):
        journal_dir = Path("examples/archive/sample/journal")
        if not journal_dir.exists():
            pytest.skip("Sample journal not found")

        result = runner.invoke(
            app,
            ["test_archive", "product", "-p", "qtcloud-asset", "-n"],
        )
        assert result.exit_code == 0

    def test_archive_nonexistent_contract(self):
        result = runner.invoke(app, ["archive", "nonexistent", "product"])
        assert result.exit_code != 0
        assert "找不到" in result.output or "Error" in result.output

    def test_archive_with_missing_journal(self):
        result = runner.invoke(app, ["archive", "test_archive", "nonexistent"])
        assert result.exit_code != 0
