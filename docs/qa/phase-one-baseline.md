# 阶段一基线与边界记录

> 核对日期：2026-08-23
> 状态：阶段一已完成；应用侧基线通过，新正式域名平台接入仍阻断阶段二
> 依据：`plan.md`、`.quanttide/` 契约和 `quanttide-platform/manifests/terraform/`

## 结论

本项目的应用侧部署边界已经确认，但新正式域名尚未完成平台接入：

- 迁移期兼容入口为 `asset.quanttide.com`，当前仍在线并指向 `qtcloud-asset`。
- 目标正式域名为 `asset.cloud.quanttide.com`，截至 2026-08-23 尚无 DNS/CDN/HTTPS 接入。
- Provider 通过平台共享 API 网关 `api.quanttide.com` 暴露，资产服务路径前缀为 `/qtcloud-asset`。
- Provider 运行时采用 Go Custom Runtime。
- Studio 生产构建通过 `--dart-define=PROVIDER_BASE_URL=...` 注入生产 API 地址，本地开发默认仍为 `http://127.0.0.1:9000`。
- 当前版本只读展示 OSS 桶和对象元数据；公开桶可生成直链，私密桶不生成访问链接。
- Docker Compose、Dockerfile 和 Kubernetes 镜像不属于当前生产发布路径。

## 平台资源边界

`quanttide-platform` 负责系统级共享资源和 API 网关分组、域名绑定。其 Terraform 输出已提供本项目计划中需要引用的共享资源：

| 输出 | 用途 |
|------|------|
| `vpc_id` | Provider 网络边界 |
| `vswitch_id` | Provider 交换机 |
| `security_group_id` | 函数计算安全组 |
| `rds_instance_id` | 共享 RDS 实例标识 |
| `rds_connection_string` | 共享 RDS 连接地址 |
| `rds_port` | 共享 RDS 端口 |
| `resource_group_id` | 应用资源组归属 |

本项目不重新定义上述系统级资源，也不修改平台仓库。平台 API 网关的路由、DNS 记录和证书绑定仍按平台仓库的脚本与 CI 流程处理。

## 当前门禁

| 门禁 | 状态 | 证据 |
|------|------|------|
| Provider 使用 Go Custom Runtime | 已验证 | Go 入口为 `src/provider/cmd/provider/main.go`；FC 函数已使用 `custom.debian12` + `./bootstrap` 发布 |
| Studio 生产地址不使用本机地址 | 通过 | `src/studio/lib/main.dart` 支持 `PROVIDER_BASE_URL` 构建参数 |
| 阶段一生产变更 | 未执行 | 本阶段仅做只读核对；没有修改 OSS、DNS、CDN、证书、权限或网关 |

## 已完成的基线修复

- Provider 默认 `PROVIDER_BASE_URL` 已统一为 `https://api.quanttide.com/qtcloud-asset`。
- Studio 已支持本地默认地址和生产构建地址。
- Provider README 已清理冲突标记，并将 Go Custom Runtime 定为当前发布方式。
- Provider 当前返回 47 个桶；契约文件显式登记 35 个桶名，存在 12 个线上新增桶待资产契约收敛。
- 根 README 已明确 Docker 仅为历史本地参考。

## 已完成的线上验证

- `https://asset.quanttide.com/` 返回 Flutter Web `index.html`，HTTP/HTTPS 均为 200；DNS CNAME 为 `asset.quanttide.com.w.kunlunaq.com`。
- `qtcloud-asset-studio` 位于 `oss-cn-hangzhou`，ACL 为 `public-read`，静态网站首页和错误页均为 `index.html`，对象数为 49。
- `qtcloud-asset` 位于 `oss-cn-hangzhou`，ACL 为 `public-read`，对象数为 54，含前端静态资源和 5 个 `provider/` 发布包。
- `asset.quanttide.com` CDN 状态为 `online`，源站为 `qtcloud-asset.oss-cn-hangzhou.aliyuncs.com`，HTTPS 已开启。
- `asset.cloud.quanttide.com` 当前无 DNS 记录、无 CDN 加速域名，HTTPS 访问失败。
- Provider `https://api.quanttide.com/qtcloud-asset/health` 和 `/buckets` 返回 200。
- `https://api.quanttide.com/qtcloud-asset/health`、`/config`、`/buckets`、对象列表和公开对象 URL 均返回 200。
- Provider 已挂载 `qtcloud-asset-provider-role`，函数环境不再保存长期 OSS AccessKey。
- `-private` 桶对象 URL 在 Provider 中被拒绝，Studio 不显示复制链接入口。

## 待处理问题

1. `asset.cloud.quanttide.com` 需要平台侧完成 CDN、DNS CNAME 和 HTTPS 证书接入，这是阶段四的前置阻断项；阶段二只完成仓库配置收敛。
2. 正式网关当前返回 `Access-Control-Allow-Origin: *`；代码默认已登记新旧 Studio 来源，但线上网关 CORS 仍需平台侧配置收紧。
3. `src/provider/requirements.txt`、Provider Dockerfile 和 Docker 清单仍保留旧运行方式引用，应在部署配置收敛时标注为废弃参考或移出当前发布流程。
4. 线上 47 个桶与契约显式登记的 35 个桶不一致，需在后续资产契约收敛中补齐或说明。

## 验证记录

已验证：

```bash
flutter test --dart-define=PROVIDER_BASE_URL=https://api.quanttide.com/qtcloud-asset test/provider_config_test.dart
go test ./...
go vet ./...
```

上述 Go 检查均通过。完整 Flutter、Go 和 CLI 质量门禁已在本阶段执行并通过；CLI 的 `clippy -D warnings` 仍有既有警告未纳入本次线上变更。
