# 样例数据

本目录存放归档流程验证的样例数据。

## 目录结构

sample/journal/product/qtcloud-asset/ 存放待归档的产品日志，包含 2026-04-06.md、2026-04-07.md 等样例日志文件。sample/archive/journal/product/qtcloud-asset/ 存放归档目标目录。sample/journal/product/qtcloud-asset/roadmap.md 存放产品路线图，表示日志已提炼。

## 使用说明

归档前检查 roadmap.md 是否存在，表示日志已提炼满足归档条件。归档时将日志文件移动到 archive 目录。归档后清理 journal 目录中的空文件。