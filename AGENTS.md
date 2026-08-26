# AGENTS.md

## 契约事实源

`.quanttide/` 目录是项目的契约事实源，所有资产和代码规则以这里为准。

| 文件 | 用途 |
|------|------|
| `.quanttide/asset/contract.yaml` | 资产组成、路径、类型 |
| `.quanttide/code/contract.yaml` | 编程规范、依赖、质量门禁 |
| `.quanttide/agent/contract.yaml` | AI 执行审核规则 |

**做任何变更前，先查阅契约文件。** 实际项目结构必须与契约一致。

## AI 执行规则

AI 助手在执行操作前，必须严格遵守 `.quanttide/agent/contract.yaml` 中的审核规则：

1. **每次对话开始时**，AI 必须读取 `.quanttide/agent/contract.yaml`
2. **分析用户请求**，识别是否涉及需要审核的操作
3. **列出操作清单**，标注每个操作的风险等级
4. **等待用户确认**，只有用户明确同意后才能执行高风险操作
5. **执行后反馈**，简要报告执行结果

详见 [`.quanttide/agent/contract.yaml`](.quanttide/agent/contract.yaml)

## 文档使用流程

详见 [CONTRIBUTING.md](CONTRIBUTING.md)。

文档各司其职：

| 文档 | 回答的问题 |
|------|-----------|
| `docs/brd/` | 为什么存在业务问题 |
| `docs/prd/` | 产品如何解决问题 |
| `docs/ixd/` | 用户如何与产品交互 |
| `docs/add/` | 技术架构是什么 |
| `docs/qa/` | 质量决策和记录 |
| `docs/user/` | 用户如何使用 |

## 工作原则

1. 契约优先：契约定义了项目应该有什么，缺失的资产需要补上
2. 变更同步：修改契约后，确保实际文件/目录已同步
3. 流程遵循：任何文档写作和维护流程遵循 CONTRIBUTING.md

## 项目执行章程

完整的分阶段开发、部署和验收计划见 [plan.md](./plan.md)。任何开发任务都应先判断所属阶段，再执行对应的阶段门禁。

### 部署边界

- 本项目不使用 Docker Compose、Dockerfile 或 Kubernetes 镜像作为当前部署路径。
- Studio 使用 Flutter Web 构建静态文件，发布到 OSS 静态站点。
- Provider 使用 Go 编译产物，以阿里云函数计算 Custom Runtime 运行。
- 平台级 VPC、资源组、共享 API 网关和证书遵循 `quanttide-platform` 的管理边界；本项目只管理应用级资源和发布配置。
- `quanttide-platform/manifests/terraform` 只作为系统级共享层使用；本项目通过 `terraform_remote_state` 或平台已导出的 `outputs` 引用 `vpc_id`、`vswitch_id`、`security_group_id`、`rds_instance_id`、`rds_connection_string`、`rds_port` 和 `resource_group_id`，不在本仓库重新定义这些共享资源。
- 平台仓库的 API 网关 Terraform 只负责网关分组和 `api.quanttide.com` 域名绑定；路由定义、DNS 记录和证书绑定按平台仓库脚本与 CI 分工处理，本项目只消费已约定好的公网入口。
- `http://127.0.0.1:8090/` 仅用于 Flutter Web 本地开发验证，不得作为生产 API 或生产页面地址。

### 开发顺序

1. 先读取 `.quanttide/agent/contract.yaml`、`.quanttide/asset/contract.yaml` 和 `.quanttide/code/contract.yaml`。
2. 先核对真实执行路径，再修改代码或配置；入口、调用方、部署清单和文档必须一起检查。
3. 优先修复上线阻断、安全边界和配置漂移，再开发非必要功能。
4. 小范围修改，避免覆盖用户已有的未提交变更、生成物或日志。
5. 修改后必须运行与变更范围匹配的测试、静态检查或构建验证。

### 配置与环境

- 本地 Provider 默认监听 `9000`，Studio 本地页面可监听 `8090`。
- Studio 的 Provider 地址必须支持环境配置；生产环境不得硬编码 `127.0.0.1`。
- Provider 使用 `PROVIDER_PORT`、`PROVIDER_BASE_URL`、`OSS_ENDPOINT`、`OSS_ACCESS_KEY_ID` 和 `OSS_ACCESS_KEY_SECRET`。
- 凭证只能通过本地安全配置、CI Secret、函数计算环境变量或 RAM 角色注入，不得写入源码、文档、静态产物、日志或 Git 历史。
- CORS 来源只允许已登记的本地开发地址和正式 Studio 域名。
- 生产 API 应通过平台共享网关访问，不在前端暴露函数计算内部地址。

### OSS 与权限

- 第一版默认只读：桶和对象元数据可以展示，上传、下载、删除和权限变更不属于默认范围。
- `-private` 桶默认不暴露内容。
- 任何对象访问链接能力都必须经过权限、有效期、审计和错误暴露评审。
- Provider 的 RAM 权限遵循最小权限原则，只授予实际需要的 OSS 读取和签名能力。

### 函数计算发布

- 函数启动配置必须与 `src/provider/cmd/provider/main.go` 的 Go 入口一致。
- 不得把 Python、FastAPI、Uvicorn 或旧的 `py311` 包路径当作当前 Provider 的默认发布方式。
- 发布前先验证函数原生健康检查，再接入 API 网关。
- 发布包应使用明确的版本或提交标识，保留上一版本以支持回滚。

### 文档与契约同步

- 修改目录、语言、框架、入口、部署方式或 API 边界后，必须同步检查 `.quanttide/`、README、部署清单和 QA 证据。
- 文档中的文件路径、测试名称和命令必须能在当前工作树中定位。
- 契约声明的路径不存在时，应先记录为阻断项，不得用“文档状态”掩盖实际缺失。
- `plan.md` 只记录当前认可的执行路线；废弃路线应明确标记，不与当前命令混用。
- 涉及 `quanttide-platform/manifests/terraform` 时，必须确认它当前导出的共享输出和职责边界，再决定本仓库是否只引用其 `outputs`，不要把平台共享资源、网关分组、DNS 或证书当成本仓库的实现目标。
- 涉及 API 网关时，必须核对平台仓库 `api-gateway.tf`、`outputs.tf`、`scripts/api-gateway/deploy.sh`、`dns.py` 和 CI 证书绑定流程的真实分工，再同步到计划和文档。

### 验证与发布门禁

- 本地验证至少覆盖 Provider `/health`、Studio 页面加载、桶列表和错误状态。
- 线上验证至少覆盖 HTTPS、API 网关路由、CORS、桶列表、对象列表和私密桶策略。
- 发布前检查 `git status`，确认没有误提交凭证、二进制、Flutter 构建目录、临时日志或调试产物。
- 任何生产部署、DNS、证书、权限、CI/CD、远程仓库或共享基础设施变更，都必须先列出操作清单并取得明确授权。
