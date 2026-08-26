use std::io::Write;
use std::path::Path;
use std::process::Command;

/// 测试辅助：创建临时项目结构
fn setup_temp_project() -> (tempfile::TempDir, ProjectPaths) {
    let tmp = tempfile::tempdir().unwrap();
    let root = tmp.path().join("project");

    // 创建契约目录
    let contract_dir = root.join(".quanttide/asset");
    std::fs::create_dir_all(&contract_dir).unwrap();
    let mut f = std::fs::File::create(contract_dir.join("contract.yaml")).unwrap();
    writeln!(
        f,
        "skills:\n  archive:\n    version: '1.0'\n    params:\n      pattern: '*.md'\n"
    )
    .unwrap();
    drop(f);

    // 创建输入目录
    let input = root.join("input");
    std::fs::create_dir_all(input.join("product1")).unwrap();
    std::fs::create_dir_all(input.join("product2")).unwrap();
    std::fs::write(input.join("product1/note.md"), "# Test 1").unwrap();
    std::fs::write(input.join("product2/note.md"), "# Test 2").unwrap();

    // 创建输出目录
    let output = root.join("output");
    std::fs::create_dir_all(&output).unwrap();

    (
        tmp,
        ProjectPaths {
            root,
            input,
            output,
        },
    )
}

struct ProjectPaths {
    root: std::path::PathBuf,
    input: std::path::PathBuf,
    output: std::path::PathBuf,
}

/// 获取 CLI 二进制路径
fn cli_binary() -> std::path::PathBuf {
    let manifest_dir = std::path::PathBuf::from(env!("CARGO_MANIFEST_DIR"));
    manifest_dir.join("target/debug/qtcloud-asset")
}

// ── 契约测试 ──

#[test]
fn test_contract_find_root() {
    use qtcloud_asset::contract::Contract;

    let tmp = tempfile::tempdir().unwrap();
    let root = tmp.path().join("project");
    std::fs::create_dir_all(root.join(".quanttide/asset")).unwrap();
    let mut f = std::fs::File::create(root.join(".quanttide/asset/contract.yaml")).unwrap();
    writeln!(f, "skills:\n  test:\n    version: '1.0'\n").unwrap();
    drop(f);

    let contract = Contract::load_at(&root).unwrap();
    assert_eq!(contract.asset_count(), 0);
    assert_eq!(contract.skill_count(), 1);
}

#[test]
fn test_contract_find_root_from_subdir() {
    use qtcloud_asset::contract::Contract;

    let tmp = tempfile::tempdir().unwrap();
    let root = tmp.path().join("project");
    std::fs::create_dir_all(root.join(".quanttide/asset")).unwrap();
    let mut f = std::fs::File::create(root.join(".quanttide/asset/contract.yaml")).unwrap();
    writeln!(f, "skills:\n  test:\n    version: '1.0'\n").unwrap();
    drop(f);

    // 从子目录查找
    let subdir = root.join("src/cli");
    std::fs::create_dir_all(&subdir).unwrap();
    std::env::set_current_dir(&subdir).unwrap();
    let contract = Contract::load().unwrap();
    assert_eq!(contract.skill_count(), 1);
}

#[test]
fn test_contract_not_found() {
    use qtcloud_asset::contract::Contract;

    let tmp = tempfile::tempdir().unwrap();
    let result = Contract::load_at(tmp.path());
    assert!(result.is_err());
}

#[test]
fn test_contract_get_skill() {
    use qtcloud_asset::contract::Contract;

    let tmp = tempfile::tempdir().unwrap();
    let root = tmp.path().join("project");
    std::fs::create_dir_all(root.join(".quanttide/asset")).unwrap();
    let mut f = std::fs::File::create(root.join(".quanttide/asset/contract.yaml")).unwrap();
    writeln!(
        f,
        "skills:\n  archive:\n    version: '1.0'\n    params:\n      pattern: '*.md'\n"
    )
    .unwrap();
    drop(f);

    let contract = Contract::load_at(&root).unwrap();
    let skill = contract.get_skill("archive").unwrap();
    assert_eq!(skill.version, "1.0");
}

#[test]
fn test_contract_get_skill_not_found() {
    use qtcloud_asset::contract::Contract;

    let tmp = tempfile::tempdir().unwrap();
    let root = tmp.path().join("project");
    std::fs::create_dir_all(root.join(".quanttide/asset")).unwrap();
    let mut f = std::fs::File::create(root.join(".quanttide/asset/contract.yaml")).unwrap();
    writeln!(f, "skills:\n  existing:\n    version: '1.0'\n").unwrap();
    drop(f);

    let contract = Contract::load_at(&root).unwrap();
    let result = contract.get_skill("missing");
    assert!(result.is_err());
    assert!(result.unwrap_err().to_string().contains("missing"));
}

// ── 文件操作测试 ──

#[test]
fn test_archive_src_not_found() {
    use qtcloud_asset::file_op;

    let result = file_op::archive_product(
        Path::new("/nonexistent/src"),
        Path::new("/nonexistent/dst"),
        "*.md",
        false,
    );
    assert!(!result.ok());
    assert!(result.error.unwrap().contains("不存在"));
}

#[test]
fn test_archive_dry_run() {
    use qtcloud_asset::file_op;

    let tmp = tempfile::tempdir().unwrap();
    let src = tmp.path().join("src");
    let dst = tmp.path().join("dst");
    std::fs::create_dir_all(&src).unwrap();
    std::fs::write(src.join("file1.md"), "content").unwrap();
    std::fs::write(src.join("file2.md"), "content").unwrap();

    let result = file_op::archive_product(&src, &dst, "*.md", true);
    assert!(result.ok());
    assert_eq!(result.moved.len(), 2);
    // dry_run 不应该实际移动
    assert!(!dst.join("file1.md").exists());
}

#[test]
fn test_archive_moves_files() {
    use qtcloud_asset::file_op;

    let tmp = tempfile::tempdir().unwrap();
    let src = tmp.path().join("src");
    let dst = tmp.path().join("dst");
    std::fs::create_dir_all(&src).unwrap();
    std::fs::write(src.join("file1.md"), "content").unwrap();
    std::fs::write(src.join("file2.md"), "content").unwrap();

    let result = file_op::archive_product(&src, &dst, "*.md", false);
    assert!(result.ok());
    assert_eq!(result.moved.len(), 2);
    assert!(dst.join("file1.md").exists());
    assert!(!src.join("file1.md").exists());
}

#[test]
fn test_archive_skips_existing() {
    use qtcloud_asset::file_op;

    let tmp = tempfile::tempdir().unwrap();
    let src = tmp.path().join("src");
    let dst = tmp.path().join("dst");
    std::fs::create_dir_all(&src).unwrap();
    std::fs::create_dir_all(&dst).unwrap();
    std::fs::write(src.join("file1.md"), "content").unwrap();
    std::fs::write(src.join("file2.md"), "content").unwrap();
    std::fs::write(dst.join("file1.md"), "existing").unwrap();

    let result = file_op::archive_product(&src, &dst, "*.md", false);
    assert!(result.ok());
    assert_eq!(result.moved, vec!["file2.md"]);
    assert_eq!(result.skipped, vec!["file1.md"]);
}

#[test]
fn test_archive_no_matching_files() {
    use qtcloud_asset::file_op;

    let tmp = tempfile::tempdir().unwrap();
    let src = tmp.path().join("src");
    let dst = tmp.path().join("dst");
    std::fs::create_dir_all(&src).unwrap();
    std::fs::write(src.join("file.txt"), "content").unwrap();

    let result = file_op::archive_product(&src, &dst, "*.md", false);
    assert!(result.ok());
    assert_eq!(result.moved.len(), 0);
    assert!(result.skipped[0].contains("无匹配"));
}

#[test]
fn test_archive_removes_empty_src() {
    use qtcloud_asset::file_op;

    let tmp = tempfile::tempdir().unwrap();
    let src = tmp.path().join("src");
    let dst = tmp.path().join("dst");
    std::fs::create_dir_all(&src).unwrap();
    std::fs::write(src.join("file1.md"), "content").unwrap();

    let result = file_op::archive_product(&src, &dst, "*.md", false);
    assert!(result.ok());
    assert!(result.source_removed);
    assert!(!src.exists());
}

// ── 工作流测试 ──

#[test]
fn test_get_products_empty() {
    use qtcloud_asset::workflow;

    let tmp = tempfile::tempdir().unwrap();
    let dir = tmp.path().join("empty");
    std::fs::create_dir_all(&dir).unwrap();

    let contract = qtcloud_asset::contract::Contract {
        root: tmp.path().to_path_buf(),
        schema: qtcloud_asset::contract::ContractSchema {
            assets: std::collections::HashMap::new(),
            skills: std::collections::HashMap::new(),
            validation: None,
        },
    };
    let result = workflow::resolve_workflow("test", &dir, &tmp.path().join("out"), None, &contract);
    assert!(result.is_err()); // skill not found
}

#[test]
fn test_get_products_not_found() {
    use qtcloud_asset::workflow;

    let tmp = tempfile::tempdir().unwrap();

    let contract = qtcloud_asset::contract::Contract {
        root: tmp.path().to_path_buf(),
        schema: qtcloud_asset::contract::ContractSchema {
            assets: std::collections::HashMap::new(),
            skills: std::collections::HashMap::new(),
            validation: None,
        },
    };
    let result = workflow::resolve_workflow(
        "test",
        &tmp.path().join("nonexistent"),
        &tmp.path().join("out"),
        None,
        &contract,
    );
    assert!(result.is_err());
}

#[test]
fn test_resolve_workflow_with_contract() {
    use qtcloud_asset::{contract::Contract, workflow};

    let tmp = tempfile::tempdir().unwrap();
    let root = tmp.path().join("project");
    let contract_dir = root.join(".quanttide/asset");
    std::fs::create_dir_all(&contract_dir).unwrap();
    let mut f = std::fs::File::create(contract_dir.join("contract.yaml")).unwrap();
    writeln!(
        f,
        "skills:\n  archive:\n    version: '1.0'\n    params:\n      pattern: '*.md'\n"
    )
    .unwrap();
    drop(f);

    let input = root.join("input");
    std::fs::create_dir_all(input.join("product1")).unwrap();
    std::fs::create_dir_all(input.join("product2")).unwrap();
    let output = root.join("output");
    std::fs::create_dir_all(&output).unwrap();

    let contract = Contract::load_at(&root).unwrap();
    let wf = workflow::resolve_workflow("archive", &input, &output, None, &contract).unwrap();

    assert_eq!(wf.name, "archive");
    assert_eq!(wf.pattern, "*.md");
    assert_eq!(wf.tasks.len(), 2);
    assert_eq!(wf.tasks[0].product, "product1");
    assert_eq!(wf.tasks[1].product, "product2");
}

#[test]
fn test_resolve_workflow_pattern_override() {
    use qtcloud_asset::{contract::Contract, workflow};

    let tmp = tempfile::tempdir().unwrap();
    let root = tmp.path().join("project");
    let contract_dir = root.join(".quanttide/asset");
    std::fs::create_dir_all(&contract_dir).unwrap();
    let mut f = std::fs::File::create(contract_dir.join("contract.yaml")).unwrap();
    writeln!(
        f,
        "skills:\n  archive:\n    version: '1.0'\n    params:\n      pattern: '*.txt'\n"
    )
    .unwrap();
    drop(f);

    let input = root.join("input");
    std::fs::create_dir_all(input.join("product1")).unwrap();
    let output = root.join("output");
    std::fs::create_dir_all(&output).unwrap();

    let contract = Contract::load_at(&root).unwrap();
    let wf =
        workflow::resolve_workflow("archive", &input, &output, Some("*.md"), &contract).unwrap();
    // 参数优先级高于契约配置
    assert_eq!(wf.pattern, "*.md");
}

#[test]
fn test_resolve_workflow_unknown_skill() {
    use qtcloud_asset::{contract::Contract, workflow};

    let tmp = tempfile::tempdir().unwrap();
    let root = tmp.path().join("project");
    let contract_dir = root.join(".quanttide/asset");
    std::fs::create_dir_all(&contract_dir).unwrap();
    let mut f = std::fs::File::create(contract_dir.join("contract.yaml")).unwrap();
    writeln!(f, "skills:\n  existing:\n    version: '1.0'\n").unwrap();
    drop(f);

    let input = root.join("input");
    std::fs::create_dir_all(&input).unwrap();
    let output = root.join("output");
    std::fs::create_dir_all(&output).unwrap();

    let contract = Contract::load_at(&root).unwrap();
    let result = workflow::resolve_workflow("nonexistent", &input, &output, None, &contract);
    assert!(result.is_err());
    assert!(result.unwrap_err().to_string().contains("nonexistent"));
}

// ── CLI 集成测试 ──

#[test]
fn test_cli_version() {
    let output = Command::new(cli_binary()).arg("version").output().unwrap();
    assert!(output.status.success());
    let stdout = String::from_utf8_lossy(&output.stdout);
    assert!(stdout.contains("qtcloud-asset-cli"));
}

#[test]
fn test_cli_help() {
    let output = Command::new(cli_binary()).arg("--help").output().unwrap();
    assert!(output.status.success());
    let stdout = String::from_utf8_lossy(&output.stdout);
    assert!(stdout.contains("run"));
    assert!(stdout.contains("scan"));
    assert!(stdout.contains("validate"));
    assert!(stdout.contains("config"));
    assert!(stdout.contains("version"));
}

#[test]
fn test_cli_config() {
    let (_tmp, paths) = setup_temp_project();

    let output = Command::new(cli_binary())
        .arg("config")
        .current_dir(&paths.root)
        .output()
        .unwrap();
    assert!(output.status.success());
    let stdout = String::from_utf8_lossy(&output.stdout);
    assert!(stdout.contains("技能"));
    assert!(stdout.contains("archive"));
}

#[test]
fn test_cli_run_dry_run() {
    let (_tmp, paths) = setup_temp_project();

    let output = Command::new(cli_binary())
        .args([
            "run",
            "-s",
            "archive",
            "-i",
            &paths.input.to_string_lossy(),
            "-o",
            &paths.output.to_string_lossy(),
            "-n", // dry-run
        ])
        .current_dir(&paths.root)
        .output()
        .unwrap();
    assert!(output.status.success());
    let stdout = String::from_utf8_lossy(&output.stdout);
    assert!(stdout.contains("预览") || stdout.contains("[预览]"));
}

#[test]
fn test_cli_run_success() {
    let (_tmp, paths) = setup_temp_project();

    let output = Command::new(cli_binary())
        .args([
            "run",
            "-s",
            "archive",
            "-i",
            &paths.input.to_string_lossy(),
            "-o",
            &paths.output.to_string_lossy(),
        ])
        .current_dir(&paths.root)
        .output()
        .unwrap();
    assert!(output.status.success());
    let stdout = String::from_utf8_lossy(&output.stdout);
    assert!(stdout.contains("完成"));

    // 验证文件已移动
    assert!(paths.output.join("product1/note.md").exists());
    assert!(!paths.input.join("product1/note.md").exists());
}

#[test]
fn test_cli_scan() {
    let (_tmp, paths) = setup_temp_project();

    let output = Command::new(cli_binary())
        .args(["scan"])
        .current_dir(&paths.input)
        .output()
        .unwrap();
    assert!(output.status.success());
    let stdout = String::from_utf8_lossy(&output.stdout);
    assert!(stdout.contains("扫描完成"));
}

#[test]
fn test_cli_scan_json() {
    let (_tmp, paths) = setup_temp_project();

    let output = Command::new(cli_binary())
        .args(["scan", "--json"])
        .current_dir(&paths.input)
        .output()
        .unwrap();
    assert!(output.status.success());
    let stdout = String::from_utf8_lossy(&output.stdout);
    // JSON 输出应包含 root_path
    assert!(stdout.contains("root_path"));
}

#[test]
fn test_cli_validate() {
    let (_tmp, paths) = setup_temp_project();

    let output = Command::new(cli_binary())
        .args(["validate"])
        .current_dir(&paths.root)
        .output()
        .unwrap();
    // 没有定义 validation policies，所以应该失败
    assert!(!output.status.success());
}

#[test]
fn test_cli_pattern_override() {
    let (_tmp, paths) = setup_temp_project();

    // 契约中配置 *.md，但指定 *.txt 应该跳过所有 .md 文件
    let output = Command::new(cli_binary())
        .args([
            "run",
            "-s",
            "archive",
            "-i",
            &paths.input.to_string_lossy(),
            "-o",
            &paths.output.to_string_lossy(),
            "-p",
            "*.txt",
        ])
        .current_dir(&paths.root)
        .output()
        .unwrap();
    assert!(output.status.success());
    let stdout = String::from_utf8_lossy(&output.stdout);
    // 无匹配 .txt 文件，但工作流应仍成功
    assert!(stdout.contains("完成"));
}
