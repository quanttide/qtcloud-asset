mod asset;
mod cli_config;

use clap::{Parser, Subcommand};

#[derive(Parser)]
#[command(
    name = "qtcloud-asset",
    version,
    about = "asset CLI 参考实现（自 qtadmin 迁移）"
)]
struct Cli {
    #[command(subcommand)]
    command: Option<asset::AssetCommands>,
}

fn main() {
    let cli = Cli::parse();
    match &cli.command {
        Some(cmd) => asset::dispatch(&asset::AssetArgs {
            command: cmd.clone(),
        }),
        None => {}
    }
}
