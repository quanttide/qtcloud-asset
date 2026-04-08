# 数字资产工作流 DSL

## ISDL Lite v1.0

V1.0 采用精简设计，专注核心需求跨平台转换和版本控制。契约结构包含 name 契约名称、version 契约版本号、asset 源资产定义（必选，包含 platform 平台标识、type 资源类型、schema.required 必填字段列表）、transform 转换规则（必选，包含 input 输入资产、output 输出资产、action 转换动作名、params 转换参数可选）、trigger 触发方式默认 manual、lifecycle 生命周期配置可选。

状态流转为 draft 到 active 到 archived。draft 草稿可编辑，active 发布执行转换，archived 归档保留历史。

触发方式包括 manual 手动触发 V1.0 默认、cron 定时触发 V2.0、event 事件触发 V2.0。适配器 V1.0 硬编码支持两种平台，feishu 支持 feishu_to_markdown，github 支持 github_upload。执行日志记录 contract_id、version、timestamp、action、result、duration_ms。

## V2.0 扩展

V2.0 规划多资产支持 assets 数组、多转换支持 transforms 链式执行、事件溯源 contract_events 和 execution_events、适配器层平台能力协商、健康检查平台 API 可达性检查、错误处理重试 DLQ 降级。

## 目录结构

docs/assets 下按 platform、asset_type、version 组织，包含 pending 目录存放 draft 状态、active 目录存放 active 状态、archive 目录存放 archived 状态。