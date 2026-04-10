# 数字资产工作流 DSL

## 当前契约结构（V1.0）

CLI 使用 `contracts.yaml` 定义路径映射，结构精简：

```yaml
contracts:
  journal_backup:
    name: 产品日志归档
    version: 1
    paths:
      journal: docs/journal
      archive: docs/archive/journal
```

工作流由 planner.py 解析：扫描 `journal/{slug}/` 下的产品子目录，为每个产品生成 `ArchiveTask`。触发方式固定为手动（CLI 命令）。

## V2.0 规划：ISDL Lite

V2.0 引入完整的 ISDL（统一数字资产描述语言），契约结构扩展为：

| 字段 | 必填 | 说明 |
|------|------|------|
| `name` | ✅ | 契约名称 |
| `version` | ✅ | 契约版本号 |
| `asset` | ✅ | 源资产定义（platform、type、schema.required） |
| `transform` | ✅ | 转换规则（input、output、action、params） |
| `trigger` | 默认 manual | 触发方式 |
| `lifecycle` | 可选 | 生命周期配置 |

状态流转：`draft` → `active` → `archived`。

触发方式：
- `manual` — 手动触发（V1.0 默认，当前实现）
- `cron` — 定时触发（V2.0）
- `event` — 事件触发（V2.0）

适配器（V2.0 硬编码支持）：
- `feishu` — 支持 `feishu_to_markdown`
- `github` — 支持 `github_upload`

执行日志记录：`contract_id`、`version`、`timestamp`、`action`、`result`、`duration_ms`。

## V2.0 扩展

- 多资产支持：`assets` 数组
- 链式转换：`transforms` 链式执行
- 事件溯源：`contract_events` 和 `execution_events`
- 适配器层平台能力协商
- 健康检查：平台 API 可达性检查
- 错误处理：重试、DLQ、降级
