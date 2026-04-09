# 配置参考

CLI 使用 `contracts.yaml` 文件定义归档规则。

## 文件位置

CLI 在以下位置查找配置文件（按顺序）：

1. 环境变量 `CONTRACTS_FILE` 指定的路径
2. `src/cli/contracts.yaml`（默认）

## 创建配置文件

### 第一步：找到正确的位置

在项目目录中找到 `src/cli/` 文件夹，你应该看到类似这样的结构：

```
qtcloud-asset/
├── src/
│   └── cli/
│       ├── app/
│       └── contracts.yaml   ← 在这里创建或编辑
```

### 第二步：创建或编辑文件

你可以用任何文本编辑器打开 `contracts.yaml`：

- **Windows**：右键文件 → 选择"打开方式" → 记事本（或 Notepad++）
- **Mac**：双击文件 → TextEdit
- **VS Code**（推荐）：右键 → 用 VS Code 打开

## 文件格式

```yaml
contracts:
  契约名称:
    name: 显示名称
    version: 版本号
    paths:
      journal: journal目录路径
      archive: archive目录路径
```

## 配置示例

```yaml
contracts:
  journal_backup:
    name: 产品日志归档
    version: 1
    paths:
      journal: docs/journal
      archive: docs/archive/journal
```

## 字段说明

| 字段 | 必须 | 说明 |
|------|------|------|
| `contracts` | 是 | 配置根节点，保持不变 |
| `契约名称` | 是 | 自定义名称，CLI 中使用 |
| `name` | 是 | 显示名称，方便理解 |
| `version` | 是 | 版本号，填 1 即可 |
| `paths.journal` | 是 | 源目录（要归档的文件在哪里） |
| `paths.archive` | 是 | 目标目录（文件要移动到哪里） |

## 路径规则

- 路径相对于项目根目录
- 支持绝对路径和相对路径
- journal 和 archive 下需要按 product 组织子目录

## 目录结构示例

```
项目目录/
├── docs/
│   ├── journal/           ← journal 根目录
│   │   └── product/      ← 产品标识（CLI 中用这个名）
│   │       └── qtcloud-asset/  ← 具体产品
│   │           └── *.md
│   └── archive/           ← archive 根目录
│       └── product/
│           └── qtcloud-asset/  ← 归档位置
```

## 常见问题

### Q: 路径写错了怎么办？

A: CLI 会报错告诉你找不到目录，仔细检查拼写和路径是否正确。

### Q: 可以同时配置多个契约吗？

A: 可以，在 `contracts:` 下面添加多个契约即可。
