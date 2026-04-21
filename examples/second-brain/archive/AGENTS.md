# AGENTS.md

本目录维护归档流程的原型验证。

## 相关文档

| 文档 | 路径 | 用途 |
|------|------|------|
| 验证指南 | CONTRIBUTING.md | 重构工作流和原则 |
| 演进路线 | ../ROADMAP.md | 验证路径和阶段目标 |
| 验证状态 | backup_product_journal.md | 当前状态和下一步 |
| 样例数据 | sample/README.md | 样例目录说明 |

## 快速索引

| 任务 | 查看文档 |
|------|---------|
| 执行重构 | ../ROADMAP.md |
| 了解验证标准 | backup_product_journal.md |
| 使用样例数据 | sample/README.md |

## 协作原则

最小干预，仅用户明确请求时执行重构。遵循规范，参照 ROADMAP.md 执行。原子验证，每次重构独立验证一个设计点。

## 重要提示

重构前确保理解当前脚本状态。重构时保持脚本可运行。验证后记录结果到验证文档。发现问题调整设计而非放弃。使用 sample/ 样例数据验证。

## 维护规则

新增验证点时更新 backup_product_journal.md。样例数据变化时更新 sample/README.md。不重复 README/CONTRIBUTING 已有内容。