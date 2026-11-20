# Recipe：检查 Agent 自动化规则

1. 用 `agent list-callback-rules` 或 `list-channel-instances` 发现真实 rule/instance ID。
2. 对用户指定对象执行 get 详情，确认 provider、状态、触发类型和目标 Agent。
3. 用户明确要求变更时，先读取 `schema agent create-callback-rule` 或对应 update 命令并校验 JSON。
4. 创建/更新后重新 get 详情，确认服务端版本和状态。
5. 需要测试时执行专用 test 命令，并把测试结果与 rule/instance ID 关联。
6. 自动化历史和实际触发仍需通过对应产品界面或任务详情核验；没有 CLI 终态时报告限制。

不要自动启用、停用、删除规则或渠道，也不要用一个成功的连接测试代替实际事件触发验证。
