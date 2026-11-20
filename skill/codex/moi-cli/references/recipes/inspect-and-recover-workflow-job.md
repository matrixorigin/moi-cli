# Recipe：检查并恢复 Workflow 作业

1. 从用户给出的 workflow/execution 或前一步运行结果取得真实 ID；不能从作业名称猜测。
2. 先根据已确认的配置/输入区分单次与Volume分发。单次用 `workflow get-run` 和 `workflow get-run-result`；Volume返回父dispatch需按下文关联子execution，不能把父ID的404当提交失败。
3. 区分 preparing/submitted/triggered/scheduled/running、completed、failed、cancelled、paused 等状态。
4. 只有服务端状态和领域文档表明可恢复，且用户明确允许时，才选择 `retry-run`、`resume-run`、`cancel-run` 或 `pause-run`。
5. 恢复命令返回新的 execution 或分支信息时保存新旧 ID，不能覆盖原始失败记录。
6. 对恢复后的 execution 重复 get-run → get-run-result 闭环，并核验目标产物。

任何未知状态先查询原作业；不能因为客户端超时就再次提交 run。

## 父 dispatch 与子 execution 的结果闭环

批量/Volume 分发可能在同一 `dispatch_job_id` 下产生多个文件级 execution；一条可以
`failed`，另一条可以 `completed`。把 `list-runs` 返回的每条 `execution_id` 分别交给
`get-run` 和 `get-run-result`，并保存 `moi_task_id`、`moi_case_id`、`case_result`、失败节点、
节点状态和 trace。结果完整后再按 `file_id`、输入版本和目标产物回对。

`dispatch_job_id` 只用于分组，不是普通 execution 查询 ID。某个父/分发 ID 查询 404 时，
先按 dispatch 展开子 execution；不能据此写成“服务端没有执行”。批量结果通过只证明本次
选定文件和本次节点合同成立，不证明写入、外发或任意副作用节点可以安全重放。

## retry 的真实语义

当前 retry-run 对终态执行创建一次全新执行，从开始重放 input/vars；不是失败节点续跑。
保存原ID、retry命令回执新ID、两侧版本和结果。原失败执行应仍可查询，不能用新成功覆盖失败记录。
默认重试会使用当前部署版本，不能承诺自动锁定原版本；执行前比较当前部署与原版本。
若业务要求固定原版本而 CLI 无对应入口，保留限制，不直接重试。

当前 dev CLI 可用 `retry-run --values-file` 提供本次新运行值，并按当前 help 选择
`--replace-saved-defaults`。field_id、授权和副作用检查与初次 run 相同；修正输入后仍要保存
新 execution ID，不能用原 ID 查询新结果。2026-11-14 QA 的纯 jq 负例在 root.gate 失败后，
以 `should_fail=false` 重试，新的 execution completed、输出正确，原 failed 记录保留。

隔离验收可用“只读SQL读取开关→jq受控报错”：首次开关0失败，修正获准的隔离开关为1，再重试一次，
核验新执行成功且原执行保留。这个只读开关样例不包含重试副作用；本轮另验了[业务唯一键保护的SQL重试](create-and-run-workflow.md#失败后重放-sql-写入)，不证明导入、任意文件写入或外发节点可安全重放。
这些节点要先完成独立幂等/去重设计与授权；不要把成功过的写入整条重做。

## 从指定节点继续

与 `retry-run` 的从头重放不同，节点级恢复先读取真实节点和服务端计划：

```sh
moi-cli -o -no-interactive -workspace-id "$workspace_id" workflow get-run-node --workflow-id "$workflow_id" --execution-id "$execution_id" --node-key "$node_key"
moi-cli -o -no-interactive -workspace-id "$workspace_id" workflow get-run-node-rerun-plan --workflow-id "$workflow_id" --execution-id "$execution_id" --node-key "$node_key"
```

node_key 来自当前执行的真实节点名/attrs（例如 `root.check_gate`），不是 WorkItem 类型 ID。
核对 plan 的 workflow/source execution/node、`source_node_execution_id`、`validation.ok`、
`available_modes` 和 `rerun_contract_hash`。计划通过只表示可提交，还不是恢复完成。
目标或合同变化时重新读取计划，不拼接节点执行 ID、不复用其它节点的 hash。

已授权且副作用范围清楚后，按计划支持的模式提交一次：

如果是顺序重发或上次结果未知，先用 `workflow list-run-rerun-branches --workflow-id ... --execution-id ...`
读取已有恢复，只按本次已记录的 `source_execution_id`、`node_key`、`source_node_execution_id` 和 `mode`
精确匹配。找到同一恢复意图的已成功分支后，先读原 rerun/新 execution，不重复创建。2026-11-14/15
生产隔离样例中，第二次创建因此被跳过，分支数保持 1。列表不含 values override 或合同 hash，
不能据此阻止一个明确的新恢复意图；该方法只证明顺序查重，不证明服务端幂等或两个并发首发只执行一次。

```sh
moi-cli -o -no-interactive -workspace-id "$workspace_id" workflow create-run-node-rerun --workflow-id "$workflow_id" --execution-id "$execution_id" --node-key "$node_key" --source-node-execution-id "$source_node_execution_id" --mode continue_from_node --expected-rerun-contract-hash "$rerun_contract_hash"
moi-cli -o -no-interactive -workspace-id "$workspace_id" workflow get-run-rerun --workflow-id "$workflow_id" --execution-id "$execution_id" --rerun-id "$rerun_id"
```

分别保存原 execution、`rerun_id`、`branch_id`、`new_workflow_execution_id`。查询 rerun 状态，
再用**新 execution ID**取详情和结果；核对版本、业务输出、实际执行节点和原记录。
当前 dev CLI 已暴露 `--values-file` 和 `--config-override-file`。只有真实计划的
`available_modes` 包含 `continue_from_node_with_overrides` 时才选择该模式；运行值仍以保存表单
field_id 为键，config override 的结构以当前 help/schema 和真实节点合同为准，不能猜字段。
旧常规 CLI 缺少 overrides 输入的历史结论不再适用于当前 dev，但也不能外推到稳定版。

```sh
moi-cli -o -no-interactive -workspace-id "$workspace_id" workflow create-run-node-rerun \
  --workflow-id "$workflow_id" --execution-id "$execution_id" --node-key "$node_key" \
  --source-node-execution-id "$source_node_execution_id" \
  --mode continue_from_node_with_overrides \
  --expected-rerun-contract-hash "$rerun_contract_hash" --values-file "$recovery_values_file"
```

2026-11-14 QA 已用纯 jq 的 root.gate 受控失败验证上述入口：计划允许两种继续模式，传入
`should_fail=false` 后 rerun succeeded，新 execution completed，trace 只有 gate 和 end，
上游 prepare 未重跑，原 failed execution 保留。这证明当前 dev 的 values override 和该只读
节点恢复；未验证 config-override-file 的线上行为，也不证明写入、外发或任意节点安全。

判断 values override 是否影响业务结果时必须读目标节点的表达式和最终 `case_result`。
`runtime_vars` 出现新值只证明运行值已进入节点上下文；若表达式仍读取 `$data.prepared.text`
等上游输出，最终文本保持原值是正确行为，不能把它写成覆盖失败或覆盖成功。需要输出变化时，
让目标表达式实际读取被覆盖的变量，或走正式 artifact 输出修订链。

`config-override-file` 只接受计划 `overrideable_config_schema` 声明的路径和类型。普通
`mowl:jq` 没有该声明时，提交 `__output` 返回 400，不能猜字段或把 values override 当成
任意节点输出编辑。artifact revision 是另一条对象与合同，二者不能互相替代。

2026-11-01 只读 jq 节点样例：服务端将请求模式记录为 `rerun_current_and_downstream`，
新分支 succeeded，新 execution completed；trace 只有目标 check_gate 和 end 节点，
原执行、上游读表/文件结果保持。只证明本次不带 overrides 的节点恢复，不能外推写入、外发或任意节点。
同一原节点传错误 hash 时返回409 `rerun_contract_mismatch`，前后执行集合不变；
合同冲突后保留错误并重新核对，不静默移除 hash 或重发创建。
查询入口 `get-run[-result]-by-id` 与带 workflow ID 的入口已在同一历史 execution 上回读一致。
`workflow history --context-id` 需要真实 authoring context；execution ID 不是 context ID，传错后的404不证明历史接口故障。

## Volume部分成功后只重试失败文件

按[批处理关联规则](catalog-workflow-knowledge-pipeline.md#3-提交一次追踪本轮文件执行)取得原dispatch及真实子execution。保存完整本轮文件集合、失败错误、已成功部分和各自版本；若列表范围不全，不据此认定全部失败文件已找到。

用户已授权修复并重试时，先解决失败条件，再比较当前部署版本与原失败版本。版本变化时不能承诺重试锁定旧版本；需要用户目标允许当前版本，或取得正式固定版本入口后再做。

```sh
moi-cli -o -no-interactive -workspace-id "$workspace_id" workflow batch-retry-failed-files --workflow-id "$workflow_id" --source-dispatch-job-id "$original_dispatch_id" --execution-ids "$failed_execution_id" --request-id "$retry_request_id"
```

显式选择真实失败execution，不能传file ID或父dispatch代替；多ID格式以当前help为准。`--all-failed`会扩大到原批全部失败部分，仅在用户目标与完整清单确实覆盖该范围时选择。request-id固定用于这一次恢复；超时先查原请求/新dispatch，不生成新键盲重发。已对同一完成请求键做一次顺序和两次并发重放，均返回原dispatch、replayed=true，执行集合仍为原三条。
2026-11-01 又以同一个全新 request-id 同时发出两个首次请求：返回同一个新dispatch，其中一个为replayed=true；
延后两次list-runs均只新增一个目标失败文件的execution，同原版本completed，case_result确认recovered=true，
原失败保留，原成功文件未重跑。该结果只证明本次两个请求与只读恢复流程，不保证任意并发规模、不同键
或外发/写入节点的业务副作用防重。

保存返回的 `source_dispatch_job_id`、新 `dispatch_job_id`、`accepted_count`。再通过list-runs精确筛新dispatch，核对只含目标失败文件，逐一get-run/result，检查新终态、版本、case_result。原失败记录应保留，原成功文件不应新增execution；提交accepted不等于恢复完成。

已通过的隔离样例是“读取原文件→只读SQL读取按file ID的开关→jq”：初批一成功一失败，仅打开失败文件的开关后定向恢复，最终原成功、原失败、新恢复共3条，同版本，新恢复completed。它证明选择范围与执行闭环，不证明导入任务失败文件恢复，也不证明任意写入/外发节点可安全重放。

## 写入节点继续：先确定写入是否已经发生

节点继续会重跑目标节点和下游，不提供自动回滚或自动业务去重。先独立读取目标数据，区分恢复点在写入之前、写入本身还是写入之后；读取真实计划/hash后只提交一次，再核对新分支的trace和目标数据。

2026-11-01 三个隔离对照已通过：普通INSERT从写节点继续，原1行17变2行34，trace确有write_event；同输入改用业务主键和`ON DUPLICATE KEY UPDATE`，仍1行17，trace也确有写节点，说明保护来自业务键。第三个流程先写入1行33、后续检查失败，SQL证明失败未回滚；修正已授权隔离开关后从`root.read_gate`继续，新分支只读开关/检查/结束，没有write_event，仍1行33。新旧execution版本一致，原结果分别保持completed/failed。

不能把唯一键样例推广为文件、通知或任意外发动作的全局幂等；不能照搬演练删行或修改真实业务数据。需要修改恢复输入时，当前dev可按真实计划使用values override，但仍须先评估重跑节点和下游副作用；稳定版是否具备相同输入以其help为准。只读状态查询超时后沿原query ID继续查，不重发写入。

## 文件写入后的恢复

文件sink也是副作用。保存真实写节点runtime_output的file_id/file_ids和目标volume_id，再用file-info、按file ID查询placement、Volume统计和完整下载核对已写内容。文件名或相同内容不是幂等键；同名文件可能并存，不能只按名字核验或删除。

2026-11-01隔离文本sink对照：写入128字节后下游检查失败，原文件仍在；修正演练开关后从`root.read_gate`继续，新trace无sink，case_result、原file/placement和下载字节不变。从原失败任务`root.write_file`继续，则新trace含sink，产生不同file ID及新placement，Volume由1文件128字节变2文件256字节，原文件仍在；两个同名文件完整下载字节/SHA256一致。三条execution同版本，原失败和前次恢复历史保持。

此前只证明该文本sink和两个恢复点；取消后文件保持及限定继续补验见下文。小型两来源ZIP另见[产物核对](catalog-workflow-knowledge-pipeline.md#小型多来源zip与抽取字段)，大文件、复制/复用已有file、不同内容覆盖和外发仍未验；相同计划重复提交的文件副作用补验见下文。先确认写入已发生，再按业务需要选择写后恢复点；获准重跑写节点时，必须分别处理新旧文件身份及新增副作用，不能承诺自动去重。

## 写后取消与取消任务的继续分支

暂停/取消不撤销已落地文件。2026-11-01写入130B后paused→cancelled，file/placement及完整下载不变；对cancelled直接`resume-run`返回409，原状态保持。取消不是暂停，不以resume拒绝判断所有节点继续都不可用。

节点继续先取得真实计划/合同hash，再按允许动作创建独立分支。含eval等待节点流程的三处计划均422 `workitem_rerun_contract_missing`；另一个仅SQL只读节点的隔离对照，取消后从已完成写后checkpoint继续成功，同版本，新trace无sink，原103B文件不变，原cancelled保留。这是不同合同流程的对照，不是修复原eval任务，也不保证任意节点/状态皆可恢复；不凭可见节点或available_actions跳过计划检查。

## 节点恢复的合同hash不是幂等键

当前正式`create-run-node-rerun` help只有原Workflow/execution/节点、来源节点执行ID、mode及expected hash，没有request/idempotency key；无typed schema的命令按help使用，不猜字段。不要套用带request-id的批量失败文件恢复合同。

2026-11-01文本sink→8只读SQL样例：原paused时取计划409 `workflow_execution_not_terminal`，同execution恢复至completed才可取计划；本例错误、跨节点和跨execution的hash均409且不新增执行。仅证明对应状态与上下文保护，未制造同源合同真实改动或TTL过期，不以此宣称所有旧计划都会失效。

同一写节点计划的两个相同请求并发提交，各创建不同rerun/branch/execution并各写一份文件：原1份142B变3份426B。完成后顺序再交同一计划仍新增一分支和文件，变4份568B，原计划/hash保持。hash不是一次性令牌或防重键。副作用次数从真实trace、各file/placement及完整下载核对，不能只按同名或相同hash去重后计数。

提交前保存本次恢复意图、真实来源、计划及授权范围；不要把同一恢复意图同时交给两个工作者提交。拿到新ID后只追踪该分支；响应未知先沿原执行集合和已写产物核对，范围不全时保留未知，不盲重发。正式入口没有幂等键时不能承诺恰好一次，也不据此断言后端所有路径都做不到。

从直接父分支的写后SQL节点继续，本例新trace只有8SQL与结束节点、无sink，文件仍4份且字节不变；source execution/branch/case与直接父对应，五执行同版本，原完整结果保持。必须取该来源自身节点计划/hash，不能把同名节点或原祖先hash复用到子分支。不同版本、任意节点及外发副作用仍需另验。
