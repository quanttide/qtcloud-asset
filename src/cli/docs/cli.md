# CLI 入口 — `cli.py`

## 职责

Typer 应用入口，定义用户可见的 CLI 命令和交互流程。

## 命令

### `archive`

```
archive [CONTRACT] [SLUG] [-p/--product PRODUCT] [--pattern PATTERN] [-n/--dry-run]
```

将产品日志从 journal 目录移动到 archive 目录。

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `contract` | 位置参数 | `journal_backup` | `contracts.yaml` 中的契约名称 |
| `slug` | 位置参数 | `product` | 产品分类标识（如 `product`、`qtadmin`） |
| `-p, --product` | 选项 | `None`（全部） | 指定单个产品 |
| `--pattern` | 选项 | `*.md` | 文件匹配 glob 模式 |
| `-n, --dry-run` | 标志 | `False` | 预览模式，不移动文件 |

## 执行流程

```
用户输入
    │
    ▼
resolve_workflow() ── 解析契约，生成 Workflow
    │
    ▼
print_workflow_summary() ── 打印摘要
    │
    ▼
for each task:
    archive_product() ── 执行归档
    │
    ▼
打印结果（✓ / ✗）+ 成功统计
    │
    ▼
提示提交 git 子模块
```

## 错误处理

- 契约不存在 / 目录不存在 → `typer.secho` 红色错误信息 + `Exit(1)`
- 部分产品失败 → 打印失败原因 + 统计 `X/Y 成功` + `Exit(1)`
- 全部成功 → `Exit(0)`

## 依赖

- `planner.py` → `resolve_workflow`, `print_workflow_summary`
- `file_operator.py` → `archive_product`
- 不直接操作文件系统
