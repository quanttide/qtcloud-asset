# 生成产品路线图流程

从产品日志生成产品路线图，将零散的日志提炼为结构化的产品蓝图。

## 前置条件

- 产品日志位于 `docs/journal/<slug>/<product-name>/` 目录下（如 `docs/journal/product/qtcloud-asset/`）
- 目标目录 `docs/roadmap/<slug>/` 会自动创建

## 流程

### 1. 收集日志

读取 `docs/journal/<slug>/` 下每个子目录（每个子目录代表一个产品）中的所有 `.md` 文件，按日期排序。

### 2. 提炼总结

对每篇日志进行内容提炼，提取：

- 产品定位和核心定义
- 功能设计和规则
- 数据处理流程
- 关键洞察和决策
- 验证方向和困惑

### 3. 生成路线图

将提炼结果写入 `docs/roadmap/<slug>/<product-name>.md`，格式要求：

- 非必要不使用表格、粗体等装饰
- 使用纯列表和编号结构
- 保持内容简洁、结构化

### 4. 合并评论反馈

如有外部评论或反馈，将其中的建议整合到路线图中，标注解决进展。

## 自动化脚本

```bash
python examples/generate_product_roadmap.py          # 处理 product 标识
python examples/generate_product_roadmap.py think    # 处理 think 标识
```

脚本遍历 `docs/journal/<slug>/` 下的每个产品子目录，将日志聚合到 `docs/roadmap/<slug>/<product-name>.md`。

## 输出位置

`docs/roadmap/<slug>/<product-name>.md`
