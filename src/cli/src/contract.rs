use anyhow::{Context, Result};
use serde::Deserialize;
use std::collections::HashMap;
use std::path::{Path, PathBuf};

const CONTRACT_REL_PATH: &str = ".quanttide/asset/contract.yaml";

/// 技能配置
#[derive(Debug, Deserialize, Clone)]
pub struct SkillConfig {
    #[serde(default)]
    pub version: String,
    #[serde(default)]
    pub entrypoint: String,
    #[serde(default)]
    pub params: HashMap<String, serde_yaml::Value>,
}

/// 资产配置
#[derive(Debug, Deserialize, Clone)]
pub struct AssetConfig {
    #[serde(rename = "type")]
    pub r#type: String,
    #[serde(default)]
    pub provider: String,
    #[serde(default)]
    pub metadata: HashMap<String, serde_yaml::Value>,
}

/// 验证策略
#[derive(Debug, Deserialize, Clone)]
pub struct Policy {
    #[serde(default)]
    pub selector: String,
    #[serde(default)]
    pub mode: String,
    #[serde(default)]
    pub required_categories: Vec<String>,
}

/// 验证配置
#[derive(Debug, Deserialize, Default)]
pub struct ValidationConfig {
    #[serde(default)]
    pub policies: Vec<Policy>,
}

/// 契约顶层结构
#[derive(Debug, Deserialize)]
pub struct ContractSchema {
    #[serde(default)]
    pub assets: HashMap<String, AssetConfig>,
    #[serde(default)]
    pub skills: HashMap<String, SkillConfig>,
    #[serde(default)]
    pub validation: Option<ValidationConfig>,
}

/// 数字资产契约
pub struct Contract {
    pub root: PathBuf,
    pub schema: ContractSchema,
}

impl Contract {
    /// 从当前目录向上查找并加载契约
    pub fn load() -> Result<Self> {
        let root = Self::find_root(std::env::current_dir()?.as_path())?;
        Self::load_at(&root)
    }

    /// 从指定根目录加载契约
    pub fn load_at(root: &Path) -> Result<Self> {
        let path = root.join(CONTRACT_REL_PATH);
        let content = std::fs::read_to_string(&path)
            .with_context(|| format!("无法读取契约文件: {}", path.display()))?;
        let schema: ContractSchema = serde_yaml::from_str(&content)
            .with_context(|| format!("契约格式错误: {}", path.display()))?;
        Ok(Self {
            root: root.to_path_buf(),
            schema,
        })
    }

    /// 向上查找包含契约文件的目录
    fn find_root(start: &Path) -> Result<PathBuf> {
        let mut current = Some(start);
        while let Some(dir) = current {
            if dir.join(CONTRACT_REL_PATH).exists() {
                return Ok(dir.to_path_buf());
            }
            current = dir.parent();
        }
        anyhow::bail!("未找到契约文件 .quanttide/asset/contract.yaml");
    }

    /// 契约文件所在根目录
    pub fn path(&self) -> &Path {
        &self.root
    }

    /// 契约文件完整路径
    pub fn contract_path(&self) -> PathBuf {
        self.root.join(CONTRACT_REL_PATH)
    }

    /// 资产数
    pub fn asset_count(&self) -> usize {
        self.schema.assets.len()
    }

    /// 技能数
    pub fn skill_count(&self) -> usize {
        self.schema.skills.len()
    }

    /// 所有资产
    pub fn assets(&self) -> impl Iterator<Item = (&str, &AssetConfig)> {
        self.schema.assets.iter().map(|(k, v)| (k.as_str(), v))
    }

    /// 所有技能
    pub fn skills(&self) -> impl Iterator<Item = (&str, &SkillConfig)> {
        self.schema.skills.iter().map(|(k, v)| (k.as_str(), v))
    }

    /// 获取指定技能
    pub fn get_skill(&self, name: &str) -> Result<&SkillConfig> {
        self.schema
            .skills
            .get(name)
            .ok_or_else(|| anyhow::anyhow!("找不到技能 '{}'", name))
    }

    /// 获取验证策略
    pub fn policies(&self) -> Vec<Policy> {
        self.schema
            .validation
            .as_ref()
            .map(|v| v.policies.clone())
            .unwrap_or_default()
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::io::Write;

    #[test]
    fn test_contract_schema_deserialize() {
        let yaml = r#"
assets:
  brd:
    type: docs
    category: brd
skills:
  archive-journal:
    version: "1.0"
    params:
      pattern: "*.md"
"#;
        let schema: ContractSchema = serde_yaml::from_str(yaml).unwrap();
        assert_eq!(schema.assets.len(), 1);
        assert_eq!(schema.assets["brd"].r#type, "docs");
        assert_eq!(schema.skills.len(), 1);
        assert_eq!(schema.skills["archive-journal"].version, "1.0");
    }

    #[test]
    fn test_contract_schema_empty() {
        let yaml = "{}";
        let schema: ContractSchema = serde_yaml::from_str(yaml).unwrap();
        assert!(schema.assets.is_empty());
        assert!(schema.skills.is_empty());
        assert!(schema.validation.is_none());
    }

    #[test]
    fn test_policy_deserialize() {
        let yaml = r#"
validation:
  policies:
    - selector: "议事档案"
      mode: ATOMIC
      required_categories:
        - 议题
        - 议程
    - selector: "**"
      mode: SCOPED
"#;
        let schema: ContractSchema = serde_yaml::from_str(yaml).unwrap();
        let policies = schema.validation.unwrap().policies;
        assert_eq!(policies.len(), 2);
        assert_eq!(policies[0].selector, "议事档案");
        assert_eq!(policies[0].mode, "ATOMIC");
        assert_eq!(policies[0].required_categories.len(), 2);
        assert_eq!(policies[1].selector, "**");
        assert_eq!(policies[1].mode, "SCOPED");
    }

    #[test]
    fn test_load_at_not_found() {
        let tmp = std::env::temp_dir().join("test_contract_not_found");
        let _ = std::fs::remove_dir_all(&tmp);
        std::fs::create_dir_all(&tmp).unwrap();
        let result = Contract::load_at(&tmp);
        assert!(result.is_err());
        let _ = std::fs::remove_dir_all(&tmp);
    }

    #[test]
    fn test_load_at_success() {
        let tmp = std::env::temp_dir().join("test_contract_load");
        let _ = std::fs::remove_dir_all(&tmp);
        std::fs::create_dir_all(tmp.join(".quanttide/asset")).unwrap();
        let mut f = std::fs::File::create(tmp.join(".quanttide/asset/contract.yaml")).unwrap();
        writeln!(f, "assets:\n  test:\n    type: code").unwrap();
        drop(f);

        let contract = Contract::load_at(&tmp).unwrap();
        assert_eq!(contract.asset_count(), 1);
        let _ = std::fs::remove_dir_all(&tmp);
    }
}
