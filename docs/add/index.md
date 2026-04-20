# 架构设计文档

## 重构意图

CLI v0.0.1 的核心错误：**把业务逻辑写进了引擎**。

### 具体错误

- `file_operator.py` 硬编码了 `archive_product` 函数
- `cli.py` 直接 import 并调用它
- 工作流路径（journal → archive）写死在代码里，而不是从契约读取

### 架构升级方式

- **移除**：删除 `archive_product` 函数和 CLI 中的直接调用
- **抽象**：`file_operator` 只提供通用能力（移动、复制、删除）
- **驱动**：CLI 读取 `contract.yaml` 中的 `skills` 配置，动态组装工作流

用户想做什么，由契约定义，不由 CLI 预设。
