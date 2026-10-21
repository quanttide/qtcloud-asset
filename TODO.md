# TODO

## [0.1.2] 身份源与发布收口

现状核对（2026-10-21）：`codex/asset-auth-no-plan-files` 已合入 `main`（该分支 0 个独有提交，`main` 已在其后前进 16 个提交），发布过程的前三步已成立；`tmp-playwright/` 目录已不存在，其验收意图已固化为 `docs/qa/phase-five-online-acceptance.md`。最新 tag 为 `v0.1.1`，`CHANGELOG.md` 仍是 `[Unreleased]`。

### Added

- [ ] `./` 0.1.2 发布：分支已合并、CI 已过，缺打 tag 与 GitHub Release。达成标志：`CHANGELOG.md` 的 `[Unreleased]` 转为带日期的版本号，打好 tag，`qtcloud-devops release status` 能查到，Release 正文附 CI 运行与线上验收链接。版本号由用户拍板
- [ ] `src/provider/` SSO 接入：接入平台 SSO 或真实身份源，替换当前 `GET /auth/login` 的未配置占位实现。前置：平台侧 SSO 接口与回调地址归属 `quanttide-platform`，需先确认可用
- [ ] `docs/qa/` SSO UI 测试：平台 SSO 接入后补充真实登录回调、会话刷新和过期态 UI 测试（原条目写的 `tmp-playwright/` 已不存在，脚本位置按当时的测试目录定）

### Changed

- [ ] `qtcloud-asset/provider/` 制品迁移评估：评估是否把该制品迁到专用 Provider 制品桶或制品仓库
- [ ] `src/studio/` 产品化确认：确认 `viewer`、`admin` 的用户管理入口和私密桶提示文案是否需要产品化优化
- [ ] `./` 域名策略：旧入口 `asset.quanttide.com` 的保留、跳转或下线另起授权流程，不与本次认证发布混做
- [ ] `./` 资源共享边界：平台共享资源仍由 `quanttide-platform` 管理，本仓库只消费已约定的公网入口
- [ ] `README.md` 文档同步：平台 SSO 完成后同步更新 `README.md`、`src/provider/README.md` 和 `docs/prd/oss-integration.md`

### Security

- [ ] `src/provider/` 登录策略：明确本地账号密码登录的内测退场条件和管理员账号轮换流程
- [ ] `src/provider/` 凭证迁移：将长期 OSS AK/SK 使用路径迁移到 RAM 角色、临时凭证或 KMS 管控链路
- [x] `src/provider/` 签名链接策略：确认私密桶和 `quanttide-terraform-state` 不生成对象访问链接，公开桶使用永久直链
- [ ] `./` 变更授权：API 网关、DNS、CDN、证书、RAM 权限和 OSS 删除操作继续按生产变更清单单独授权
