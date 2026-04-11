# 量潮数字资产云

数字资产管理基础设施，提供资产发现、关系抽取、AI 交付约束验证和数字资产定价能力。

## 架构

```
Studio (Flutter Web) ← HTTP → Provider (FastAPI) ← SDK → 阿里云 (FC/OSS)
                            ↑
                       CLI (管理工具)
```

## 目录结构

```
├── .quanttide/           # 项目契约
│   ├── asset/
│   │   └── contract.yaml # 资产组成定义
│   └── code/
│       └── contract.yaml # 代码规则定义
├── docs/                 # 项目文档
│   ├── brd/              # 商业需求文档
│   ├── prd/              # 产品需求文档
│   ├── ixd/              # 交互设计文档
│   ├── add/              # 架构设计文档
│   ├── qa/               # 质量保证文档
│   └── user/             # 用户文档
├── src/                  # 源代码
│   ├── cli/              # CLI 管理工具
│   ├── studio/           # Flutter Web 客户端
│   └── provider/         # FastAPI 服务端
├── manifests/            # 部署清单
│   ├── docker/           # Docker 部署配置
│   └── iac/              # 阿里云基础设施 (Terraform)
├── tests/                # 测试代码
└── examples/             # 示例与原型验证
```

## 快速开始

### 本地开发

```bash
# 启动客户端和服务端
docker compose -f manifests/docker/docker-compose.yml up --build

# 客户端: http://localhost:8080
# 服务端: http://localhost:9000
```

### CLI 工具

```bash
cd src/cli
uv run pytest -v        # 运行测试
ruff check .            # 代码检查
```

## 契约定义

项目使用 `.quanttide/` 目录定义所有契约：

| 文件 | 用途 |
|------|------|
| `.quanttide/asset/contract.yaml` | 资产组成、路径、类型 |
| `.quanttide/code/contract.yaml` | 编程规范、依赖、质量门禁 |

详见 [CONTRIBUTING.md](CONTRIBUTING.md)。

## 部署

使用 Terraform 部署到阿里云：

```bash
cd manifests/iac
terraform init
terraform apply
```

支持环境：`dev` / `staging` / `prod`

## 文档

- [商业需求](docs/brd/) - 业务痛点和动机
- [产品需求](docs/prd/) - 产品功能和设计
- [交互设计](docs/ixd/) - 界面布局和交互流程
- [架构设计](docs/add/) - 技术方案和架构决策
- [用户文档](docs/user/) - 安装和使用指南
