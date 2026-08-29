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
| `/auth/login` | GET | 发起 SSO 登录；真实飞书/Lark SSO 待平台接入 |
| `/auth/login` | POST | 本地账号密码登录；支持本地管理员账号和带初始密码的受管用户 |
| `/auth/callback` | GET | 登录回调，校验 state 后创建服务端会话 |
| `/auth/me` | GET | 返回当前登录用户 |
| `/auth/logout` | POST | 撤销当前会话 |
| `/admin/users` | GET | `admin` 查看用户列表、角色、状态和最后登录时间 |
| `/admin/users` | POST | `admin` 邀请用户，写入账号、姓名、角色和初始密码哈希 |
| `/admin/users/{id}/role` | PATCH | `admin` 修改用户角色 |
| `/admin/users/{id}/disable` | POST | `admin` 禁用用户并撤销其会话 |
| `/admin/users/{id}/sessions/revoke` | POST | `admin` 撤销指定用户所有会话 |
| `/buckets` | GET | 登录后列出 OSS 桶；`viewer` 隐藏 `-private` 和 `quanttide-terraform-state`，`admin` 查看全部 |
| `/buckets/{name}/objects` | GET | 登录后列出桶内对象（只读） |
| `/buckets/{name}/object-url?key=…&expires=…` | GET | 登录后生成公开桶直链；私密桶和 `quanttide-terraform-state` 禁止生成对象链接 |
| `/shares` | POST | 登录后为公开桶的一个或多个文件和/或文件夹创建只读分享页 |
| `/shares/{token}` | GET | 公开读取分享页元数据；无须登录 |
| `/shares/{token}/objects?prefix=…` | GET | 公开浏览分享页授权前缀下的对象；无须登录 |
| `/shares/{token}/object-url?key=…` | GET | 公开生成分享页授权文件的直链；无须登录 |
| `/shares/{token}/download` | GET | 公开将分享范围内的文件打包为 ZIP 下载；无须登录 |
| `/shares/{token}` | DELETE | 创建者或管理员撤销分享；需要登录 |

认证状态使用服务端会话，浏览器只保存 `HttpOnly` Cookie。当前默认 SSO 身份源是未配置占位实现，未接入平台 SSO 时 `GET /auth/login` 会返回 503，避免误开放登录入口。内测可启用 `AUTH_MODE=local` 使用单管理员账号密码登录；管理员在用户管理页邀请新用户时必须填写初始密码，Provider 只保存 PBKDF2-SHA256 哈希值，用户列表和接口响应不回显密码材料。

账号记录使用平台共享 RDS 的 `users` 表，默认 `USER_STORE=rds`；只有本地开发或测试才应显式设置 `USER_STORE=memory`，否则 Provider 重启后用户管理记录会丢失。生产 PostgreSQL 用户迁移脚本位于 `internal/storage/auth_users_postgres.sql`，分享记录使用同一个 RDS 的 `folder_shares` 表，分享迁移脚本位于 `internal/storage/folder_shares_postgres.sql`。Provider 正常启动时只连接并检查 RDS，不会自动执行迁移；仅在受控上线时分别显式设置一次 `USER_MIGRATION=users-postgres-v1` 和 `SHARE_MIGRATION=folder-shares-postgres-v1`，才会执行对应的固定幂等 PostgreSQL DDL，成功后应移除这两个一次性变量。Provider 默认同时写入内存审计存储和结构化 stdout JSON，事件名为 `qtcloud_asset_audit`；生产函数计算通过 FC `logConfig` 将 stdout 持久化到 SLS。历史 MySQL 认证草案位于 `internal/storage/auth_audit_schema.sql`，不作为当前生产迁移入口。

业务权限按 `viewer` 和 `admin` 两级执行。`viewer` 的桶清单隐藏 `-private` 桶和 `quanttide-terraform-state`，只能访问公开桶对象元数据和公开链接；`admin` 可以查看全部桶、访问私密桶对象元数据，但任何角色都不能为私密桶生成对象链接。公开桶直链返回 `expires_in=0`。未登录返回 401，已登录但权限不足或账号被禁用返回 403。管理写接口会校验浏览器 `Origin`，未登记来源返回 403；无 `Origin` 的服务端/CLI 调用保留给后续 CLI 接入。

分享只允许白名单内、且 OSS 当前 ACL 为 `public-read` 的桶，单次分享固定一个桶，可包含多个以 `/` 结尾的对象前缀和明确文件 key。`POST /shares` 使用 `prefixes` 和 `keys` 表示分享目标，最多 128 个目标；公开页面只读，不提供上传、删除或权限变更。文件对象链接仍由 Provider 按分享 token、前缀或明确 key 范围和当前桶 ACL 重新校验。Provider 提供 `/shares/{token}/download` ZIP 接口，保留原始对象路径和文件名（包括隐藏文件）；Studio 分享页为规避网关大响应超时，当前优先逐个读取分享对象并在浏览器端生成 ZIP，单次浏览器打包总大小最多 100 MiB。正式入口默认使用 RDS；仅显式设置 `SHARE_STORE=memory` 时才启用不持久的内存实现，用于本地开发和测试。

## 配置

| 环境变量 | 默认值 | 说明 |
|---------|--------|------|
| `PROVIDER_PORT` | `9000` | 监听端口 |
| `PROVIDER_BASE_URL` | `https://api.quanttide.com/qtcloud-asset` | 服务基础 URL |
| `STUDIO_ORIGIN` | `https://asset.cloud.quanttide.com` | 正式 Studio 来源；默认 CORS 白名单同时保留 `https://asset.quanttide.com` |
| `AUTH_MODE` | `sso` | 认证模式；内测账号密码登录设为 `local` |
| `LOCAL_AUTH_ACCOUNT` | （空） | 本地登录账号 |
| `LOCAL_AUTH_EMAIL` | （空） | 兼容旧配置的邮箱字段，可为空 |
| `LOCAL_AUTH_NAME` | （空） | 本地账号展示名称；为空时使用账号 |
| `LOCAL_AUTH_ROLE` | `admin` | 本地账号角色，仅支持 `viewer` 或 `admin` |
| `LOCAL_AUTH_PASSWORD_HASH` | （空） | 本地账号密码哈希，格式 `pbkdf2_sha256$iterations$salt$hash` |
| `OSS_ENDPOINT` | `https://oss-cn-hangzhou.aliyuncs.com` | OSS 服务端点 |
| `OSS_ACCESS_KEY_ID` | （空） | 本地开发用 AccessKey ID；缺失时回退到 `ALIBABA_CLOUD_ACCESS_KEY_ID` |
| `OSS_ACCESS_KEY_SECRET` | （空） | 本地开发用 AccessKey Secret；缺失时回退到 `ALIBABA_CLOUD_ACCESS_KEY_SECRET` |
| `OSS_SESSION_TOKEN` | （空） | 本地开发用临时凭证令牌 |
| `ALIBABA_CLOUD_SECURITY_TOKEN` | （空） | 函数计算 RAM 角色注入的临时凭证令牌 |
| `RDS_DRIVER` | `postgres` | RDS 使用的 `database/sql` 驱动；支持 `mysql`、`postgres`、`postgresql`、`pgx`，生产共享 RDS 使用 `postgres` |
| `RDS_CONNECTION_STRING` | （空） | 平台共享 RDS 连接串；生产分享功能必填 |
| `USER_STORE` | `rds` | 用户管理存储模式；生产使用 `rds`，本地测试可显式使用 `memory` |
| `USER_MIGRATION` | （空） | 受控一次性用户表迁移标识；仅支持 `users-postgres-v1`，成功后必须移除 |
| `SHARE_STORE` | `rds` | 分享存储模式；生产使用 `rds`，本地测试可显式使用 `memory` |
| `SHARE_MIGRATION` | （空） | 受控一次性迁移标识；仅支持 `folder-shares-postgres-v1`，成功后必须移除 |
| `SHARE_TOKEN_ENCRYPTION_KEY` | （空） | 32 字节 AES-256 密钥的 base64 值；生产分享功能必填 |
| `SHAREABLE_BUCKETS` | `qtcloud-asset-studio` | 允许创建分享的 OSS 桶名逗号分隔白名单 |

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
