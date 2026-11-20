# Recipe：创建或运行 Workflow

## 编写并部署原生 MOWL

先确定输入、业务规则、预期输出、执行单元和节点副作用。验收使用唯一名称与合成输入，不运行
发送消息、外部写入或修改已有业务资产的图。需要 WorkItem 时，先用
`workflow workitem-catalog` / `workitem-catalog-item --node-id` 读取真实节点 ID、版本和输入输出；
纯 `mowl:*` 节点不需要编造 version。

MOWL JSON 顶层包含 `name`、`root`、`nets`。每个 net 必须有连通的 `mowl:start` 和 `mowl:end`；
节点的 `varMap.__input__` / `__save__` 是 JSON 对象，`__output` 是 jq 字符串。输入模板使用
`{{ .data }}`、`{{ .vars }}`、`{{ .state }}`；jq 输出、保存和路由表达式使用 `$data`、`$vars`、
`$state`，不要混写成 `.vars` / `.state`。

最小单 jq 图：

```json
{
  "name": "isolated_mowl_example",
  "root": "root",
  "nets": {
    "root": {
      "nodes": {
        "__start": {
          "id": "mowl:start",
          "outConnector": {"type": "simple", "nodes": ["answer"]}
        },
        "answer": {
          "id": "mowl:jq",
          "inConnector": {"type": "simple", "nodes": ["__start"]},
          "outConnector": {"type": "simple", "nodes": ["__end"]},
          "varMap": {
            "__output": "{ok: true, text: $vars.text}"
          }
        },
        "__end": {
          "id": "mowl:end",
          "inConnector": {"type": "simple", "nodes": ["answer"]}
        }
      }
    }
  }
}
```

将图、字段、默认值和运行值分别保存为本地 JSON。字段的 `field_id` 是默认值和运行值对象的键，
`bind_to` 才决定绑到 `input.<path>` 或 `vars.<path>`：

```json
{"fields":[{"field_id":"text","label":"Text","widget":"text","json_type":"string","bind_to":"vars.text"}]}
```

```json
{"text":"saved value"}
```

`schema ... --validate-file` 只验证 CLI 输入结构；MOWL 图和依赖由服务端 deploy 校验。部署并回读：

```sh
moi-cli -o -no-interactive -workspace-id "$workspace_id" workflow deploy \
  --name "$workflow_name" --workflow-file "$workflow_file" \
  --execution-mode one_shot --runtime-fields-file "$fields_file" \
  --default-values-file "$defaults_file"
moi-cli -o -no-interactive -workspace-id "$workspace_id" workflow get --workflow-id "$workflow_id"
moi-cli -o -no-interactive -workspace-id "$workspace_id" workflow latest-published-version --workflow-id "$workflow_id"
```

没有字段或默认值时省略相应文件。保存 `workflow.id` / `deployment.workflow_app_id`、definition
和 version ID；部署 ready 只表示已保存并可运行，不是业务测试通过。cron、Volume 和 dynamic
service 的配置见[触发器与 API 服务 Recipe](configure-workflow-triggers-and-service.md)。

## 在同一个 App 上部署 MOWL 新版

用户明确授权修改指定 App 时，先读取并保存原图、版本、字段、默认值和触发方式。修改原 MOWL
源文件后，使用原 App ID 部署；不要新建同名 App，也不要为普通 Agent 会话伪造候选：

```sh
moi-cli -o -no-interactive -workspace-id "$workspace_id" workflow deploy \
  --workflow-id "$workflow_id" --name "$workflow_name" \
  --workflow-file "$workflow_file" --execution-mode one_shot \
  --runtime-fields-file "$fields_file" --default-values-file "$defaults_file"
```

保留未要求改变的配置；如果目标触发方式不是 one_shot，使用已核实的原配置而非照抄示例。
核对返回的 App ID 未变、`previous_workflow_version_id` 是原版本，再用 `versions` 和
`latest-published-version` 确认新版 published。最后实际运行，检查 execution 的
`moi_workflow_version_id` 和输出；旧 execution 仍应归属旧版。

2026-11-14 QA 的隔离 jq 验收已从 v1 更新到 v2，并验证新旧默认值、显式运行值和版本绑定。
触发方式更新也可能生成后续版本，不能拿这里的 v2 ID 继续代表最终 App 当前版本。

## 用现有 MOWL 创建独立副本

当前没有独立 `copy` 命令时，可用正式 `workflow get` 取得源 App 的完整 `workflow` 图、
`input_form.fields`、`default_values` 和 `execution_mode`，本地改成唯一新名称后用普通
`workflow deploy` 创建新 App。不要使用源 App ID，否则会变成更新源工作流。

响应里的 `in_connector/out_connector/var_map` 要转换为部署输入的
`inConnector/outConnector/varMap`。`varMap` 中 JSON 字符串可按其真实类型还原为对象；
jq 表达式仍是字符串，不能一律二次编码或全部改成对象。保留源字段、默认值和触发方式，
但不复制源 App/version/execution ID。

部署后分别读取源和副本，规范化名称后比较图、节点参数、输入字段、默认值和执行方式，
确认两个 App ID 不同；再实际运行副本并核对本次 `case_result`。2026-11-17 隔离副本与源
这些字段全部一致，运行结果同为 42。该方法只适用于已确认的 manual MOWL，不能冒充
Workflow Agent 候选或任意 YAML/动态服务的一键复制。

如果手工 `create-version` / `publish-version` 后 `latest-published-version` 已是目标 v2，但 App 的
`latest_workflow_version_id` 仍停在 v1、运行返回 400，可以在同一 App 上用上述普通 `workflow deploy`
恢复。保存原图、字段和默认值并传入原 App ID，部署后同时核对新恢复版本、App 指针、execution 绑定版本
和实际输出。2026-11-14/15 生产隔离复测中，普通 deploy 生成 v3 后，App 和 execution 均指向 v3，
v2 业务输出正确。该方法是受影响 App 的恢复路径，不表示手工发布原路径已经修复，也不能只凭
`latest-published-version` 判断指针已恢复。

## 用 YAML DSL 创建独立 Workflow

用户要求新建时，先确定输入、预期输出和节点副作用。验收用独立名称和合成输入；不运行会
发送消息、调用外部写入或修改现有业务资产的流程。读取 `workitem-catalog-item` 的真实节点
版本和输入要求，不能从显示名称猜 node ID 或复制已有业务流程中的凭证。

已确认的 YAML DSL 结构示例：

```yaml
workflow:
  name: isolated_clean_example
  root: root
root:
  chain:
    - work_item:
        name: clean
        id: moi:parser.clean.text
        version: v1
        input:
          text: "  Codex    MOI  "
        save:
          cleaned: ".text"
    - jq:
        name: verify
        expr: '{text: $state.cleaned, matched: ($state.cleaned == "Codex MOI")}'
```

先将 DSL 保存到本地文件，再把其**内容**作为 `--dsl-yaml` 的单个参数传入：

```sh
moi-cli -o -no-interactive -workspace-id "$workspace_id" workflow workitem-catalog-item --node-id moi:parser.clean.text
moi-cli -o -no-interactive -workspace-id "$workspace_id" workflow deploy --name "$name" --dsl-yaml "$dsl_yaml"
```

`deploy` 创建可手动运行的定义，返回 `workflow.id`、`workflow.status` 和
`deployment.workflow_version_id`；不等于执行完成。检查 ready、实际版本和配置后按下面流程
运行一次，最终检查 `case_result` 的业务断言。不要从变量名推断输出已经正确。

若已有工作流详情因未知字段解码失败，记录具体命令和字段；可只读检查已知版本，但不得据此
运行尚未确认配置的旧工作流，更不能通过重建旧资源规避问题。

## 部署时传默认输入

dev CLI 的 `workflow deploy --help` 已提供默认值和运行字段输入；执行前仍以当前二进制 help
和 schema 为准，不把结构校验当作服务端业务校验。

```sh
moi-cli -o -no-interactive -workspace-id "$workspace_id" workflow deploy --name "$name" --dsl-yaml "$dsl_yaml" --default-values-file "$defaults_file"
```

对象键是实际 `input_form.fields[].field_id`，由 `bind_to` 映射到执行变量，不直接凭变量名
猜键。例如自动表单将 `{{ .vars.source_ref }}` 映为 `vars_source_ref`，其默认值可以是已确认的
`{"kind":"volume","resource_type":"volume","volume_id":123}`。123 须换为本轮真实非空 Volume。

部署后回读 `workflow get` 的 input_form/default_values、ready与正式版本，再运行。每文件运行需要后续SQL使用file ID时，可在 source读取节点save `selected_file_id: ".file_ids[0]"`，后续使用 `{{ .state.selected_file_id }}`；直接在DSL写 `{{ .vars.file_id }}`会生成额外必填用户字段，并可能在分发前拦住运行。不要填假file ID规避校验。

## 在同一个 App 上更新正式候选

适用：目标明确属于 Workflow 编辑器的候选审阅生命周期。普通 Agent 会话的 MOWL 更新沿
“在同一个 App 上部署 MOWL 新版”，
不需要候选。此路径使用 `deploy-candidate --workflow-id`，要求真实 Workflow Agent 候选和服务端
验收，不能把手工 MOWL/DSL 或普通 submit-candidate 的返回包装成 acceptance。

1. 固定原 App ID，保存原 definition/version、输入默认值和预期输出。向真实 system Workflow
   Agent（`agent_id: workflow`、`agent_workspace_id: system`）提出更新目标。按
   [Agent 入口](../domains/agent.md)和 `moi-cli schema agent a2a-stream` 准备
   `agent a2a-stream --json-file`：请求 method 为 `message/stream`，`params.message` 包含
   唯一 messageId、user role 和目标文本；`params.metadata` 显式给出 workspace_id 与
   `conversation_purpose: workflow`。模型选择遵循[模型规则](../domains/model.md)。
   保存服务端返回的 contextId 和 task ID。
2. 如果任务处于计划确认，读取真实计划。计划在已有授权内时，通过新 messageId、同一
   `params.message.contextId` 的正式消息
   发送文本 JSON `{"type":"workflow_plan_decision","decision":"approved","plan_task_id":"原计划task ID"}`。
   新增副作用超出授权时先补齐授权。保存续接后返回的真实 task ID，不把原计划任务当成
   已完成候选任务；流式调用超时后查询原 task，不重新发送创建消息。
3. 用 `agent a2a --json-file` 调用 `tasks/get` 查询该候选 task。只有任务 completed，且可见
   artifacts 中确实有 `metadata.data_type: workflow.candidate` 时才继续。读取该 artifact 的
   `artifactId` 和其 data 中的 `proposal_id`，检查候选图、输入要求、测试状态和真实测试输出。
   若模型最终文字中的 ID 不同，使用服务端 artifact 数据。只有 draft、未完成测试或没有
   proposal 时保留当前状态，不能伪造后续验收。
4. 接受用户范围内的候选，将下面命令的**完整原始返回 JSON**保存为 acceptance 文件；不要
   添加日志信封、删字段、改变 ID 或自行填 accepted_at。幂等键用于本次验收；失联先查原 task
   和已有验收收据，不生成新键盲目重复提交。

   ```sh
   moi-cli -o -no-interactive -workspace-id "$workspace_id" workflow accept-candidate --task-id "$candidate_task_id" --artifact-id "$artifact_id" --proposal-id "$proposal_id" --idempotency-key "$acceptance_key"
   moi-cli -o -no-interactive -workspace-id "$workspace_id" workflow deploy-candidate --workflow-id "$workflow_id" --name "$name" --acceptance-file "$acceptance_file"
   ```

5. 核对部署返回 `workflow.id` 仍为原 App，`deployment.previous_workflow_version_id` 为预期旧版，
   新版 ID/版本号已变化；用 `workflow versions --workflow-id` 与
   `workflow latest-published-version --workflow-id` 回读新版 published。候选 provenance 必须
   对应原 task/artifact；发布成功本身还不代表新版运行成功。
6. 按下文运行一次，确认 `get-run` 的 `moi_workflow_version_id` 等于新部署版本，检查本次真实
   输出，并在需要时回读旧 execution，确认旧结果仍能归属旧版。手工 create-version/publish-version
   的 App 指针问题仍单独保留，不能因 candidate 路径通过便判定它已修复。

`tasks/get` 的请求结构示例（文件中填入真实 ID）：

```json
{"jsonrpc":"2.0","id":"read-candidate-task","method":"tasks/get","agent_id":"workflow","agent_workspace_id":"system","params":{"id":"真实候选task ID"}}
```

### 自然语言创建与修改后的正式运行验收

候选测试与正式运行是两道独立检查。训练数据集、信息提取、多模态和 RAG 四类候选都要保存
真实 task/artifact/proposal，候选测试通过后接受并部署，再用真实隔离输入运行；不能只凭
`test_status=passed` 或 Agent 的完成说明把整个功能点标为通过。

2026-11-17 的生产隔离复核中，训练数据 3 条样例保留 2 条并丢弃 1 条；信息提取返回两个
订单及金额、来源 file ID；多模态读出图片标题、唯一标记、表格和框线。RAG 第一次正式运行
发现 `vector_table=schema.table` 被整体拿去匹配 `information_schema.tables.table_name`，
得到空库名并拼出 `FROM .schema.table`。随后在同一 Workflow Agent 会话中自然语言修改候选，
新候选同时覆盖带库名与裸表名测试，再部署为原 App v2；正式执行绑定 v2，回答每车 8 箱、
53 箱至少 7 辆，5 条引用与真实 model/source/file/chunk 一致。

这类解析修复应让带库名输入直接拆出 schema 和裸表名，只有裸表名才回查 information_schema；
检索与索引自检查询共用同一份全限定表名。表单准备可能规范化知识库句柄，不能只看测试输入
字面值就认定生产节点收到相同形状；需要分别核对候选测试、部署版本、正式执行和来源分段。
旧失败执行保留，不能用新版成功改写旧结果。

## 运行已配置 Workflow

适用：用户明确要求运行现有、已配置且可运行的 Workflow。

1. `workflow list` 找到用户选择的 workflow，保存真实 `workflow_id`。
2. `workflow get` 检查配置、输入、版本和 runnable 状态。若遇到已知 `unknown field "session_id"`
   解码错误，按[领域说明](../domains/workflow.md#发现和详情)保留问题，用本次明确的部署收据、
   版本图与已确认配置交叉核对；证据不足时不运行。
3. 先检查实际输入。省略 values 文件使用保存默认值；需要本次显式值时，以表单 field_id 为键
   构造 JSON，并执行一次正式 run。需要替换保存默认值时按当前 help 显式传
   `--replace-saved-defaults`；不要发明 `input_payload` / `vars_payload` 参数。

   ```sh
   moi-cli -o -no-interactive -workspace-id "$workspace_id" workflow run \
     --workflow-id "$workflow_id" --values-file "$values_file" --replace-saved-defaults
   ```

   Volume 非空且本次文件集合已确认后才能提交。保存原始返回；`--trigger-now` 只控制本次
   dispatch，不是业务完成状态。
4. 单次执行用真实 execution ID 查询；Volume 按[批处理规则](../domains/workflow.md#单次执行与-volume-批处理必须分开)用 `list-runs` 的 `dispatch_job_id` 关联子 execution，再逐个查状态，有界轮询。
5. 成功终态后用 `workflow get-run-result` 获取 case_result、case_error、node_states。
6. 按请求核验输出 artifact、表、文件或其它结果资产。

已观察到 get-run 使用 `execution.status`，结果使用 `result.status`，完成值为 `completed`；
不要套用 SQL 的 `success`。历史 execution/result 可只读检查合同，不证明本次新运行完成。
get-run 参数和 trace 可能含敏感内容，默认只投影 ID、状态与节点数量。

仅执行用户已授权的新建、更新、发布或运行目标，不能因一次运行失败自动扩展为删除或重跑；
结果失败时进入[作业恢复Recipe](inspect-and-recover-workflow-job.md)，先核对retry的新执行/当前版本语义和副作用。

## 按授权控制本次执行

需要暂停、恢复或取消时，使用已确认的单次或子 execution ID；Volume 的分发任务 ID 不适用于下列命令：

```sh
moi-cli -o -no-interactive -workspace-id "$workspace_id" workflow pause-run --workflow-id "$workflow_id" --execution-id "$execution_id"
moi-cli -o -no-interactive -workspace-id "$workspace_id" workflow get-run --workflow-id "$workflow_id" --execution-id "$execution_id"
moi-cli -o -no-interactive -workspace-id "$workspace_id" workflow resume-run --workflow-id "$workflow_id" --execution-id "$execution_id"
moi-cli -o -no-interactive -workspace-id "$workspace_id" workflow cancel-run --workflow-id "$workflow_id" --execution-id "$execution_id"
```

这些是分别按目标选择的操作，不是依次执行的固定脚本。暂停后回读 paused，恢复应沿用同一
execution/task/case/version，之后继续有界等待并核验结果。取消后回读 cancelled，并用结果接口
记录取消前已有的输出，不能把部分输出当成成功；它不撤销已发生的外部副作用。若暂停返回
completed，说明请求未使执行停住，不能记为暂停成功；终态判断和 trace 限制见
[领域说明](../domains/workflow.md#暂停恢复与取消执行)。

## 模板编辑、导出与导入

模板是独立对象，template ID 不是 Workflow App ID。先读对应 help；常规0.1.8已验证
`create-workflow-template`、`get-workflow-template`、`update-workflow-template` 和
`delete-workflow-template`。创建/更新用 `--name`、`--dsl-yaml` 内容，更新带真实 `--template-id`；
若 schema 报 CLI_SCHEMA_NOT_FOUND，不能只凭这一点判定该命令没有能力，继续核对正式 help 和输入合同。

导出可以把 get 返回的权威 dsl_yaml 原文保存到新本地文件；导入时读这个文件的内容交给
`workflow deploy`。核对模板回读、文件内容和部署版本，再只运行一次，读取本次结果。
本次从更新后模板导出再部署，真实输出48而非旧版42。无需另造“export”命令。
这只证明模板和手动运行路径；cron、Volume 和 dynamic service 另按
[触发器与 API 服务 Recipe](configure-workflow-triggers-and-service.md)验收，不能由模板成功推导。
清理只删除本次新建模板与App，不因删除模板顺带删除已有App。

## 失败后重放 SQL 写入

retry-run 会产生新的 execution，并可能重新执行已有写入节点，不是从失败节点自动续跑。
重试默认采用当前部署版本；本次两侧版本相同是因为没有更换部署，不能外推为自动锁定原失败版本。
执行前按[作业恢复Recipe](inspect-and-recover-workflow-job.md)比较当前部署与原失败版本。
先保存原execution/version、节点状态和目标表快照，修正已查明的前置问题，再判断重放是否获准且安全。
本次受控流程先写一条金额77的事件，后续节点故意失败；独立查询仍为1条77。
在同一业务唯一键下执行 upsert 后重试，同版本新execution完成且仍是1条77；原failed记录保留。
安全来自这个样例的显式业务键，不是Workflow全局自动去重或自动回滚。
消息发送、外部系统写入和取消后的补偿需要各自合同；没有恢复方案时不盲目retry。

## 类型合并、报表空集合与失败定位

正式工作项 `moi:data.records.join v1` 与 `moi:text.render.blocks v1` 已完成小集合组合验收。左连接保留左侧顺序，数字键与字符串键分别匹配，null键未匹配；先写独立预期，再核对逐行类型、数量和输出。render内部的where/sort/limit可处理筛选、稳定排序、limit为0和空数组，但它们不是Workflow层的if/foreach。没有正式DSL合同和真实验收时，不承诺通用条件分支或循环。

连接未命中产生的null可以存在于结构结果中，但本轮render的标量占位符拒绝null，返回`DATA_TRANSFORM_TYPE_MISMATCH`及实际字段名。按明确业务规则决定保留、排除或替换，不能擅自把null改空串。本轮嵌套绑定`data: {rows: "{{ .state.joined }}"}`把结构数组原样交给节点；这是限定结构的证据，不推导任意表达式语法。

failed时读取已完成节点的runtime_output及失败节点的runtime_input/error；最终case_result可能不存在。本轮重复lookup键和null占位符负例均保留前置输出，后置节点无trace且查询404；只说明这些节点未观察到执行，不证明外部副作用自动回滚。

本轮另一份对象构造表达式DSL部署返回500且没有App/version ID，根因和是否创建对象均未知。正式list/apps仅返回前20项、total更大且无分页输入时，名称没出现不能证明未创建；不重传、不猜ID，也不能声称全部清理。另行授权的嵌套绑定变体通过输入核验，不会修复原500或解决其未知副作用。
