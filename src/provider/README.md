# QtCloud Asset Provider

量潮数字资产云服务端，基于 Go 构建，提供数字资产治理的 HTTP API。

## 架构

```
┌─────────────────────────────────┐
│         API Layer              │
│  /health  /  /config           │
│  /buckets  /buckets/{name}/... │
├─────────────────────────────────┤
│       Service Layer            │
│  BucketService（只读发现）       │
├─────────────────────────────────┤
│     Repository Layer           │
│  OssAdapter（SourceAdapter 首个实现）│
├─────────────────────────────────┤
│        Schemas                 │
└─────────────────────────────────┘
```

## 快速开始

```bash
cd src/provider

# 直接运行
go run ./cmd/provider

# 构建二进制
go build -o bin/provider ./cmd/provider
./bin/provider

# 构建 Docker
docker build -t qtcloud-asset-provider .
docker run -p 9000:9000 qtcloud-asset-provider
```

## API

| 端点 | 方法 | 说明 |
|------|------|------|
| `/health` | GET | 健康检查 |
| `/` | GET | 服务信息 |
| `/config` | GET | 服务配置 |
| `/auth/login` | GET | 发起登录；真实飞书/Lark SSO 待平台接入 |
| `/auth/callback` | GET | 登录回调，校验 state 后创建服务端会话 |
| `/auth/me` | GET | 返回当前登录用户 |
| `/auth/logout` | POST | 撤销当前会话 |
| `/admin/users` | GET | `admin` 查看用户列表、角色、状态和最后登录时间 |
| `/admin/users` | POST | `admin` 预授权邀请用户，写入邮箱、姓名和角色 |
| `/admin/users/{id}/role` | PATCH | `admin` 修改用户角色 |
| `/admin/users/{id}/disable` | POST | `admin` 禁用用户并撤销其会话 |
| `/admin/users/{id}/sessions/revoke` | POST | `admin` 撤销指定用户所有会话 |
| `/buckets` | GET | 登录后列出 OSS 桶；`viewer` 隐藏 `-private` 和 `quanttide-terraform-state`，`admin` 查看全部 |
| `/buckets/{name}/objects` | GET | 登录后列出桶内对象（只读） |
| `/buckets/{name}/object-url?key=…&expires=…` | GET | 登录后生成公开桶直链；`admin` 可为私密桶生成限时签名 URL |

认证状态使用服务端会话，浏览器只保存 `HttpOnly` Cookie。当前默认身份源是未配置占位实现，未接入平台 SSO 时 `/auth/login` 会返回 503，避免误开放登录入口。

账号、会话和审计的仓库侧模型位于 `internal/auth`。RDS 表结构草案位于 `internal/storage/auth_audit_schema.sql`，当前仅作为平台共享 RDS 接入前的审阅稿，不会在 Provider 启动时自动执行。

业务权限按 `viewer` 和 `admin` 两级执行。`viewer` 的桶清单隐藏 `-private` 桶和 `quanttide-terraform-state`，只能访问公开桶对象元数据和公开链接；`admin` 可以查看全部桶、访问私密桶对象元数据、生成限时签名 URL，并通过后端管理接口邀请、禁用、改角色和撤销会话。私密桶签名 URL 默认有效期为 86400 秒，最大 604800 秒；公开桶直链返回 `expires_in=0`。未登录返回 401，已登录但权限不足或账号被禁用返回 403。管理写接口会校验浏览器 `Origin`，未登记来源返回 403；无 `Origin` 的服务端/CLI 调用保留给后续 CLI 接入。

## 配置

| 环境变量 | 默认值 | 说明 |
|---------|--------|------|
| `PROVIDER_PORT` | `9000` | 监听端口 |
| `PROVIDER_BASE_URL` | `https://api.quanttide.com/qtcloud-asset` | 服务基础 URL |
| `STUDIO_ORIGIN` | `https://asset.cloud.quanttide.com` | 正式 Studio 来源；默认 CORS 白名单同时保留 `https://asset.quanttide.com` |
| `OSS_ENDPOINT` | `https://oss-cn-hangzhou.aliyuncs.com` | OSS 服务端点 |
| `OSS_ACCESS_KEY_ID` | （空） | 本地开发用 AccessKey ID；缺失时回退到 `ALIBABA_CLOUD_ACCESS_KEY_ID` |
| `OSS_ACCESS_KEY_SECRET` | （空） | 本地开发用 AccessKey Secret；缺失时回退到 `ALIBABA_CLOUD_ACCESS_KEY_SECRET` |
| `OSS_SESSION_TOKEN` | （空） | 本地开发用临时凭证令牌 |
| `ALIBABA_CLOUD_SECURITY_TOKEN` | （空） | 函数计算 RAM 角色注入的临时凭证令牌 |

## 开发

```bash
# 运行测试
go test ./...

# 代码检查
go vet ./...
```

## 发布边界

当前发布方式是阿里云函数计算 Go Custom Runtime。Docker 文件仅保留为历史本地参考，不属于当前生产部署路径。

```bash
# 目标运行时由 manifests/iac/modules/fc 配置
GOOS=linux GOARCH=amd64 go build -o bootstrap ./cmd/provider
```
