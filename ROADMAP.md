# 数字资产云路线图

## MVP 演进路线

### Phase 1: 本地优先的"资产Git"（已有验证）

- 形态：本地 Python 脚本
- 已有案例：
  - `backup_product_journal.py`：将产品日志从 journal 目录移动到 archive 目录，实现资产的本地版本化管理
  - `generate_product_roadmap.py`：读取产品日志，调用本地 Ollama 模型生成结构化蓝图
- 验证结论：本地脚本可以完成数字资产的采集、转换、归档全流程，资产形态为 Markdown 文件，通过 Git 管理版本
- 下一步：提供可视化界面，替代命令行操作

### Phase 2: 单向的"资产管道"（待验证）

- 形态：Phase 1 基础上加入"连接器"
- 功能：配置 Pipeline，把本地整理好的 JSON 任务清单一键推送到 GitHub Issues，或把本地 Markdown 发布到飞书文档
- 待验证：跨平台转换需求真实性，以及格式映射的损耗用户能否接受

### Phase 3: 双向同步的"云蓝图"（远期）

- 形态：SaaS 云平台
- 功能：双向同步、Webhook 监听平台变更、团队协作、云端版本库

## 待办

1. 跨平台连接器：实现飞书多维表格到 GitHub Issues 的单向同步
2. ISDL Schema 定义：验证中间标准Schema能否覆盖目标场景的数据结构
3. 格式映射损耗：验证跨平台转换后的格式丢失用户能否接受
4. 可视化界面：将现有 CLI 脚本升级为 TUI/GUI，降低使用门槛
