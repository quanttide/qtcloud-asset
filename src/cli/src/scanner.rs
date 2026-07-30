use anyhow::Result;
use serde::Serialize;
use std::path::Path;

/// 单个资产的信息
#[derive(Debug, Serialize)]
pub struct AssetInfo {
    pub name: String,
    pub path: String,
    #[serde(rename = "type")]
    pub asset_type: String,
    pub size: u64,
    pub files_count: usize,
    pub categories: Vec<String>,
}

/// 扫描结果
#[derive(Debug, Serialize)]
pub struct ScanOutput {
    pub root_path: String,
    pub total_dirs: usize,
    pub total_files: usize,
    pub assets: Vec<AssetInfo>,
    pub synced_at: String,
}

/// 扫描目录，返回资产清单
pub fn scan_directory(path: &Path) -> Result<ScanOutput> {
    if !path.exists() {
        anyhow::bail!("目录不存在: {}", path.display());
    }
    if !path.is_dir() {
        anyhow::bail!("不是目录: {}", path.display());
    }

    let mut assets = Vec::new();
    let mut total_dirs = 0usize;
    let mut total_files = 0usize;

    for entry in std::fs::read_dir(path)? {
        let entry = entry?;
        let entry_path = entry.path();
        if !entry_path.is_dir() {
            continue;
        }

        total_dirs += 1;
        let (files_count, size) = dir_stats(&entry_path)?;
        total_files += files_count;

        let categories: Vec<String> = std::fs::read_dir(&entry_path)
            .ok()
            .into_iter()
            .flatten()
            .filter_map(|e| e.ok())
            .filter(|e| e.path().is_dir())
            .map(|e| e.file_name().to_string_lossy().to_string())
            .collect();

        assets.push(AssetInfo {
            name: entry.file_name().to_string_lossy().to_string(),
            path: entry_path.to_string_lossy().to_string(),
            asset_type: guess_asset_type(&entry_path),
            size,
            files_count,
            categories,
        });
    }

    Ok(ScanOutput {
        root_path: path.to_string_lossy().to_string(),
        total_dirs,
        total_files,
        assets,
        synced_at: String::new(),
    })
}

/// 获取目录的文件数和总大小
fn dir_stats(path: &Path) -> Result<(usize, u64)> {
    let mut files_count = 0usize;
    let mut total_size = 0u64;

    if path.is_dir() {
        for entry in walkdir::WalkDir::new(path).into_iter().filter_map(|e| e.ok()) {
            if entry.file_type().is_file() {
                files_count += 1;
                if let Ok(meta) = entry.metadata() {
                    total_size += meta.len();
                }
            }
        }
    }

    Ok((files_count, total_size))
}

/// 根据目录名推测资产类型
fn guess_asset_type(path: &Path) -> String {
    let name_lower = path.file_name()
        .map(|n| n.to_string_lossy().to_lowercase())
        .unwrap_or_default();

    if name_lower.contains("议事档案") || name_lower.contains("archive") {
        return "议事档案".to_string();
    }
    if name_lower.contains("journal") {
        return "日志".to_string();
    }
    if ["prd", "brd", "ixd", "add", "qa"].iter().any(|kw| name_lower.contains(kw)) {
        return "文档".to_string();
    }
    if name_lower.contains("docs") || name_lower.contains("doc") {
        return "文档".to_string();
    }
    "通用".to_string()
}

#[derive(clap::Args)]
pub struct ScanArgs {
    /// 要扫描的目录（默认当前目录）
    #[arg(short = 'i', long, default_value = ".")]
    pub input: String,

    /// 详细输出
    #[arg(short, long)]
    pub verbose: bool,

    /// JSON 格式输出
    #[arg(long)]
    pub json: bool,
}

pub fn execute(args: ScanArgs) -> Result<()> {
    let path = std::path::Path::new(&args.input);

    let output = scan_directory(path)?;

    if args.json {
        println!("{}", serde_json::to_string_pretty(&output)?);
        return Ok(());
    }

    crate::render::print_header("资产扫描");
    crate::render::print_info(&format!("扫描目录: {}", path.display()));
    crate::render::print_success(&format!(
        "扫描完成: {} 个资产目录，{} 个文件",
        output.total_dirs, output.total_files
    ));

    if args.verbose {
        crate::render::print_asset_table(&output.assets);
    }

    Ok(())
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_guess_asset_type_archive() {
        let p = Path::new("/some/议事档案");
        assert_eq!(guess_asset_type(p), "议事档案");
    }

    #[test]
    fn test_guess_asset_type_archive_english() {
        let p = Path::new("/some/archive-data");
        assert_eq!(guess_asset_type(p), "议事档案");
    }

    #[test]
    fn test_guess_asset_type_journal() {
        let p = Path::new("/some/journal");
        assert_eq!(guess_asset_type(p), "日志");
    }

    #[test]
    fn test_guess_asset_type_docs() {
        let p = Path::new("/some/prd");
        assert_eq!(guess_asset_type(p), "文档");
    }

    #[test]
    fn test_guess_asset_type_other() {
        let p = Path::new("/some/random");
        assert_eq!(guess_asset_type(p), "通用");
    }

    #[test]
    fn test_scan_directory_not_found() {
        let result = scan_directory(Path::new("/nonexistent/path"));
        assert!(result.is_err());
    }

    #[test]
    fn test_scan_directory_empty() {
        let tmp = std::env::temp_dir().join("test_scan_empty");
        let _ = std::fs::remove_dir_all(&tmp);
        std::fs::create_dir_all(&tmp).unwrap();
        let output = scan_directory(&tmp).unwrap();
        assert_eq!(output.total_dirs, 0);
        assert_eq!(output.total_files, 0);
        let _ = std::fs::remove_dir_all(&tmp);
    }

    #[test]
    fn test_scan_directory_with_entries() {
        let tmp = std::env::temp_dir().join("test_scan_entries");
        let _ = std::fs::remove_dir_all(&tmp);
        std::fs::create_dir_all(&tmp.join("sub1")).unwrap();
        std::fs::create_dir_all(&tmp.join("sub2")).unwrap();
        std::fs::write(tmp.join("sub1").join("file1.md"), "test").unwrap();
        std::fs::write(tmp.join("sub1").join("file2.md"), "test").unwrap();

        let output = scan_directory(&tmp).unwrap();
        assert_eq!(output.total_dirs, 2);
        assert_eq!(output.total_files, 2);
        assert_eq!(output.assets.len(), 2);

        let _ = std::fs::remove_dir_all(&tmp);
    }
}
