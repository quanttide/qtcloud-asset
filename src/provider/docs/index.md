# Provider 服务端

量潮数字资产云 Provider 服务端，基于 FastAPI 构建，提供数字资产治理的 HTTP API 接口。

## 实现思路

Provider 采用分层架构，将数字资产治理的核心能力封装为 RESTful API：

```
┌─────────────┐
│   Routers   │  HTTP 路由层，处理请求/响应
├─────────────┤
│  Services   │  业务逻辑层，实现治理流程
├─────────────┤
│ Repositories│  数据访问层，对接存储后端
├─────────────┤
│   Schemas   │  数据模型层，定义 Pydantic 模型
└─────────────┘
```

## 核心能力

Provider 将 CLI 的本地治理能力扩展为服务端能力：

- **契约管理**：读取和解析 `.quanttide/asset/contract.yaml`
- **发现服务**：扫描多源资产（本地、GitHub、飞书）
- **验证服务**：校验资产完整性和一致性
- **快照服务**：生成 Catalog 快照并归档

## 技术栈

- **框架**：FastAPI
- **配置**：PyYAML
- **部署**：阿里云函数计算（FC）

## 目录结构

```
src/provider/
├── app/
│   ├── main.py          # FastAPI 应用入口
│   ├── routers/         # API 路由
│   ├── services/        # 业务服务
│   ├── repositories/    # 数据仓库
│   └── schemas/         # Pydantic 模型
├── tests/               # 单元测试
└── docs/                # 文档
```

## 与 CLI 的关系

Provider 是 CLI 的服务端延伸：

- CLI 面向本地开发，直接操作文件系统
- Provider 面向团队协作，提供远程 API
- 两者共享契约定义和治理逻辑
