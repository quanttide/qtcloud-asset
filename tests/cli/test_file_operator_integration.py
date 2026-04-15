# Integration Tests — File Operator Module
# 使用真实的 assets/fixtures 进行文件系统端到端测试

from __future__ import annotations

import shutil
import tempfile
from pathlib import Path

import pytest

from app.file_operator import (
    ArchiveResult,
    FileResult,
    archive_product,
)


class TestArchiveProductIntegration:
    """archive_product 函数集成测试"""

    def test_archive_product_with_real_files(self, tmp_path: Path):
        """使用真实文件测试归档功能"""
        src_dir = tmp_path / "src" / "product1"
        src_dir.mkdir(parents=True)
        (src_dir / "note.md").write_text("# Test Note")
        (src_dir / "draft.txt").write_text("draft content")
        
        dst_dir = tmp_path / "dst" / "product1"
        
        result = archive_product(
            src_dir=src_dir,
            dst_dir=dst_dir,
            pattern="*.md",
        )
        
        assert result.ok
        assert len(result.moved) == 1
        assert "note.md" in result.moved
        assert (dst_dir / "note.md").exists()
        assert not (src_dir / "note.md").exists()

    def test_archive_product_preserves_content(self, tmp_path: Path):
        """验证文件内容在归档后保持不变"""
        content = "# Test Content\n\nThis is a test."
        src_dir = tmp_path / "src" / "product1"
        src_dir.mkdir(parents=True)
        (src_dir / "note.md").write_text(content)
        
        dst_dir = tmp_dir = tmp_path / "dst" / "product1"
        
        result = archive_product(
            src_dir=src_dir,
            dst_dir=dst_dir,
            pattern="*.md",
        )
        
        assert result.ok
        assert (dst_dir / "note.md").read_text() == content

    def test_archive_product_dry_run(self, tmp_path: Path):
        """测试预览模式不实际移动文件"""
        src_dir = tmp_path / "src" / "product1"
        src_dir.mkdir(parents=True)
        (src_dir / "note.md").write_text("# Test")
        
        dst_dir = tmp_path / "dst" / "product1"
        
        result = archive_product(
            src_dir=src_dir,
            dst_dir=dst_dir,
            pattern="*.md",
            dry_run=True,
        )
        
        assert result.ok
        assert len(result.moved) == 1
        assert (src_dir / "note.md").exists()
        assert not dst_dir.exists()

    def test_archive_product_removes_empty_source(self, tmp_path: Path):
        """测试归档后删除空源目录"""
        src_dir = tmp_path / "src" / "product1"
        src_dir.mkdir(parents=True)
        (src_dir / "note.md").write_text("# Test")
        
        dst_dir = tmp_path / "dst" / "product1"
        
        result = archive_product(
            src_dir=src_dir,
            dst_dir=dst_dir,
            pattern="*.md",
        )
        
        assert result.ok
        assert result.source_removed
        assert not src_dir.exists()

    def test_archive_product_keeps_nonempty_source(self, tmp_path: Path):
        """测试保留非空源目录"""
        src_dir = tmp_path / "src" / "product1"
        src_dir.mkdir(parents=True)
        (src_dir / "note.md").write_text("# Test")
        (src_dir / "keep.txt").write_text("keep this")
        
        dst_dir = tmp_path / "dst" / "product1"
        
        result = archive_product(
            src_dir=src_dir,
            dst_dir=dst_dir,
            pattern="*.md",
        )
        
        assert result.ok
        assert not result.source_removed
        assert src_dir.exists()
        assert (src_dir / "keep.txt").exists()

    def test_archive_product_skips_existing_files(self, tmp_path: Path):
        """测试跳过已存在的目标文件"""
        src_dir = tmp_path / "src" / "product1"
        src_dir.mkdir(parents=True)
        (src_dir / "note.md").write_text("# Source")
        
        dst_dir = tmp_path / "dst" / "product1"
        dst_dir.mkdir(parents=True)
        (dst_dir / "note.md").write_text("# Existing")
        
        result = archive_product(
            src_dir=src_dir,
            dst_dir=dst_dir,
            pattern="*.md",
        )
        
        assert len(result.skipped) == 1
        assert "note.md" in result.skipped
        assert (dst_dir / "note.md").read_text() == "# Existing"

    def test_archive_product_no_matching_files(self, tmp_path: Path):
        """测试无匹配文件时正常返回"""
        src_dir = tmp_path / "src" / "product1"
        src_dir.mkdir(parents=True)
        (src_dir / "note.txt").write_text("# Text file")
        
        dst_dir = tmp_path / "dst" / "product1"
        
        result = archive_product(
            src_dir=src_dir,
            dst_dir=dst_dir,
            pattern="*.md",
        )
        
        assert result.total == 0
        assert not result.ok is False

    def test_archive_product_source_not_exists(self, tmp_path: Path):
        """测试源目录不存在时返回错误"""
        src_dir = tmp_path / "nonexistent"
        dst_dir = tmp_path / "dst"
        
        result = archive_product(
            src_dir=src_dir,
            dst_dir=dst_dir,
            pattern="*.md",
        )
        
        assert not result.ok
        assert result.error is not None
        assert "不存在" in result.error


class TestArchiveResultIntegration:
    """ArchiveResult 数据类集成测试"""

    def test_archive_result_ok_property(self):
        """测试 ok 属性"""
        result_ok = ArchiveResult(product="test", total=1, moved=["file.md"])
        assert result_ok.ok
        
        result_error = ArchiveResult(product="test", error="some error")
        assert not result_error.ok
        
        result_failed = ArchiveResult(product="test", failed=["file.md"])
        assert not result_failed.ok

    def test_archive_result_with_multiple_files(self, tmp_path: Path):
        """测试多文件归档"""
        src_dir = tmp_path / "src" / "product1"
        src_dir.mkdir(parents=True)
        for i in range(3):
            (src_dir / f"note{i}.md").write_text(f"# Note {i}")
        
        dst_dir = tmp_path / "dst" / "product1"
        
        result = archive_product(
            src_dir=src_dir,
            dst_dir=dst_dir,
            pattern="*.md",
        )
        
        assert result.ok
        assert result.total == 3
        assert len(result.moved) == 3
        assert len(result.skipped) == 0
        assert len(result.failed) == 0


class TestRealFixturesWorkflow:
    """使用真实 fixtures 的完整归档工作流测试"""

    def test_journal_to_archive_workflow(self, fixtures_root: Path, tmp_path: Path):
        """模拟 journal 到 archive 的归档流程"""
        journal_dir = fixtures_root / "docs" / "journal"
        work_dir = tmp_path / "work"
        
        test_journal = work_dir / "journal"
        test_journal.mkdir(parents=True)
        test_archive = work_dir / "archive"
        
        (test_journal / "note1.md").write_text("# Note 1")
        (test_journal / "note2.md").write_text("# Note 2")
        
        result = archive_product(
            src_dir=test_journal,
            dst_dir=test_archive / "journal",
            pattern="*.md",
        )
        
        assert result.ok
        assert len(result.moved) == 2
        assert (test_archive / "journal" / "note1.md").exists()
        assert (test_archive / "journal" / "note2.md").exists()
        assert not test_journal.exists()

    def test_archive_dir_structure_preserved(self, fixtures_root: Path, tmp_path: Path):
        """测试归档目录结构保持不变"""
        work_dir = tmp_path / "work"
        
        src_dir = work_dir / "docs" / "journal" / "product1"
        src_dir.mkdir(parents=True)
        (src_dir / "note.md").write_text("# Product Note")
        
        dst_dir = work_dir / "docs" / "archive" / "journal"
        
        result = archive_product(
            src_dir=src_dir,
            dst_dir=dst_dir,
            pattern="*.md",
        )
        
        assert result.ok
        expected_path = work_dir / "docs" / "archive" / "journal" / "note.md"
        assert expected_path.exists()
