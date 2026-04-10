# CLI 设计文档契约

## 模块定位

cli.py、planner.py、file_operator.py 的设计文档

## 资产定义

| 资产 | 类型 | 分类 | 路径 | 描述 |
|------|------|------|------|------|
| CLI 设计文档 | docs | cli_docs | src/cli/docs | cli.py、planner.py、file_operator.py 的设计文档 |
| CLI 实现代码 | code | cli | src/cli/app | 三层解耦架构：cli.py、planner.py、file_operator.py |
| CLI 测试代码 | code | cli_tests | src/cli/tests | planner 和 file_operator 的单元测试 |

## 资产详情

### CLI 设计文档 (cli_docs)

- **类型**: docs
- **分类**: cli_docs
- **路径**: `src/cli/docs`
- **描述**: cli.py、planner.py、file_operator.py 的设计文档

### CLI 实现代码 (cli_code)

- **类型**: code
- **分类**: cli
- **路径**: `src/cli/app`
- **描述**: 三层解耦架构

### CLI 测试代码 (cli_tests)

- **类型**: code
- **分类**: cli_tests
- **路径**: `src/cli/tests`
- **描述**: planner 和 file_operator 的单元测试

## 三层架构

```
cli.py（入口层）
    │  解析命令行参数
    │  调用 planner → file_operator
    │  格式化输出
    ▼
planner.py（配置层）
    │  加载 contracts.yaml
    │  扫描目录
    │  生成 Workflow
    ▼
file_operator.py（操作层）
    逐文件操作
    失败自动回滚
```

## CLI 接口

```
qtcloud-asset --input=<源> --contract=<契约> --output=<目标>
```

| 参数 | 说明 |
|------|------|
| `-i, --input` | 数据源目录 |
| `-c, --contract` | 契约名称 |
| `-o, --output` | 输出目标目录 |
| `-p, --pattern` | 文件匹配模式（默认 `*.md`） |
| `-n, --dry-run` | 预览模式 |
| `-v, --verbose` | 详细输出 |
