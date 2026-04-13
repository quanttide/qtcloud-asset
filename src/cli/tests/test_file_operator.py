#!/usr/bin/env python3
"""操作层测试"""

from __future__ import annotations

import tempfile
from pathlib import Path

import pytest
from app.file_operator import (
    ArchiveResult,
    archive_product,
)


@pytest.fixture
def temp_dirs():
    with tempfile.TemporaryDirectory() as tmpdir:
        tmp = Path(tmpdir)
        src = tmp / "src"
        dst = tmp / "dst"
        src.mkdir()
        yield {"src": src, "dst": dst, "tmp": tmp}


@pytest.fixture
def src_with_files(temp_dirs):
    src = temp_dirs["src"]
    (src / "file1.md").write_text("content1")
    (src / "file2.md").write_text("content2")
    (src / "file3.txt").write_text("ignore me")
    return temp_dirs


class TestArchiveResult:
    def test_ok_true_when_no_errors(self):
        result = ArchiveResult(product="test")
        assert result.ok is True

    def test_ok_false_when_has_error(self):
        result = ArchiveResult(product="test", error="some error")
        assert result.ok is False

    def test_ok_false_when_has_failed_files(self):
        result = ArchiveResult(product="test", failed=["file.txt"])
        assert result.ok is False


class TestArchiveProduct:
    def test_returns_error_when_src_not_exists(self, tmp_path):
        nonexistent = tmp_path / "nonexistent"
        result = archive_product(nonexistent, tmp_path / "dst")
        assert result.error is not None
        assert "不存在" in result.error

    def test_dry_run_moves_nothing(self, src_with_files):
        result = archive_product(
            src_with_files["src"],
            src_with_files["dst"],
            pattern="*.md",
            dry_run=True,
        )
        assert result.total == 2
        assert len(result.moved) == 2
        assert result.error is None
        assert not (src_with_files["dst"] / "file1.md").exists()

    def test_archive_moves_files(self, src_with_files):
        result = archive_product(
            src_with_files["src"],
            src_with_files["dst"],
            pattern="*.md",
        )
        assert result.ok is True
        assert len(result.moved) == 2
        assert (src_with_files["dst"] / "file1.md").exists()
        assert not (src_with_files["src"] / "file1.md").exists()

    def test_skips_existing_files(self, src_with_files):
        dst = src_with_files["dst"]
        dst.mkdir()
        (dst / "file1.md").write_text("existing")

        result = archive_product(
            src_with_files["src"],
            dst,
            pattern="*.md",
        )
        assert "file1.md" in result.skipped
        assert "file2.md" in result.moved

    def test_no_matching_files(self, tmp_path):
        src = tmp_path / "src"
        src.mkdir()
        (src / "file.txt").write_text("not a md")

        result = archive_product(src, tmp_path / "dst", pattern="*.md")
        assert result.total == 0
        assert "无匹配" in result.skipped[0]

    def test_removes_empty_src_dir(self, tmp_path):
        src = tmp_path / "src"
        src.mkdir()
        (src / "file1.md").write_text("content1")
        (src / "file2.md").write_text("content2")
        dst = tmp_path / "dst"

        result = archive_product(src, dst, pattern="*.md")
        assert result.source_removed is True
        assert not src.exists()
