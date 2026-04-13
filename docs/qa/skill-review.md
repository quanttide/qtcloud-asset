# Skill 审查规范

> 集成 Skill 审查器到 CI，确保所有 SKILL.md 符合质量标准。

## 审查维度

| 维度 | 规则 | 级别 |
|------|------|------|
| 结构完整性 | 必须包含"用途"、"触发"、"执行步骤" | error |
| 代码块可执行性 | 代码块不能为空或只有注释 | warning |
| 变量一致性 | 模板变量应有赋值定义 | warning |
| 安全性 | 检测危险操作（rm -rf, force push 等） | warning |
| 可恢复性 | 应包含失败回退方案 | warning |

## CI 集成

```yaml
# GitHub Actions 示例
- name: Review Skills
  run: |
    for skill in examples/skill/sample/*/SKILL.md .agents/skills/*/SKILL.md; do
      if [ -f "$skill" ]; then
        python3 examples/skill/app/skill_reviewer.py "$skill"
      fi
    done
```

## 审查器位置

- **实现**: `examples/skill/app/skill_reviewer.py`
- **设计文档**: `examples/skill/docs/skill_reviewer.md`

## 使用方式

```bash
# 审查单个 Skill
python3 examples/skill/app/skill_reviewer.py examples/skill/sample/release/SKILL.md

# 审查所有 Skill
find . -name "SKILL.md" -exec python3 examples/skill/app/skill_reviewer.py {} \;
```

## 审查报告示例

```
============================================================
Skill 审查报告: examples/skill/sample/release/SKILL.md
============================================================

章节 (6): 用途, 触发, 执行步骤, 子模块发布, 标签格式, 注意事项
代码块: 5
变量: {major}, {minor}, {patch}, {version}, {module}

发现问题: 0 错误, 1 警告

  ⚠ [warning] unassigned_variable: 变量 {version} 在 14 处使用但未赋值

============================================================
✅ 审查通过（有警告但无错误）
============================================================
```
