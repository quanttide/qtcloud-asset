# 架构设计文档

## 重构意图

CLI v0.0.1 错误地将 `archive_product` 硬编码为内置功能。这是平台用户应该自己定义的工作流。

重构目标：移除 `archive_product`，CLI 只提供通用引擎，工作流由用户通过契约的 `skills` 定义。
