# 数字资产工作流 DSL

## ISDL Lite v1.0

V1.0 采用精简设计，专注核心需求：跨平台转换 + 版本控制。

### 契约结构

```yaml
name: 飞书文档同步到GitHub
version: 1

# 资产定义（必选）
asset:
  platform: feishu
  type: document
  schema:
    required:
      - content
      - title

# 转换规则（必选）
transform:
  name: 转Markdown
  input: feishu.document
  output: github.markdown
  action: feishu_to_markdown
  params:
    extract_frontmatter: true
    strip_styles: true

# 触发（默认 manual）
trigger: manual

# 生命周期（简化版）
lifecycle:
  initial: draft
  states:
    - draft      # 草稿
    - active     # 活跃/发布
    - archived   # 归档
```

### 字段说明

| 字段 | 必选 | 说明 |
|------|------|------|
| name | 是 | 契约名称 |
| version | 是 | 契约版本号 |
| asset | 是 | 源资产定义 |
| asset.platform | 是 | 平台标识（feishu/github） |
| asset.type | 是 | 资源类型 |
| asset.schema.required | 是 | 必填字段列表 |
| transform | 是 | 转换规则 |
| transform.action | 是 | 转换动作名 |
| transform.params | 否 | 转换参数 |
| trigger | 否 | 触发方式，默认 manual |
| lifecycle | 否 | 生命周期配置 |
| lifecycle.initial | 否 | 初始状态，默认 draft |
| lifecycle.states | 否 | 状态列表 |

### 状态流转

```
draft → active → archived
```

- draft：草稿，可编辑
- active：发布，执行转换
- archived：归档，保留历史

### 触发方式

| 类型 | 说明 |
|------|------|
| manual | 手动触发（v1.0 默认） |
| cron | 定时触发（v2.0） |
| event | 事件触发（v2.0） |

### 适配器

v1.0 硬编码支持两种平台：

| 平台 | 支持的 action |
|------|---------------|
| feishu | feishu_to_markdown |
| github | github_upload |

### 执行日志

```yaml
# 执行记录
logs:
  - contract_id: xxx
    version: 1
    timestamp: 2024-01-01T00:00:00Z
    action: feishu_to_markdown
    result: success
    duration_ms: 1500
```

---

## V2.0 扩展（规划）

| 特性 | 说明 |
|------|------|
| 多资产 | assets 数组支持 |
| 多转换 | transforms 链式执行 |
| 事件溯源 | contract_events + execution_events |
| 适配器层 | 平台能力协商 |
| 健康检查 | 平台 API 可达性检查 |
| 错误处理 | 重试、DLQ、降级 |

---

## 目录结构

```
docs/assets/
└── <platform>/
    └── <asset_type>/
        └── <version>/
            ├── pending/      # draft 状态
            ├── active/       # active 状态
            └── archive/      # archived 状态
```
