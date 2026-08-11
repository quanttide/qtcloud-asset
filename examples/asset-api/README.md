# asset API 参考实现（自 qtadmin provider 迁移）

## 资源与路由
- /api/v1/qtcloud/resources（CRUD）

## 结构
model（QtCloudResource）+ api（ResourceHandler）+ store（filestore）

## 共享基础设施
store/response 已随附复制；config/version/cmd（服务装配）在 qtadmin git 历史 `src/provider/` 中可恢复。

## 验证
go build ./... && go test ./...
