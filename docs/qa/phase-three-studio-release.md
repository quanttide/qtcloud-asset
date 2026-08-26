# 阶段三 Studio 发布记录

> 发布日期：2026-08-24
> 状态：阶段三已完成
> 目标桶：`oss://qtcloud-asset-studio/`

## 发布结果

使用生产 API 地址构建并上传 Flutter Web：

```bash
flutter build web --release \
  --dart-define=PROVIDER_BASE_URL=https://api.quanttide.com/qtcloud-asset \
  --dart-define=STUDIO_ORIGIN=https://asset.cloud.quanttide.com
```

上传结果：

- 39 个文件
- 10 个目录对象
- `42,410,990` bytes
- 远端对象总数为 49
- 远端更新时间为 `2026-08-24 12:31:42 +0800`
- OSS 静态网站首页和错误页均配置为 `index.html`
- `https://qtcloud-asset-studio.oss-cn-hangzhou.aliyuncs.com/index.html` 返回 HTTP 200

## 关键对象

| 对象 | ETag | 本地 MD5 |
|:-----|:-----|:---------|
| `index.html` | `F79DEB9E8A3CB1B40066B0B881C415BC` | `F79DEB9E8A3CB1B40066B0B881C415BC` |
| `main.dart.js` | `2978DFAC153D54D61F30826518B26CEB` | `2978DFAC153D54D61F30826518B26CEB` |
| `manifest.json` | `80304D1019333F6CBB9033D1D9681C62` | `80304D1019333F6CBB9033D1D9681C62` |

## 发布门禁

- 构建产物包含生产 API 地址 `https://api.quanttide.com/qtcloud-asset`。
- 未发现 `127.0.0.1`、内部函数地址或凭证标记。
- 未发现 `.zip`、`.log`、`.pem` 或 `.key` 文件。
- Studio 桶中未发现 Provider 发布包。
- 未修改 DNS、CDN、HTTPS 证书或 API 网关。

## 后续刷新

- 2026-08-24 线上正式域名已再次刷新 `main.dart.js`，当前 CDN 访问到的 ETag 为 `E144C4197AF464B2D85599A0B3CEB8DD`。
- 该刷新属于后续修复发布，不改变本阶段初始发布记录中的 `index.html`、`manifest.json` 与发布门禁结论。
