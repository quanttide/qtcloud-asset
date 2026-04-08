# AGENTS.md

本目录维护契约驱动架构的原型验证脚本。

## 相关文档

| 文档 | 路径 | 用途 |
|------|------|------|
| 验证指南 | CONTRIBUTING.md | 重构工作流和原则 |
| 演进路线 | ROADMAP.md | 验证路径和阶段目标 |
| AI 转换验证 | generate_product_roadmap.md | 当前状态和下一步 |
| 归档验证 | backup_product_journal.md | 当前状态和下一步 |
| 文档格式 | docs/CONTRIBUTING.md | Markdown 写作规范 |

## 快速索引

| 任务 | 查看文档 |
|------|---------|
| 执行重构 | ROADMAP.md |
| 了解验证标准 | generate_product_roadmap.md 或 backup_product_journal.md |
| 编写验证文档 | docs/CONTRIBUTING.md |

## 协作原则

最小干预，仅用户明确请求时执行重构。遵循规范，参照 ROADMAP.md 和验证文档执行。原子验证，每次重构独立验证一个设计点。

## 重要提示

重构前确保理解当前脚本状态。重构时保持脚本可运行。验证后记录结果到验证文档。发现问题调整设计而非放弃。迭代演进逐步接近完整架构。

## 维护规则

新增验证脚本时创建对应的验证文档。ROADMAP.md 新增阶段目标时更新。验证文档记录当前状态和下一步重构目标。不重复 README/CONTRIBUTING 已有内容。