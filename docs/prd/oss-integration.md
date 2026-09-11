# OSS 资产接入契约

> 状态：已落地并持续维护（初始决策 2026-08-15；当前复核 2026-09-11）
> 作者：郝子腾
> 日期：2026-08-15
> 关联：33 周会「对象存储可视化平台方案」提案 / BRD「数字资产全景图」/ PRD「资产发现」

## 一、目的与定位

本文档回答一个问题：**「对象存储（OSS）要不要接入 qtcloud-asset？接入后边界在哪？」**

结论先行：**以只读资产发现与全景展示为核心，不开放 OSS 对象上传、删除和权限变更。** 后续版本已补充账号权限、分享元数据写入、分享撤销和公开分享下载，但这些能力不改变 OSS 对象本身的只读边界。

本文档记录初始边界、冲突与决策点；实现已在后续阶段落地，当前状态以代码、CI 和 QA 记录为准。

## 二、背景：为什么要接

1. **BRD 三大场景之一「数字资产全景图」**：公司核心资产散落在各平台，关联关系存在人脑中，管理者看不清全局。
2. **PRD `asset_discovery`**：自动扫描和注册数字资产，识别依赖关系。跨平台自动注册和依赖关系建模仍未实现，本项目已完成 OSS 发现试点。
3. **33 周会提案**：当时对象存储主要依靠 CLI 操作，查看文件、上传资源不直观，非技术人员无法直接了解资源状态。
4. **张果（量潮科技群）审议意见**：建议不只考虑对象存储，把 GitHub、飞书一起考虑——「目前资产主要就是这三个平台」。

## 三、历史基线盘点（截至 2026-08-15）

### 3.1 项目契约登记的 OSS 桶（共 40 个，region 均为 oss-cn-hangzhou）

| 用途分类 | 桶名 | 数量 |
|---------|------|------|
| Studio（前端静态站点） | qtaccount-studio、qtadmin-studio、qtclass-studio、qtcloud-business-studio、qtcloud-data-studio、qtcloud-delib-studio、qtcloud-econ-studio、qtcloud-execute-studio、qtcloud-health-studio、qtcloud-human-studio、qtcloud-org-studio、qtcloud-product-studio、qtcloud-project-studio、qtcloud-secret-studio、qtcloud-studio、qtdata-studio、qthealth-studio、qtrecurit-studio | 18 |
| Private（私密数据） | qtadmin-private、qtclass-private、qtcloud-private、qtconsult-private、qtdata-private、qtrecruit-private | 6 |
| Site（站点） | qtbusiness-site、qtdocs-site、qtfounder-site、qthealth-site、qtrecurit-site、qtweb-site | 6 |
| Provider（后端） | qtadmin-provider | 1 |
| 其他 | qtcloud-learn-admin、qtcloud-learn-data、qtcloud-secret-data、quanttide-terraform-state | 4 |

命名规律：`{产品线}-{用途}`，用途四类 —— `-studio` / `-private` / `-site` / `-provider`。此规律可直接映射为资产目录结构。

### 3.2 初始发现：terraform 中的桶尚未创建

`manifests/iac/variables.tf` 定义了 `qtcloud-asset-studio`（默认值），但它不在当时的 40 个桶登记中。该桶已在后续基线核对中确认存在，并已作为当前 Studio 发布目标。

### 3.3 当前实现摘要

- Studio 正式入口为 `https://asset.cloud.quanttide.com`，发布桶为 `qtcloud-asset-studio`。
- Provider 使用 Go Custom Runtime，通过 `https://api.quanttide.com/qtcloud-asset` 对外提供服务。
- Provider 已提供 `viewer` / `admin` 权限、RDS 持久化、分享和分享下载能力。
- 平台真实 SSO 尚未接入，当前 SSO 入口仍是占位实现。

## 四、初始范围与边界

### 4.1 范围内（第一阶段，只读）

| 能力 | 说明 | 对应 PRD/提案 |
|------|------|---------------|
| 桶清单展示 | 登录后列出桶清单，含名称、region、存储类型、创建时间；`viewer` 隐藏 `-private` 和 `quanttide-terraform-state`，`admin` 查看全部 | 资产发现 / 提案「文件可视化查看」 |
| 桶内文件浏览 | 列对象（ListObjects），显示文件名、大小、修改时间 | 提案「文件可视化查看」 |
| 资产全景总览 | 按 `-studio/-private/-site/-provider` 分组聚合 | PRD `graph.md` 资产全景 |
| 资产目录建模 | 把桶/对象注册为契约资产条目 | PRD `asset_contract` |

### 4.2 初始范围外与当前状态

| 能力 | 原因 |
|------|------|
| OSS 对象上传 / 原始对象下载 / 删除 | 当前仍不开放；公开分享页的单文件下载和 ZIP 下载已在后续阶段实现 |
| 私密对象访问链接 / 有效期控制 | 明确不做；公开桶保留永久直链，私密桶只展示对象元数据 |
| 权限控制（管理员/内部/普通） | 当前已落地 `viewer` / `admin` 两级；平台真实 SSO 仍待接入 |
| CDN 刷新、静态网站部署 | 提案「后续扩展方向」，非第一版 |
| GitHub / 飞书适配器 | 张果建议，属第二阶段（`SourceAdapter` 已预留接口） |

### 4.3 铁律

1. **`viewer` 不展示 `-private` 桶和 `quanttide-terraform-state`**；`admin` 可查看这些桶的元数据，但任何角色都不能生成私密对象访问链接。
2. **不修改 OSS 对象**：不开放 OSS 对象上传、删除、权限变更；分享元数据的创建和撤销属于 Provider 业务数据操作。
3. **凭证不进代码库**：阿里云 AK/SK 走环境变量或密钥管理，禁止硬编码（呼应 `agent/contract.yaml` security_operations）。

## 五、历史冲突记录与决策

### 冲突 1：Provider 技术栈，契约与实现脱节

- 初始记录中，`.quanttide/code/contract.yaml` 声明 Provider 为 **Python + FastAPI + aliyun-oss2**。
- 初始记录中的实际代码已为 **Go**（`go.mod`、`internal/api/handler.go`）。
- 初始记录中的 `src/provider/README.md` 存在未解决的 git 冲突，且 Python 与 Go 两种口径并存。

**✅ 已决策（2026-08-15）**：采纳**方案 X**——以 Go 为准，修订 code 契约与 README（Go 有官方 `aliyun-oss-go-sdk`）。理由：实际代码已是 Go，改动最小。

### 冲突 2：OSS Adapter 与 `repository.SourceAdapter` 的关系

`internal/repository/repository.go` 已预留 `SourceAdapter` 接口（注释明确「implement for filesystem, GitHub, Feishu」），但未实现。OSS 不在其原始列举中。

**✅ 已决策（2026-08-15）**：OSS 适配器作为 `SourceAdapter` 的**第一个实现**（`OssAdapter`），纳入多源发现体系，为后续 GitHub/Feishu 适配器立好范式。

### 冲突 3：产品契约中的「范围外」

`.quanttide/product/contract.yaml` 的 `out_of_scope` 写了：
- 「跨平台适配器（飞书/GitHub 等）V2.0 规划，当前仅支持本地文件系统」。

**✅ 已决策（2026-08-15）**：明确表述为「资产发现的只读数据源试点」，不算全面启动跨平台适配器，避免与 V2.0 规划冲突。

## 六、历史落地路径与当前结果

以下阶段记录描述了当时的执行路线，不再表示「尚未实现」。

### 阶段一：契约对齐（先行）
1. 修订 `.quanttide/code/contract.yaml`：Provider 统一为 Go，补充 `aliyun-oss-go-sdk` 依赖。
2. 清理 `src/provider/README.md` 的 git 冲突。
3. 在 `.quanttide/asset/contract.yaml` 增加 `oss_buckets` 资产条目（40 个桶）。
4. 本文档由草案转为正式（合并进 `docs/prd/`）。

### 阶段二：只读接入（本次最小可跑）
1. `OssAdapter` 实现 `SourceAdapter.Discover()`，调用 ListBuckets / ListObjects。
2. Provider 新增 `GET /buckets`、`GET /buckets/{name}/objects` 只读端点。
3. Studio 首页渲染桶列表 + 分组总览。

### 阶段三：全景与多源（后续）
1. 桶/对象注册为契约资产，建立关联关系。
2. 接入 GitHub、飞书适配器，实现三平台资产全景。

## 七、历史风险与未决事项

| 事项 | 责任方 | 状态 |
|------|--------|------|
| Provider 技术栈统一（Go vs Python） | 数据工程部 | ✅ 已决策：以 Go 为准 |
| `-private` 桶和 `quanttide-terraform-state` 的可见范围 | Plan A 临时决策 | `viewer` 桶清单隐藏且不可访问对象级接口；`admin` 可查看全部桶并访问授权对象元数据 |
| 私密对象链接策略与权限分级 | 已决策 | 私密桶和 `quanttide-terraform-state` 不生成访问链接；公开桶使用永久直链 |
| 项目落地要素（负责人/排期/验收人） | 待周会决议 | 待明确 |
| 凭证管理方式（环境变量/密钥服务） | 数据工程部 | 待确认 |

## 八、初始决策记录

以下三点已于 2026-08-15 由郝子腾确认拍板：

1. **Provider 技术栈**：✅ 采纳「以 Go 为准，修订契约」的方案 X。
2. **接入边界**：✅ 认可「第一阶段只读、私密桶先不碰、不碰写操作」。
3. **落地顺序**：✅ 先做「契约对齐（阶段一）」，再进入只读接入（阶段二）。

当时的后续行动是进入**阶段一「基线与边界确认」**。该阶段已完成；当前未闭合事项包括平台真实 SSO 接入和阶段七分享功能的完整管理员生命周期验收。
