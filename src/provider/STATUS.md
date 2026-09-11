# Provider 状态报告

> 更新日期：2026-09-11
> 位置：`src/provider/`
> 技术栈：Go 1.25、`net/http`
> 生产运行方式：阿里云函数计算 Go Custom Runtime

## 当前状态

Provider 已完成从 Python/FastAPI 到 Go 的迁移，当前包含以下能力：

- 健康检查和服务配置接口
- 本地账号登录、会话、JWT 验证和退出登录
- `viewer` / `admin` 两级权限
- 管理员用户邀请、角色修改、停用和会话撤销
- OSS 桶列表、对象列表和公开桶对象链接
- 文件/文件夹分享、分享对象浏览、对象链接、ZIP 下载和撤销
- RDS 用户与分享记录持久化，以及结构化审计日志

主要分层为 `api/`、`auth/`、`service/`、`repository/`、`storage/`、`schema/` 和 `config/`。

## 当前配置

- 默认监听端口：`9000`
- 默认 Provider API：`https://api.quanttide.com/qtcloud-asset`
- 默认 Studio 来源：`https://asset.cloud.quanttide.com`
- 默认用户和分享存储：RDS
- 本地开发可显式使用 `AUTH_MODE=local` 和内存存储

## 未闭合事项

- 默认 SSO 仍是占位实现，平台真实身份源尚未接入。
- 阶段七的完整管理员分享生命周期仍需真实线上会话验收。
- 长期 OSS 凭证路径仍待继续收敛到 RAM 角色、临时凭证或 KMS 管控链路。
