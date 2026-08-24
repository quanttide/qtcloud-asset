# OSS 资产接入契约

> 状态：已决策（2026-08-15），进入阶段一「契约对齐」
> 作者：郝子腾
> 日期：2026-08-15
> 关联：33 周会「对象存储可视化平台方案」提案 / BRD「数字资产全景图」/ PRD「资产发现」

## 一、目的与定位

本文档回答一个问题：**「对象存储（OSS）要不要接入 qtcloud-asset？接入后边界在哪？」**

结论先行：**要接，且只做「只读资产发现与全景展示」这一小步，不做写操作。** 它是 PRD 中 `asset_discovery`（资产发现）能力的第一次真实落地，也是 33 周会提案的第一版交付。

本文档定义边界、冲突与决策点；实现代码在阶段二之后另行编写。

## 二、背景：为什么要接

1. **BRD 三大场景之一「数字资产全景图」**：公司核心资产散落在各平台，关联关系存在人脑中，管理者看不清全局。
2. **PRD `asset_discovery`**：自动扫描和注册数字资产，识别依赖关系。当前 `components: []`，尚未实现。
3. **33 周会提案**：对象存储目前靠 CLI 操作，查看文件、上传资源不直观，非技术人员无法直接了解资源状态。
4. **张果（量潮科技群）审议意见**：建议不只考虑对象存储，把 GitHub、飞书一起考虑——「目前资产主要就是这三个平台」。

## 三、现状盘点（截至 2026-08-15）

### 3.1 项目契约登记的 OSS 桶（共 40 个，region 均为 oss-cn-hangzhou）

| 用途分类 | 桶名 | 数量 |
|---------|------|------|
| Studio（前端静态站点） | qtaccount-studio、qtadmin-studio、qtclass-studio、qtcloud-business-studio、qtcloud-data-studio、qtcloud-delib-studio、qtcloud-econ-studio、qtcloud-execute-studio、qtcloud-health-studio、qtcloud-human-studio、qtcloud-org-studio、qtcloud-product-studio、qtcloud-project-studio、qtcloud-secret-studio、qtcloud-studio、qtdata-studio、qthealth-studio、qtrecurit-studio | 18 |
| Private（私密数据） | qtadmin-private、qtclass-private、qtcloud-private、qtconsult-private、qtdata-private、qtrecruit-private | 6 |
| Site（站点） | qtbusiness-site、qtdocs-site、qtfounder-site、qthealth-site、qtrecurit-site、qtweb-site | 6 |
| Provider（后端） | qtadmin-provider | 1 |
| 其他 | qtcloud-learn-admin、qtcloud-learn-data、qtcloud-secret-data、quanttide-terraform-state | 4 |

命名规律：`{产品线}-{用途}`，用途四类 —— `-studio` / `-private` / `-site` / `-provider`。此规律可直接映射为资产目录结构。

### 3.2 关键发现：terraform 中的桶尚未创建

`manifests/iac/variables.tf` 定义了 `qtcloud-asset-studio`（默认值），但它不在当前 40 个桶的契约登记中。说明资产云自己的 Studio 桶仍需在发布前确认，不能把默认变量当作已存在资产。

## 四、范围与边界（本次草案的核心）

### 4.1 范围内（第一阶段，只读）

| 能力 | 说明 | 对应 PRD/提案 |
|------|------|---------------|
| 桶清单展示 | 列出 40 个桶，含名称、region、存储类型、创建时间 | 资产发现 / 提案「文件可视化查看」 |
| 桶内文件浏览 | 列对象（ListObjects），显示文件名、大小、修改时间 | 提案「文件可视化查看」 |
| 资产全景总览 | 按 `-studio/-private/-site/-provider` 分组聚合 | PRD `graph.md` 资产全景 |
| 资产目录建模 | 把桶/对象注册为契约资产条目 | PRD `asset_contract` |

### 4.2 范围外（明确不做，留待后续决策）

| 能力 | 原因 |
|------|------|
| 文件上传 / 下载 / 删除 | 涉及写操作与权限，周会决议「待明确」 |
| 访问链接生成 / 有效期控制 | 刘婧怡、涂雅芳提出，权限方案未定 |
| 权限控制（管理员/内部/普通） | 周会决议待明确，涉及私密桶安全 |
| CDN 刷新、静态网站部署 | 提案「后续扩展方向」，非第一版 |
| GitHub / 飞书适配器 | 张果建议，属第二阶段（`SourceAdapter` 已预留接口） |

### 4.3 铁律

1. **`-private` 桶只读展示元数据，不暴露内容**（或按后续权限方案再放开）。
2. **不动写操作**：第一阶段任何写/删除/改权限能力一律不实现。
3. **凭证不进代码库**：阿里云 AK/SK 走环境变量或密钥管理，禁止硬编码（呼应 `agent/contract.yaml` security_operations）。

## 五、与项目现状的冲突（必须先行决策）

### 冲突 1：Provider 技术栈，契约与实现脱节

- `.quanttide/code/contract.yaml` 声明 Provider 为 **Python + FastAPI + aliyun-oss2**。
- 实际代码为 **Go**（`go.mod`、`internal/api/handler.go`）。
- `src/provider/README.md` 存在**未解决的 git 冲突**（`<<<<<<< HEAD` / `>>>>>>> origin/main`），且 README 内 Python 与 Go 两种口径并存。

**✅ 已决策（2026-08-15）**：采纳**方案 X**——以 Go 为准，修订 code 契约与 README（Go 有官方 `aliyun-oss-go-sdk`）。理由：实际代码已是 Go，改动最小。

### 冲突 2：OSS Adapter 与 `repository.SourceAdapter` 的关系

`internal/repository/repository.go` 已预留 `SourceAdapter` 接口（注释明确「implement for filesystem, GitHub, Feishu」），但未实现。OSS 不在其原始列举中。

**✅ 已决策（2026-08-15）**：OSS 适配器作为 `SourceAdapter` 的**第一个实现**（`OssAdapter`），纳入多源发现体系，为后续 GitHub/Feishu 适配器立好范式。

### 冲突 3：产品契约中的「范围外」

`.quanttide/product/contract.yaml` 的 `out_of_scope` 写了：
- 「跨平台适配器（飞书/GitHub 等）V2.0 规划，当前仅支持本地文件系统」。

**✅ 已决策（2026-08-15）**：明确表述为「资产发现的只读数据源试点」，不算全面启动跨平台适配器，避免与 V2.0 规划冲突。

## 六、落地路径（分阶段，非本次实现）

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

## 七、风险与未决事项

| 事项 | 责任方 | 状态 |
|------|--------|------|
| Provider 技术栈统一（Go vs Python） | 数据工程部 | ✅ 已决策：以 Go 为准 |
| `-private` 桶的可见范围 | 待周会决议 | 待明确（当前策略：不碰） |
| 访问链接有效期与权限分级 | 刘婧怡/涂雅芳提出 | 待明确（当前策略：不做） |
| 项目落地要素（负责人/排期/验收人） | 待周会决议 | 待明确 |
| 凭证管理方式（环境变量/密钥服务） | 数据工程部 | 待确认 |

## 八、决策记录

以下三点已于 2026-08-15 由郝子腾确认拍板：

1. **Provider 技术栈**：✅ 采纳「以 Go 为准，修订契约」的方案 X。
2. **接入边界**：✅ 认可「第一阶段只读、私密桶先不碰、不碰写操作」。
3. **落地顺序**：✅ 先做「契约对齐（阶段一）」，再进入只读接入（阶段二）。

后续行动：进入**阶段一「基线与边界确认」**，核对 Go Custom Runtime、生产域名、平台共享资源边界和 40 个桶资产登记。
