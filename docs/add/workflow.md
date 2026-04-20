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
