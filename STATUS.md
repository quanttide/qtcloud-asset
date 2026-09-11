# qtcloud-asset 状态报告

> 更新日期：2026-09-11
> 当前分支：`codex/share-download-release-20260830`
> 最新 commit：`19fce62`

## 当前状态

项目当前由三个协作模块组成：

| 模块 | 当前状态 |
|------|----------|
| CLI (`src/cli`) | Rust/Cargo 主实现，支持本地归档、扫描、校验、配置、版本和 OSS 子命令；旧 Python 文件仍保留为遗留代码 |
| Provider (`src/provider`) | Go 1.25 `net/http` 服务，提供认证、用户管理、桶/对象浏览、公开对象链接和分享接口；生产路径为函数计算 Go Custom Runtime |
| Studio (`src/studio`) | Flutter Web 应用，支持登录、桶列表、对象浏览、排序分页、公开链接、文件/文件夹分享和分享下载 |

## 当前发布入口

- Studio 正式入口：`https://asset.cloud.quanttide.com`
- Studio 发布桶：`qtcloud-asset-studio`
- Provider API：`https://api.quanttide.com/qtcloud-asset`
- Provider 发布包：`oss://qtcloud-asset/provider/`
- 旧入口 `https://asset.quanttide.com` 继续作为兼容入口保留

## 当前未闭合事项

- 平台 SSO 或真实身份源尚未替换当前占位实现。
- 阶段七功能已发布，但管理员创建、浏览、撤销分享的完整线上生命周期仍需补做真实会话验收。
- Provider 发布包是否迁移到专用制品桶或制品仓库仍待决定。
- 旧 Python CLI 文件和历史 Docker/Kubernetes 配置尚未清理。

阶段性 QA 文档保留各自验证日期；需要判断当前结论时，以较新的 QA 记录和当前代码、CI 配置为准。
