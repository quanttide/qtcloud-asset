# AGENTS.md

## 契约事实源

`.quanttide/` 目录是项目的契约事实源，所有资产和代码规则以这里为准。

| 文件 | 用途 |
|------|------|
| `.quanttide/asset/contract.yaml` | 资产组成、路径、类型 |
| `.quanttide/code/contract.yaml` | 编程规范、依赖、质量门禁 |

**做任何变更前，先查阅契约文件。** 实际项目结构必须与契约一致。

## 文档使用流程

详见 [CONTRIBUTING.md](CONTRIBUTING.md)。

文档各司其职：

| 文档 | 回答的问题 |
|------|-----------|
| `docs/brd/` | 为什么存在业务问题 |
| `docs/prd/` | 产品如何解决问题 |
| `docs/ixd/` | 用户如何与产品交互 |
| `docs/add/` | 技术架构是什么 |
| `docs/qa/` | 质量决策和记录 |
| `docs/user/` | 用户如何使用 |

## 子模组

本项目使用 Git 子模组管理关联仓库：

| 子模组 | 路径 | 用途 |
|--------|------|------|
| qtcloud-asset | `apps/qtcloud-asset` | 云平台应用 |
| handbook | `docs/handbook` | 资产管理手册 |
| context | `docs/context` | 上下文文档 |
| tutorial | `docs/tutorial` | 教程文档 |

**子模组操作规范：**
- 拉取更新：`git submodule update --remote <path>`
- 修改子模组：进入子模组目录提交后，回到主仓库更新引用
- 初始化克隆：`git clone --recurse-submodules <repo-url>`

## 工作原则

1. 契约优先：契约定义了项目应该有什么，缺失的资产需要补上
2. 变更同步：修改契约后，确保实际文件/目录已同步
3. 流程遵循：任何文档写作和维护流程遵循 CONTRIBUTING.md
4. 子模组一致：子模组的变更必须在子模组内提交，主仓库仅更新引用
