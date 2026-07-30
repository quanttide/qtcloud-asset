# QtCloud Asset Provider

量潮数字资产云服务端，基于 Go 构建，提供数字资产治理的 HTTP API。

## 架构

```
┌─────────────────────────────────┐
│         API Layer              │
│  /health  /  /config           │
├─────────────────────────────────┤
│       Service Layer            │  ← 待实现
├─────────────────────────────────┤
│     Repository Layer           │  ← 待实现
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

## 配置

| 环境变量 | 默认值 | 说明 |
|---------|--------|------|
| `PROVIDER_PORT` | `9000` | 监听端口 |
| `PROVIDER_BASE_URL` | `https://api.asset.quanttide.com` | 服务基础 URL |

## 开发

```bash
# 运行测试
go test ./...

# 代码检查
go vet ./...
```

## 部署

Provider 部署为 Docker 容器：

```bash
docker build -t qtcloud-asset-provider .
docker tag qtcloud-asset-provider registry.example.com/qtcloud-asset-provider:latest
docker push registry.example.com/qtcloud-asset-provider:latest
```
