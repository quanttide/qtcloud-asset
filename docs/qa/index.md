# 质量控制文档

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

### 测试优先级

1. **契约解析器**：验证 YAML 解析、schema 校验
2. **执行器**：状态机推进、事件记录
3. **适配器**：平台 API 调用（可 mock）
4. **端到端**：完整工作流

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
    contract = load_contract("generate_roadmap")
    assert contract["name"] == "产品日志生成路线图"
    assert contract["transform"]["action"] == "journal_to_roadmap"
```

### 路径配置测试

```python
def test_resolve_paths():
    contract = load_contract("generate_roadmap")
    paths = resolve_paths(contract)
    assert paths["journal"].exists()
```

## 执行器测试

### 状态机测试

```python
def test_state_transitions():
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
    engine.execute("contract_id", assets)
    events = get_events("contract_id")
    assert len(events) > 0
    assert events[0]["action"] == "execute"
```

## 适配器测试

### 本地文件系统适配器

```python
def test_local_fs_load():
    adapter = LocalFSAdapter()
    assets = adapter.load("docs/journal/product/qtcloud")
    assert len(assets) > 0
```

### LLM 适配器

```python
def test_llm_transform(mocker):
    mocker.patch("subprocess.run")
    adapter = LLMAdapter()
    result = adapter.transform(journal_content, params)
    assert result is not None
```

## 质量标准

| 指标 | 目标 |
|------|------|
| 代码覆盖率 | > 70% |
| 契约解析测试 | 必须通过 |
| 状态机测试 | 覆盖所有状态转换 |
| 适配器测试 | 每个适配器至少 3 个用例 |

## CI/CD 集成

```bash
# 测试命令
pytest tests/ -v --cov=src

# 契约校验
python -c "from contracts import validate; validate()"
```

## 测试数据管理

### fixtures 原则

- 使用最小化数据
- 不包含真实敏感信息
- 定期更新以反映契约变更

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

| 错误 | 原因 | 解决 |
|------|------|------|
| `ContractNotFound` | contracts.yaml 路径错误 | 检查 CONTRACTS_FILE 路径 |
| `PathNotFound` | journal 目录不存在 | 创建测试数据 |
| `LLMError` | llm CLI 不可用 | 安装 llm CLI |
