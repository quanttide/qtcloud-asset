# 验证指南

## 验证工作流

原型验证遵循渐进式重构策略。首先阅读 backup_product_journal.md 了解当前状态和验证目标，然后执行脚本验证基础归档能力，接着按照 ROADMAP.md 第一阶段重构引入契约配置，重构后验证契约配置是否驱动归档执行。

## 重构原则

每次重构只验证一个设计点。重构时保持脚本可运行。验证后记录结果到验证文档。发现问题调整设计而非放弃。

## 环境配置

backup_product_journal.py 依赖本地文件系统。sample/ 目录提供了完整的样例数据结构。归档前确保 sample/journal 和 sample/archive 目录存在。

## 样例数据

sample/journal/product/qtcloud-asset/ 存放两个样例日志和一个路线图文件。sample/archive/journal/product/qtcloud-asset/ 是归档目标目录。归档逻辑是移动日志文件到 archive 目录并清理 journal 目录。