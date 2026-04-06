# 质量控制文档

## 设计意图

本系统的核心设计是 **Contract 驱动的、松耦合的转换管道**：
- 契约（Contract）是唯一的 Single Source of Truth
- 转换过程通过适配器实现平台解耦
- 状态机确保工作流的可控性与审计追踪

测试策略的设计目的，就是**通过测试来保障和验证这些设计**。

## 测试策略

### 测试金字塔

```
       ╱╲
      ╱  ╲        E2E 测试（少量，覆盖核心流程）
     ╱────╲
    ╱      ╲      集成测试（验证组件协作）
   ╱────────╲
  ╱          ╲    单元测试（核心逻辑）
 ╱────────────╲
```

**为何如此分层：**
- 单元测试确保 `Contract Parser`、`State Machine` 等核心单元的健壮性 — **设计的基石**
- 集成测试验证 `Parser -> Engine -> Adapter` 的协作 — **设计的关键集成点**
- 少量 E2E 测试验证从契约定义到资产生成的全流程 — **设计的端到端价值体现**

### 测试优先级

1. **契约解析器**：验证 YAML 解析、schema 校验 — **系统入口必须可靠**
2. **执行器**：状态机推进、事件记录 — **核心业务流程必须完整无误**
3. **适配器**：平台 API 调用（可 mock）— **确保新平台能通过标准接口集成**
4. **端到端**：完整工作流 — **验证设计的端到端价值**

## 测试结构

```
tests/
├── unit/                 # 单元测试
│   ├── test_contract_parser.py
│   └── test_state_machine.py
├── integration/          # 集成测试
│   ├── test_transform.py
│   └── test_adapters.py
├── e2e/                  # 端到端测试
│   └── test_full_workflow.py
└── fixtures/             # 测试数据
    ├── contracts/        # 契约 YAML 示例
    └── assets/           # 资产数据
```

## 契约测试

### 契约解析测试

```python
def test_parse_contract():
    """
    验证契约的核心配置项能被正确解析。
    设计契约：transform.action 是驱动后续所有转换行为的关键约定。
    """
    contract = load_contract("generate_roadmap")
    assert contract["name"] == "产品日志生成路线图"
    assert contract["transform"]["action"] == "journal_to_roadmap"
```

### 路径配置测试

```python
def test_resolve_paths():
    """
    验证系统对契约中路径约定的处理能力。
    设计假设：系统依赖于外部（文件系统）的特定结构。
    """
    contract = load_contract("generate_roadmap")
    paths = resolve_paths(contract)
    assert paths["journal"].exists()
```

## 执行器测试

### 状态机测试

```python
def test_state_transitions():
    """
    验证执行引擎的状态机设计。
    设计契约：契约提交后状态必须为 'pending_review'，审批后必须转为 'active'。
    这确保了工作流的可控性与审计追踪能力。
    """
    # draft -> pending_review
    engine.submit("contract_id")
    assert state == "pending_review"
    
    # pending_review -> active
    engine.approve("contract_id")
    assert state == "active"
```

### 事件溯源测试

```python
def test_event_logging():
    """
    验证事件驱动的最终一致性设计。
    通过回放事件重建状态，是审计和故障恢复的基础。
    """
    engine.execute("contract_id", assets)
    events = get_events("contract_id")
    assert len(events) > 0
    assert events[0]["action"] == "execute"
```

## 适配器测试

### 本地文件系统适配器

```python
def test_local_fs_load():
    """
    验证 LocalFSAdapter 对契约路径约定的实现。
    确保资产建模、版本快照、事件路由的内核设计能落地。
    """
    adapter = LocalFSAdapter()
    assets = adapter.load("docs/journal/product/qtcloud")
    assert len(assets) > 0
```

### LLM 适配器

```python
def test_llm_transform(mocker):
    """
    验证 LLM 适配器的标准接口。
    确保 AI transform 能通过统一接口集成，且错误可隔离。
    """
    mocker.patch("subprocess.run")
    adapter = LLMAdapter()
    result = adapter.transform(journal_content, params)
    assert result is not None
```

## 质量标准

| 指标 | 目标 | 保障的设计质量 |
|------|------|----------------|
| 代码覆盖率 | > 70% | 可维护性、变更信心 — 核心逻辑变更可通过测试快速反馈 |
| 契约解析测试 | 必须通过 | 健壮性、可用性 — 系统入口必须可靠 |
| 状态机测试 | 覆盖所有状态转换 | 正确性、可追溯性 — 核心业务流程必须完整且无误 |
| 适配器测试 | 每个适配器至少 3 个用例 | 可扩展性、隔离性 — 确保新平台适配器能通过标准接口集成 |

## CI/CD 集成

```bash
# 测试命令
pytest tests/ -v --cov=src

# 契约校验
python -c "from contracts import validate; validate()"
```

**质量门禁设计**：每次合并请求时运行，强制执行"质量门禁"原则，确保不符合测试标准和覆盖率的代码无法进入主分支。

## 测试数据管理

### fixtures 原则

- **使用最小化数据**：为了测试效率和聚焦问题，符合单元测试的设计理念
- **不包含真实敏感信息**：保障测试资产的安全性和可共享性，使测试套件能在任何环境安全运行
- **目录结构映射真实契约**：模拟生产环境的资产组织方式，使集成测试更贴近真实场景

### 目录结构

```
tests/fixtures/
├── contracts/
│   ├── generate_roadmap.yaml
│   └── journal_backup.yaml
└── assets/
    └── journal/
        └── product/
            └── qtcloud/
                └── 2026-04-06.md
```

## 故障排查

### 常见测试失败

| 错误 | 原因 | 设计关联 |
|------|------|----------|
| `ContractNotFound` | contracts.yaml 路径错误 | 系统入口依赖契约配置，必须正确配置 |
| `PathNotFound` | journal 目录不存在 | 系统依赖外部文件系统的特定结构，环境需匹配 |
| `LLMError` | llm CLI 不可用 | 适配器层需要外部工具可用，需确保依赖安装 |
