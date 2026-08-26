# 新人使用手册

本文面向第一次接手量潮数字资产云的新同学，目标是在本地把 Provider、Studio 和 CLI 跑起来，并理解哪些配置可以改、哪些配置不能写入仓库。

## 准备环境

本项目当前开发链路包含三部分：

- Provider：Go 服务，默认监听 `9000`。
- Studio：Flutter Web 前端，开发时通常监听 `8090`。
- CLI：Rust 命令行工具，用于本地资产扫描和 OSS 只读接口验证。

需要提前安装：

| 工具 | 用途 | 验证命令 |
|:--|:--|:--|
| Go | 运行 Provider | `go version` |
| Flutter | 运行 Studio | `flutter --version` |
| Rust / Cargo | 运行 CLI | `cargo --version` |
| Git | 版本管理 | `git --version` |

推荐先在项目根目录执行：

```bash
git status --short --branch
```

确认当前分支和工作区状态。如果只是本地试跑，不要提交 `.env`、构建目录、截图、测试结果或临时日志。

## 配置原则

凭证只能放在本机环境变量、安全配置、CI Secret、函数计算环境变量或 RAM 角色里，不能写进源码、文档、静态构建产物、日志或 Git 历史。

本地开发常用配置如下：

| 变量 | 必填 | 说明 |
|:--|:--:|:--|
| `PROVIDER_PORT` | 否 | Provider 监听端口，默认 `9000` |
| `PROVIDER_BASE_URL` | 否 | Provider 对外基础地址，生产为 `https://api.quanttide.com/qtcloud-asset` |
| `STUDIO_ORIGIN` | 否 | 正式 Studio 来源，默认 `https://asset.cloud.quanttide.com` |
| `OSS_ENDPOINT` | 否 | OSS 端点，默认 `https://oss-cn-hangzhou.aliyuncs.com` |
| `OSS_ACCESS_KEY_ID` | 本地需要 | 本地开发用 AccessKey ID |
| `OSS_ACCESS_KEY_SECRET` | 本地需要 | 本地开发用 AccessKey Secret |
| `OSS_SESSION_TOKEN` | 临时凭证需要 | 本地开发用临时凭证 token |
| `AUTH_MODE` | 否 | 默认 `sso`；本地账号密码登录用 `local` |
| `LOCAL_AUTH_EMAIL` | local 需要 | 本地登录邮箱 |
| `LOCAL_AUTH_NAME` | 否 | 本地登录展示名，空值时使用邮箱 |
| `LOCAL_AUTH_ROLE` | 否 | 本地登录角色，`viewer` 或 `admin`，默认 `admin` |
| `LOCAL_AUTH_PASSWORD_HASH` | local 需要 | PBKDF2-SHA256 密码哈希，不写明文密码 |

Windows PowerShell 示例：

```powershell
$env:OSS_ENDPOINT = "https://oss-cn-hangzhou.aliyuncs.com"
$env:AUTH_MODE = "local"
$env:LOCAL_AUTH_EMAIL = "your.name@example.com"
$env:LOCAL_AUTH_ROLE = "admin"
$env:LOCAL_AUTH_PASSWORD_HASH = "pbkdf2_sha256$..."
```

不要把真实 `OSS_ACCESS_KEY_ID`、`OSS_ACCESS_KEY_SECRET` 或 `LOCAL_AUTH_PASSWORD_HASH` 贴进聊天、提交信息或 Markdown 文档。

## 生成本地密码哈希

本地账号密码登录不读取明文密码，只读取 `LOCAL_AUTH_PASSWORD_HASH`。可以用 Provider 内置 Go 函数在测试或临时小工具中生成 PBKDF2-SHA256 哈希。生成后只保存哈希，不保存明文。

如果只是验证链路，也可以请已有维护者提供一次性内测哈希值。不要在仓库里新增真实密码或真实哈希。

## 启动 Provider

在项目根目录执行：

```bash
cd src/provider
go run ./cmd/provider
```

启动成功后，新开一个终端验证：

```bash
curl http://127.0.0.1:9000/health
```

预期返回类似：

```json
{"status":"ok","service":"qtcloud-asset-provider"}
```

如果 `/health` 正常，但 `/buckets` 返回 401，这是正确行为，说明账号门禁已生效。

## 启动 Studio

新开终端，执行：

```bash
cd src/studio
flutter run -d web-server --web-port 8090 --dart-define=PROVIDER_BASE_URL=http://127.0.0.1:9000
```

浏览器打开 Flutter 输出的本地地址，通常是：

```text
http://127.0.0.1:8090/
```

未登录时会看到登录页。使用本地 `AUTH_MODE=local` 配置的邮箱和密码登录。登录成功后应能看到桶列表、分类、搜索和对象浏览入口。

## 运行 CLI

CLI 位于 `src/cli`。常用命令：

```bash
cd src/cli
cargo run -- --help
cargo test
```

如需验证 OSS 只读接口，可根据 CLI 帮助使用 `oss` 子命令。认证后 Provider 默认会对未登录请求返回 401，对权限不足返回 403，CLI 提示也应按这个语义理解。

## 常用验证命令

改代码或准备提交前，至少按变更范围运行对应检查。

Provider：

```bash
cd src/provider
go test ./...
go vet ./...
```

Studio：

```bash
cd src/studio
flutter analyze
flutter test --coverage
```

CLI：

```bash
cd src/cli
cargo test
cargo clippy --all-targets -- -D warnings
```

全仓库发布文档状态：

```bash
qtcloud-devops plan status
qtcloud-devops release status
git diff --check
```

当前 `qtcloud-devops plan audit` 会自动尝试修复 Markdown，遇到 ROADMAP 或 TODO 被改成说明文字、代码块时，不要直接提交，先人工复核。

## 生产边界

以下事项属于生产或平台边界，不能在普通本地开发中顺手执行：

- `git push`、创建 PR、打 tag 或发布 GitHub Release。
- 部署 Provider 或 Studio 到生产。
- 修改 API 网关、DNS、CDN、证书、RAM 权限或函数计算环境变量。
- 删除 OSS 对象、清空目录、下线旧域名或移动 Provider 制品。
- 修改 `.env`、credentials、secrets、AccessKey、Token。

这些操作必须先列清单并取得明确授权。

## 常见问题

### 登录页提示服务尚未配置

默认 `AUTH_MODE=sso` 时，`GET /auth/login` 是 SSO 占位入口，未接入平台 SSO 会返回 503。内测请使用 `AUTH_MODE=local` 并配置 `LOCAL_AUTH_EMAIL` 与 `LOCAL_AUTH_PASSWORD_HASH`。

### 桶列表返回 401

这是账号门禁的正确结果。先登录，再访问 `/buckets`。

### viewer 看不到私密桶

这是权限设计。`viewer` 隐藏 `-private` 桶和 `quanttide-terraform-state`，访问私密桶对象能力应返回 403。

### 不确定能不能改旧域名

不要顺手改。`asset.quanttide.com` 当前仍是兼容入口，是否保留、跳转或下线需要单独授权。

## 下一步阅读

- [配置参考](config.md)
- [快速开始](quickstart.md)
- [质量保证记录](../qa/README.md)
- [Provider README](../../src/provider/README.md)
- [发布记录](../../CHANGELOG.md)
