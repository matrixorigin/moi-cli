# Workflow 领域

## 发现和详情

```sh
moi-cli -o -no-interactive -workspace-id "$workspace_id" workflow list
moi-cli -o -no-interactive -workspace-id "$workspace_id" workflow get --workflow-id "$workflow_id"
```

从用户选择或详情返回中取得 workflow ID，检查是否已配置、可运行和需要哪些输入。名称不是
execution ID，也不能作为运行参数。

若 `get` 报 `unknown field "session_id"`，这是该次 CLI 对详情响应的解码失败，不能解释为
未登录、App 不存在或运行失败。保留原错误；可以使用已确认的部署收据、`versions`、
`latest-published-version` 和执行绑定交叉核对。只有这些证据足够确认本次版本、输入及依赖时
才继续已授权运行；否则停止，不重建旧资源规避详情问题。

## 运行和结果

```sh
moi-cli -o -no-interactive -workspace-id "$workspace_id" workflow run --workflow-id "$workflow_id"
moi-cli -o -no-interactive -workspace-id "$workspace_id" workflow get-run --workflow-id "$workflow_id" --execution-id "$execution_id"
moi-cli -o -no-interactive -workspace-id "$workspace_id" workflow get-run-result --workflow-id "$workflow_id" --execution-id "$execution_id"
```

单次执行使用真实 `execution_id` 查询状态和结果。结果中的 `case_result`、`case_error`、
`node_states` 必须与本次 execution 关联；只有 completed 且业务输出符合断言才能报告成功。

## 单次执行与 Volume 批处理必须分开

先读取配置和默认输入，判断本次是单次执行还是 Volume 文件分发。**不能仅凭返回字段名或
`execution_mode=one_shot` 分类。** 生产实测中，Volume 批处理也返回
`workflow_run.execution_id`，其值实际是分发任务 ID，直接传给 `get-run*` 会返回 404。
这不证明文件没处理，也不能据此重跑。

批处理追踪使用正式 CLI：

```sh
moi-cli -o -no-interactive -workspace-id "$workspace_id" workflow list-runs --workflow-id "$workflow_id"
```

保存原始返回字段和值，在 `executions[]` 中按 `dispatch_job_id == run 返回的 ID` 精确匹配。
每条匹配记录的 `execution_id` 才是可用于 `get-run`、`get-run-result` 和执行控制的文件执行 ID。
同时核对 `file_id`、`moi_workflow_version_id`、`moi_case_id`，不按名称、列表顺序或最近时间
猜测归属，也不发明统一的 `file_execution_id` 服务字段。

有界等待期间若尚无匹配记录，保留“已提交，尚未观察到子执行”，不要重新提交。列表分页时
按当前 help 读取足够范围；无法穷尽则报告观察限制。`list-file-executions` 当前不接筛选参数，
返回 `{}` 不能证明没有执行；`list-run-import-files` 也必须使用真实子 execution。

批处理成功至少要求：实际输入 Volume 在运行前非空；已知输入文件集合与本轮子执行一致；
每个文件完成或有明确的跳过/失败记录；结果、产物及 Knowledge source 沿 file ID 和本轮 case
关联。只有父任务状态、历史成功或 source 数量不构成本轮成功。
失败文件重试须在授权内，`batch-retry-failed-files` 使用已证实的分发任务 ID 作为
`--source-dispatch-job-id`，并传失败子 execution ID；不传原文件、case 或成功文件 ID。

## 创建、版本与部署指向

Agent 会话编写、部署和实测原生 MOWL 使用普通 `workflow deploy --workflow-file`，不经过
Workflow 页面的候选接受流程。图、运行字段、默认值和测试输入的写法见
[创建与运行 Recipe](../recipes/create-and-run-workflow.md)。YAML DSL 仍是另一种定义格式，
`--dsl-yaml` 接收内容而非文件路径，且与 `--workflow-file` 互斥；不要把两种格式混用。
解析档位、附件、音视频、清洗和 XML 的结果断言见
[解析节点核验 Recipe](../recipes/verify-parser-workitems.md)。

部署返回的 `workflow.id` / `deployment.workflow_app_id` 是后续 `get`、`update`、`run` 和
`deploy --workflow-id` 使用的 App ID。`deployment.workflow_def_id` 与
`deployment.workflow_version_id` 分别是定义和版本身份，不能互相替代。

修改已保存 MOWL 时，先 `workflow get` 保存原 App、版本、字段、默认值和触发配置，再用同一个
App ID 部署新版：

```sh
moi-cli -o -no-interactive -workspace-id "$workspace_id" workflow deploy \
  --workflow-id "$workflow_id" --name "$workflow_name" \
  --workflow-file "$workflow_file" --execution-mode one_shot \
  --runtime-fields-file "$fields_file" --default-values-file "$defaults_file"
```

同名但未传 `--workflow-id` 是新建，不是更新。部署后同时核对部署回执的
`previous_workflow_version_id`、`versions`、`latest-published-version`、App 的
`latest_workflow_version_id` 和本次 execution 的 `moi_workflow_version_id`。2026-11-14 QA
已实测同一 App 从 v1 更新到 v2，v1 deprecated、v2 published，新执行绑定 v2 且业务输出为新版；
这只证明普通 MOWL deploy 路径，不代表旧手工 create-version/publish-version 的历史问题已修复。

cron、Volume trigger 和 dynamic service 的配置及调用见
[触发器与 API 服务 Recipe](../recipes/configure-workflow-triggers-and-service.md)。
`workflow update` 只改名称、描述或默认值时，省略触发字段；不要无意重提交它们。触发方式更新
复用保存的图，但本轮 QA 的 cron、volume_trigger、恢复 one_shot 三次更新分别产生了新版，
最终为 v5。每次配置后都必须重新读取 latest version 和 `trigger_summary`，不能沿用更新前版本 ID。

Workflow 编辑器候选生命周期是独立入口。`deploy-candidate --workflow-id` 仍要求真实 Workflow
Agent 候选和服务端验收；不要把普通 MOWL 文件伪造成 acceptance，也不要把候选路径当成 Agent
会话部署的前置条件。候选更新和跨 App 限制见
[正式候选更新 Recipe](../recipes/create-and-run-workflow.md#在同一个-app-上更新正式候选)。

暂停、恢复、重试、取消、删除、发布版本、动态服务、模板和自定义算子是独立操作。不能把
运行失败自动转换成 retry，也不能把 `moi_task_id` 当成 `execution_id`。

## 暂停、恢复与取消执行

使用 `pause-run`、`resume-run`、`cancel-run`，同时传原 `workflow-id` 和 `execution-id`；
它们控制一次执行，与应用级 `pause`、`resume`、`enable` 不同。按用户已授权范围操作，
以 `get-run` 回读状态为准，不以命令退出码或 available_actions 代替终态。

暂停已实测能使刚触发的执行保持 paused，恢复后沿用原 execution/task/case/version 并完成；
取消已实测覆盖 triggered、running、paused。暂停不等于终止；取消的成功状态是 cancelled，
不能报告业务执行完成。对已完成执行发 pause-run 可能只返回 completed，不能算暂停生效。

取消结果可用 `get-run-result` 检查，但 trace 中部分 span 仍可能显示 running；不能据此声称
所有 trace 都已收尾，也不能承诺外部写入、消息或其他副作用会自动撤销。恢复后仍需核验实际
结果；重试遵循[失败恢复](../contracts/failure-recovery.md)，不能把 retry 当成节点断点续跑。

## 批量重试的请求重放

在第一次 `batch-retry-failed-files` 前持久化本次 request ID、原dispatch和精确失败execution集合。
结果未知时用相同请求键和相同内容查询/重放，不换键提交；核对返回的dispatch、accepted_count、
replayed和原子execution集合。键的防重范围与节点业务副作用防重是两件事。
本次已完成请求的一次顺序重放和两次并发重放均返回原dispatch、replayed=true，执行集合仍为原三条。
2026-11-01 又完成同一全新键的双请求同时首发：只新增一个目标文件execution，同版本完成、结果正确，
延后两次列表执行集合均为四条。完整条件见[恢复Recipe](../recipes/inspect-and-recover-workflow-job.md#volume部分成功后只重试失败文件)。
不据此承诺任意并发规模、不同键或外部副作用恰好执行一次。
