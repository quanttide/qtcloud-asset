use crate::scanner::AssetInfo;

/// 打印成功信息
pub fn print_success(message: &str) {
    println!("✅ {}", message);
}

/// 打印错误信息
pub fn print_error(message: &str) {
    eprintln!("❌ {}", message);
}

/// 打印警告信息
pub fn print_warning(message: &str) {
    println!("⚠️  {}", message);
}

/// 打印普通信息
pub fn print_info(message: &str) {
    println!("ℹ️  {}", message);
}

/// 打印标题
pub fn print_header(message: &str) {
    println!("\n=== {} ===", message);
}

/// 打印资产表格
pub fn print_asset_table(assets: &[AssetInfo]) {
    if assets.is_empty() {
        print_warning("未发现任何资产");
        return;
    }

    println!("\n资产清单:");
    for asset in assets {
        println!("  - {} ({})", asset.name, asset.asset_type);
    }
}

/// 打印验证报告
pub fn print_validate_report(passed: usize, failed: usize, total: usize) {
    if failed == 0 {
        println!("\n✅ 所有验证通过！({}/{})", passed, total);
    } else {
        println!("\n⚠️  有 {} 项验证失败 ({}/{} 通过)", failed, passed, total);
    }
}

/// 打印工作流摘要
pub fn print_workflow_summary(name: &str, pattern: &str, tasks_count: usize, dry_run: bool) {
    let mode = if dry_run { "预览" } else { "执行" };
    println!("\n[{}] 技能: {}", mode, name);
    println!("[{}] 模式: {}", mode, pattern);
    println!("[{}] 产品数: {}", mode, tasks_count);
}
