# 阿里云基础设施代码

## 概述

使用 Terraform 管理阿里云资源（ECS、OSS、VPC 等）。

## 快速开始

```bash
# 初始化
terraform init

# 预览变更
terraform plan

# 应用变更
terraform apply

# 销毁资源
terraform destroy
```

## 环境说明

| 环境 | 用途 | 配置文件 |
|------|------|----------|
| dev | 开发测试 | `environments/dev/` |
| staging | 预发布验证 | `environments/staging/` |
| prod | 生产环境 | `environments/prod/` |

## 模块

| 模块 | 描述 |
|------|------|
| `modules/ecs/` | 云服务器实例 |
| `modules/oss/` | 对象存储桶 |
| `modules/vpc/` | 专有网络 |

## 状态管理

Terraform 状态存储在阿里云 OSS 中，确保团队共享状态和锁定。
