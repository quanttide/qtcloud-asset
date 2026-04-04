# 归档产品日志流程

将已处理的产品日志从工作区移动到归档站，保持工作区整洁。

## 前置条件

- 产品日志位于 `docs/journal/product/<product-name>/` 目录下
- 归档目标目录 `docs/archive/journal/product/<product-name>/` 已存在
- 日志内容已提炼到路线图（`docs/roadmap/product/<product-name>.md`）

## 流程

### 1. 确认已提炼

确保日志中的核心想法已总结到对应的产品路线图中。

### 2. 移动文件

将 `docs/journal/product/<product-name>/` 下的所有 `.md` 文件移动到 `docs/archive/journal/product/<product-name>/`。

### 3. 清理空目录

如果 `docs/journal/product/<product-name>/` 目录已空，删除该目录。

## 手动操作示例

```bash
# 移动文件
mv docs/journal/product/<product-name>/*.md docs/archive/journal/product/<product-name>/

# 删除空目录
rmdir docs/journal/product/<product-name>
```

## 归档结构

归档站保持与 journal 相同的目录结构：

```
docs/archive/journal/product/
├── qtcloud-think/
│   ├── 2026-03-15.md
│   ├── 2026-03-31.md
│   └── 2026-04-04.md
├── qtadmin/
└── ...
```

## 注意事项

- 归档后不轻易删除文件
- 保持归档目录结构与源模块一致
- 归档前确认路线图已更新
