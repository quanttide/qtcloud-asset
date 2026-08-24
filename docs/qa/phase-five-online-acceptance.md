# 阶段五线上验收记录

> 验证日期：2026-08-24  
> 状态：⏸️ 阻断  
> 正式入口：`https://asset.cloud.quanttide.com`

## 已验证

### 正式入口可访问

- 现实需求：用户可以通过正式域名打开 Studio。
- 判定标准：页面返回 HTTP 200，标题为「量潮资产云」。
- 存证：`curl -I -L https://asset.cloud.quanttide.com`；浏览器标题为「量潮资产云」。
- 状态：✅ 符合

### Provider 基础链路

- 现实需求：Studio 可以访问 Provider。
- 判定标准：`/health` 和 `/buckets` 返回 HTTP 200。
- 存证：`curl` 请求 `https://api.quanttide.com/qtcloud-asset/health` 和 `/buckets`。
- 状态：✅ 符合

### 私密桶对象保护

- 现实需求：私密桶不暴露对象内容和访问链接。
- 判定标准：私密桶对象列表接口返回 HTTP 403；Studio 不显示对象浏览和复制链接入口。
- 存证：`/buckets/qtadmin-private/objects` 返回 `403 Forbidden`；`src/studio/test/widget_test.dart` 覆盖私密桶和 `quanttide-terraform-state`。
- 状态：✅ 符合

## 仓库侧修复

### 排序状态

- 现实需求：创建时间排序和桶名排序可以独立关闭、开启和切换方向。
- 判定标准：支持两项都关闭、单独开启、同时开启；排序条件变化后回到第一页。
- 存证：`src/studio/lib/main.dart` 的 `BucketSortMode` 和 `BucketListView`；`src/studio/test/widget_test.dart` 的四态排序与分页测试。
- 状态：✅ 符合

### 对象续页

- 现实需求：大桶浏览不能静默遗漏超过单页上限的对象。
- 判定标准：Studio 沿 Provider 返回的 `next_marker` 拉取所有续页。
- 存证：`fetchAllObjectPages` 及其 Flutter 测试。
- 状态：✅ 符合

### 对象 URL 编码

- 现实需求：对象名包含空格或保留字符时，复制的公开链接仍可用。
- 判定标准：对象 key 按 URL 路径编码，保留目录分隔符。
- 存证：`src/provider/internal/repository/oss_sort_test.go` 的 URL 编码测试。
- 状态：✅ 符合

### 线上发布刷新

- 现实需求：仓库修复后的 Studio bundle 已重新上线。
- 判定标准：正式域名 `main.dart.js` 与最新构建一致。
- 存证：`https://asset.cloud.quanttide.com/main.dart.js` 返回 `ETag: "E144C4197AF464B2D85599A0B3CEB8DD"`；`index.html` 仍为 `F79DEB9E8A3CB1B40066B0B881C415BC`。
- 状态：✅ 符合

### Studio 覆盖率

- 现实需求：Studio 质量门禁达到契约要求。
- 判定标准：`flutter test --coverage` 的覆盖率不低于 80%。
- 存证：`coverage/lcov.info`，当前结果 `LH:296 / LF:559`，约 `52.95%`。
- 状态：⏸️ 待实现

## 当前阻断

### 网关 CORS 尚未按来源收口

- 现状：线上 API 返回 `Access-Control-Allow-Origin: *`。
- 影响：当前 Studio 使用无凭据 GET 请求仍可工作，但与 Provider 已登记的新旧 Studio 精确来源白名单不一致。
- 后续：由平台侧将网关 CORS 配置为 `https://asset.cloud.quanttide.com` 和 `https://asset.quanttide.com`；涉及生产共享网关配置，需单独授权。

### Studio 覆盖率未达门槛

- 现状：`flutter test --coverage` 当前结果约 `52.95%`。
- 影响：仍低于契约要求的 `80%`，不满足质量门禁。
- 后续：补齐未覆盖分支并重新跑测试覆盖率。

### 旧入口 DNS

- 现状：`asset.quanttide.com` 保持兼容入口，旧 DNS 记录未删除。
- 后续：确认兼容周期和跳转策略后，再单独审批删除或调整。
