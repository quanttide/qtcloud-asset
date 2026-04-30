# QtCloud CLI 使用说明

## 快速开始

在项目根目录下，双击运行 `qtcloud.bat`

或者在命令行中：
```bash
cd C:\Users\雨下雨停\q-tech\apps\qtcloud-asset
qtcloud.bat
```

## 命令用法

```
qtcloud.bat              # 直接运行（默认执行归档）
qtcloud.bat scan         # 扫描当前目录的资产
qtcloud.bat scan -i ./examples  # 扫描指定目录
qtcloud.bat config       # 查看契约配置
qtcloud.bat run --dry-run      # 预览模式（不实际执行）
qtcloud.bat run -i ./journal -o ./archive  # 指定输入输出目录
```

## 选项

| 选项 | 说明 |
|------|------|
| `-i, --input` | 输入目录 |
| `-o, --output` | 输出目录 |
| `-s, --skill` | 技能名称 |
| `-p, --pattern` | 文件匹配模式 |
| `-n, --dry-run` | 预览模式 |
| `-v, --verbose` | 详细输出 |
| `-c, --contract` | 契约文件路径 |

## 命令

- **run** — 归档工作流
- **scan** — 扫描资产
- **validate** — 验证契约
- **config** — 查看配置
