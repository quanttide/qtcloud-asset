use anyhow::Result;
use std::path::Path;

/// 单个产品的归档结果
#[derive(Debug)]
pub struct ArchiveResult {
    pub product: String,
    pub total: usize,
    pub moved: Vec<String>,
    pub skipped: Vec<String>,
    pub failed: Vec<String>,
    pub source_removed: bool,
    pub error: Option<String>,
}

impl ArchiveResult {
    pub fn ok(&self) -> bool {
        self.error.is_none() && self.failed.is_empty()
    }
}

/// 归档单个产品目录
///
/// 将源目录中匹配 pattern 的文件移动到目标目录。
/// 失败时自动回滚已移动的文件。
pub fn archive_product(
    src_dir: &Path,
    dst_dir: &Path,
    pattern: &str,
    dry_run: bool,
) -> ArchiveResult {
    let product = src_dir
        .file_name()
        .map(|n| n.to_string_lossy().to_string())
        .unwrap_or_default();
    let mut result = ArchiveResult {
        product,
        total: 0,
        moved: Vec::new(),
        skipped: Vec::new(),
        failed: Vec::new(),
        source_removed: false,
        error: None,
    };

    // 检查源目录
    if !src_dir.exists() {
        result.error = Some(format!("源目录不存在: {}", src_dir.display()));
        return result;
    }

    // 收集文件
    let pattern_glob =
        glob::Pattern::new(pattern).unwrap_or_else(|_| glob::Pattern::new("*").unwrap());

    let entries: Vec<_> = match std::fs::read_dir(src_dir) {
        Ok(rd) => rd
            .filter_map(|e| e.ok())
            .filter(|e| e.file_type().map(|t| t.is_file()).unwrap_or(false))
            .map(|e| e.path())
            .filter(|p| {
                p.file_name()
                    .map(|n| pattern_glob.matches(&n.to_string_lossy()))
                    .unwrap_or(false)
            })
            .collect(),
        Err(e) => {
            result.error = Some(format!("读取源目录失败: {}", e));
            return result;
        }
    };

    result.total = entries.len();
    if entries.is_empty() {
        result.skipped.push(format!("(无匹配 {} 文件)", pattern));
        return result;
    }

    // 预览模式
    if dry_run {
        result.moved = entries
            .iter()
            .map(|p| {
                p.file_name()
                    .map(|n| n.to_string_lossy().to_string())
                    .unwrap_or_default()
            })
            .collect();
        return result;
    }

    // 创建目标目录
    if let Err(e) = std::fs::create_dir_all(dst_dir) {
        result.error = Some(format!("无法创建目标目录: {}", e));
        return result;
    }

    // 逐文件移动
    for f in &entries {
        let fname = f
            .file_name()
            .map(|n| n.to_string_lossy().to_string())
            .unwrap_or_default();
        let dst = dst_dir.join(&fname);

        if dst.exists() {
            result.skipped.push(fname);
            continue;
        }

        match move_file(f, &dst) {
            Ok(()) => result.moved.push(fname),
            Err(e) => {
                result.failed.push(fname.clone());
                result.error = Some(format!("移动失败: {} — {}", fname, e));
            }
        }
    }

    // 失败时回滚
    if !result.failed.is_empty() {
        rollback(src_dir, dst_dir, &result.moved);
        result.moved.clear();
        return result;
    }

    // 清理空源目录
    if std::fs::read_dir(src_dir)
        .map(|mut rd| rd.next().is_none())
        .unwrap_or(false)
    {
        let _ = std::fs::remove_dir(src_dir);
        result.source_removed = true;
    }

    result
}

/// 移动文件（复制后删除源）
fn move_file(src: &Path, dst: &Path) -> Result<()> {
    std::fs::copy(src, dst)?;
    std::fs::remove_file(src)?;
    Ok(())
}

/// 尽力回滚：将已移动的文件退回源目录
fn rollback(src_dir: &Path, dst_dir: &Path, moved: &[String]) {
    // 确保源目录存在
    let _ = std::fs::create_dir_all(src_dir);
    for name in moved {
        let dst = dst_dir.join(name);
        let src = src_dir.join(name);
        if dst.exists() && !src.exists() {
            let _ = std::fs::copy(&dst, &src);
            let _ = std::fs::remove_file(&dst);
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::io::Write;

    #[test]
    fn test_archive_product_src_not_found() {
        let result = archive_product(
            Path::new("/nonexistent/src"),
            Path::new("/nonexistent/dst"),
            "*.md",
            false,
        );
        assert!(!result.ok());
        assert!(result.error.unwrap().contains("源目录不存在"));
    }

    #[test]
    fn test_archive_product_dry_run() {
        let tmp = std::env::temp_dir().join("test_archive_dry");
        let _ = std::fs::remove_dir_all(&tmp);
        std::fs::create_dir_all(tmp.join("src")).unwrap();
        let mut f = std::fs::File::create(tmp.join("src").join("test.md")).unwrap();
        writeln!(f, "content").unwrap();

        let result = archive_product(
            &tmp.join("src"),
            &tmp.join("dst"),
            "*.md",
            true, // dry_run
        );
        assert!(result.ok());
        assert_eq!(result.moved.len(), 1);
        assert_eq!(result.moved[0], "test.md");
        // dry-run 不应该实际移动文件
        assert!(!tmp.join("dst").join("test.md").exists());

        let _ = std::fs::remove_dir_all(&tmp);
    }

    #[test]
    fn test_archive_product_move() {
        let tmp = std::env::temp_dir().join("test_archive_move");
        let _ = std::fs::remove_dir_all(&tmp);
        std::fs::create_dir_all(tmp.join("src")).unwrap();
        std::fs::create_dir_all(tmp.join("dst")).unwrap();
        let mut f = std::fs::File::create(tmp.join("src").join("test.md")).unwrap();
        writeln!(f, "content").unwrap();

        let result = archive_product(&tmp.join("src"), &tmp.join("dst"), "*.md", false);
        assert!(result.ok());
        assert_eq!(result.moved.len(), 1);
        // 目标文件应存在
        assert!(tmp.join("dst").join("test.md").exists());
        // 源文件应不存在
        assert!(!tmp.join("src").join("test.md").exists());

        let _ = std::fs::remove_dir_all(&tmp);
    }

    #[test]
    fn test_archive_product_no_match() {
        let tmp = std::env::temp_dir().join("test_archive_nomatch");
        let _ = std::fs::remove_dir_all(&tmp);
        std::fs::create_dir_all(tmp.join("src")).unwrap();
        std::fs::write(tmp.join("src").join("test.txt"), "content").unwrap();

        let result = archive_product(&tmp.join("src"), &tmp.join("dst"), "*.md", false);
        assert!(result.ok());
        assert_eq!(result.moved.len(), 0);
        assert!(result.skipped[0].contains("无匹配"));

        let _ = std::fs::remove_dir_all(&tmp);
    }

    #[test]
    fn test_move_file() {
        let tmp = std::env::temp_dir().join("test_move_file");
        let _ = std::fs::remove_dir_all(&tmp);
        std::fs::create_dir_all(&tmp).unwrap();
        let src = tmp.join("src.txt");
        let dst = tmp.join("dst.txt");
        std::fs::write(&src, "content").unwrap();

        move_file(&src, &dst).unwrap();
        assert!(dst.exists());
        assert!(!src.exists());

        let _ = std::fs::remove_dir_all(&tmp);
    }

    #[test]
    fn test_rollback() {
        let tmp = std::env::temp_dir().join("test_rollback");
        let _ = std::fs::remove_dir_all(&tmp);
        // 只创建 dst 目录，不创建 src 目录（模拟源已被清理的场景）
        std::fs::create_dir_all(tmp.join("dst")).unwrap();
        std::fs::write(tmp.join("dst").join("file.md"), "content").unwrap();

        rollback(&tmp.join("src"), &tmp.join("dst"), &["file.md".to_string()]);
        // 回滚后，文件应回到源目录
        assert!(tmp.join("src").join("file.md").exists());
        assert!(!tmp.join("dst").join("file.md").exists());

        let _ = std::fs::remove_dir_all(&tmp);
    }
}
