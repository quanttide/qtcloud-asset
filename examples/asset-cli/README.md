# asset CLI 参考实现（自 qtadmin 迁移）

qtadmin CLI 的 asset 职能模块参考实现（Rust），供 qtcloud-asset 产品线参考。

## 命令

| 命令 | 说明 |
|------|------|
| `archive` | 将 journal 日志归档（N 天前，默认 3），git 提交归档 |
| `quality` | 资产内容质量评估（叙事/知识/认知三维度，LLM 驱动，需 `DEEPSEEK_API_KEY`） |
| `status` | 资产结构合规检查（文件/格式/提交规范），失败退出码 1 |

## 结构

```
src/
├── main.rs            # 入口（qtcloud-asset 命令）
├── cli_config.rs      # profile 路径与 API key（QTRECURIT_PROFILE/DEEPSEEK_API_KEY）
└── asset/
    ├── mod.rs         # 命令分发
    ├── archive.rs     # journal 日志归档（582 行）
    ├── git_utils.rs   # git 工具（外部 git 命令）
    ├── quality.rs     # 质量评估（reqwest 调 DeepSeek API）
    └── status.rs      # 结构合规检查（1044 行）
```

## 与产品线 CLI 的定位差异

qtcloud-asset 产品线已有 CLI（`src/cli`：scanner/workflow/validate/render，资产发现与交付约束验证方向）。
本参考实现是 qtadmin 时代的「日志归档/合规检查/质量评估」工具，定位不同，作为历史能力参考。

## 验证

```bash
cargo build
cargo test    # 56 个测试
```

数据路径语义：`QTRECURIT_PROFILE`/`QTRECURIT_DATA` 环境变量，默认 `../../data/profile`（与 qtadmin 兼容）。
