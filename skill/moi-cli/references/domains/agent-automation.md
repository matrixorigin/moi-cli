# Agent 自动化领域

自动化菜单包含定时/API任务、运行记录、callback规则和channel实例。定时/API任务通过正式通用
资源 `agent-automation-tasks`、`agent-automation-runs` 管理，完整流程见[任务 Recipe](../recipes/configure-agent-automation.md)。
不能因 CLI 没有同名独立子命令就认定功能不存在。

Callback 和 channel 另用：

```text
moi-cli -o -no-interactive -workspace-id "$workspace_id" agent list-callback-rules
moi-cli -o -no-interactive -workspace-id "$workspace_id" agent get-callback-rule --rule-id "$rule_id"
moi-cli -o -no-interactive -workspace-id "$workspace_id" agent list-channel-instances --provider "$provider"
moi-cli -o -no-interactive -workspace-id "$workspace_id" agent get-channel-instance --provider "$provider" --instance-id "$instance_id"
```

创建/更新/删除规则和渠道实例，以及测试、启用、停用和 OAuth 都是外部副作用或权限敏感
操作。typed JSON 先用 `schema` 和 `--validate-file`；连接凭证只放入用户本机安全输入，
不进入日志和报告。定时/API配置、实际触发、执行结果和外部事件分别验收；不要从一种成功外推另一种。
2026-10-30 已通过真实Cron；2026-10-31使用用户提供并在本机安全配置的PAT，真实API调用返回527、
trigger_type=api，最终仅一次运行并归档。此成功限固定指令；带业务参数/严格schema的失败条件见下文。
完整UI、渠道事件和所有认证方式仍保持 `partial`。

## 自动化 API 的认证边界

先分清 native 登录与 UC PAT。当前 native moi-cli 通过 Astra 登录发送 Bearer；自动化 API
策略虽名为 `workspace_api_key`，实际接收的是同一UC用户的有效PAT，经UC校验身份、
`aistudio-api` audience、`aistudio:user` scope，再校验工作区权限和目标 `agent.use`。
不能据策略名承诺这是仅限某个workspace/service的key。

历史上native权限查询成功而invoke返回401，PAT模式本机缺api-key；补充正式PAT后已实际调用成功。
因此这类401先辨别认证通道，不能一律解释为没登录；固定任务成功也不证明动态参数链路可用。

已有正式配置的PAT时，先由其正式安全配置owner确认凭据来自预期的同一账号；
以下只读命令仅验证PAT在目标工作区可用及返回的角色/能力，不返回可比较的UC主体ID，
不能据能力相同声称已经证明同一用户。归属无法确认时不继续写入。随后按Recipe启用并单次触发隔离任务：

```text
moi-cli -auth pat -endpoint "$endpoint" -o -no-interactive -workspace-id "$workspace_id" iam current-principal-capabilities
moi-cli -auth pat -endpoint "$endpoint" -o -no-interactive -workspace-id "$workspace_id" workflow invoke-dynamic-service --json-file "$invoke_file"
```

dev 的 invoke JSON 使用实际返回的 `service_name` 及所需 `payload`（小写字段）；不能按任务名称猜服务名。
稳定版 0.1.8 使用 `ServiceName` / `Payload`，两版字段不能混用。两条
`invoke-dynamic-service` / `invoke-workflow-dynamic-service` 路径不同，但归同一handler和认证边界，
换命令名不能解决此401。成功仍须回读run/result，核对trigger_type=api和业务输出。

2026-11-15 新配置的生产PAT已再次完成固定指令调用：run `24893446-86df-4cff-9958-fc08b4f819b2`
返回 `final_text=527`、`trigger_type=api`、`output_validation_status=passed`，最终仅一次运行并归档。
缺少必填 b 的请求也在提交前返回400 `b is required`，未产生业务 run。固定指令与缺参拒绝已通过，
但合法 a/b 的外层 envelope 校验和无 schema 严格 JSON 输出仍失败，不能合并成一个“API全部恢复”。

持久配置字段为 `api_key`，没有 `auth` 字段；`-auth pat` 是每次调用的全局参数。
配置PAT后auto/moi仍使用Astra native登录，不会自动切PAT；需要PAT的入口显式指定-auth pat。
CLI只消费PAT，不提供创建/轮换命令。正式 `config set-api-key` 只有位置参数，无隐藏输入、
stdin/fd入口；不要让用户把key贴进聊天、截图、命令历史或验收日志，也不读取已有凭证文件。
缺PAT时说明需通过账号凭据管理owner及本机安全配置流程提供，不要求重新Astra登录。

正式沙箱会自动注入runtime PAT与签名proof，但当前invoke路由使用普通requireIdentity，
明确不接受runtime proof；本次只核对合同，没有发送已知不支持的runtime请求。
不能删除proof降成默认角色、混用Bearer和X-API-Key、搬移token或伪造header。
支持native Bearer/runtime proof需要产品owner修改认证合同并测试部署，不能由Skill绕过。

## 带参数 API 的当前失败边界

2026-10-31进行了两个独立负向/正向场景，均保留真实run/result，不以配置保存作成功：

- `trigger.input_schema`是JSON Schema。缺必填b的Payload被400拒绝，随后runs为0，输入拒绝正确。
  相同任务提供合法a/b后返回500；回读出现真实failed run，错误为WORKFLOW_SUBMIT_FAILED，
  内层INPUT_VALIDATION_FAILED把含payload的整层automation输入当成业务schema根对象校验。
  这不是缺PAT，也不是缺少a/b；不能把业务schema改成包络结构、删校验或重发来假装修复。
- 独立任务不声明input_schema，真实Payload包含a/b，但最终回答称未收到数值；
  result为OUTPUT_CONTRACT_VALIDATION_FAILED，structured_output为空。此场景的动态参数业务结果未通过；
  进一步复核确认payload已保存在运行输入快照中，但没有投影进模型消息：无显式message_parts时只传
  instruction_text，模型因此只能回答未收到a/b。严格JSON失败是该投影缺陷的下游结果，现有证据不单独
  证明输出校验器还有第二个独立bug。模型自述不能替代真实运行链证据。

`output_contract`本身是JSON Schema，`{}`允许任何输出，返回validation passed不等于拿到结构化JSON。
要求JSON业务结果时核对structured_output和字段值；非空schema测试未通过时保留失败，不能换成空schema宣称通过。
API遇到500也先查原task的runs，可能已有failed run；原请求没有返回run_id时从对应task运行列表关联，不能重发生成重复执行。
当前只可承诺已验证的固定指令API；需要动态参数或严格输入/输出约束的任务应报告此限制并等待产品修复，
不转低层接口、混用身份或静默降级。

## Team 和动态 Payload 的根因边界

2026-10-31 的独立根因复核确认：Team 两轮任务都能 `completed/committed`，但第二轮 runtime
session 的 turn 仍为 1，只能回答当前消息，不能自动继承第一轮结果。显式把前文放进第二轮
输入可用，因此不是缺少 workspace、provider session 或用户需要补填的字段；待产品 owner
沿“completion 保存 → ingress 冻结 → provider 请求”核对续传链路。Skill 不应要求用户把上一轮
答案重复粘贴来伪装自动历史已修复。

动态 API 的合法 `input_schema={a,b}` 请求确实进入了 run，但服务端把 payload 包在
automation envelope 中，后续 workflow 用整个根对象校验业务 schema，得到
`WORKFLOW_SUBMIT_FAILED/INPUT_VALIDATION_FAILED`。无 input_schema 的独立任务又在 runtime
适配层丢失结构化 payload，最终 `structured_output` 为空。两条证据都指向服务端输入适配，
不是 CLI 少了参数、PAT 失效或重新登录可解决；保留固定指令 API 的可用范围，动态参数和严格
JSON 结果继续标为 `partial`。

2026-11-02 正式0.1.8的 `agent list-channel-instances` 在不带实例上下文时返回
`ErrIDRequired`，即使命令名是 list，也不能把它当作无条件的全量列表入口。先读取当前二进制 help，
确认 provider/instance 的真实发现合同；省略 `--provider` 时报告“请求缺少必要筛选条件”，不是把整个能力判成 CLI 缺口。
按 `--provider wecom|slack|feishu` 可读取结构化列表；返回空列表只证明当前 provider 下没有实例，不能外推 OAuth、连接验证或真实事件已通过。
只有 help 没有任何可用筛选输入，或按合同输入仍无法发出请求时，才登记为 `cli_missing`；不猜 ID、不创建连接。

2026-11-15 dev 实测中，省略 provider 时 CLI 会明确提示先运行 `list-connector-catalog`；目录可返回实际
provider 及 `configured=false`，再按 wecom/slack/feishu 筛选均得到标准空列表。`list-callback-rules`
也返回 `items=[]、total=0、limit=50、offset=0`。这些只读发现与空列表语义已通过，但仍不证明 OAuth、
真实外部事件、回调投递或外部连接成功。
