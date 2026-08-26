# 阶段四域名与 CDN 接入记录

> 验证日期：2026-08-24
> 状态：阶段四已完成
> 正式入口：`https://asset.cloud.quanttide.com`

## 接入结果

- `asset.cloud.quanttide.com` 已完成 CDN、DNS 和 HTTPS 接入。
- CDN 源站为 `qtcloud-asset-studio.oss-cn-hangzhou.aliyuncs.com`。
- HTTPS 证书覆盖 `asset.cloud.quanttide.com` 和 `*.cloud.quanttide.com`。
- 兼容入口 `https://asset.quanttide.com` 保持不变，未删除旧 DNS 记录。

## 验证证据

```bash
curl -I -L https://asset.cloud.quanttide.com
curl -i -H "Origin: https://asset.cloud.quanttide.com" \
  https://api.quanttide.com/qtcloud-asset/health
```

- 正式入口返回 HTTP 200，响应包含 CDN 缓存链路标识。
- 返回的 `ETag` 为 `F79DEB9E8A3CB1B40066B0B881C415BC`，与阶段三发布的 `index.html` 一致。
- Provider `/health` 返回 HTTP 200 和 `{"status":"ok","service":"qtcloud-asset-provider"}`。
- API 网关返回 `Access-Control-Allow-Origin: *`；当前 Studio 仅使用无凭据的 GET 请求，因此可满足当前跨域调用。若后续引入 Cookie、浏览器凭据或认证会话，应改为精确来源白名单并返回 `Access-Control-Allow-Credentials`。

## 后续事项

1. 执行阶段五线上验收，覆盖页面加载、桶列表、对象列表、排序、私密桶限制及浏览器控制台。
2. 旧的 `asset.quanttide.com` OSS 直连 DNS 记录是否删除，需在兼容入口保留策略确认后单独审批和执行。
