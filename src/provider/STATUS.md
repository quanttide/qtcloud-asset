# Provider 状态报告

> 更新日期：2026-08-16
> 位置：`src/provider/`
> 技术栈：Go 1.22
> 最新版本：无 tag（Go 重写后未发布）

## 版本历史

| 版本 | 日期 | 内容 |
|------|------|------|
| — | 2026-07-30 | Go 全量重写（替换 Python FastAPI） |

## 当前状态

- Go 重写完成，分层结构：`api/`（handler）、`service/`、`repository/`、`schema/`、`config/`
- API 骨架已就绪：`/health`、`/config`
- Service 与 Repository 层待实现

## 规划进度

见根 `ROADMAP.md`：Provider 相关能力依赖目标 1（数字资产契约）与目标 3（Studio 资产浏览）的 API 需求，尚未排期。

## 注意事项

- `src/provider/README.md` 含未解决的合并冲突标记（`<<<<<<< HEAD`），需人工清理
- 产品契约标记 Provider 为「已搁置」（QA 决策 Q006），当前以 CLI 形态运行
