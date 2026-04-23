---
name: devops-release
description: 版本发布技能，遵循语义化版本规范，用于创建 Git 标签、更新 CHANGELOG 并发布 GitHub Release。当需要发布新版本、打标签、更新发布说明时使用。
---

# 版本发布技能

用于 quanttide-gallery-of-business-entity 仓库的版本发布流程。

## 发布前检查

1. 检查未发布内容:
   ```bash
   git log <last-version>..HEAD --oneline
   ```

2. 确认发布就绪:
   - [ ] 所有测试通过
   - [ ] 文档已更新
   - [ ] 无未提交的变更
   - [ ] CHANGELOG已更新
   - [ ] 版本号符合语义化版本规范

## 更新 CHANGELOG

1. 编辑 `CHANGELOG.md`:
   ```markdown
   # CHANGELOG

   ## [Unreleased]

   ## [<version>] - <date>

   ### Added
   - 新增功能列表

   ### Changed
   - 修改内容列表

   ### Fixed
   - 修复问题列表

   ### Removed
   - 移除内容列表
   ```

2. 内容要求:
   - 简洁明了，避免冗余描述
   - 使用动词开头：「新增」「修改」「修复」「移除」
   - 每条记录一行，不超过50字
   - 按重要性排序

3. 提交并推送:
   ```bash
   git add CHANGELOG.md
   git commit -m "chore: update CHANGELOG for <version>"
   git push origin main
   ```

## 创建 Git 标签

1. 创建附注标签:
   ```bash
   git tag -a <version> -m "Release version <version-no-v>"
   ```
   其中 `<version-no-v>` 为不带 `v` 前缀的版本号（如 `0.0.2`）

2. 推送标签到远程:
   ```bash
   git push origin <version>
   ```

## 创建 GitHub Release

使用 GitHub CLI 创建 Release:

```bash
gh release create <version> \
  --title "<version>" \
  --generate-notes
```

如需更新 Release notes（使用 `--generate-notes` 时可能不符合 CHANGELOG 格式）：
```bash
gh api repos/<owner>/<repo>/releases/<release-id> \
  -X PATCH \
  -f body="### Added

- 内容从CHANGELOG提取

### Changed

- 内容从CHANGELOG提取"
```

Release URL: `https://github.com/quanttide/quanttide-gallery-of-business-entity/releases/tag/<version>`

## 发布后确认

- [ ] Git标签已创建并推送
- [ ] GitHub Release已创建
- [ ] Release notes内容准确
- [ ] 主仓库子模块引用已更新
- [ ] 版本号符合语义化版本规范

验收方式：
```bash
# 检查标签是否存在
git tag -l v*

# 检查标签格式是否为 vX.Y.Z
git tag -l "v[0-9]*.[0-9]*.[0-9]*"

# 检查 CHANGELOG 对应版本是否存在
grep "## \[v" CHANGELOG.md
```

版本号规则（语义化版本: `主版本.次版本.修订号`）：

- **主版本（Major）**: 不兼容的API修改
- **次版本（Minor）**: 向后兼容的功能新增
- **修订号（Patch）**: 向后兼容的问题修复

预发布版本格式：

- Alpha版本：`v0.0.1-alpha.1`
- Beta版本：`v0.0.1-beta.1`
- RC版本：`v0.0.1-rc.1`
- Release版本：`v0.0.1`

## 规范来源

版本发布操作遵循量潮科技工程标准的版本发布标准
[https://github.com/quanttide/quanttide-specification-of-business-entity/blob/v0.1.1/devops/release.md](https://github.com/quanttide/quanttide-specification-of-business-entity/blob/v0.1.1/devops/release.md)
