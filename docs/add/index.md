# 数字资产云技术架构

## 设计理念

架构的核心目标是：**让意图可追溯，让变更可控制**。

### 契约即接口

契约是模块与外界的唯一约定。通过 Pydantic Schema 定义强类型模型，AI 和人类都不需要猜测配置格式：

```
ContractSchema
├── assets: Dict[str, AssetConfig]   # 资产声明
└── skills: Dict[str, SkillConfig]   # 技能定义
    └── version / entrypoint / params
```

所有 Schema 设置 `frozen=True`，传递一个明确信号：这是静态配置，运行时不可变。

### 分层通信

CLI 采用四层解耦设计，每层通过类型而非隐式行为通信：

```
cli.py（入口层）
    │ 解析命令行参数
    │ 调用 contract → workflow → file_operator
    │ 格式化输出
    ▼
contract.py（契约层）
    │ 自动向上查找 .quanttide/asset/contract.yaml
    │ 加载并验证为 ContractSchema（Pydantic）
    │ 对外只提供 get_skill() / get_asset()
    ▼
workflow.py（工作流层）
    │ 读取契约中的 SkillConfig
    │ 扫描输入目录
    │ 生成 Workflow { ArchiveTask }
    ▼
file_operator.py（操作层）
       逐文件移动（copy2 + unlink）
       失败自动回滚
       清理空源目录
```

分层原则：
- **入口层**不操作文件系统，只编排和展示
- **契约层**不依赖外部路径，自动查找契约文件
- **工作流层**不执行写操作，纯数据转换
- **操作层**不依赖 YAML 或 CLI 参数，只接收具体 `Path`

### 唯一事实源

所有测试数据、样例配置集中在 `assets/fixtures/`。其他位置不重复存储，通过 conftest.py fixture 引用。消除"该用哪份数据"的歧义。

### QA 即交付证明

质量保证文档采用「准则 → 判定 → 证据」结构。不记录测试过程，只回答一个问题：**我们证明了什么？**

## 当前实现（V1.0）

### 代码结构

```
src/cli/app/
├── cli.py           # 入口层：Typer 应用
├── contract.py      # 契约层：Contract + ContractSchema
├── workflow.py      # 工作流层：Workflow + ArchiveTask
└── file_operator.py # 操作层：archive_product

src/cli/tests/
├── conftest.py      # 统一引用 assets/fixtures
├── test_contract.py # 契约层测试
├── test_workflow.py # 工作流层测试
└── test_file_operator.py  # 操作层测试

assets/fixtures/     # 唯一事实源：样例数据
├── .quanttide/asset/contract.yaml  # 样例契约
└── docs/
    ├── journal/     # 样例日志
    └── archive/     # 样例归档
```

### 契约查找逻辑

`Contract.find_root()` 从当前目录向上遍历，直到找到 `.quanttide/asset/contract.yaml`。这使得 CLI 可以在项目任意子目录运行，无需指定契约路径。

## V2.0 规划：引擎化演进

### 核心原则

1. **核外挂载，内核稳定** — 内核只做资产建模、版本快照、事件路由。跨平台适配器（飞书、GitHub）作为独立插件
2. **中间标准 Schema** — 定义 ISDL（统一数字资产描述语言），版本控制的是 ISDL 而非底层存储格式
3. **事件驱动最终一致性** — 跨平台不做强一致同步，通过异步消息队列处理状态同步

### 演进路径

**Step 1: 契约驱动（✅ 已完成）**
- 契约定义从 YAML 而非硬编码
- Pydantic Schema 提供强类型验证

**Step 2: 抽取通用引擎（规划中）**
- 抽取 `engine.py`，支持可插拔的 transform 链
- 每个执行步骤记录结构化事件

**Step 3: 适配器抽象（规划中）**
- `LocalFSAdapter`（✅ 已验证）
- `FeishuAdapter`、`GithubAdapter`

**Step 4: 内容寻址版本控制（规划中）**
- 通过哈希值标识资产版本
- 支持 diff、回滚、fork
