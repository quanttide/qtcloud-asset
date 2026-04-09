# 安装指南

## 环境要求

- Python 3.12 或更高版本
- pip 包管理器（通常随 Python 一起安装）

## 什么是终端？

终端是一个用来输入命令的程序，就像文件夹用来管理文件一样，终端用来管理系统。

### 打开终端的方法

**Windows：**
- 按 `Win + R`，输入 `cmd`，回车
- 或右键点击开始菜单，选择"终端"或"命令提示符"

**Mac：**
- 按 `Command + 空格`，搜索"终端"，回车

**Linux：**
- 按 `Ctrl + Alt + T`

打开后，你会看到一个黑色窗口，里面有闪烁的光标。

## 第一步：确认 Python 版本

在终端中输入以下命令，按回车执行：

```bash
python --version
```

你应该看到类似这样的输出：

```
Python 3.14.3
```

如果显示 3.12 或更高版本，继续下一步。

如果提示"不是内部命令"，尝试：

```bash
python3 --version
```

## 第二步：安装依赖

在终端中运行：

```bash
pip install typer pyyaml
```

看到类似 `Successfully installed` 的文字说明安装成功。

## 验证安装

安装完成后，运行：

```bash
python --version
pip show typer
```

两个命令都能正常输出说明安装成功。

## 卸载

如果不需要了，运行：

```bash
pip uninstall typer pyyaml
```
