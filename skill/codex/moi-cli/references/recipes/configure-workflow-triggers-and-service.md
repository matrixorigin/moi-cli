# Recipe：配置 Workflow 触发器与 API 服务

适用：已经部署并实测的 Workflow 需要未来定时运行、Volume 新文件触发或 API 调用。
触发配置不是图节点，也不代表已经立即处理数据。先完成
[创建、同 App 更新和运行验收](create-and-run-workflow.md)，保存真实 App ID、当前版本和默认值。

## 配置 cron 并恢复手动模式

定时配置使用五字段或六字段表达式；需要明确时区时使用 `CRON_TZ=<IANA timezone>`：

```sh
moi-cli -o -no-interactive -workspace-id "$workspace_id" workflow update \
  --workflow-id "$workflow_id" --execution-mode cron \
  --cron-expression 'CRON_TZ=Asia/Shanghai 0 7 * * *'
moi-cli -o -no-interactive -workspace-id "$workspace_id" workflow get --workflow-id "$workflow_id"
```

核对 `execution_mode`、`cron_expression`、`trigger_summary`、保存默认值和新的
`latest_workflow_version_id`。配置成功不等于已到点执行；需要实际调度行为时，另按用户授权观察
新生 execution，不能用手动 run 冒充 cron 触发。

只需临时验证配置或不再需要调度时，恢复手动模式并再次回读：

```sh
moi-cli -o -no-interactive -workspace-id "$workspace_id" workflow update \
  --workflow-id "$workflow_id" --execution-mode one_shot
moi-cli -o -no-interactive -workspace-id "$workspace_id" workflow get --workflow-id "$workflow_id"
```

最终应看到 `execution_mode=one_shot`、`cron_expression` 为空（缺失或 null），且没有本轮配置
产生的非预期执行。
本轮 QA 在 cron 配置后恢复了 one_shot，并用最终 v5 手动执行验证图和默认值仍可用。

## 配置 Volume trigger

先用 `catalog volume-info --volume-id "$volume_id"` 确认真实 Volume 和授权范围。配置文件结构以
`moi-cli schema workflow deploy --argument volume-trigger-file` 为准：

```json
{"volume_id":12345,"enabled":false,"vars_json":"{\"text\":\"sample\",\"should_fail\":false}"}
```

将 `12345` 和值替换为本轮真实配置；`vars_json` 是 JSON 字符串，不是嵌套对象。禁用示例只用于
隔离配置验收，不能证明新文件会自动触发。正式启用前必须明确目标 Volume、文件范围和副作用。

```sh
moi-cli -o -no-interactive -workspace-id "$workspace_id" workflow update \
  --workflow-id "$workflow_id" --execution-mode volume_trigger \
  --volume-trigger-file "$volume_trigger_file"
moi-cli -o -no-interactive -workspace-id "$workspace_id" workflow get --workflow-id "$workflow_id"
```

核对 `execution_mode=volume_trigger`、`trigger_summary.volume_id` 和新版 ID。当前 QA 对
`enabled:false` 的回读只显示 `mode/configured/volume_id`，没有显式回显 `enabled`；不要把字段
缺席单独当成已禁用的充分证据。临时验收后立即恢复 one_shot；正式启用则需保留服务端明确的
启停证据并观察真实文件 execution。

Volume 新文件触发按文件分发，不能把父 dispatch 当普通 execution。实际运行时沿
[领域批处理规则](../domains/workflow.md#单次执行与-volume-批处理必须分开)展开子 execution，
逐个读取结果；配置回读和父 dispatch completed 都不是全部文件成功。

2026-11-17 生产隔离验收使用 `enabled:true` 的直接 Volume 挂载，而不是只保存 disabled
配置。上传新文件后，按返回的分发 ID 在 `list-runs` 中精确匹配 `dispatch_job_id`，再用
子 execution 读取 completed 和业务结果 42；随后恢复 one_shot。五字段和六字段 Cron 也在
同一轮分别保存回读。真实触发与 Cron 语法保存是不同证据，不能互相替代，也不能从一次
触发外推全部 Cron 边界或跨版本组合。

## 发布并调用 dynamic service

服务配置包含 `service_name`、JSON 字符串形式的 `input_schema` / `output_schema`、
`result_mode`，需要时再加 `runtime_spec_json`。同步等待结果使用 `oneshot`：

```json
{"service_name":"isolated-workflow-api","input_schema":"{}","output_schema":"{}","result_mode":"oneshot"}
```

用独立 MOWL 图部署：

```sh
moi-cli -o -no-interactive -workspace-id "$workspace_id" workflow deploy \
  --name "$workflow_name" --workflow-file "$workflow_file" \
  --execution-mode dynamic_service --dynamic-service-file "$service_file"
moi-cli -o -no-interactive -workspace-id "$workspace_id" workflow get --workflow-id "$workflow_id"
```

保存 App ID、服务版本 ID 和 `deployment.dynamic_service.invoke.body_example`。回读
`trigger_summary.service_name`、schema、执行方式和状态；服务名、App ID、definition ID 与
service version ID 是不同引用。

正式调用使用 dev CLI 的小写请求字段：

```json
{"service_name":"isolated-workflow-api","payload":{"purpose":"example"}}
```

```sh
moi-cli -o -no-interactive -workspace-id "$workspace_id" workflow invoke-workflow-dynamic-service \
  --json-file "$invoke_file"
```

需要指定版本时按部署回执的 body example 添加 `version_number`。检查返回 `status=COMPLETED`、实际 `result` 和
本次 `case_id`；部署 ready 或 schema 合法不等于调用成功。本轮 QA 已验证 dynamic service 回读
及 oneshot 调用，返回固定合成 payload；未覆盖异步 result mode、文件上传、schema 复杂约束和
高并发调用。
