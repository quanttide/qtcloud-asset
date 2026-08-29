# 量潮资产云站点规范化迁移计划

## 背景

量潮云产品矩阵中的前端控制台域名通常使用 `{产品}.cloud.quanttide.com`，静态站点桶通常使用 `*-studio` 或 `*-site` 后缀。量潮资产云当前线上入口为 `asset.quanttide.com`，静态前端和 Provider 发布包主要混放在 `qtcloud-asset` 桶中。

当前仓库和线上资源已经具备迁移基础：

- `asset.quanttide.com` 当前可用，作为现有兼容入口。
- `qtcloud-asset-studio` 桶已存在，里面已有一份旧的 Flutter Web 静态构建。
- `qtcloud-asset` 桶当前同时包含前端静态资源和 `provider/` 下的后端发布包。
- Provider 已通过 `https://api.quanttide.com/qtcloud-asset` 对外提供 API。

本计划用于把资产云收敛到平台统一命名和发布结构。

## 目标

正式入口调整为：

```text
asset.cloud.quanttide.com
  -> CDN
  -> qtcloud-asset-studio
  -> https://api.quanttide.com/qtcloud-asset
  -> qtcloud-asset-provider
```

兼容入口保留为：

```text
asset.quanttide.com
```

兼容入口在迁移期内不下线。它可以继续指向新版前端，也可以跳转到 `asset.cloud.quanttide.com`，具体方式由平台域名和 CDN 边界确认后决定。

## 命名决策

Studio 桶使用 `qtcloud-asset-studio`。

原因：

- 当前产品是管理控制台，不是公开官网。
- 控制台需要调用 Provider API，符合 `*-studio` 语义。
- `qtcloud-asset-studio` 已存在，不需要新建同名资源。

暂不创建 `qtcloud-asset-site`。

原因：

- 当前没有资产云官网、产品介绍页或公开落地页需求。
- 先补 `site` 会增加空桶、DNS、CDN 和发布链路，容易制造混乱。
- 未来如果要做公开官网，再单独规划 `qtcloud-asset-site`。

Provider 发布包暂不放入 Studio 桶。

短期可以继续保留在 `qtcloud-asset/provider/`，但需要记录为待收敛项。中长期应由平台确认是否迁到专用 Provider 发布桶或制品仓库，避免和公开静态站点资源混放。

## 排序交互决策

桶列表的「创建时间」和 `A→Z / Z→A` 是两个独立排序开关，不应做成互斥按钮。

需要支持四种状态：

1. 两个排序都不开启，保留 Provider 返回顺序。
2. 只开启创建时间排序，按新到旧或旧到新排列。
3. 只开启桶名排序，按 `A→Z` 或 `Z→A` 排列。
4. 同时开启两个排序，以创建时间作为主排序，桶名作为同一创建时间下的次级排序。

任一排序条件变化后，分页回到第一页。界面需要清楚展示每个排序开关是否启用，以及当前方向。

## 阶段一 基线确认

> **状态：已完成（2026-08-23）**  
> 结论：应用侧基线核对完成；`asset.cloud.quanttide.com` 尚未完成平台侧 DNS、CDN 和证书接入，阶段二不得将其视为已可发布的正式入口。

目标是确认资源现实状态，不做生产变更。

工作内容：

1. 确认 `qtcloud-asset-studio` 桶存在、区域、ACL、静态网站配置和对象列表。
2. 确认 `qtcloud-asset` 桶中前端资源和 `provider/` 发布包的现状。
3. 确认 `asset.quanttide.com` 当前 CDN、源站、证书和 DNS 指向。
4. 向平台仓库核对 `asset.cloud.quanttide.com` 是否已被占用。
5. 核对平台手册中 `*.cloud.quanttide.com`、`*-studio`、`*-site` 的职责边界。
6. 记录当前可回滚入口、对象 ETag 和 CDN 刷新任务号。

阶段门禁：

- [阻断] 确认 `asset.cloud.quanttide.com` 可以作为资产云正式域名：未通过。2026-08-23 查询不到 DNS 记录，CDN 中也不存在该加速域名，HTTPS 访问失败。
- [通过] 确认本项目只管理应用级 Studio 发布，不接管平台级共享资源。平台仓库负责 API 网关分组、`api.quanttide.com` 域名绑定、DNS 和证书 CI；本项目只消费其公网入口。
- [通过] 确认 `qtcloud-asset-studio` 可作为新版 Studio 发布目标。桶位于 `oss-cn-hangzhou`，ACL 为 `public-read`，静态网站首页和错误页均为 `index.html`，已有 49 个 Flutter Web 对象。

阶段一核对记录：

- `qtcloud-asset-studio`：存在于 `oss-cn-hangzhou`，对象数 49，关键对象包括 `index.html`、`main.dart.js`、`manifest.json`；`index.html` ETag 为 `F6CDF9CF6945D364D3034592E677652C`。
- `qtcloud-asset`：存在于 `oss-cn-hangzhou`，对象数 54，同时包含 Flutter Web 静态资源和 `provider/` 下 5 个 Provider 发布包；`index.html` ETag 为 `F79DEB9E8A3CB1B40066B0B881C415BC`。
- `asset.quanttide.com`：DNS CNAME 为 `asset.quanttide.com.w.kunlunaq.com`；阿里云 CDN 状态为 `online`，源站为 `qtcloud-asset.oss-cn-hangzhou.aliyuncs.com`，HTTPS 已开启；访问首页返回 200。
- `asset.cloud.quanttide.com`：DNS 查询结果为不存在，CDN 查询无记录，HTTPS 访问失败；未发现可记录的 CDN 刷新任务号。
- Provider：`https://api.quanttide.com/qtcloud-asset/health` 和 `/buckets` 均返回 200；当前网关 CORS 返回 `Access-Control-Allow-Origin: *`，尚未按新正式域名收紧。
- 资产清单：Provider 当前返回 47 个 OSS 桶；`.quanttide/asset/contract.yaml` 显式登记 35 个桶名，线上多出的 12 个桶为 `qtclass-site`、`qtclass-video`、`qtcloud-agent-site`、`qtcloud-agent-studio`、`qtcloud-asset`、`qtcloud-asset-studio`、`qtcloud-course-admin`、`qtcloud-execute-provider`、`qtcloud-execute-site`、`qtcrowd-site`、`qtfiction-site`、`qtmedia-site`。

阶段一后续阻断项：

1. 由平台侧确认并接入 `asset.cloud.quanttide.com` 的 CDN、DNS CNAME 和 HTTPS 证书。
2. 在新域名确定后，将 Studio 的 CI、Terraform 默认值、Provider `StudioOrigins` 和相关文档统一切换。
3. 在进入阶段二前，先决定 `asset.quanttide.com` 是继续作为兼容入口还是跳转到新域名。

## 阶段二 仓库配置收敛

> **状态：已完成（2026-08-24）**  
> 结论：仓库侧已切换到 `qtcloud-asset-studio` 和 `asset.cloud.quanttide.com` 目标；不执行生产发布，正式域名平台接入仍留在阶段四。

目标是让代码和 CI 指向新的 Studio 桶与正式域名。

工作内容：

1. 将 Studio 发布目标从 `qtcloud-asset` 改为 `qtcloud-asset-studio`。
2. 保持 Provider API 地址为 `https://api.quanttide.com/qtcloud-asset`。
3. 将文档中的正式 Studio 域名从 `asset.quanttide.com` 改为 `asset.cloud.quanttide.com`。
4. 将 `asset.quanttide.com` 标记为兼容入口。
5. 检查 `.github/workflows/deploy.yml` 是否运行 `flutter test` 和 `flutter analyze`。
6. 检查构建产物中不得出现 `127.0.0.1`、AccessKey、Secret 或内部函数地址。
7. 同步 README、QA 证据和部署说明。

阶段门禁：

- [通过] CI 构建仍注入生产 API 地址 `https://api.quanttide.com/qtcloud-asset`，并执行 `flutter test`、`flutter analyze`。
- [通过] Studio 发布目标只指向 `qtcloud-asset-studio`；Provider 发布包仍写入 `qtcloud-asset/provider/`。
- [通过] 构建流程增加本地地址、凭证标记、内部函数地址和无关发布文件扫描。
- [通过] Provider 默认 CORS 白名单包含 `https://asset.cloud.quanttide.com` 和兼容入口 `https://asset.quanttide.com`；`STUDIO_ORIGIN` 支持环境覆盖。
- [通过] README、Provider 配置说明、Terraform 说明和 QA 记录已同步正式入口与平台边界。

阶段二变更记录：

- `.github/workflows/deploy.yml`：切换 Studio 桶，增加 Flutter 测试、分析和构建产物安全扫描。
- `manifests/iac/`：Studio 默认桶和域名切换；Provider 制品桶独立为 `qtcloud-asset`；移除本项目对 Studio/Provider DNS 记录的管理，避免越过平台边界；Studio 桶启用 `prevent_destroy`。
- `src/provider/`：默认 Studio 来源切换到新域名，保留旧域名和本地开发来源；增加配置和 CORS 测试。

阶段二未执行：

- 未运行 Terraform apply。
- 未上传 Studio 构建产物。
- 未修改 DNS、CDN、证书、API 网关或生产权限。

## 阶段三 Studio 发布

> **状态：已完成（2026-08-24）**  
> 结论：生产 API 配置的 Flutter Web 构建已发布到 `qtcloud-asset-studio`；正式域名仍未切换，阶段四平台接入阻断继续保留。

目标是在不切换正式域名的前提下，把最新前端发布到 `qtcloud-asset-studio`。

工作内容：

1. 使用生产 API 地址构建 Flutter Web。
2. 上传 `build/web` 到 `oss://qtcloud-asset-studio/`。
3. 确认不上传 Provider zip、调试日志、凭证或无关生成物。
4. 校验 `index.html`、`main.dart.js`、`manifest.json` 和资源目录。
5. 检查 OSS 静态网站首页和错误页配置。
6. 记录上传后的关键对象 ETag 和更新时间。

阶段门禁：

- `qtcloud-asset-studio` 中是最新 Studio 前端构建。
- 构建产物请求 `https://api.quanttide.com/qtcloud-asset`。
- 桶内不混入 Provider 发布包。

阶段三发布记录：

- 发布时间：`2026-08-24 12:31:42 +0800`
- 上传结果：39 个文件、10 个目录对象，共 `42,410,990` bytes
- Studio 桶对象总数：49
- `index.html` ETag：`F79DEB9E8A3CB1B40066B0B881C415BC`
- `main.dart.js` ETag：`2978DFAC153D54D61F30826518B26CEB`
- `manifest.json` ETag：`80304D1019333F6CBB9033D1D9681C62`
- OSS 静态网站首页和错误页均为 `index.html`，对象 HTTPS 访问返回 200。
- 构建产物安全扫描通过，未发现本地地址、内部函数地址、凭证标记或无关发布文件。

## 阶段四 域名和 CDN 接入

> **状态：已完成（2026-08-24）**  
> 结论：`asset.cloud.quanttide.com` 已完成 CDN、DNS 和 HTTPS 接入，正式域名可用；旧入口 `asset.quanttide.com` 继续保留为兼容入口。

目标是让 `asset.cloud.quanttide.com` 指向 `qtcloud-asset-studio` 并通过 HTTPS 访问。

工作内容：

1. 为 `asset.cloud.quanttide.com` 配置 CDN 加速域名。
2. CDN 源站指向 `qtcloud-asset-studio.oss-cn-hangzhou.aliyuncs.com`。
3. 绑定覆盖 `*.cloud.quanttide.com` 或对应域名的 HTTPS 证书。
4. DNS 将 `asset.cloud` 记录 CNAME 到 CDN 返回域名。
5. 在 Provider CORS 中加入 `https://asset.cloud.quanttide.com`。
6. 保留 `https://asset.quanttide.com` 兼容入口。
7. 确认是否需要把旧入口跳转到新入口。

阶段门禁：

- `https://asset.cloud.quanttide.com` 返回 200。
- HTTPS 证书域名匹配，无浏览器安全告警。
- 页面资源不出现跨域错误。
- Provider `/health` 和 `/buckets` 从新域名页面可用。

阶段四接入记录：

- `asset.cloud.quanttide.com` 已指向 CDN 并可通过 HTTPS 访问。
- 证书已覆盖 `asset.cloud.quanttide.com` 和 `*.cloud.quanttide.com`。
- Provider CORS 仍允许新旧 Studio 来源；当前前端请求未使用凭据式模式，因此暂不需要额外收口。
- `asset.quanttide.com` 作为兼容入口继续保留；旧的直连 DNS 记录是否删除，待单独确认后再执行。

## 阶段五 线上验收

> **状态：阻断（2026-08-24）**  
> 结论：新域名、HTTPS、首页、Provider 健康检查和私密桶后端限制已验证；仓库侧已补齐排序状态、对象续页和 URL 编码修复，并已重新发布到 `qtcloud-asset-studio`，当前正式域名 `main.dart.js` ETag 为 `E144C4197AF464B2D85599A0B3CEB8DD`。当前 Studio 覆盖率约 `52.95%`，仍低于契约要求的 `80%`。线上网关仍返回 `Access-Control-Allow-Origin: *`，需要平台侧按已登记来源收口。

目标是确认新域名完整链路可用。

工作内容：

1. 验证首页标题为「量潮资产云」。
2. 验证桶列表、分类、搜索、分页和排序。
3. 验证创建时间排序和 `A→Z / Z→A` 可分别开启、同时开启和同时关闭。
4. 验证非私密桶对象列表可浏览。
5. 验证私密桶只展示元数据，不暴露对象内容或访问链接。
6. 验证 `https://api.quanttide.com/qtcloud-asset/health` 返回 200。
7. 检查浏览器控制台无关键资源 404、CORS 错误或证书错误。

阶段门禁：

- [部分通过] 新域名首页返回 200，页面标题为「量潮资产云」，Provider `/health` 和 `/buckets` 可用。
- [部分通过] 新排序四态、对象续页和元数据桶链接保护已在仓库测试通过，线上正式域名已刷新到最新 `main.dart.js`，`index.html` ETag 仍为 `F79DEB9E8A3CB1B40066B0B881C415BC`。
- [待实现] Studio 覆盖率仍低于 80%，当前约 52.95%。
- [通过] 旧域名兼容入口仍保持可访问。
- [通过] 线上 JS 已使用生产 API 地址，并未发现本地地址、内部函数地址或凭证标记。
- [阻断] 线上 API 网关仍返回 `Access-Control-Allow-Origin: *`，未与 Provider 的精确来源白名单保持一致。

## 阶段六 清理和回滚准备

目标是完成迁移后的资产边界收敛，并保留可回滚方案。

工作内容：

1. 记录新旧域名、CDN、OSS 桶和关键对象 ETag。
2. 保存旧版 `qtcloud-asset` 前端对象清单，必要时支持回滚。
3. 明确 `qtcloud-asset` 后续职责：保留 Provider 发布包、迁出 Provider 包，或退役前端资源。
4. 更新 QA 文档，记录迁移验收结果。
5. 更新 `plan.md` 和 README，使正式入口指向 `asset.cloud.quanttide.com`。
6. 标记 `asset.quanttide.com` 的兼容保留周期。

回滚方式：

1. DNS 或 CDN 回退到 `asset.quanttide.com` 当前可用入口。
2. Studio 发布目标临时回退到 `qtcloud-asset`。
3. 使用上一版 `main.dart.js`、`index.html` 和 `manifest.json` 的 ETag 进行对象级核对。

## 风险和待确认项

- `asset.cloud.quanttide.com` 的 DNS、CDN 和证书属于平台边界，需要平台侧确认。
- `qtcloud-asset-studio` 的 ACL、静态网站配置和 CDN 源站访问方式需要复核。
- Provider CORS 需要同时覆盖新域名和迁移期旧域名。
- `asset.quanttide.com` 是保留为别名还是跳转，需要产品和平台一起确认。
- Provider 发布包是否继续放在 `qtcloud-asset/provider/`，需要后续单独收敛。

## 不在本计划内

- 不创建 `qtcloud-asset-site`。
- 不变更 Provider API 路径 `/qtcloud-asset`。
- 不修改平台共享 API 网关、VPC、RDS、资源组等基础设施归属。
- 不引入 Docker、Kubernetes 或新的运行方式。
- 不开放 OSS 写入、删除、权限管理或私密桶对象访问能力。
