# AGENTS.md

本目录维护路线图生成的原型验证。

## 相关文档

| 文档 | 路径 | 用途 |
|------|------|------|
| 验证指南 | CONTRIBUTING.md | 重构工作流和原则 |
| 演进路线 | ../ROADMAP.md | 验证路径和阶段目标 |
| 验证状态 | generate_product_roadmap.md | 当前状态和下一步 |

## 快速索引

| 任务 | 查看文档 |
|------|---------|
| 执行重构 | ../ROADMAP.md |
| 了解验证标准 | generate_product_roadmap.md |

## 协作原则

最小干预，仅用户明确请求时执行重构。遵循规范，参照 ROADMAP.md 执行。原子验证，每次重构独立验证一个设计点。

## 重要提示

重构前确保理解当前脚本状态。重构时保持脚本可运行。验证后记录结果到验证文档。发现问题调整设计而非放弃。配置 Ollama 环境后验证。

## 维护规则

新增验证点时更新 generate_product_roadmap.md。不重复 README/CONTRIBUTING 已有内容。