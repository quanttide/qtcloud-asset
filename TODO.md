# TODO

## [0.1.2] 身份源与发布收口

### Added
- [ ] `./` 发布前准备：推送 `codex/asset-auth-no-plan-files` 分支并创建 Pull Request 到 `main`。
- [ ] `./` PR 检查：在 PR 中确认不包含 `plan.md`、`plans.md` 或 `plan-a.md`。
- [ ] `./` 合并条件：等 CI 跑完后再合并到 `main`。
- [ ] `./` 发布后操作：合并后运行 `qtcloud-devops release status` 和必要的 release audit。
- [ ] `src/provider/` SSO 接入：接入平台 SSO 或真实身份源，替换当前 `GET /auth/login` 的未配置占位实现。
- [ ] `tmp-playwright/phase5-ui.spec.js` 验收脚本：固化为正式线上验收脚本或纳入 CI smoke 流程。
- [ ] `tmp-playwright/` SSO UI 测试：平台 SSO 接入后补充真实登录回调、会话刷新和过期态 UI 测试。

### Changed
- [ ] `qtcloud-asset/provider/` 制品迁移评估：评估是否把该制品迁到专用 Provider 制品桶或制品仓库。
- [ ] `src/studio/` 产品化确认：确认 `viewer`、`admin` 的用户管理入口和私密桶提示文案是否需要产品化优化。
- [ ] `./` 域名策略：旧入口 `asset.quanttide.com` 的保留、跳转或下线另起授权流程，不与本次认证发布混做。
- [ ] `./` 资源共享边界：平台共享资源仍由 `quanttide-platform` 管理，本仓库只消费已约定的公网入口。
- [ ] `CHANGELOG.md` 版本发布：发布前把 `[Unreleased]` 转为实际版本号和日期。
- [ ] `docs/qa/` 文档补全：发布后把 GitHub Release、CI 运行和线上验收链接补回。
- [ ] `README.md` 文档同步：平台 SSO 完成后同步更新 `README.md`、`src/provider/README.md` 和 `docs/prd/oss-integration.md`。

### Security
- [ ] `src/provider/` 登录策略：明确本地账号密码登录的内测退场条件和管理员账号轮换流程。
- [ ] `src/provider/` 凭证迁移：将长期 OSS AK/SK 使用路径迁移到 RAM 角色、临时凭证或 KMS 管控链路。
- [x] `src/provider/` 签名链接策略：确认私密桶和 `quanttide-terraform-state` 不生成对象访问链接，公开桶使用永久直链。
- [ ] `./` 变更授权：API 网关、DNS、CDN、证书、RAM 权限和 OSS 删除操作继续按生产变更清单单独授权。
