# CLI 入口 — `cli.py`

## 职责

Typer 应用入口，定义用户可见的 CLI 命令和交互流程。

## 命令

### `run`

```
qtcloud-asset --input=<源> --contract=<契约> --output=<目标> [-p/--pattern] [-n/--dry-run] [-v/--verbose]
```

数据转换：输入 → 契约(转换) → 输出

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `-i, --input` | 路径 | 是 | 数据源目录 |
| `-c, --contract` | 字符串 | 是 | `contracts.yaml` 中的契约名称 |
| `-o, --output` | 路径 | 是 | 输出目标目录 |
| `-p, --pattern` | 字符串 | 否 | 文件匹配模式，默认 `*.md` |
| `-n, --dry-run` | 标志 | 否 | 预览模式，不执行实际操作 |
| `-v, --verbose` | 标志 | 否 | 详细输出 |

## 使用示例

```bash
# 基本用法
qtcloud-asset -i ./docs/journal -c archive -o ./docs/archive

# 预览模式
qtcloud-asset -i ./docs/journal -c archive -o ./docs/archive --dry-run

# 详细输出
qtcloud-asset -i ./docs/journal -c archive -o ./docs/archive -v

# 指定文件模式
qtcloud-asset -i ./docs -c archive -o ./output --pattern "*.txt"
```

## 执行流程

```
用户输入
    │
    ▼
resolve_workflow_simple() ── 解析契约，生成 Workflow
    │
    ▼
for each task:
    archive_product() ── 执行归档
    │
    ▼
打印结果（OK / FAIL）+ 成功统计
```

## 错误处理

- 契约不存在 / 目录不存在 → 红色错误信息 + `Exit(1)`
- 部分产品失败 → 打印失败原因 + 统计 `X/Y 成功` + `Exit(1)`
- 全部成功 → `Exit(0)`

## 依赖

- `planner.py` → `resolve_workflow_simple`, `print_workflow_summary`
- `file_operator.py` → `archive_product`
- 不直接操作文件系统
