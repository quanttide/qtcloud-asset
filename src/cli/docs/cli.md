# CLI 入口 — `cli.py`

## 职责

Typer 应用入口，提供契约验证和操作计划执行的用户接口。

**CLI 不做工作流引擎，只负责解析参数、调用底层模块、格式化输出。**

## 命令

### `validate`

```
qtcloud-asset validate --contract <契约路径>
```

验证 `contract.yaml` 格式是否正确。

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `-c, --contract` | 路径 | 否 | 契约文件路径，默认自动查找 |

### `plan`

```
qtcloud-asset plan --contract <契约路径> [--context <JSON>]
```

解析契约，生成并打印操作计划（不执行实际操作）。

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `-c, --contract` | 路径 | 否 | 契约文件路径 |
| `--context` | JSON | 否 | 上下文参数，用于路径变量替换 |

### `execute`

```
qtcloud-asset execute --contract <契约路径> [--context <JSON>] [--dry-run]
```

解析契约并执行操作计划。

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `-c, --contract` | 路径 | 否 | 契约文件路径 |
| `--context` | JSON | 否 | 上下文参数 |
| `-n, --dry-run` | 标志 | 否 | 预览模式 |
| `-v, --verbose` | 标志 | 否 | 详细输出 |

## 执行流程

```
用户输入
    │
    ▼
validate() ── 验证契约格式
    │
    ▼
resolve_plan() ── 解析为操作计划
    │
    ▼
for each operation:
    move_file() / copy_file() / delete_file() / scan_knowledge_base()
    │
    ▼
打印结果（OK / FAIL）+ 成功统计
```

## 错误处理

- 契约不存在 / 格式错误 → 红色错误信息 + `Exit(1)`
- 部分操作失败 → 打印失败原因 + 统计 `X/Y 成功` + `Exit(1)`
- 全部成功 → `Exit(0)`

## 依赖

- `workflow.py` → `resolve_plan`
- `workflow.py` → `move_file`, `copy_file`, `delete_file`, `scan_knowledge_base`
- 不直接操作文件系统，不硬编码业务逻辑
