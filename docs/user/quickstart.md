# 快速开始

## 目标

将示例文件从 journal 目录归档到 archive 目录。

## 基础概念

### 什么是目录？

目录就是文件夹。本教程中"目录"和"文件夹"是同一个意思。

### 什么是路径？

路径就是文件或文件夹在电脑中的地址。

例如：
- `D:\github clone hub\qtcloud-asset\docs` 是 Windows 上的一个路径
- `/home/user/projects` 是 Mac 或 Linux 上的一个路径

### 常用命令

| 命令 | 作用 | 示例 |
|------|------|------|
| `cd 目录` | 进入目录 | `cd docs` |
| `cd ..` | 返回上级目录 | 返回上一级 |
| `dir` (Windows) | 查看当前目录内容 | 查看有什么文件 |
| `ls` (Mac/Linux) | 查看当前目录内容 | 查看有什么文件 |
| `pwd` | 查看当前目录路径 | 确认你在哪里 |

## 操作步骤

### 1. 进入项目目录

打开终端，运行：

```bash
cd qtcloud-asset
```

如果终端提示找不到，可能需要输入完整路径：

**Windows：**
```bash
cd D:\github clone hub\qtcloud-asset
```

**Mac/Linux：**
```bash
cd /home/你的用户名/qtcloud-asset
```

### 2. 确认位置

运行以下命令，确认你在正确的目录：

```bash
dir
```

你应该能看到 `src`、`docs`、`examples` 等文件夹。

### 3. 查看帮助

```bash
python -m src.cli.app.cli archive --help
```

看到帮助信息说明一切正常。

### 4. 预览归档（dry-run）

**重要：预览模式不会真的移动文件，只是让你看看会发生什么。**

```bash
python -m src.cli.app.cli test_archive product -p qtcloud-asset -n
```

解释：
- `test_archive` - 使用哪个配置
- `product` - 产品分类目录名
- `-p qtcloud-asset` - 指定具体产品
- `-n` 或 `--dry-run` - 预览模式

### 5. 执行归档

确认预览结果后，去掉 `-n` 参数执行实际归档：

```bash
python -m src.cli.app.cli test_archive product -p qtcloud-asset
```

### 6. 查看结果

运行以下命令查看 archive 目录：

**Windows：**
```bash
dir examples\archive\archive\product\qtcloud-asset
```

**Mac/Linux：**
```bash
ls examples/archive/archive/product/qtcloud-asset
```

## 下一步

- 查看 [归档命令详解](archive.md)
- 查看 [配置参考](config.md)
