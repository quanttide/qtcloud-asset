# Provider 服务端设计

基于多源资产治理需求，Provider 采用分层架构设计，将 CLI 的本地治理能力扩展为服务端 API。

## 架构设计

### 分层结构

```
┌─────────────────────────────────────────┐
│              API Layer                  │
│  ┌─────────┐ ┌─────────┐ ┌──────────┐ │
│  │ Contract│ │Discovery│ │ Snapshot │ │
│  │  API    │ │  API    │ │   API    │ │
│  └────┬────┘ └────┬────┘ └────┬─────┘ │
└───────┼───────────┼───────────┼────────┘
        │           │           │
┌───────┴───────────┴───────────┴────────┐
│           Service Layer                │
│  ┌─────────┐ ┌─────────┐ ┌──────────┐ │
│  │ Contract│ │Discovery│ │ Snapshot │ │
│  │ Service │ │ Service │ │ Service  │ │
│  └────┬────┘ └────┬────┘ └────┬─────┘ │
└───────┼───────────┼───────────┼────────┘
        │           │           │
┌───────┴───────────┴───────────┴────────┐
│         Repository Layer               │
│  ┌─────────┐ ┌─────────┐ ┌──────────┐ │
│  │Filesystem│ │ GitHub  │ │  Feishu  │ │
│  │   Repo  │ │  Repo   │ │   Repo   │ │
│  └─────────┘ └─────────┘ └──────────┘ │
└─────────────────────────────────────────┘
```

### 各层职责

**API Layer**：HTTP 路由和请求处理
- Contract API：契约读取、解析、更新
- Discovery API：触发发现、查询发现状态
- Snapshot API：生成快照、查询历史

**Service Layer**：业务逻辑编排
- Contract Service：契约验证、版本管理
- Discovery Service：多源发现调度、结果聚合
- Snapshot Service：快照生成、一致性校验

**Repository Layer**：数据源适配
- Filesystem Repository：本地文件系统访问
- GitHub Repository：GitHub API 封装
- Feishu Repository：飞书 API 封装

## 核心模块设计

### 1. 契约管理模块

```python
# 读取项目契约
GET /api/v1/projects/{project_id}/contract

# 更新契约
PUT /api/v1/projects/{project_id}/contract

# 验证契约格式
POST /api/v1/projects/{project_id}/contract/validate
```

### 2. 发现模块

```python
# 触发发现任务
POST /api/v1/projects/{project_id}/discovery

# 查询发现状态
GET /api/v1/projects/{project_id}/discovery/{task_id}

# 获取发现结果
GET /api/v1/projects/{project_id}/discovery/{task_id}/results
```

发现任务支持多源并发：

```json
{
  "sources": ["filesystem", "github", "feishu"],
  "strategy": "parallel",
  "callback_url": "https://example.com/webhook"
}
```

### 3. 快照模块

```python
# 生成快照
POST /api/v1/projects/{project_id}/snapshots

# 查询快照列表
GET /api/v1/projects/{project_id}/snapshots

# 获取快照详情
GET /api/v1/projects/{project_id}/snapshots/{snapshot_id}

# 对比两个快照
POST /api/v1/projects/{project_id}/snapshots/compare
```

## 多源适配器设计

### 统一接口

```python
class SourceAdapter(ABC):
    @abstractmethod
    async def discover(self, asset_def: AssetDef) -> DiscoveredAsset:
        pass
    
    @abstractmethod
    async def validate(self, asset: DiscoveredAsset) -> ValidationResult:
        pass
    
    @abstractmethod
    async def get_version(self, uri: str) -> VersionInfo:
        pass
```

### 适配器实现

**Filesystem Adapter**
- 本地文件系统扫描
- 文件内容哈希计算
- 权限检查

**GitHub Adapter**
- GitHub API 调用
- Commit 信息获取
- Webhook 监听变更

**Feishu Adapter**
- 飞书开放平台 API
- 文档元数据获取
- 版本历史查询

## 数据模型

### Project（项目）

```python
class Project(BaseModel):
    id: str
    name: str
    contract_url: str  # 契约文件地址
    created_at: datetime
    updated_at: datetime
```

### Discovery Task（发现任务）

```python
class DiscoveryTask(BaseModel):
    id: str
    project_id: str
    status: TaskStatus  # pending / running / completed / failed
    sources: List[str]
    results: Dict[str, SourceResult]
    created_at: datetime
    completed_at: Optional[datetime]
```

### Snapshot（快照）

```python
class Snapshot(BaseModel):
    id: str
    project_id: str
    contract_version: str
    assets: Dict[str, AssetSnapshot]
    consistency_report: ConsistencyReport
    created_at: datetime
```

## 存储设计

### 元数据存储

- PostgreSQL：项目、任务、快照元数据
- Redis：任务状态缓存、分布式锁

### 快照存储

- 对象存储（OSS/S3）：快照文件长期存储
- 本地缓存：最近快照加速读取

### 契约存储

- 直接读取项目仓库中的 `contract.yaml`
- 支持 GitHub、GitLab 等代码托管平台

## 部署架构

```
┌─────────────────────────────────────────┐
│           阿里云函数计算 FC              │
│  ┌───────────────────────────────────┐  │
│  │        FastAPI Application        │  │
│  │  ┌─────────┐ ┌─────────┐ ┌─────┐ │  │
│  │  │ Contract│ │Discovery│ │Snap │ │  │
│  │  └─────────┘ └─────────┘ └─────┘ │  │
│  └───────────────────────────────────┘  │
└─────────────────────────────────────────┘
                   │
    ┌──────────────┼──────────────┐
    v              v              v
┌────────┐   ┌──────────┐   ┌──────────┐
│  RDS   │   │  Redis   │   │   OSS    │
│PostgreSQL│  │  Cache   │   │ Snapshot │
└────────┘   └──────────┘   └──────────┘
```

## 与 CLI 的协作

```
Developer                    Provider
    │                          │
    │  1. 编辑 contract.yaml   │
    │─────────────────────────>│
    │                          │
    │  2. 推送代码              │
    │─────────────────────────>│
    │                          │
    │  3. 触发发现              │
    │  POST /discovery         │
    │─────────────────────────>│
    │                          │
    │  4. 查询发现结果          │
    │  GET /discovery/{id}     │
    │<─────────────────────────│
    │                          │
    │  5. 生成快照              │
    │  POST /snapshots         │
    │─────────────────────────>│
    │                          │
    │  6. 获取快照报告          │
    │  GET /snapshots/{id}     │
    │<─────────────────────────│
```

## 扩展性设计

### 新增数据源

1. 实现 `SourceAdapter` 接口
2. 注册到 `AdapterRegistry`
3. 配置数据源参数

### 自定义验证规则

1. 实现 `ValidationRule` 接口
2. 注册到 `ValidationEngine`
3. 在契约中引用规则

### 插件机制

```python
# 发现后处理插件
class PostDiscoveryPlugin:
    async def process(self, result: DiscoveryResult) -> None:
        # 自定义处理逻辑
        pass
```
