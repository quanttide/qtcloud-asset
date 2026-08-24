//! OSS 子命令的参数定义与执行入口。

use anyhow::Result;
use clap::{Args, Subcommand};

use crate::oss::{self, Client};
use crate::render;

/// OSS 相关子命令。
#[derive(Subcommand)]
pub enum OssCommand {
    /// 列出所有 OSS 桶
    List(OssListArgs),
    /// 列出桶内对象
    Ls(OssLsArgs),
    /// 生成对象访问链接
    Url(OssUrlArgs),
}

#[derive(Args)]
pub struct OssListArgs {
    /// 排序字段：name / created
    #[arg(long, value_parser = ["name", "created"])]
    sort: Option<String>,

    /// 排序方向：asc / desc
    #[arg(long, value_parser = ["asc", "desc"])]
    order: Option<String>,

    /// Provider 服务地址
    #[arg(long, default_value = oss::default_provider_url())]
    provider_url: String,
}

#[derive(Args)]
pub struct OssLsArgs {
    /// 桶名
    bucket: String,

    /// 对象 key 前缀（可选，用于过滤）
    #[arg(long)]
    prefix: Option<String>,

    /// 排序字段：key / size / date
    #[arg(long, value_parser = ["key", "size", "date"])]
    sort: Option<String>,

    /// 排序方向：asc / desc
    #[arg(long, value_parser = ["asc", "desc"])]
    order: Option<String>,

    /// 每页数量
    #[arg(long)]
    limit: Option<i64>,

    /// Provider 服务地址
    #[arg(long, default_value = oss::default_provider_url())]
    provider_url: String,
}

#[derive(Args)]
pub struct OssUrlArgs {
    /// 桶名
    bucket: String,

    /// 对象 key
    key: String,

    /// 链接有效期（秒），公开桶忽略此参数
    #[arg(long, default_value_t = 86400)]
    expires: i64,

    /// Provider 服务地址
    #[arg(long, default_value = oss::default_provider_url())]
    provider_url: String,
}

/// 分发 OSS 子命令。
pub fn execute(cmd: OssCommand) -> Result<()> {
    match cmd {
        OssCommand::List(args) => list(args),
        OssCommand::Ls(args) => ls(args),
        OssCommand::Url(args) => url(args),
    }
}

fn list(args: OssListArgs) -> Result<()> {
    let client = Client::new(&args.provider_url);
    let buckets = client.list_buckets(args.sort.as_deref().unwrap_or(""), args.order.as_deref().unwrap_or(""))?;

    render::print_header("OSS 桶列表");
    render::print_info(&format!("共 {} 个桶", buckets.len()));
    for b in &buckets {
        render::print_bucket(b);
    }
    Ok(())
}

fn ls(args: OssLsArgs) -> Result<()> {
    let client = Client::new(&args.provider_url);
    let (objects, next_marker, truncated) = client.list_objects(
        &args.bucket,
        args.prefix.as_deref().unwrap_or(""),
        args.sort.as_deref().unwrap_or(""),
        args.order.as_deref().unwrap_or(""),
        args.limit.unwrap_or(0),
        "",
    )?;

    render::print_header(&format!("桶 {} 的对象", args.bucket));
    render::print_info(&format!("共 {} 个对象", objects.len()));
    for o in &objects {
        render::print_object(o);
    }

    if truncated {
        match next_marker {
            Some(m) => render::print_info(&format!("还有更多对象，下一页 marker: {m}")),
            None => render::print_info("还有更多对象"),
        }
    }
    Ok(())
}

fn url(args: OssUrlArgs) -> Result<()> {
    let client = Client::new(&args.provider_url);
    let link = client.object_url(&args.bucket, &args.key, args.expires)?;

    render::print_header("对象访问链接");
    println!("{link}");
    Ok(())
}
