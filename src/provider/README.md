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
| `/buckets` | GET | 列出 OSS 桶（只读） |
| `/buckets/{name}/objects` | GET | 列出桶内对象（只读） |
| `/buckets/{name}/object-url?key=…&expires=…` | GET | 生成公开桶直链；私密桶不生成访问链接 |

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
