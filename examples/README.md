# 示例与原型验证

本目录包含示例项目和原型验证脚本。

## 目录结构

| 目录 | 用途 |
|------|------|
| `demo/` | CLI 工具演示项目（独立数据） |
| `archive/` | 归档流程原型验证 |
| `roadmap/` | 路线图生成原型验证 |
| `prototype/` | 原型页面演示 |

## 数据隔离原则

- `examples/demo/` — 演示示例，使用独立虚构数据
- `assets/fixtures/` — 测试夹具，仅供自动化测试使用

## 快速开始

### 演示项目

查看 `demo/README.md` 了解 CLI 归档功能演示。

### 原型验证

运行 `archive/backup_product_journal.py` 归档产品日志。

运行 `roadmap/generate_product_roadmap.py` 生成产品路线图。