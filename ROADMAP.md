# 数字资产云路线图

## 当前阶段：文档体系对齐（v0.1.x）

主仓库已建立 BRD→PRD→ADD 三层文档标准，资产云需要将现有文档对齐到新格式。

### 文档对齐

- [ ] **docs/index.md** — 产品简介（已完稿 ✅）
- [ ] **BRD 模块重写** — graph、harness、pricing、skill-composition 按新格式（场景-困境-救援）统一
- [ ] **PRD 模块重写** — 对应 BRD 模块，按用户故事 + Given-When-Then 格式统一
- [ ] **IXD 文档补齐** — 当前 ixd/ 目录仅索引页，需补充交互设计文档（等待主仓库 product-ixd SKILL）
- [ ] **ADD 文档补齐** — 当前缺少 add/ 目录，需为每个模块补充架构设计文档
- [ ] **QA 文档对齐** — 当前 qa/ 是架构决策记录，需按未来 product-qa SKILL 范式转换
- [ ] **README / CONTRIBUTING / AGENTS** — 按认知角色重组（参考 qtcloud-hr）

### 检验标准

- 所有模块对齐完成后通过 `myst build --html` 无报错
- 读者能从 BRD 理解痛点，从 PRD 理解需求，从 ADD 理解方案

---

## 产品开发目标

### 目标 1 — 重新设计数字资产契约

当前契约是简单的路径映射（`journal → archive`），需要演进为**声明式资产地图**：

- [ ] **统一契约 Schema** — 定义 `assets` 和 `skills` 的标准结构，对齐主仓库 `.quanttide/asset/contract.yaml`
- [ ] **契约解析器** — 读取 `.quanttide/asset/contract.yaml`，扫描资产目录，生成资产清单
- [ ] **Skill 执行引擎** — CLI 支持 `qtcloud-asset execute --skill=archive-journal`
- [ ] **契约验证** — 检查契约中声明的资产路径是否存在、是否可访问
- [ ] **契约 diff** — 对比契约版本差异，显示资产增减、技能变更

### 目标 2 — 重新设计质量控制文档

当前 QA 文档是架构决策记录（Q001-Q008），需要建立**完整的质量治理体系**：

- [ ] **质量指标体系** — 代码质量、文档质量、契约质量
- [ ] **准入/准出标准** — 每个功能从 exploring → validating → released 的门槛
- [ ] **自动化质量检查** — 集成到 CI
- [ ] **质量报告** — 定期生成各维度评分和趋势
