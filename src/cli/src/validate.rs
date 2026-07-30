use anyhow::Result;
use serde::Serialize;
use std::path::Path;

use crate::contract::{Contract, Policy};

/// 单个验证结果
#[derive(Debug, Serialize)]
pub struct ValidationResult {
    pub name: String,
    pub passed: bool,
    pub rules_passed: usize,
    pub rules_failed: usize,
    pub details: Vec<ValidationDetail>,
}

/// 验证详情
#[derive(Debug, Serialize)]
pub struct ValidationDetail {
    pub rule: String,
    pub status: String,
}

/// 验证报告
#[derive(Debug, Serialize)]
pub struct ValidationReport {
    pub total_assets: usize,
    pub passed_assets: usize,
    pub failed_assets: usize,
    pub results: Vec<ValidationResult>,
}

/// 本地资产信息
struct LocalAsset {
    name: String,
    categories: Vec<String>,
}

/// 从目录加载资产信息
fn load_assets_from_dir(base_dir: &Path) -> Result<Vec<LocalAsset>> {
    let mut assets = Vec::new();
    if !base_dir.exists() {
        return Ok(assets);
    }
    for entry in std::fs::read_dir(base_dir)? {
        let entry = entry?;
        let entry_path = entry.path();
        if !entry_path.is_dir() {
            continue;
        }
        let categories: Vec<String> = std::fs::read_dir(&entry_path)
            .ok()
            .into_iter()
            .flatten()
            .filter_map(|e| e.ok())
            .filter(|e| e.path().is_dir())
            .map(|e| e.file_name().to_string_lossy().to_string())
            .collect();

        assets.push(LocalAsset {
            name: entry.file_name().to_string_lossy().to_string(),
            categories,
        });
    }
    Ok(assets)
}

/// 匹配 selector
fn match_selector(name: &str, selector: &str) -> bool {
    if selector == "**" {
        return true;
    }
    if selector.ends_with("/**") {
        let prefix = selector.trim_end_matches("/**");
        return name.starts_with(prefix);
    }
    name == selector
}

/// 验证单个资产
fn validate_asset(name: &str, categories: &[String], policies: &[Policy]) -> ValidationResult {
    let mut result = ValidationResult {
        name: name.to_string(),
        passed: true,
        rules_passed: 0,
        rules_failed: 0,
        details: Vec::new(),
    };

    for policy in policies {
        if !match_selector(name, &policy.selector) {
            continue;
        }

        if policy.mode == "ATOMIC" {
            for cat in &policy.required_categories {
                if categories.iter().any(|c| c == cat) {
                    result.details.push(ValidationDetail {
                        rule: format!("必须包含分类 '{}'", cat),
                        status: "pass".to_string(),
                    });
                    result.rules_passed += 1;
                } else {
                    result.details.push(ValidationDetail {
                        rule: format!("必须包含分类 '{}'", cat),
                        status: "fail".to_string(),
                    });
                    result.rules_failed += 1;
                    result.passed = false;
                }
            }
        } else if policy.mode == "SCOPED" {
            result.details.push(ValidationDetail {
                rule: "资产存在".to_string(),
                status: "pass".to_string(),
            });
            result.rules_passed += 1;
        }

        // 首位命中
        break;
    }

    result
}

/// 验证目录中的资产是否符合契约要求
pub fn validate_directory(directory: &Path, contract: &Contract) -> Result<ValidationReport> {
    let policies = contract.policies();
    if policies.is_empty() {
        anyhow::bail!("契约文件中未定义验证策略");
    }

    let assets = load_assets_from_dir(directory)?;
    let mut report = ValidationReport {
        total_assets: assets.len(),
        passed_assets: 0,
        failed_assets: 0,
        results: Vec::new(),
    };

    for asset in assets {
        let result = validate_asset(&asset.name, &asset.categories, &policies);
        if result.passed {
            report.passed_assets += 1;
        } else {
            report.failed_assets += 1;
        }
        report.results.push(result);
    }

    Ok(report)
}

#[derive(clap::Args)]
pub struct ValidateArgs {
    /// 要验证的目录（默认当前目录）
    #[arg(short = 'i', long, default_value = ".")]
    pub input: String,

    /// 契约文件路径（默认自动查找）
    #[arg(short, long)]
    pub contract: Option<String>,

    /// 详细输出
    #[arg(short, long)]
    pub verbose: bool,

    /// JSON 格式输出
    #[arg(long)]
    pub json: bool,
}

pub fn execute(args: ValidateArgs) -> Result<()> {
    let path = Path::new(&args.input);

    let contract = if let Some(ref cp) = args.contract {
        Contract::load_at(Path::new(cp).parent().unwrap_or(Path::new(".")))
    } else {
        Contract::load()
    }?;

    let report = validate_directory(path, &contract)?;

    if args.json {
        println!("{}", serde_json::to_string_pretty(&report)?);
        return Ok(());
    }

    crate::render::print_header("契约验证");
    crate::render::print_info(&format!("验证目录: {}", path.display()));
    crate::render::print_validate_report(
        report.passed_assets,
        report.failed_assets,
        report.total_assets,
    );

    if args.verbose {
        for result in &report.results {
            let status = if result.passed { "✅" } else { "❌" };
            println!("  {} {}", status, result.name);
            for detail in &result.details {
                if detail.status == "fail" {
                    println!("      - {}", detail.rule);
                }
            }
        }
    }

    if report.failed_assets > 0 {
        std::process::exit(1);
    }

    Ok(())
}

#[cfg(test)]
mod tests {
    use super::*;

    fn make_policy(selector: &str, mode: &str, cats: Vec<&str>) -> Policy {
        Policy {
            selector: selector.to_string(),
            mode: mode.to_string(),
            required_categories: cats.into_iter().map(|s| s.to_string()).collect(),
        }
    }

    #[test]
    fn test_match_selector_all() {
        assert!(match_selector("anything", "**"));
    }

    #[test]
    fn test_match_selector_exact() {
        assert!(match_selector("议事档案", "议事档案"));
        assert!(!match_selector("其他档案", "议事档案"));
    }

    #[test]
    fn test_match_selector_prefix() {
        assert!(match_selector("议事档案/议题", "议事档案/**"));
        assert!(match_selector("议事档案/议题/规则", "议事档案/**"));
        assert!(!match_selector("其他档案/议题", "议事档案/**"));
    }

    #[test]
    fn test_validate_asset_atomic_pass() {
        let categories = vec!["议题".to_string(), "议程".to_string()];
        let policies = vec![make_policy("test", "ATOMIC", vec!["议题", "议程"])];
        let result = validate_asset("test", &categories, &policies);
        assert!(result.passed);
        assert_eq!(result.rules_passed, 2);
        assert_eq!(result.rules_failed, 0);
    }

    #[test]
    fn test_validate_asset_atomic_fail() {
        let categories = vec!["议题".to_string()];
        let policies = vec![make_policy("test", "ATOMIC", vec!["议题", "决策"])];
        let result = validate_asset("test", &categories, &policies);
        assert!(!result.passed);
        assert_eq!(result.rules_passed, 1);
        assert_eq!(result.rules_failed, 1);
    }

    #[test]
    fn test_validate_asset_scoped() {
        let categories: Vec<String> = vec![];
        let policies = vec![make_policy("test", "SCOPED", vec![])];
        let result = validate_asset("test", &categories, &policies);
        assert!(result.passed);
        assert_eq!(result.rules_passed, 1);
    }

    #[test]
    fn test_validate_asset_no_match() {
        let categories: Vec<String> = vec![];
        let policies = vec![make_policy("other", "SCOPED", vec![])];
        let result = validate_asset("test", &categories, &policies);
        // 无匹配策略时，passed 保持默认 true
        assert!(result.passed);
    }

    #[test]
    fn test_validate_directory_no_policies() {
        let contract = crate::contract::Contract {
            root: std::path::PathBuf::from("."),
            schema: crate::contract::ContractSchema {
                assets: std::collections::HashMap::new(),
                skills: std::collections::HashMap::new(),
                validation: None,
            },
        };
        let result = validate_directory(Path::new("."), &contract);
        assert!(result.is_err());
    }
}
