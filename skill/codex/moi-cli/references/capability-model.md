# MOI 业务能力模型

用途：把 24 个产品菜单内的有限业务能力抽象为可编排对象，而不是把 UI 控件数或 CLI 命令数当成覆盖数。
本文件是每个领域和 Recipe 的共同 schema；具体对象状态机、命令组合和当前缺口分别擅长在
domain 和 Recipe 文档中维护。

## 1. 能力记录 schema

每个能力卡片都必须具备以下合同事实。卡片存在及静态组织形态均不代表线上验收通过：

```yaml
capability_id: data-ingestion
business_goal: 把用户选定的本地文件导入指定的目标表，并确认目标表可读
menu_refs: 概览 / 导入任务 / Catalog / SQL 编辑器
objects:
  workspace: { id_field: id, discover: workspace list, primary_id: workspace_id }
  connector: { id_field: connector_id, discover: connector list/get }
  upload: { id_fields: [conn_file_ids, file_ids], discover: connector local-upload }
  import_task: { id_field: task_id, discover: connector tasks/get-task }
  table: { id_field: table_id, discover: catalog tree/list-databases/database-children }
  query: { id_field: query_id/statement_id, discover: sql execute/describe-query }
lifecycle:
  uploaded: file_id exists
  task_created: task created and not rejected
  task_terminal: run succeeded/failed/cancelled/partial
  target_verified: catalog table-info or bounded SQL result matches expected file scope
inputs:
  required: [workspace_id, source_file_path, destination_rule, user_selection]
  forbidden: [guessed ids, name-derived ids, credentials in chat]
commands:
  discover: [workspace list, connector list, catalog tree]
  execute: [connector local-upload, connector create-task, connector run-task-now]
  verify: [connector get-task, connector task-runs, catalog table-info, sql execute]
recovery:
  duplicate_guard: 超时后先 get task / task-runs，不重复上传或 create
  evidence_ids: [workspace_id, file_id, task_id, table_id, query_id, statement_id]
completion_evidence:
  minimum: 本次任务终态成功，目标表存在且 SQL/Catalog 核验与文件范围一致
  failure_evidence: 失败终态可以结束观察，但不能证明业务目标完成
  not_evidence: UI 按钮存在、任务 accepted、上传收据、同名对象
risk:
  level: write-async
  permissions: workspace resource create/import
  user_requirements: 明确文件、目标、覆盖语义
static_support:  # 文档层级/组合形态；不是验收状态
  level: L2
  shape: composed  # read_only|single_domain|composed|controlled
observations:  # 同能力可同时有多项成功、失败、缺口；不设能力级status
  - observation_id: data-ingestion:example:regular:observation-1
    scenario_id: example
    status: unverified  # passed|failed|partial|blocked|cli_missing|unverified
    conditions: {scope: 本次明确目标和配置}
    cli: {channel: regular, version: 0.1.8}  # candidate必须独立记录
    recorded_on: '2026-11-02'
    validated_on: null
    date_note: 尚未执行；记录日期不是验证日期
    evidence_kind: not_run  # reported_live|cli_contract|not_run
    result: 未验证本场景
    limits: [不得借其它场景通过外推]
    sources: [{source_id: source-key, record: 来源内的场景或记录标识}]
```

## 2. 工作节点 vs 业务对象

功能清单中的节点、字段和筛选器先作为组合原料。节点也可支撑独立用户目标，例如解析并交付文件；
是否纳入能力由目标和完成合同决定，不由它属于 WorkItem 还是 API 决定：

1. 有稳定业务对象，以及可发现、可核验的真实 ID。
2. 有明确状态迁移或同步状态字段，能判断 accepted/running/succeeded/failed/cancelled。
3. 有业务结果或目标端证据，能把提交事实、运行事实和完成事实分开。

同一个“存至 MOI 表”节点可参与多个业务目标；同一目标也可有多个实现。不要按节点数量统计能力。

## 3. ID 和关系禁止项

- 每个 ID 只在所属对象和接口中使用；`file_id` 不是 `source_id`，`task_id` 不是
  `execution_id`，`query_id` 不是 `statement_id`。
- 组合 Recipe 只能连接对象间的 server/library-level reference，例如 source file、
  destination table、workflow execution、knowledge source、chart result；不能靠名称匹配。
- UI 菜单名称、节点展示名、截图顺序和列表位置只能用于发现上下文，不能作为 ID。
- Additional command、alias 和未导出的 Product SDK method 只能作为实质命令面证据；
  不能替代 schema reference 和终态查询。

## 4. 能力分层

### L0：对象发现

只回答目标是否存在、当前主体能否看见、真实 ID 是什么。
常见证据：list/tree/get/detail/schema/metadata、current-principal/roles、allowed-actions。

### L1：单域闭环

一个 domain 内完成用户目标，并能在不跨 domain 猜测的情况下核验结果。
例如 SQL 查询取结果、Workflow 运行取 execution result、Knowledge source job 到 source 状态。

### L2：跨域组合闭环

多个对象按真实 ID 传递形成业务闭环。
例如 local file → Connector task → table → bounded SQL；file → knowledge source/job → searchable
source；dashboard → chart → chart result；publish → subscription → target workspace visibility。

### L3：受控治理闭环

包含成员、IAM、Data Share、Compute 生命周期、删除、覆盖、发布、跨 workspace、凭证或外部渠道。
除完成证据外，还必须有影响范围、当前主体能力、pre-flight/impact 或等价核验。

## 5. 唯一状态源与查询

[capability-cards.json](capability-cards.json) 同时拥有能力合同、静态支持、24 菜单映射和场景观测。
`schema_version=1` 保留既有合同，`registry_version=1` 标明场景状态结构。目录、覆盖矩阵、菜单目录与
实测范围的生成标记内内容均从它派生；手写说明留在标记外，不另存可独立修改的当前状态。
历史实测正文作为来源保留，不自动随当前状态改写。

`static_support.level/shape` 说明文档如何组织能力：`read_only`、`single_domain`、`composed`、`controlled`。
风险与授权仍用卡片 `risk` 及执行合同；静态形态不含 `ready/full`，不代表可执行或已通过。

场景状态有六种：

- `passed`：该来源报告的业务目标在所列条件和渠道下完成，仍受限制约束。
- `failed`：实际执行没有达成该目标或出现已证实的错误/副作用；其它条件成功不能覆盖。
- `partial`：该场景只完成部分目标，剩余结果、合同或观测仍未确认；明确列出实际失败，不用部分完成淡化失败。
- `blocked`：缺少具体前置或授权，本场景未执行；不等于产品能力失败。
- `cli_missing`：正式 CLI 缺少所需输入、作用域或稳定操作入口；不是缺用户信息。
- `unverified`：尚未取得该场景结论，或查询条件未命中；不是隐含通过。

每条 `observations` 必须记录场景 ID、精确条件、结果、限制、验证日期及其来源说明、记录日期、
CLI 渠道与版本、证据类型和来源内的记录标识。缺少逐项验证日期时使用 `validated_on: null`，
在 `date_note` 说明未知；`recorded_on` 只记录登记日。`not_run` 不允许产生通过/失败或验证日期。
`reported_live` 表示迁移自已有真实测试报告，不意味着登记工具执行了线上复测。

查询默认只选 `regular`；候选开发构建需显式 `--channel candidate`，即使自报版本数字更小，也不能冒充常规发布。
只给能力 ID 用于发现；返回 `unverified`、`needs_scope_match=true` 及可选场景。指定场景/条件未命中时同样返回该结果。
仅匹配到一组状态时才返回其状态；匹配多个状态则保留各状态，不压成能力级结论。
`all_matched_passed` 也仅指返回的场景和过滤条件，不证明用户未描述的范围。

从 Skill 根目录执行以下本地脚本，不调用 MOI：

```sh
python3 scripts/capability_status.py query agent-automation --scenario automation-revisions-ar-04
python3 scripts/capability_status.py query sql-query --condition 'scope=原异步零行statement结果及CSV'
python3 scripts/capability_status.py query workflow-lifecycle --scenario deploy-volume-defaults --channel candidate
python3 scripts/capability_status.py validate
python3 scripts/capability_status.py render
python3 scripts/capability_status.py render --check
```

`--root /absolute/path/to/moi-cli` 可放在子命令之后选择独立包。没有任何仓库证据目录也可查询、校验及渲染。
脚本只用 Python 标准库，不访问网络、产品、凭据或外部证据路径。

## 6. 来源、历史与派生视图保护

`sources` 有两种来源语义：`packaged_reference` 指向包内 references 文件，校验其存在；
`external_evidence` 包含 `origin/locator/availability=external_not_bundled`，只描述外部来源，不要求文件随包存在。
原历史正文与 complex 场景源文件的摘要用于追溯登记基线，不把摘要当成新的线上回执。

复测新增观测，不覆盖原观测。只有相同能力、场景、条件和 CLI 渠道的显式 `supersedes` 可以替代旧状态；
不得跨条件/渠道替代，不允许自指、循环或同场景渠道的重复有效记录。不同条件使用不同场景 ID。
默认视图只取有效记录；`query --history` 仍显示原失败与替代关系。消除已知失败需同范围、有明确验证日期的
外部真实复测来源；未复测的静态修改、候选渠道或其它成功均不能使它消失。

四份 Markdown 视图只渲染 `capability-status:begin/end` 标记内内容，保留外部手写说明。
`render --check` 对陈旧内容返回非零且不写文件；`validate` 拒绝非法状态、重复场景、非法替代、
缺来源/日期说明和能力级状态。两项均通过只表示本地登记一致，不表示产品新通过。

可用性、风险和任务执行范围仍分别记录，详见[业务目标编排](contracts/task-composition.md)。
