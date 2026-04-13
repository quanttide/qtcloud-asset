# 数字资产云路线图

### 目标 1 — 重新设计数字资产契约

当前契约是简单的路径映射（`journal → archive`），需要演进为**声明式资产地图**：

- [ ] **统一契约 Schema** — 定义 `assets`（资产组成）和 `skills`（操作技能）的标准结构，对齐主仓库 `.quanttide/asset/contract.yaml` 的 `type`/`category` 分层
- [ ] **契约解析器** — 读取 `.quanttide/asset/contract.yaml`，扫描资产目录，生成资产清单（类似 `git ls-files` + YAML 声明的交集）
- [ ] **Skill 执行引擎** — CLI 支持 `qtcloud-asset execute --skill=archive-journal`，从契约中读取 `skills.*.commands` 并执行
- [ ] **契约验证** — 检查契约中声明的资产路径是否存在、是否可访问，输出验证报告
- [ ] **契约 diff** — 对比契约版本差异，显示资产增减、技能变更

### 目标 2 — 重新设计质量控制文档

当前 QA 文档是架构决策记录（Q001-Q008），需要建立**完整的质量治理体系**：

- [ ] **质量指标体系** — 定义代码质量（测试覆盖率、lint 通过率）、文档质量（完整性、一致性）、契约质量（资产可达率、技能成功率）
- [ ] **准入/准出标准** — 每个功能从 exploring → validating → released 的明确门槛
- [ ] **自动化质量检查** — 集成 Skill 审查器（`skill_reviewer.py`）到 CI，自动检查 SKILL.md 完整性
- [ ] **质量报告** — 定期生成质量报告，包含各维度评分和趋势
- [ ] **契约质量关联** — 资产成熟度评分与契约执行成功率联动

### 待办

1. 分析现有契约结构，设计统一 Schema
2. 实现契约解析器（扫描 + 验证 + 报告）
3. 重构 CLI 为 Skill 驱动模式
4. 设计质量控制文档体系
5. 集成质量检查到 CI/CD
