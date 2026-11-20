# Recipe：Agent 定时与 API 自动化

使用正式 `agent resource-*`，资源名为 `agent-automation-tasks` 和 `agent-automation-runs`。
`schema agent resource-post` 是开放对象，不证明业务字段合法。先固定 Agent/owner、用户指令、触发器、
模型、权限及外部副作用范围；默认 DeepSeek、省略 effort。

## 创建和变更

新任务先 disabled，最小 API 配置如下；字符串占位符要替换为实际输入：

```json
{
  "id":"本次独立任务ID", "name":"用户选择的名称",
  "agent_workspace_id":"实际owner", "agent_id":"实际AgentID",
  "model":"deepseek-v4.1-flash", "instruction_text":"本次授权要求",
  "trigger":{"mode":"api"},
  "auth_policy":{"type":"workspace_api_key"},
  "output_contract":{}, "status":"disabled"
}
```

Cron 使用 `trigger:{"mode":"cron","cron_expression":"合法cron"}`，不带 API auth_policy。
非空 tool_policy_ref/runtime_policy_ref/approval_policy_ref 当前会被业务校验拒绝，不能据字段名编造支持。
需要动态参数、输入校验或JSON输出时，先读[当前失败边界](../domains/agent-automation.md#带参数-api-的当前失败边界)。
固定指令API已成功不代表这些条件已通过；不能用空output_contract的passed冒充结构化输出通过。

```text
moi-cli -o -no-interactive -workspace-id "$workspace_id" agent resource-post --resource agent-automation-tasks --json-file "$request_file"
moi-cli -o -no-interactive -workspace-id "$workspace_id" agent resource-get --resource agent-automation-tasks --tail-path "/$automation_id"
moi-cli -o -no-interactive -workspace-id "$workspace_id" agent resource-patch --resource agent-automation-tasks --tail-path "/$automation_id" --json-file "$patch_file"
moi-cli -o -no-interactive -workspace-id "$workspace_id" agent resource-get --resource agent-automation-tasks --tail-path "/$automation_id/runs"
```

返回可能包在 `Value` 中。保存 exact id，回读 trigger/model/status；只有用户已要求启用时才改 active。
启用会配置真实 Workflow，不能把它当无副作用保存。API 必须从 api_summary 取得真实 service_name、版本和认证要求，
不按任务名拼地址；登录态和触发器要求的 workspace_api_key 需分别核实。401 不能一律解释为用户没登录。
认证路径不匹配时先查[领域边界](../domains/agent-automation.md)。现有native登录不能替代此入口的UC PAT；
缺PAT不重复登录，有正式安全配置且其owner确认账号归属后，显式 `-auth pat` 先查工作区能力，再单次invoke；
能力查询不含可比较的UC主体ID，不能用它证明同一用户；不读取/复制已有密钥或切身份绕过。

## 触发、结果和停用

Cron 需读取真实 next_trigger_at，观察服务端 run.trigger_type=cron；手动 `/runs` 调用不能证明定时器触发。
测试用频繁 cron 必须设明确结束时间，首条运行出现后即停用，并回读 disabled，避免下一轮；真实用户任务按用户期望保持状态。

运行出现后保存 run_id；分别查询：

```text
moi-cli -o -no-interactive -workspace-id "$workspace_id" agent resource-get --resource agent-automation-runs --tail-path "/$run_id"
moi-cli -o -no-interactive -workspace-id "$workspace_id" agent resource-get --resource agent-automation-runs --tail-path "/$run_id/result"
moi-cli -o -no-interactive -workspace-id "$workspace_id" agent resource-get --resource agent-automation-runs --tail-path "/$run_id/events"
```

核对 succeeded、final_text/output、output_validation_status 和错误。run/result 没有 runtime_task_id 时不能猜任务ID，
也不能以此声称校验了底层模型全部运行参数。配置和空 runs 列表不能证明触发成功。
超时或500先查对应task的runs和原run；本轮invoke500仍产生真实failed run，不能直接重发。新的请求ID可能导致重复业务执行。归档用 `resource-delete --resource agent-automation-tasks --tail-path /ID`，
返回 archived 并非硬删，最后检查 enabled 状态与运行数量。

生命周期验收分别保存创建、disable、enable、Cron 修改和归档回执。每次修改后读取
task.version、trigger 和 status；归档后同时检查活动列表不再出现、归档列表仍可查询，
以及历史 run/result 是否按业务语义保留。2026-11-17 隔离任务已完成启停、Cron 修改、
版本增长和归档。归档是软删除，不应把“历史仍保留”误判为删除失败，也不能把它写成
物理擦除全部元数据。

缺少 `trigger.input_schema` 必填字段时，提交前应直接拒绝且不产生业务 run；2026-11-15 缺 b 返回400
`b is required`，属于预期拒绝。合法 a/b 若进入 run 后仍出现 `INPUT_VALIDATION_FAILED`，必须沿真实 run
读取错误，不能把缺参拒绝和合法输入失败混为一谈。结果查询使用 `agent-automation-runs/{run_id}/result`，
任务级 `/result` 404不能代替终态，也不应提前归档；只有原 run 明确结束后才停用归档。

外部事件、callback、channel、OAuth 是另一分支；见[自动化领域](../domains/agent-automation.md)。

需要每轮读取当前知识时，按[工作流运行时检索](run-runtime-knowledge-workflow.md)核对 Agent 的正式知识绑定、已发布版本、provider/runtime binding、真实工具调用和正式引用。不能只把KB ID写在instruction里当作绑定，也不能仅凭final_text声称已检索。

## 修订配置与新版实际执行分别验收

修订后回读task.version、trigger和instruction，再沿新run核automation_task_version、trigger_config_snapshot及实际答案。task为v2而last_run_summary仍v1，只说明旧运行还在，不能证明新版已执行。公开resource-patch caller未建立expected_version/CAS合同，不猜字段或承诺陈旧写保护。

频繁cron测试保存绝对截止、服务端next_trigger_at和每次观测时间；首条目标新run出现即停用。到期无新run先停用并独立回读，再有限只读检查原runs有无晚到记录；不以手动触发替代cron验收，也不自动重新启用延长测试。

2026-11-02样例v1真实cron成功；disabled时修订指令和cron为v2，启用后120秒内23次轮询仍仅旧run，due后约31秒已停用，后续归档及延迟只读仍无第二run。新版实际执行判partial、原因未知，不能外推所有修订任务均不通，也不能以配置保存或旧答案冒充通过。归档后重新active返回400且对象保持archived v2；旧run/result/events完整保持。归档不是硬删，空output_contract校验passed也不是严格JSON输出通过。

2026-11-15 再次复现同类边界：v1 返回 `SKILL045_A`，修订 v2 的 `next_trigger_at` 正确但 200 秒内无新版 run，
随后停用归档。另建独立 B 任务使用 `*/2 * * * *` 后，真实 cron run 返回 `SKILL045_B`，result/events
完整并已归档。修订不生效时，先按截止时间停用旧任务；只有用户明确允许替换，才用新 ID 创建目标配置
并核验实际触发。该替换路径不证明原修订已修复，真实任务还必须单独核对历史、输入、权限和外部副作用，
不能简单改名或复用旧 ID 后宣称无缝迁移。
