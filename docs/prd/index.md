# 数字资产云

## 产品概述

量潮数字资产云（qtcloud-asset）解决三大核心业务问题：资产关系不透明、AI 交付质量难验证、无形资产缺乏定价依据。

## 用户故事索引

| 故事 | 文档 | 优先级 | 对应 BRD 场景 |
|------|------|--------|---------------|
| 资产全景可视化 | graph.md | Must | 资产关系不透明 |
| 资产变更追踪 | graph.md | Should | 资产变化难追踪 |
| 交付自动验证 | harness.md | Must | 智能体交付质量无法验证 |
| 约束规则管理 | harness.md | Should | 约束规则难维护 |
| 资产估值报告 | pricing.md | Must | 无形资产缺乏定价依据 |
| 定价数据整合 | pricing.md | Should | 定价数据分散难整合 |
| 契约执行引擎 | workflow.md | Must | 治理操作自动化 |
| 原子技能编排 | skill-composition.md | Must | 治理场景硬编码、难以灵活组合 |

## 已验证能力

- 本地脚本管理 Markdown 资产：通过 `backup_product_journal.py` 验证了资产的采集、移动、归档流程
- 本地 AI 转换管道：通过 `generate_product_roadmap.py` 验证了调用本地 Ollama 模型将日志转换为结构化蓝图的可行性
- 目录结构即资产模型：journal/archive/roadmap 三目录结构验证了"资产建模、版本快照、事件路由"的内核设计
