# Studio 状态报告

> 更新日期：2026-08-16
> 位置：`src/studio/`
> 技术栈：Flutter (Dart)
> 最新版本：无 tag

## 版本历史

| 版本 | 日期 | 内容 |
|------|------|------|
| — | 2026-04-17 | 随 v0.0.1 发布：Flutter Web 应用骨架 |

## 当前状态

- 骨架：`main.dart`（导航栏）+ `screens/asset_contract_screen.dart`（数字资产契约页）
- 资产契约页按约束力层级（宪法层/法律层/法理层）分类展示资产，与 `.gitmodules` 对齐
- 基础 widget 测试存在（`test/widget_test.dart`）

## 规划进度

见根 `ROADMAP.md` 目标 3（Studio：资产浏览，参照 qtfounder asset 页模式）：

- 资产契约落地、资产目录引擎、通用资产页面、只读浏览、现有页面升级——均未开始
- `asset_contract_screen.dart` 待按契约驱动重构，与目标 1（契约解析器）衔接
