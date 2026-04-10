# 用户文档

量潮数字资产云 CLI 帮助你管理文档资产，包括归档、同步等操作。

## 快速导航

| 文档 | 说明 | 操作顺序 |
|------|------|----------|
| [README](README.md) | 首页，看这个开始 | 第一步 |
| [index](index.md) | 目录，不知道看哪篇时看这里 | 随时查看 |
| [install](install.md) | 安装指南，先装好环境 | 第二步 |
| [quickstart](quickstart.md) | 快速入门，跟着做一遍就会了 | 第三步 |
| [archive](archive.md) | 归档命令详解，想深入了解时看 | 忘了就查 |
| [config](config.md) | 配置参考，想自定义时看 | 忘了就查 |

## 命令一览

| 命令 | 说明 |
|------|------|
| `archive` | 归档产品日志，从 journal 移动到 archive |

## AI 使用指南

### 操作步骤

1. 打开终端，进入项目目录
2. 运行 `python -m src.cli.app.cli archive -n` 预览归档
3. 确认无误后，运行 `python -m src.cli.app.cli archive` 执行归档

### AI 工作逻辑

```
你发送指令
    ↓
AI 理解你的意思
    ↓
AI 打开终端
    ↓
AI 执行相关命令
    ↓
AI 把结果展示给你
```

### 常用指令模板

| 你说的话 | AI 会做什么 |
|----------|--------------|
| `预览归档` | 执行带 `-n` 的命令 |
| `执行归档` | 执行不带 `-n` 的命令 |
| `帮我归档 product 下 qtcloud-asset` | 使用指定参数执行 |
| `看看 archive 命令的帮助` | 执行 `--help` |
