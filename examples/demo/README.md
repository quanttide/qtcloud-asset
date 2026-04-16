# 演示项目：个人笔记管理系统

这是一个使用量潮数字资产云 CLI 工具的演示项目。

## 场景

假设你是一个学习者，每天记录学习笔记。时间久了，笔记越来越多，需要定期归档整理。

## 目录结构

```
demo/
├── .quanttide/
│   └── asset/
│       └── contract.yaml    # 契约配置，定义归档规则
├── journal/
│   └── my-notes/          # 待归档的笔记
│       ├── 2026-04-10-reading-notes.md
│       ├── 2026-04-12-knowledge-system.md
│       └── 2026-04-14-weekly-summary.md
└── archive/
    └── my-notes/          # 归档目标目录（空的）
```

## 契约配置

`contract.yaml` 定义了归档技能：

```yaml
skills:
  archive-notes:
    title: 笔记归档
    description: 将 journal/my-notes 中的笔记移动到 archive/my-notes
    version: "1.0"
    params:
      pattern: "*.md"
```

## 使用方法

### 1. 预览归档

```bash
cd examples/demo
python -m app.cli run -s archive-notes -i journal -o archive -n
```

### 2. 执行归档

```bash
cd examples/demo
python -m app.cli run -s archive-notes -i journal -o archive
```

### 3. 查看结果

归档后，笔记会从 `journal/my-notes/` 移动到 `archive/my-notes/`。

## 注意事项

- 本演示项目使用独立的数据，与 `assets/fixtures/` 测试数据完全隔离
- 归档后 `journal/my-notes/` 目录会被清空
- 如果需要重新演示，可以重新创建 journal 目录和文件
