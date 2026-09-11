# 阿里云基础设施代码

## 概述

使用 Terraform 管理应用级阿里云资源（Studio OSS 桶和可选的函数计算资源）。平台级 API 网关、DNS 和证书不在本目录管理。

## 架构

```
函数计算（FC）→ OSS 存储
      ↓
 触发器（HTTP/OSS/定时）
```

## 快速开始

```bash
# 初始化
terraform init

# 预览变更
terraform plan

# 应用变更（需另行授权）
terraform apply

# 销毁资源
terraform destroy
```

## 当前发布边界

- Studio 发布桶：`qtcloud-asset-studio`
- Studio 正式域名：`asset.cloud.quanttide.com`
- Provider API：`https://api.quanttide.com/qtcloud-asset`
- Provider 发布包桶：`qtcloud-asset`，对象前缀为 `provider/`
- Provider `APP_SECRET_KEY`、`AUTH_JWT_PUBLIC_JWK`：在函数计算运行时环境变量中配置，不写入 Terraform 变量或状态文件；由 Provider 发布工作流在发布代码后合并更新
- 平台仓库负责 API 网关、DNS 和证书；本项目不创建或修改这些资源

`qtcloud-asset-studio` 已存在时，不要直接执行创建计划。先导入现有桶：

```bash
terraform import alicloud_oss_bucket.studio qtcloud-asset-studio
```

## 环境说明

当前仓库没有 `environments/dev/`、`environments/staging/` 或
`environments/prod/` 子目录。环境差异通过 Terraform 变量、工作目录和
CI/CD 配置注入；执行前应确认当前变量文件和目标账号。

## 模块

| 模块 | 描述 |
|------|------|
| `modules/fc/` | 函数计算服务和函数 |
| `modules/trigger/` | 触发器（HTTP/OSS/定时） |
| `modules/oss/` | 对象存储桶 |
| `modules/vpc/` | 专有网络（函数计算网络配置） |

## 状态管理

Terraform 状态存储在阿里云 OSS 中，确保团队共享状态和锁定。
