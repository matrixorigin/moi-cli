# Recipe：配置 Agent 运行环境

1. 发现可用 compute、model 和 provider backend；确认用户指定的运行环境和模型。
2. 对已存在资源先 get/detail，确认状态、owner、网络和可用 endpoint。
3. 只有用户明确要求时创建或更新 compute/backend/endpoint；保存所有返回 ID。
4. 用 `agent executor-health`、模型/router 查询和运行环境详情核验依赖状态。
5. 在 Agent 执行或 A2A 调用后，用 task/result 详情核验实际使用的环境和模型。
6. 任一依赖 pending、disabled 或权限拒绝时停止后续编排，报告缺失依赖和 ID。

不要把 backend probe、compute runtime 或 executor health 单独解释为 Agent 业务成功。
