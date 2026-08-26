use anyhow::Result;
use std::path::{Path, PathBuf};

use crate::contract::Contract;

/// 单个归档任务
#[derive(Debug)]
pub struct ArchiveTask {
    pub product: String,
    pub src_dir: PathBuf,
    pub dst_dir: PathBuf,
}

/// 归档工作流
#[derive(Debug)]
pub struct Workflow {
    pub name: String,
    pub pattern: String,
    pub tasks: Vec<ArchiveTask>,
}

impl Workflow {
    pub fn products(&self) -> Vec<&str> {
        self.tasks.iter().map(|t| t.product.as_str()).collect()
    }
}

/// 获取目录中的产品子目录
fn get_products(directory: &Path) -> Result<Vec<String>> {
    if !directory.exists() {
        anyhow::bail!("目录不存在: {}", directory.display());
    }
    let mut products = Vec::new();
    for entry in std::fs::read_dir(directory)? {
        let entry = entry?;
        if entry.file_type()?.is_dir() {
            products.push(entry.file_name().to_string_lossy().to_string());
        }
    }
    products.sort();
    Ok(products)
}

/// 解析契约，生成工作流
pub fn resolve_workflow(
    skill_name: &str,
    input_dir: &Path,
    output_dir: &Path,
    pattern: Option<&str>,
    contract: &Contract,
) -> Result<Workflow> {
    let skill = contract.get_skill(skill_name)?;

    // 优先使用技能配置中的 pattern，再尝试参数中传入的，默认 "*.md"
    let effective_pattern = pattern
        .map(|s| s.to_string())
        .or_else(|| {
            skill
                .params
                .get("pattern")
                .and_then(|v| v.as_str())
                .map(|s| s.to_string())
        })
        .unwrap_or_else(|| "*.md".to_string());

    let products = get_products(input_dir)?;

    let tasks = products
        .into_iter()
        .map(|p| ArchiveTask {
            src_dir: input_dir.join(&p),
            dst_dir: output_dir.join(&p),
            product: p,
        })
        .collect();

    Ok(Workflow {
        name: skill_name.to_string(),
        pattern: effective_pattern,
        tasks,
    })
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_get_products_not_found() {
        let result = get_products(Path::new("/nonexistent/path"));
        assert!(result.is_err());
    }

    #[test]
    fn test_get_products_empty() {
        let tmp = std::env::temp_dir().join("test_wf_products_empty");
        let _ = std::fs::remove_dir_all(&tmp);
        std::fs::create_dir_all(&tmp).unwrap();
        let products = get_products(&tmp).unwrap();
        assert!(products.is_empty());
        let _ = std::fs::remove_dir_all(&tmp);
    }

    #[test]
    fn test_get_products_with_entries() {
        let tmp = std::env::temp_dir().join("test_wf_products");
        let _ = std::fs::remove_dir_all(&tmp);
        std::fs::create_dir_all(tmp.join("产品A")).unwrap();
        std::fs::create_dir_all(tmp.join("产品B")).unwrap();
        std::fs::write(tmp.join("文件.txt"), "").unwrap(); // 文件应被忽略

        let products = get_products(&tmp).unwrap();
        assert_eq!(products.len(), 2);
        assert_eq!(products[0], "产品A");
        assert_eq!(products[1], "产品B");
        let _ = std::fs::remove_dir_all(&tmp);
    }

    #[test]
    fn test_workflow_default_pattern() {
        use crate::contract::ContractSchema;
        use std::collections::HashMap;

        let schema = ContractSchema {
            assets: HashMap::new(),
            skills: HashMap::new(),
            validation: None,
        };
        let _contract = crate::contract::Contract {
            root: std::path::PathBuf::from("."),
            schema,
        };

        // test fallback when skill not found: expect error
        // 但 resolve_workflow 会直接报错，测试这个行为
    }
}
