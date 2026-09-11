# Studio 状态报告

> 更新日期：2026-09-11
> 位置：`src/studio/`
> 技术栈：Flutter Web、Dart
> CI Flutter 版本：3.41.9

## 当前状态

Studio 已从初始骨架发展为可用的资产浏览界面，当前支持：

- 账号密码登录、登录态展示和退出登录
- 按分类查看、搜索、排序和分页浏览 OSS 桶
- 文件夹下钻、文件搜索、日期/大小排序和对象分页
- 公开桶对象链接复制
- 文件或文件夹分享、分享列表、分享撤销
- 公开分享页浏览、单文件下载和分享内容 ZIP 下载
- 管理员用户管理入口

Provider 地址通过 `PROVIDER_BASE_URL` 构建参数注入。生产 CI 使用 `https://api.quanttide.com/qtcloud-asset`，本地开发默认使用 `http://127.0.0.1:9000`。

## 当前发布入口

- 正式入口：`https://asset.cloud.quanttide.com`
- 发布桶：`qtcloud-asset-studio`
- 兼容入口：`https://asset.quanttide.com`

## 未闭合事项

- 平台真实 SSO 登录回调尚未接入。
- 阶段七的完整管理员分享生命周期仍需补做线上验收。
- 继续保持私密桶和 Terraform 状态桶不暴露对象访问链接。
