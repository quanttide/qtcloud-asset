# Integration Tests — Contract Module
# 使用真实的 assets/fixtures 契约文件进行端到端测试

from __future__ import annotations

from pathlib import Path

import pytest

from app.contract import AssetConfig, Contract, ContractSchema, SkillConfig


class TestContractIntegration:
    """契约模块集成测试"""

    def test_load_contract_from_fixtures(self, fixtures_root: Path):
        """从 fixtures 目录加载契约配置"""
        contract = Contract(root=fixtures_root)
        
        assert contract.root == fixtures_root
        assert isinstance(contract.config, ContractSchema)

    def test_contract_has_assets(self, fixtures_root: Path):
        """验证契约包含资产定义"""
        contract = Contract(root=fixtures_root)
        
        assets = contract.config.assets
        assert "journal" in assets
        assert "archive" in assets

    def test_contract_has_skills(self, fixtures_root: Path):
        """验证契约包含技能定义"""
        contract = Contract(root=fixtures_root)
        
        skills = contract.config.skills
        assert "archive-journal" in skills

    def test_get_asset_returns_asset_config(self, fixtures_root: Path):
        """get_asset 返回正确的 AssetConfig"""
        contract = Contract(root=fixtures_root)
        
        journal_asset = contract.get_asset("journal")
        assert isinstance(journal_asset, AssetConfig)
        assert journal_asset.type == "docs"

    def test_get_skill_returns_skill_config(self, fixtures_root: Path):
        """get_skill 返回正确的 SkillConfig"""
        contract = Contract(root=fixtures_root)
        
        skill = contract.get_skill("archive-journal")
        assert isinstance(skill, SkillConfig)
        assert skill.version == "1.0"

    def test_get_asset_raises_on_unknown(self, fixtures_root: Path):
        """获取不存在的资产时抛出 KeyError"""
        contract = Contract(root=fixtures_root)
        
        with pytest.raises(KeyError) as exc_info:
            contract.get_asset("nonexistent")
        assert "nonexistent" in str(exc_info.value)

    def test_get_skill_raises_on_unknown(self, fixtures_root: Path):
        """获取不存在的技能时抛出 KeyError"""
        contract = Contract(root=fixtures_root)
        
        with pytest.raises(KeyError) as exc_info:
            contract.get_skill("nonexistent")
        assert "nonexistent" in str(exc_info.value)

    def test_asset_config_is_frozen(self, fixtures_root: Path):
        """验证 AssetConfig 是不可变对象"""
        contract = Contract(root=fixtures_root)
        asset = contract.get_asset("journal")
        
        with pytest.raises(Exception):
            asset.category = "changed"

    def test_skill_config_is_frozen(self, fixtures_root: Path):
        """验证 SkillConfig 是不可变对象"""
        contract = Contract(root=fixtures_root)
        skill = contract.get_skill("archive-journal")
        
        with pytest.raises(Exception):
            skill.description = "changed"


class TestFindRootIntegration:
    """find_root 方法集成测试"""

    def test_find_root_from_fixtures_dir(self, fixtures_root: Path):
        """从 fixtures 目录向上查找契约根目录"""
        root = Contract.find_root(start=fixtures_root)
        
        assert root == fixtures_root
        assert (root / ".quanttide" / "asset" / "contract.yaml").exists()

    def test_find_root_from_subdir(self, fixtures_root: Path):
        """从子目录向上查找契约根目录"""
        subdir = fixtures_root / "docs" / "journal"
        root = Contract.find_root(start=subdir)
        
        assert root == fixtures_root
        assert (root / ".quanttide" / "asset" / "contract.yaml").exists()

    def test_find_root_raises_when_not_found(self, tmp_path: Path):
        """找不到契约文件时抛出 FileNotFoundError"""
        with pytest.raises(FileNotFoundError) as exc_info:
            Contract.find_root(start=tmp_path)
        assert "未找到契约文件" in str(exc_info.value)
