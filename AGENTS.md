# AGENTS.md

## 契约事实源

`.quanttide/` 目录是项目的契约事实源，所有资产和代码规则以这里为准。

| 文件 | 用途 |
|------|------|
| `.quanttide/asset/contract.yaml` | 资产组成、路径、类型 |
| `.quanttide/code/contract.yaml` | 编程规范、依赖、质量门禁 |
| `.quanttide/agent/contract.yaml` | AI 执行审核规则 |

**做任何变更前，先查阅契约文件。** 实际项目结构必须与契约一致。

## AI 执行规则

AI 助手在执行操作前，必须严格遵守 `.quanttide/agent/contract.yaml` 中的审核规则：

1. **每次对话开始时**，AI 必须读取 `.quanttide/agent/contract.yaml`
2. **分析用户请求**，识别是否涉及需要审核的操作
3. **列出操作清单**，标注每个操作的风险等级
4. **等待用户确认**，只有用户明确同意后才能执行高风险操作
5. **执行后反馈**，简要报告执行结果

详见 [`.quanttide/agent/contract.yaml`](.quanttide/agent/contract.yaml)

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

## 工作原则

1. 契约优先：契约定义了项目应该有什么，缺失的资产需要补上
2. 变更同步：修改契约后，确保实际文件/目录已同步
3. 流程遵循：任何文档写作和维护流程遵循 CONTRIBUTING.md
