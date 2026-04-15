# Integration Tests — Workflow Module
# 使用真实的 assets/fixtures 进行端到端工作流测试

from __future__ import annotations

from pathlib import Path

import pytest

from app.workflow import (
    ArchiveTask,
    Workflow,
    _get_products,
    print_workflow_summary,
    resolve_workflow,
)


class TestWorkflowIntegration:
    """工作流模块集成测试"""

    def test_resolve_workflow_with_real_fixtures(self, fixtures_root: Path):
        """使用真实 fixtures 解析工作流"""
        journal_dir = fixtures_root / "docs" / "journal"
        archive_dir = fixtures_root / "docs" / "archive"
        
        workflow = resolve_workflow(
            skill_name="archive-journal",
            input_dir=journal_dir,
            output_dir=archive_dir,
            contract_root=fixtures_root,
        )
        
        assert isinstance(workflow, Workflow)
        assert workflow.name == "archive-journal"
        assert workflow.pattern == "*.md"

    def test_workflow_tasks_from_real_structure(self, fixtures_root: Path):
        """验证工作流任务与真实目录结构匹配"""
        journal_dir = fixtures_root / "docs" / "journal"
        archive_dir = fixtures_root / "docs" / "archive"
        
        workflow = resolve_workflow(
            skill_name="archive-journal",
            input_dir=journal_dir,
            output_dir=archive_dir,
            contract_root=fixtures_root,
        )
        
        products = workflow.products
        assert len(products) == 0
        assert isinstance(workflow.tasks, list)

    def test_workflow_with_pattern_override(self, fixtures_root: Path):
        """测试模式覆盖功能"""
        journal_dir = fixtures_root / "docs" / "journal"
        archive_dir = fixtures_root / "docs" / "archive"
        
        workflow = resolve_workflow(
            skill_name="archive-journal",
            input_dir=journal_dir,
            output_dir=archive_dir,
            pattern="*.txt",
            contract_root=fixtures_root,
        )
        
        assert workflow.pattern == "*.txt"

    def test_workflow_products_property(self, fixtures_root: Path):
        """测试 products 属性"""
        journal_dir = fixtures_root / "docs" / "journal"
        archive_dir = fixtures_root / "docs" / "archive"
        
        workflow = resolve_workflow(
            skill_name="archive-journal",
            input_dir=journal_dir,
            output_dir=archive_dir,
            contract_root=fixtures_root,
        )
        
        assert isinstance(workflow.products, list)


class TestGetProductsIntegration:
    """_get_products 函数集成测试"""

    def test_get_products_from_real_journal_dir(self, journal_dir: Path):
        """从真实的 journal 目录获取产品列表"""
        products = _get_products(journal_dir)
        
        assert isinstance(products, list)

    def test_get_products_from_empty_dir(self, tmp_path: Path):
        """处理空目录"""
        empty_dir = tmp_path / "empty"
        empty_dir.mkdir()
        
        products = _get_products(empty_dir)
        
        assert products == []

    def test_get_products_raises_on_nonexistent(self, tmp_path: Path):
        """目录不存在时抛出异常"""
        with pytest.raises(FileNotFoundError) as exc_info:
            _get_products(tmp_path / "nonexistent")
        assert "目录不存在" in str(exc_info.value)


class TestWorkflowClasses:
    """Workflow 数据类集成测试"""

    def test_archive_task_creation(self):
        """测试 ArchiveTask 创建"""
        task = ArchiveTask(
            product="test-product",
            src_dir=Path("/src/test-product"),
            dst_dir=Path("/dst/test-product"),
        )
        
        assert task.product == "test-product"
        assert isinstance(task.src_dir, Path)
        assert isinstance(task.dst_dir, Path)

    def test_workflow_with_tasks(self):
        """测试 Workflow 与 ArchiveTask 配合"""
        tasks = [
            ArchiveTask(
                product="product1",
                src_dir=Path("/src/product1"),
                dst_dir=Path("/dst/product1"),
            ),
            ArchiveTask(
                product="product2",
                src_dir=Path("/src/product2"),
                dst_dir=Path("/dst/product2"),
            ),
        ]
        
        workflow = Workflow(
            name="test-workflow",
            pattern="*.md",
            tasks=tasks,
        )
        
        assert len(workflow.tasks) == 2
        assert workflow.products == ["product1", "product2"]


class TestPrintWorkflowSummary:
    """print_workflow_summary 集成测试"""

    def test_print_summary_execution_mode(self, capsys):
        """测试执行模式输出"""
        workflow = Workflow(
            name="test-workflow",
            pattern="*.md",
            tasks=[
                ArchiveTask(
                    product="product1",
                    src_dir=Path("/src/product1"),
                    dst_dir=Path("/dst/product1"),
                ),
            ],
        )
        
        print_workflow_summary(workflow, dry_run=False)
        captured = capsys.readouterr()
        
        assert "执行" in captured.out
        assert "test-workflow" in captured.out
        assert "product1" in captured.out

    def test_print_summary_dry_run_mode(self, capsys):
        """测试预览模式输出"""
        workflow = Workflow(
            name="test-workflow",
            pattern="*.md",
            tasks=[
                ArchiveTask(
                    product="product1",
                    src_dir=Path("/src/product1"),
                    dst_dir=Path("/dst/product1"),
                ),
            ],
        )
        
        print_workflow_summary(workflow, dry_run=True)
        captured = capsys.readouterr()
        
        assert "预览" in captured.out
