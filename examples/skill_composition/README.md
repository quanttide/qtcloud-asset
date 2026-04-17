# 原子技能编排示例项目

本项目展示如何使用原子技能编排功能，通过配置文件定义工作流。

## 项目结构

```
skill_composition/
├── .quanttide/
│   └── asset/
│       └── contract.yaml    # 契约配置
├── skills/
│   ├── list_files.yaml     # 技能：列举文件
│   ├── move_files.yaml     # 技能：移动文件
│   ├── copy_files.yaml     # 技能：复制文件
│   ├── verify.yaml        # 技能：验收结果
│   └── compress.yaml      # 技能：压缩文件
├── workflows/
│   ├── archive.yaml       # 工作流：归档
│   ├── backup.yaml        # 工作流：备份
│   └── migrate.yaml       # 工作流：迁移
└── data/
    ├── journal/           # 源数据目录
    └── archive/           # 归档目标目录
```

## 快速开始

### 1. 查看可用技能

技能定义在 `skills/` 目录下，每个 YAML 文件定义一个原子技能。

### 2. 查看工作流配置

工作流定义在 `workflows/` 目录下，展示如何组合技能。

### 3. 组合新工作流

创建新的 YAML 文件，参考现有工作流格式进行配置。

## 核心概念

### 原子技能

原子技能是系统提供的基础功能单元，不可拆分：

| 技能 ID | 名称 | 功能 |
|---------|------|------|
| list_files | 列举文件 | 列出目录下所有文件 |
| move_files | 移动文件 | 将文件移动到目标目录 |
| copy_files | 复制文件 | 将文件复制到目标目录 |
| verify | 验收结果 | 验证操作是否成功 |
| compress | 压缩文件 | 将文件压缩为 ZIP |

### 工作流

工作流是原子技能的有序组合：

```yaml
workflow:
  name: "归档"
  description: "将日志文件归档到 archive 目录"
  skills:
    - id: list_files
      params:
        pattern: "*.md"
        source: data/journal
    - id: move_files
      params:
        source: data/journal
        destination: data/archive
    - id: verify
      params:
        check: "destination_exists"
```

### 公式

| 工作流 | 公式 |
|--------|------|
| 归档 | 列举文件 → 移动文件 → 验收结果 |
| 迁移 | 列举文件 → 复制文件 → 权限变更 → 验收结果 |
| 备份 | 列举文件 → 复制文件 → 压缩文件 → 验收结果 |

## 使用示例

### 示例 1：自定义归档规则

创建 `workflows/my-archive.yaml`：

```yaml
workflow:
  name: "我的归档"
  description: "归档特定前缀的日志文件"
  skills:
    - id: list_files
      params:
        pattern: "2026-*.md"
        source: data/journal
    - id: move_files
      params:
        source: data/journal
        destination: data/archive
    - id: verify
      params:
        check: "count_match"
```

### 示例 2：组合多个技能

创建 `workflows/full-backup.yaml`：

```yaml
workflow:
  name: "完整备份"
  description: "备份并压缩所有数据"
  skills:
    - id: list_files
      params:
        pattern: "*"
        source: data/journal
    - id: copy_files
      params:
        source: data/journal
        destination: data/archive
    - id: compress
      params:
        source: data/archive
        output: "backup-2026-04-17.zip"
    - id: verify
      params:
        check: "archive_created"
```

## 相关文档

- BRD: `docs/brd/skill-composition.md`
- PRD: `docs/prd/skill-composition.md`
