# ROADMAP

## 飞书知识库递归下载

### 已完成

- [x] 问题定位：API 用错了，`/wiki/v2/spaces/:space_id/nodes` 才是获取子节点列表
- [x] 验证本地 lark-cli 命令（--help + --dry-run）

### 进行中

- [ ] 验证 API 实际可用性（--dry-run）
- [ ] 根据验证结果调整方案

### 待处理

- [ ] 实现递归下载脚本
- [ ] 处理快捷方式节点

详见 [docs/ROADMAP.md](docs/ROADMAP.md)