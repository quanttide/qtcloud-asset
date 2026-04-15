# Integration Tests — CLI Module
# 使用真实的 assets/fixtures 进行 CLI 端到端测试

from __future__ import annotations

import os
import shutil
from pathlib import Path

import pytest
from typer.testing import CliRunner

from app.cli import app


class TestCLIIntegration:
    """CLI 模块集成测试"""

    def test_cli_help_command(self):
        """测试帮助命令"""
        runner = CliRunner()
        result = runner.invoke(app, ["--help"])
        
        assert result.exit_code == 0
        assert "--input" in result.stdout
        assert "--skill" in result.stdout
        assert "--output" in result.stdout

    def test_cli_run_help(self):
        """测试 run 子命令帮助"""
        runner = CliRunner()
        result = runner.invoke(app, ["run", "--help"])
        
        assert result.exit_code == 0

    def test_cli_no_args_shows_help(self):
        """测试无参数时显示帮助"""
        runner = CliRunner()
        result = runner.invoke(app, [])
        
        assert result.exit_code != 0


class TestCLIRealWorkflowIntegration:
    """CLI 真实工作流集成测试"""

    @pytest.fixture
    def real_fixtures_project(self, fixtures_root: Path, tmp_path: Path):
        """创建使用真实 fixtures 的测试项目"""
        project_root = tmp_path / "project"
        project_root.mkdir()
        
        contract_dir = project_root / ".quanttide" / "asset"
        contract_dir.mkdir(parents=True)
        
        fixtures_contract = fixtures_root / ".quanttide" / "asset" / "contract.yaml"
        if fixtures_contract.exists():
            shutil.copy(fixtures_contract, contract_dir / "contract.yaml")
        else:
            (contract_dir / "contract.yaml").write_text("""
skills:
  archive-journal:
    version: "1.0"
    params:
      pattern: "*.md"
""")
        
        input_dir = project_root / "input"
        input_dir.mkdir()
        
        (input_dir / "product1").mkdir()
        (input_dir / "product1" / "note1.md").write_text("# Note 1")
        (input_dir / "product1" / "note2.md").write_text("# Note 2")
        (input_dir / "product1" / "draft.txt").write_text("draft")
        
        (input_dir / "product2").mkdir()
        (input_dir / "product2" / "note.md").write_text("# Note")
        
        output_dir = project_root / "output"
        output_dir.mkdir()
        
        return {
            "root": project_root,
            "input": input_dir,
            "output": output_dir,
        }

    def test_cli_archive_real_workflow(self, real_fixtures_project):
        """测试真实的归档工作流"""
        runner = CliRunner()
        
        with runner.isolated_filesystem():
            os.chdir(real_fixtures_project["root"])
            
            result = runner.invoke(
                app,
                [
                    "-s", "archive-journal",
                    "-i", str(real_fixtures_project["input"]),
                    "-o", str(real_fixtures_project["output"]),
                ],
            )
            
            assert result.exit_code == 0
            assert (real_fixtures_project["output"] / "product1" / "note1.md").exists()
            assert (real_fixtures_project["output"] / "product1" / "note2.md").exists()
            assert (real_fixtures_project["output"] / "product2" / "note.md").exists()

    def test_cli_dry_run_no_actual_move(self, real_fixtures_project):
        """测试预览模式不实际移动文件"""
        runner = CliRunner()
        
        with runner.isolated_filesystem():
            os.chdir(real_fixtures_project["root"])
            
            result = runner.invoke(
                app,
                [
                    "-s", "archive-journal",
                    "-i", str(real_fixtures_project["input"]),
                    "-o", str(real_fixtures_project["output"]),
                    "-n",
                ],
            )
            
            assert result.exit_code == 0
            assert (real_fixtures_project["input"] / "product1" / "note1.md").exists()
            assert not (real_fixtures_project["output"] / "product1").exists()

    def test_cli_verbose_mode(self, real_fixtures_project):
        """测试详细输出模式"""
        runner = CliRunner()
        
        with runner.isolated_filesystem():
            os.chdir(real_fixtures_project["root"])
            
            result = runner.invoke(
                app,
                [
                    "-s", "archive-journal",
                    "-i", str(real_fixtures_project["input"]),
                    "-o", str(real_fixtures_project["output"]),
                    "-v",
                ],
            )
            
            assert result.exit_code == 0

    def test_cli_pattern_override(self, real_fixtures_project):
        """测试模式覆盖"""
        runner = CliRunner()
        
        with runner.isolated_filesystem():
            os.chdir(real_fixtures_project["root"])
            
            result = runner.invoke(
                app,
                [
                    "-s", "archive-journal",
                    "-i", str(real_fixtures_project["input"]),
                    "-o", str(real_fixtures_project["output"]),
                    "-p", "*.txt",
                ],
            )
            
            assert result.exit_code == 0
            assert (real_fixtures_project["output"] / "product1" / "draft.txt").exists()


class TestCLIErrorHandlingIntegration:
    """CLI 错误处理集成测试"""

    def test_cli_unknown_skill(self, tmp_path: Path):
        """测试未知技能"""
        runner = CliRunner()
        
        project_root = tmp_path / "project"
        project_root.mkdir()
        input_dir = project_root / "input"
        input_dir.mkdir()
        output_dir = project_root / "output"
        output_dir.mkdir()
        
        with runner.isolated_filesystem():
            os.chdir(project_root)
            
            result = runner.invoke(
                app,
                [
                    "-s", "unknown-skill",
                    "-i", str(input_dir),
                    "-o", str(output_dir),
                ],
            )
            
            assert result.exit_code == 1

    def test_cli_missing_input_dir(self, tmp_path: Path):
        """测试输入目录不存在"""
        runner = CliRunner()
        
        project_root = tmp_path / "project"
        project_root.mkdir()
        output_dir = project_root / "output"
        output_dir.mkdir()
        
        with runner.isolated_filesystem():
            os.chdir(project_root)
            
            result = runner.invoke(
                app,
                [
                    "-s", "archive-journal",
                    "-i", str(project_root / "nonexistent"),
                    "-o", str(output_dir),
                ],
            )
            
            assert result.exit_code in (1, 2)


class TestCLIExitCodesIntegration:
    """CLI 退出码集成测试"""

    def test_exit_code_0_on_success(self, fixtures_root: Path, tmp_path: Path):
        """成功时返回退出码 0"""
        runner = CliRunner()
        
        project_root = tmp_path / "project"
        project_root.mkdir()
        
        contract_dir = project_root / ".quanttide" / "asset"
        contract_dir.mkdir(parents=True)
        (contract_dir / "contract.yaml").write_text("""
skills:
  archive-journal:
    version: "1.0"
    params:
      pattern: "*.md"
""")
        
        input_dir = project_root / "input"
        input_dir.mkdir()
        (input_dir / "product1").mkdir()
        (input_dir / "product1" / "note.md").write_text("# Test")
        
        output_dir = project_root / "output"
        output_dir.mkdir()
        
        with runner.isolated_filesystem():
            os.chdir(project_root)
            
            result = runner.invoke(
                app,
                [
                    "-s", "archive-journal",
                    "-i", str(input_dir),
                    "-o", str(output_dir),
                ],
            )
            
            assert result.exit_code == 0
