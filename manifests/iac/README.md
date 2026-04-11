# 阿里云基础设施代码

## 概述

使用 Terraform 管理阿里云 FaaS 资源（函数计算、API 网关、OSS 等）。

## 架构

```
API 网关 → 函数计算（FC）→ OSS 存储
                ↓
           触发器（HTTP/OSS/定时）
```

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
| `modules/fc/` | 函数计算服务和函数 |
| `modules/api-gateway/` | API 网关配置 |
| `modules/trigger/` | 触发器（HTTP/OSS/定时） |
| `modules/oss/` | 对象存储桶 |
| `modules/vpc/` | 专有网络（函数计算网络配置） |

## 状态管理

Terraform 状态存储在阿里云 OSS 中，确保团队共享状态和锁定。
