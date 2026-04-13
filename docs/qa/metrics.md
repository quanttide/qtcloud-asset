# 质量指标体系

> 定义各维度的量化指标和阶段准入门槛。

## 代码质量

| 指标 | 探索 | 验证 | 发布 | 检查方式 |
|------|------|------|------|----------|
| 测试覆盖率 | ≥ 50% | ≥ 80% | ≥ 90% | `pytest --cov=app --cov-report=term-missing` |
| lint 通过率 | 无 error | 无 error/warning | 无 error/warning | `ruff check .` |
| 格式化合规 | — | — | — | `ruff format . --check` |

## 文档质量

| 指标 | 说明 | 检查方式 |
|------|------|----------|
| 完整性 | BRD/PRD/ADD/QA/User 文档齐全 | 人工审查 |
| 一致性 | 契约声明与实际资产一致 | 契约解析器扫描 |
| 格式合规 | 遵循量潮文档格式标准 | `docs-format` Skill |

## 契约质量

| 指标 | 说明 | 检查方式 |
|------|------|----------|
| 资产可达率 | 契约声明的资产路径实际可访问 | 契约验证器 |
| 技能成功率 | Skill 执行成功率 ≥ 95% | CLI 执行统计 |
| 契约版本一致性 | 契约版本与代码版本对齐 | 契约 diff |

## 准入门槛详解

### placeholder → exploring
- 目录结构已建立
- README 已撰写，说明产品定位

### exploring → validating
- 核心功能原型已验证
- 测试覆盖率 ≥ 50%
- lint 无 error
- 契约已定义，资产可达

### validating → released
- 测试覆盖率 ≥ 80%
- lint 无 error/warning
- Skill 审查全部通过
- 契约可解析且执行成功率 ≥ 95%
- 文档完整（BRD/PRD/ADD/User）

### released 维护
- 每次发布前运行全量质量检查
- 修复 regressions 优先于新功能
