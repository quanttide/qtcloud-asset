# 归档命令

将产品日志从 journal 目录移动到 archive 目录，保持工作区整洁。

## 基本用法

```bash
python -m src.cli.app.cli archive [契约名称] [产品标识] [选项]
```

## 参数说明

| 参数 | 必须 | 说明 | 默认值 |
|------|------|------|--------|
| 契约名称 | 否 | contracts.yaml 中定义的契约名 | journal_backup |
| 产品标识 | 否 | 产品分类目录名 | product |

## 选项

| 选项 | 说明 |
|------|------|
| `-p, --product` | 指定要归档的产品名称 |
| `--pattern` | 文件匹配模式，默认 `*.md` |
| `-n, --dry-run` | 预览模式，不实际移动文件 |

## 示例

### 预览归档

```bash
python -m src.cli.app.cli test_archive product -p qtcloud-asset -n
```

输出：
```
[预览] 契约: test_archive  标识: product
[预览] 产品 (1): qtcloud-asset
[预览] 模式: *.md
  [预览] qtcloud-asset: 2026-04-06.md, 2026-04-07.md

完成: 1/1 成功
```

### 归档单个产品

```bash
python -m src.cli.app.cli test_archive product -p qtcloud-asset
```

### 归档所有产品

```bash
python -m src.cli.app.cli test_archive product
```

### 使用自定义文件模式

只归档 txt 文件：

```bash
python -m src.cli.app.cli test_archive product --pattern "*.txt"
```

## 工作原理

1. 读取 contracts.yaml 获取 journal 和 archive 路径
2. 扫描 journal/product 下所有匹配的文件
3. 移动文件到 archive/product 目录
4. 如果源目录变空，自动删除空目录
5. 提示你在 git 子模块中提交更改
