# 数字资产云路线图

## MVP 演进路线

### Phase 1: 本地优先的"资产Git"（已完成）

- 形态：本地 Python 脚本
- 已有案例：
  - `backup_product_journal.py`：将产品日志从 journal 目录移动到 archive 目录，实现资产的本地版本化管理
  - `generate_product_roadmap.py`：读取产品日志，调用本地 Ollama 模型生成结构化蓝图
- 验证结论：本地脚本可以完成数字资产的采集、转换、归档全流程，资产形态为 Markdown 文件，通过 Git 管理版本
- 文档沉淀：
  - `docs/prd/index.md`：产品蓝图
  - `docs/prd/contract.md`：契约产品需求
  - `docs/prd/workflow.md`：工作流产品需求
  - `docs/add/index.md`：技术架构
  - `docs/add/contract.md`：契约技术设计（三大工程公理、生命周期状态机、契约即证据）
  - `docs/add/workflow.md`：工作流技术设计（ISDL 结构、适配器层、事件溯源）

### Phase 2: 单向的"资产管道"（进行中）

- 形态：Phase 1 基础上加入"连接器"
- 功能：配置 Pipeline，把本地整理好的 JSON 任务清单一键推送到 GitHub Issues，或把本地 Markdown 发布到飞书文档
- 当前进展：
  - ISDL 核心结构已定义（assets、transforms、triggers、lifecycle）
  - 适配器层设计已完成（能力协商、插件化）
  - 契约三大工程公理已确立
- 待验证：跨平台转换需求真实性，以及格式映射的损耗用户能否接受

### Phase 3: 双向同步的"云蓝图"（远期）

- 形态：SaaS 云平台
- 功能：双向同步、Webhook 监听平台变更、团队协作、云端版本库

## 待办

1. 跨平台连接器：实现飞书多维表格到 GitHub Issues 的单向同步（Phase 2 核心）
2. 格式映射损耗：验证跨平台转换后的格式丢失用户能否接受
3. 可视化界面：将现有 CLI 脚本升级为 TUI/GUI，降低使用门槛
4. 契约引擎原型：基于 ISDL 实现最小可执行引擎（解析、校验、执行、事件记录）
