# QtCloud Asset CLI

量潮数字资产云 CLI — 资产管理工具。

## 安装

### 方式一：从源码构建

需要 Rust 工具链（1.75+）：

```bash
cd src/cli
cargo build --release
./target/release/qtcloud-asset --help
```

### 方式二：下载预编译二进制

从 [GitHub Releases](https://github.com/quanttide/qtcloud-asset/releases) 下载对应平台的压缩包：

| 平台 | 架构 | 文件 |
|------|------|------|
| Linux | x86_64 | `qtcloud-asset-x86_64-unknown-linux-gnu.tar.gz` |
| Linux | ARM64 | `qtcloud-asset-aarch64-unknown-linux-gnu.tar.gz` |
| macOS | Intel | `qtcloud-asset-x86_64-apple-darwin.tar.gz` |
| macOS | Apple Silicon | `qtcloud-asset-aarch64-apple-darwin.tar.gz` |
| Windows | x86_64 | `qtcloud-asset-x86_64-pc-windows-msvc.tar.gz` |

## 用法

```
量潮数字资产云 CLI — 资产管理工具

Usage: qtcloud-asset <COMMAND>

Commands:
  run       一键执行资产管理工作流（归档）
  scan      扫描目录，列出所有资产
  validate  验证资产是否符合契约要求
  config    查看契约配置
  oss       管理 OSS 对象存储（复用 Provider 接口）
  version   显示版本和预发布阶段
  help      Print this message or the help of the given subcommand(s)

Options:
  -h, --help     Print help
  -V, --version  Print version
```

### run — 归档工作流

```bash
# 默认归档当前目录下的 journal → output
qtcloud-asset run

# 指定输入输出目录
qtcloud-asset run -i ./journal -o ./archive

# 指定文件匹配模式
qtcloud-asset run -p "*.md" --dry-run

# 详细模式
qtcloud-asset run -v
```

### scan — 扫描资产

```bash
# 扫描当前目录
qtcloud-asset scan

# 扫描指定目录
qtcloud-asset scan -i ./docs

# 详细输出（列出资产）
qtcloud-asset scan -v

# JSON 格式输出
qtcloud-asset scan --json
```

### validate — 验证资产

```bash
# 验证当前目录
qtcloud-asset validate

# 验证指定目录
qtcloud-asset validate -i ./output

# 指定契约文件
qtcloud-asset validate -c ./contract.yaml

# JSON 格式输出
qtcloud-asset validate --json
```

### config — 查看契约

```bash
# 显示契约配置
qtcloud-asset config

# 列出资产
qtcloud-asset config -a list
```

### version — 版本信息

```bash
qtcloud-asset version
```

### oss — 管理 OSS 对象存储

`oss` 子命令作为 Provider 的 HTTP 客户端，复用其只读接口，不直接访问阿里云 OSS，也不持有 AK/SK。需先启动 Provider（默认监听 `http://127.0.0.1:9000`）。

```bash
# 列出所有 OSS 桶
qtcloud-asset oss list

# 列出桶内对象
qtcloud-asset oss ls <桶名>

# 列出桶内对象，按 key 前缀过滤
qtcloud-asset oss ls <桶名> --prefix docs/

# 按创建时间倒序排列桶
qtcloud-asset oss list --sort created --order desc

# 按文件大小倒序排列对象
qtcloud-asset oss ls <桶名> --sort size --order desc

# 分页：每页 100 个对象
qtcloud-asset oss ls <桶名> --limit 100

# 生成公开桶对象访问链接（私密桶不支持）
qtcloud-asset oss url <桶名> <对象key>

# 兼容旧参数（公开桶忽略有效期）
qtcloud-asset oss url <桶名> <对象key> --expires 604800
```

排序与分页参数：

| 命令 | 排序字段（`--sort`） | 说明 |
|------|---------------------|------|
| `oss list` | `name` / `created` | 按桶名或创建时间排序 |
| `oss ls` | `key` / `size` / `date` | 按 key、大小或修改日期排序 |

`--order` 取值 `asc`（升序，默认）/ `desc`（降序）；`--limit` 指定每页数量（走 Provider 分页）。

所有 `oss` 子命令支持 `--provider-url` 覆盖默认的 Provider 地址：

```bash
qtcloud-asset oss list --provider-url http://localhost:9000
```

## 契约系统

CLI 通过 `.quanttide/asset/contract.yaml` 驱动，契约定义了两类配置：

### 资产定义

```yaml
assets:
  brd:
    type: docs
    path: docs/brd
    description: 商业需求文档
```

### 技能定义

```yaml
skills:
  archive-journal:
    version: "1.0"
    params:
      pattern: "*.md"
```

### 验证策略

```yaml
validation:
  policies:
    - selector: "议事档案"
      mode: ATOMIC
      required_categories:
        - 议题
        - 议程
    - selector: "**"
      mode: SCOPED
```

支持两种验证模式：

| 模式 | 说明 |
|------|------|
| `ATOMIC` | 资产必须包含所有 `required_categories` |
| `SCOPED` | 资产存在即通过 |

## 开发

```bash
# 运行测试
cargo test

# 代码检查
cargo clippy

# 格式化
cargo fmt
```

## 架构

```
┌──────────────────────────────────────────┐
│  clap 入口 → 命令分发                     │
│                                           │
│  contract.rs — .quanttide/asset/contract  │
│  scanner.rs  — 目录扫描                   │
│  file_op.rs  — 文件移动/回滚              │
│  workflow.rs — 契约→工作流解析             │
│  validate.rs — 声明式验证                 │
│  oss.rs      — Provider HTTP 客户端        │
│  oss_cmd.rs  — oss 子命令定义             │
│  render.rs   — 终端输出                   │
└──────────────────────────────────────────┘
```

## 变更记录

### v0.1.0 (alpha)

- 基础 CLI 骨架：run / scan / validate / config / version
- 契约加载：读取 `.quanttide/asset/contract.yaml`
- 目录扫描：递归扫描 + 资产类型推测
- 归档工作流：文件移动 + 失败回滚 + 空目录清理
- 声明式验证：ATOMIC / SCOPED 策略
- JSON 输出：支持 `--json` 标志
- 单元测试：31 个测试覆盖全部模块
- CI/CD：GitHub Actions 测试 + 交叉编译发布
