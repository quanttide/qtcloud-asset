---
name: docs-validate
description: 验证文档能否正常编译或渲染。
---

# 文档验证技能

验证各类文档能否正常编译或渲染。

## 验证流程

### 1. 检查工具

根据文档格式选择验证工具：

```bash
# Myst 文档
which myst

# 其他格式按需检查
```

### 2. 编译尝试

进入文档目录，执行编译命令：

```bash
cd <docs-dir>
myst build --site
```

### 3. 检查结果

- 无警告/错误
- 编译输出存在

### 正确配置

```yaml
project:
  title: 标题
  authors:
    - name: 名称
  toc:
    - file: index.md
    - title: 章节
      children:
        - file: path/to.md
```

## 验证要点

- 工具已安装
- 配置文件正确
- 源文件存在
- 编译输出正常