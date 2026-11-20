# Provider Backend 领域

```sh
moi-cli -o -no-interactive -workspace-id "$workspace_id" provider-backend list --kind llm
moi-cli -o -no-interactive -workspace-id "$workspace_id" provider-backend get --kind llm --id "$backend_id"
moi-cli -o -no-interactive -workspace-id "$workspace_id" provider-backend list-endpoints --kind llm --backend-id "$backend_id"
moi-cli -o -no-interactive -workspace-id "$workspace_id" provider-backend router --kind llm
```

`kind` 必须是实际支持的 `llm|embedding|parser`。创建、更新、删除、endpoint 状态切换和
router 更新均是配置写入；地址、模型和认证参数不能写入报告或日志。probe 只证明探测结果，
不证明 Agent、Knowledge 或 Workflow 已能使用该 backend。

## 当前配置输入与运行环境范围

常规0.1.8 create只有kind/name，update只有kind/id，update-router只有kind；
地址、模型、认证或router策略等SDK options尚未接出。不能用空update证明已完成配置，
也不为验收改生产共用router。健康检查、配置保存和目标通道真实调用分别判断。

`agent resource-* --resource runtime-policy-profiles` 是正式工作区资源，返回可能包在Value中。
隔离策略可保持draft，更新元数据并独立回读version，最后改archived；该owner无delete方法。
不绑定或激活策略，就只能证明配置生命周期，不能证明预算、网络或数据限制已实际强制。
SDK的Sandbox/Runner及网络接口属于另一根作用域，常规CLI未发现相应命令；
workspace resource命令不能代替，不拼路径穿越或raw HTTP绕行。
