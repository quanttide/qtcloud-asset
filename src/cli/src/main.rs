use clap::{Parser, Subcommand};
use qtcloud_asset::{contract, file_op, oss_cmd, render, scanner, validate, workflow};

const VERSION: &str = "0.1.0";
const STAGE: &str = "alpha";

#[derive(Parser)]
#[command(
    name = "qtcloud-asset",
    about = "量潮数字资产云 CLI — 资产管理工具",
    version = VERSION,
)]
struct Cli {
    #[command(subcommand)]
    command: Command,
}

#[derive(Subcommand)]
enum Command {
    /// 一键执行资产管理工作流（归档）
    Run(RunArgs),
    /// 扫描目录，列出所有资产
    Scan(scanner::ScanArgs),
    /// 验证资产是否符合契约要求
    Validate(validate::ValidateArgs),
    /// 查看契约配置
    Config(ConfigArgs),
    /// 管理 OSS 对象存储（复用 Provider 接口）
    #[command(subcommand)]
    Oss(oss_cmd::OssCommand),
    /// 显示版本和预发布阶段
    Version,
}

#[derive(clap::Args)]
struct RunArgs {
    /// 技能名称（默认 archive-journal）
    #[arg(short, long, default_value = "archive-journal")]
    skill: String,

    /// 输入目录（默认当前目录）
    #[arg(short = 'i', long, default_value = ".")]
    input: String,

    /// 输出目录（默认 ./output）
    #[arg(short = 'o', long, default_value = "output")]
    output: String,

    /// 文件匹配模式（默认从契约读取）
    #[arg(short, long)]
    pattern: Option<String>,

    /// 预览模式
    #[arg(short = 'n', long)]
    dry_run: bool,

    /// 详细输出
    #[arg(short, long)]
    verbose: bool,
}

#[derive(clap::Args)]
struct ConfigArgs {
    /// 操作: show（显示契约）/ list（列出资产）
    #[arg(short, long, default_value = "show")]
    action: String,

    /// 契约文件路径（默认自动查找）
    #[arg(short, long)]
    contract: Option<String>,
}

fn main() -> anyhow::Result<()> {
    let cli = Cli::parse();
    match cli.command {
        Command::Run(args) => execute_run(args),
        Command::Scan(args) => scanner::execute(args),
        Command::Validate(args) => validate::execute(args),
        Command::Config(args) => execute_config(args),
        Command::Oss(cmd) => oss_cmd::execute(cmd),
        Command::Version => Ok(execute_version()),
    }
}

fn execute_version() {
    println!("qtcloud-asset-cli v{VERSION}");
    println!("阶段: {STAGE}");
}

fn execute_run(args: RunArgs) -> anyhow::Result<()> {
    use std::path::Path;

    let input_dir = Path::new(&args.input);
    let output_dir = Path::new(&args.output);

    render::print_header("资产管理工作流");

    let contract = contract::Contract::load()?;

    let wf = workflow::resolve_workflow(
        &args.skill,
        input_dir,
        output_dir,
        args.pattern.as_deref(),
        &contract,
    )?;

    if args.verbose {
        render::print_workflow_summary(
            &wf.name,
            &wf.pattern,
            wf.tasks.len(),
            args.dry_run,
        );
    }

    let mut success_products = 0usize;
    let mut failed_products = 0usize;

    for task in &wf.tasks {
        let result = file_op::archive_product(
            &task.src_dir,
            &task.dst_dir,
            &wf.pattern,
            args.dry_run,
        );

        if result.ok() {
            if args.dry_run {
                println!("  [预览] {}: {}", task.product, result.moved.join(", "));
            } else {
                let label = format!("归档 {} 个文件", result.moved.len());
                render::print_success(&format!("{}: {}", task.product, label));
            }
            success_products += 1;
        } else {
            let err = result.error.as_deref().unwrap_or("未知错误");
            render::print_error(&format!("{}: {}", task.product, err));
            failed_products += 1;
        }
    }

    let total = wf.tasks.len();
    if failed_products > 0 {
        render::print_warning(&format!(
            "完成: {}/{} 成功，{} 失败",
            success_products, total, failed_products
        ));
        std::process::exit(1);
    } else {
        render::print_success(&format!("完成: {}/{} 全部成功", success_products, total));
    }

    Ok(())
}

fn execute_config(args: ConfigArgs) -> anyhow::Result<()> {
    render::print_header("契约配置");

    let contract = if let Some(ref path) = args.contract {
        contract::Contract::load_at(std::path::Path::new(path).parent().unwrap_or(std::path::Path::new(".")))
    } else {
        contract::Contract::load()
    }?;

    if args.action == "show" {
        render::print_info(&format!("契约位置: {}", contract.path().display()));
        render::print_info(&format!("资产数: {}", contract.asset_count()));
        render::print_info(&format!("技能数: {}", contract.skill_count()));

        println!("\n资产列表:");
        for (name, asset) in contract.assets() {
            println!("  - {}: {}", name, asset.r#type);
        }

        println!("\n技能列表:");
        for (name, skill) in contract.skills() {
            println!("  - {}: v{}", name, skill.version);
        }
    } else if args.action == "list" {
        render::print_info("资产配置:");
        for (name, _) in contract.assets() {
            println!("  {}", name);
        }
    }

    Ok(())
}
