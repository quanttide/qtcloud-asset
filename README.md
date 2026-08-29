# 量潮数字资产云

可视化对象存储，把散落在阿里云 OSS 上的桶和文件以可视化方式呈现，让非技术人员无需 CLI 也能查看资源状态。这是数字资产全景（BRD 场景之一）的第一次真实落地。

## 架构

```
Studio (Flutter Web) ← HTTP → Provider (Go) ← SDK → 阿里云 OSS
                            ↑
                       CLI (Rust 管理工具)
```

Provider 通过 `OssAdapter` 只读发现 OSS 桶与对象，Studio 负责可视化展示。CLI 负责本地资产扫描、验证与归档。

## 目录结构

```
├── .quanttide/           # 项目契约
│   ├── asset/
│   │   └── contract.yaml # 资产组成定义
│   ├── code/
│   │   └── contract.yaml # 代码规则定义
│   └── agent/
│       └── contract.yaml # AI 执行审核规则
├── docs/                 # 项目文档
│   ├── brd/              # 商业需求文档
│   ├── prd/              # 产品需求文档
│   ├── ixd/              # 交互设计文档
│   ├── qa/               # 质量保证文档
│   └── user/             # 用户文档
├── src/                  # 源代码
│   ├── cli/              # Rust CLI 管理工具
│   ├── studio/           # Flutter Web 客户端
│   └── provider/         # Go 服务端
├── manifests/            # 部署清单
│   ├── docker/           # Docker 部署配置
│   └── iac/              # 阿里云基础设施 (Terraform)
├── tests/                # 测试代码
└── examples/             # 示例与原型验证
```

## 快速开始

### 本地开发

无需 Docker，直接本地跑即可。先启动 Provider，再启动 Studio：

```bash
# 1. 启动服务端（需配置 OSS 凭证，见下方环境变量）
cd src/provider
go run ./cmd/provider

# 2. 启动客户端（另开终端）
cd src/studio
flutter run -d web-server
```

Provider 默认监听 `http://localhost:9000`，Studio 连接 `http://127.0.0.1:9000`，界面地址见 flutter 启动日志（如 `http://127.0.0.1:8090`）。

Provider 需要以下环境变量：

| 环境变量 | 说明 |
|---------|------|
| `OSS_ACCESS_KEY_ID` | 阿里云 AccessKey ID |
| `OSS_ACCESS_KEY_SECRET` | 阿里云 AccessKey Secret |
| `OSS_ENDPOINT` | OSS 端点，默认 `https://oss-cn-hangzhou.aliyuncs.com` |
| `AUTH_MODE` | 认证模式，默认 `sso`；内测账号密码登录使用 `local` |
| `LOCAL_AUTH_ACCOUNT` | 本地登录账号，仅 `AUTH_MODE=local` 时使用 |
| `LOCAL_AUTH_EMAIL` | 兼容旧配置的邮箱字段，可为空 |
| `LOCAL_AUTH_PASSWORD_HASH` | PBKDF2-SHA256 密码哈希，不写入明文密码 |

### CLI 工具

```bash
cd src/cli
cargo run -- --help      # 查看命令
cargo test               # 运行测试
cargo clippy             # 代码检查
```

CLI 支持本地资产扫描、验证与归档，也支持通过 `oss` 子命令复用 Provider 接口管理 OSS：

```bash
cargo run -- oss list                # 列出所有桶
cargo run -- oss ls <桶名>           # 列出桶内对象
cargo run -- oss url <桶名> <key>    # 生成公开桶对象访问链接
```

## 已实现能力

- Provider 只读端点：按角色列出桶、列出对象、生成公开桶直链；`viewer` 隐藏 `-private` 与 `quanttide-terraform-state`，`admin` 可查看全部但不能生成私密桶对象链接
- Provider 账号门禁：SSO 占位入口、本地账号密码登录、`viewer` / `admin` 鉴权、管理员用户管理接口
- Provider 审计：结构化 `qtcloud_asset_audit` stdout 日志，生产函数计算可通过 SLS 持久查询
- Studio 桶列表：按用途分类筛选、搜索、创建时间排序
- Studio 文件浏览：目录层级下钻、搜索、日期/大小排序
- Studio 登录页：账号密码登录、当前用户展示、退出登录和管理员入口
- Studio 复制对象链接：仅公开桶返回直链；私密桶只展示对象元数据，不提供链接分享
- Studio 文件/文件夹分享：公开桶支持选择单个或多个文件、文件夹生成一个只读分享页链接；分享页支持单文件下载和 ZIP 下载全部；私密桶不显示分享入口
- CLI `oss` 子命令：命令行方式复用上述只读能力

## 契约定义

项目使用 `.quanttide/` 目录定义所有契约：

| 文件 | 用途 |
|------|------|
| `.quanttide/asset/contract.yaml` | 资产组成、路径、类型 |
| `.quanttide/code/contract.yaml` | 编程规范、依赖、质量门禁 |
| `.quanttide/agent/contract.yaml` | AI 执行审核规则 |

详见 [CONTRIBUTING.md](CONTRIBUTING.md)。

## 部署

### 生产入口

- 正式 Studio：`https://asset.cloud.quanttide.com`
- 兼容入口：`https://asset.quanttide.com`，当前暂不处理旧入口下线
- Provider API：`https://api.quanttide.com/qtcloud-asset`

Studio 的 CI 发布目标是 `qtcloud-asset-studio`。当前正式入口已完成 DNS、CDN、HTTPS、CORS 和登录后浏览器验收；阶段六已记录新域名关键对象 ETag、旧桶前端对象清单和回滚证据。Provider 发布包短期仍保留在 `qtcloud-asset/provider/`，不上传到 Studio 桶。

### 历史 Docker 参考

当前生产路径不使用 Docker Compose、Dockerfile 或 Kubernetes 镜像。下面的命令仅保留作历史本地参考，不作为 Studio 或 Provider 的发布方式。

```bash
docker compose -f manifests/docker/docker-compose.yml up --build
```

- Studio: http://localhost:8080
- Provider: http://localhost:9000

### 阿里云基础设施

应用级 Terraform 仅管理 Studio OSS 桶和可选的 FC 应用资源。API 网关、DNS 和证书由 `quanttide-platform` 管理，本项目不执行平台级域名接入。

首次接管已存在的 Studio 桶前，应先完成 Terraform state import，再执行 `terraform plan`：

```bash
cd manifests/iac
terraform init
terraform import alicloud_oss_bucket.studio qtcloud-asset-studio
terraform plan
```

支持环境：`dev` / `staging` / `prod`

## 文档

- [商业需求](docs/brd/) - 业务痛点和动机
- [产品需求](docs/prd/) - 产品功能和设计
- [交互设计](docs/ixd/) - 界面布局和交互流程
- [质量保证](docs/qa/) - 质量决策和记录
- [用户文档](docs/user/) - 安装和使用指南

路线图见 [ROADMAP.md](ROADMAP.md)，发布记录见 [CHANGELOG.md](CHANGELOG.md)，剩余事项见 [TODO.md](TODO.md)。
