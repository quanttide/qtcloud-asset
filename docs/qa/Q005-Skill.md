# Q005: Skill 审查

**[ 需求来源 ]**：ROADMAP Phase 2 要求自动化质量检查，集成 Skill 审查器到 CI。
**[ 核心契约 ]**：`examples/skill/app/skill_reviewer.py`

---

## 1. 现实需求：审查单个 Skill
* **验收点**：`python skill_reviewer.py SKILL.md` 输出审查报告。
* **验证资产**：`examples/skill/app/skill_reviewer.py`
* **状态**：✅ *Passed*

## 2. 现实需求：检查结构完整性
* **验收点**：缺少"用途"、"触发"、"执行步骤"任一章节时报错。
* **验证资产**：`examples/skill/sample/release/SKILL.md`
* **状态**：✅ *Passed*

## 3. 现实需求：检测空代码块
* **验收点**：代码块为空或只有注释时警告。
* **验证资产**：审查器 `parse_code_blocks()` 函数
* **状态**：✅ *Passed*

## 4. 现实需求：检测危险操作
* **验收点**：`rm -rf`、`git push --force` 等命令被警告。
* **验证资产**：审查器 `DANGEROUS_PATTERNS` 常量
* **状态**：✅ *Passed*

## 5. 现实需求：检测失败回退方案
* **验收点**：无回滚/恢复关键词时警告。
* **验证资产**：审查器 `_check_recoverability()` 函数
* **状态**：✅ *Passed*

## 6. 现实需求：CI 集成
* **验收点**：GitHub Actions 中自动审查所有 SKILL.md。
* **验证资产**：待实现 CI workflow
* **状态**：⏸️ *Pending*
