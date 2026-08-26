# 阶段六清理和回滚准备记录

> 验证日期：2026-08-27
>
> 状态：已完成
> 正式入口：`https://asset.cloud.quanttide.com`

## 新域名正式化

- 现实需求：资产云正式入口固定到 `asset.cloud.quanttide.com`，并保留可审计的 CDN、DNS、证书和静态对象证据。
- 判定标准：DNS 指向平台 CDN CNAME；CDN 状态为 `online`；源站为 `qtcloud-asset-studio.oss-cn-hangzhou.aliyuncs.com`；HTTPS 证书覆盖 `asset.cloud.quanttide.com`；首页和关键静态对象返回 200。
- 存证：`Resolve-DnsName asset.cloud.quanttide.com` 返回 CNAME `asset.cloud.quanttide.com.w.kunlunaq.com`，A 记录 `111.62.97.181`，AAAA 记录 `2409:8c04:1005:1112:3::f`。`aliyun cdn DescribeCdnDomainDetail --DomainName asset.cloud.quanttide.com` 返回 `DomainStatus=online`，`CdnType=web`，`Scope=domestic`，源站类型 `oss`，源站 `qtcloud-asset-studio.oss-cn-hangzhou.aliyuncs.com`，端口 `443`。`aliyun cdn DescribeDomainCertificateInfo --DomainName asset.cloud.quanttide.com` 返回证书域名 `quanttide.com,*.cloud.quanttide.com,*.quanttide.com`，状态 `on`，到期时间 `2026-11-22T23:59:59Z`。
- 状态：✅ 符合

## 静态对象回滚证据

- 现实需求：如果新版 Studio 需要回滚，可以按对象 ETag 精确核对当前线上版本。
- 判定标准：记录首页、主 JS、manifest、bootstrap、构建标识和关键资源对象的 ETag、更新时间和对象数量。
- 存证：`aliyun oss ls oss://qtcloud-asset-studio/` 返回对象数 `49`。关键对象为 `index.html` ETag `F79DEB9E8A3CB1B40066B0B881C415BC`，`main.dart.js` ETag `3CB211BCEC3A1AF0E2910383AD6BF2D0`，`manifest.json` ETag `80304D1019333F6CBB9033D1D9681C62`，`flutter_bootstrap.js` ETag `82AE86A88BC18D072216EB53ABFA1C82`，`assets/AssetManifest.bin.json` ETag `69A99F98C8B1FB8111C5FB961769FCD8`，`.last_build_id` ETag `2D49E5CDFFA18DFDDACA8581DF18885A`，`version.json` ETag `BB9229FB947481675794D25D8F296544`。
- 状态：✅ 符合

## API 和 CORS 回滚证据

- 现实需求：新域名页面可以稳定访问 Provider，并且跨域策略按正式入口精确放行。
- 判定标准：Provider `/health` 返回 200；新域名来源的认证预检返回 200；未登录 `/auth/me` 返回 401 且回显正式入口来源。
- 存证：`curl -s https://api.quanttide.com/qtcloud-asset/health` 返回 `{"status":"ok","service":"qtcloud-asset-provider"}`。`OPTIONS /auth/login` 携带 `Origin: https://asset.cloud.quanttide.com` 返回 200，并返回 `Access-Control-Allow-Origin: https://asset.cloud.quanttide.com`、`Access-Control-Allow-Credentials: true`、`Access-Control-Allow-Methods: GET,POST,PATCH,OPTIONS`。`GET /auth/me` 携带同源返回 401 `authentication required`，并回显同一 CORS 来源。
- 状态：✅ 符合

## 旧域名处理策略

- 现状：用户明确要求先不处理旧域名，因此本轮不删除、不停用、不修改 `asset.quanttide.com` 相关 DNS、CDN 或 CORS 配置。
- 影响：旧入口仍作为兼容入口存在；阶段六的新域名正式化和回滚证据可以先闭合，旧入口退役另起授权流程。
- 后续：待用户重新确认旧域名退役时，再单独列出 DNS、CDN、CORS 和文档更新清单并执行。

## qtcloud-asset 旧桶清单

- 现实需求：`qtcloud-asset` 旧桶曾混放前端静态对象和 Provider 发布包，阶段六需要把对象边界留痕，避免后续清理时误删可回滚基线或发布制品。
- 判定标准：旧前端静态对象、Provider 发布包数量、大小和关键 ETag 已保存；本轮不删除、不移动、不覆盖任何对象。
- 存证：`aliyun oss ls oss://qtcloud-asset/ -a` 返回对象总数 `62`。其中旧前端静态对象 `49` 个，合计 `42,410,991` bytes；Provider 发布包 `13` 个，合计 `77,097,787` bytes。完整清单见 [阶段六 qtcloud-asset 对象清单](./phase-six-qtcloud-asset-inventory.md)。
- 归属判断：`qtcloud-asset` 不再作为正式 Studio 发布桶；旧前端静态对象只作为迁移期回滚证据保留。`provider/` 短期继续作为 Provider 发布包制品存放位置，后续如迁到专用 Provider 桶或制品仓库，需要另起授权流程。
- 状态：✅ 符合

## 待收敛项

1. 决定是否把 `tmp-playwright/phase5-ui.spec.js` 固化为正式线上验收脚本。
2. 平台 SSO/真实身份源接入仍作为后续增强，不阻断当前新域名正式化。
3. 如未来要迁移或清理 `qtcloud-asset/provider/` 制品，需要单独列出 OSS、权限和回滚操作清单后执行。
