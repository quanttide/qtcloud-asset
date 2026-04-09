# 贡献指南

欢迎贡献本项目！

## 目录结构

- `src/provider/app/` — 源码
- `src/provider/tests/` — 单元测试
- `tests/` — 集成测试

## 测试规范

测试文件与源码模块一一对应：`tests/<模块名>/test_<源文件名>.py`

示例：
- `src/provider/app/repositories/file_operator.py` → `src/provider/tests/repositories/test_file_operator.py`
- `src/provider/app/services/planner.py` → `src/provider/tests/services/test_planner.py`

覆盖率目标：80%+

## 提交流程

1. Fork 仓库
2. 创建分支：`git checkout -b feature/xxx`
3. 进行修改
4. 提交更改：遵循Git提交规范
5. 推送分支：`git push origin branch-name`
6. 创建 Pull Request

## 编写规范

遵循量潮科技文档格式标准：删除不必要的格式元素，优先用段落和标题；能用列表就不表格，能用文字就不列表；全文格式元素尽量少；同一概念全程使用相同名称。

具体规范：https://github.com/quanttide/quanttide-specification-of-business-entity/blob/v0.1.1/docs/format.md
