# MOI Canonical Capability Catalog

这是有限业务目标的发现目录，21 个 capability ID 保持稳定。菜单、UI 控件、节点或命令数量不代表业务覆盖率。
新增接口先归入已有能力；一个能力可以映射多个菜单，一个菜单也可以映射多个能力。
具体输入输出、完成断言和恢复要求见 [能力卡片](capability-cards.json)。

唯一状态源是卡片中的 `observations`，下表由它生成。`static_support` 只说明文档的层级和组合形态，
没有能力级“通过”状态。计数按场景/条件/CLI 渠道统计，不能解释为功能完成率。
同一能力中的成功、失败和缺口同时保留；候选构建成功不会升级常规 CLI。

<!-- capability-status:begin -->
来源：`capability-cards.json`（SHA256 `2619e43c50d1c1373c34d7e2fbddd1623896c9a3633857147fb11e135e7aca02`）。此段由 `scripts/capability_status.py render` 生成，修改JSON后重生成；手写说明留在标记外。

| capability_id | 业务目标 | 对象生命周期 | 层级/静态形态 | 风险 | regular场景状态 | candidate场景状态 | 最小证据 |
|---|---|---|---|---|---|---|---|
| catalog-discovery | 发现并确认 Catalog 数据资产 | database/table/volume/file: listed → inspected | L1/single_domain | read | passed×2, partial×1, cli_missing×2 | passed×7 | 真实资源 ID 与详情/schema |
| sql-query | 对固定 workspace 资产执行查询并取结果 | query/statement: submitted → running → result/failed | L1/single_domain | read | passed×14, failed×4, partial×1 | passed×16, failed×6, partial×1 | query/statement ID + result metadata/preview |
| source-staging | 把用户选定的文件放入可导入源位置 | file/source: selected → uploaded → addressable | L1/single_domain | write | passed×1, partial×1 | passed×2 | file/source ID、大小/类型和可读详情 |
| data-ingestion | 把 source 导入目标表并确认数据可读 | import task: created → running → terminal；table: created/updated | L2/composed | write | passed×6, failed×2, partial×1 | passed×7, failed×2, partial×2 | task terminal + target table + bounded SQL |
| data-export | 把源表或文件导出到指定目标并核验产物 | export task: created → running → terminal；artifact: produced → verified | L2/composed | write | passed×1, blocked×1 | passed×2, blocked×1 | task state + artifact hash/目标侧可见 |
| workflow-lifecycle | 创建、版本化、运行并核验 Workflow | workflow: draft → versioned → runnable；execution: submitted → terminal | L2/composed | write | passed×6, failed×2, cli_missing×1 | passed×11, failed×2 | workflow/version/execution ID + result/node details |
| job-recovery | 诊断并恢复失败或未知的异步作业 | job/execution: pending/running → terminal/unknown | L2/composed | controlled_write | passed×2, failed×1 | passed×3, failed×1, partial×1 | 原对象状态、错误原因、恢复动作结果 |
| knowledge-preparation | 建立知识源并确认可检索 | knowledge: created；source job: queued → prepared/failed；source: indexed/disabled | L2/composed | write | passed×8, failed×2, partial×1, cli_missing×1 | passed×20, failed×1 | source job terminal + source detail + retrieval evidence |
| dashboard-analytics | 创建图表并执行查询得到结果 | dashboard/chart: draft → configured；chart run: submitted → result/failed | L2/composed | write | passed×8, failed×1, partial×1 | passed×9, failed×1, partial×1 | chart/run ID + result metadata/preview |
| data-share | 发布数据并在目标工作区确认可见 | publish: draft → published；subscription: requested → active/failed | L3/controlled | permission_sensitive | passed×1, failed×1, partial×1 | passed×3, partial×1 | publish/subscription IDs + target visibility |
| workspace-context | 固定工作区和当前主体上下文 | workspace: discovered → selected；principal: identified | L0/read_only | read | passed×1 | passed×1 | workspace ID、principal/roles 返回 |
| connector-lifecycle | 创建或验证数据连接 | connector: draft → validated → active/failed | L1/single_domain | permission_sensitive | blocked×1, cli_missing×1 | passed×1, blocked×1 | connector 详情和服务端验证结果 |
| compute-lifecycle | 管理计算资源并确认运行状态 | compute: creating → active/idle → suspended/deleting/error | L3/controlled | permission_sensitive | passed×1, blocked×1 | passed×1, blocked×1 | resource/runtime/preflight details |
| audit-observation | 查询操作和 SQL 审计事实 | audit/query history: recorded → filtered → inspected | L0/read_only | read | passed×1, failed×1, partial×2, cli_missing×1 | passed×5 | log/query detail with server timestamps |
| member-lifecycle | 受控邀请、更新、禁用或删除成员 | member: invited/active/suspended/deleted | L3/controlled | permission_sensitive | blocked×1 | blocked×1 | member detail + audit record + effective visibility |
| iam-authorization | 查看和变更角色、binding、grant | role/binding/grant: observed → changed → verified | L3/controlled | permission_sensitive | passed×1, failed×2, partial×1, blocked×1 | passed×5, blocked×1 | Core/PDP response + binding/grant reread |
| agent-conversation | 向 Agent 发起对话/A2A 并取得结果 | conversation/task: submitted → running → result/failed | L2/composed | write | passed×2, failed×2 | passed×2, failed×2, partial×1 | task/conversation/result IDs + result evidence |
| agent-package-resource | 构建、验证、加载 Agent package 并确认资源绑定 | package: built → verified → loaded；resource: bound → usable | L2/composed | controlled_write | passed×2, failed×1 | passed×4 | package verification + loaded/resource detail + run check |
| agent-skill-tool | 管理 Skill、Tool、Plugin 包和连接 | package/connection: discovered → installed/verified | L3/controlled | permission_sensitive | passed×6, failed×2, blocked×1, cli_missing×1 | passed×8, failed×1, blocked×1, cli_missing×1 | package/connection detail and safe verification |
| agent-runtime | 关联 compute、model、provider 与执行器 | dependency: discovered → healthy/disabled；task: uses snapshot | L3/controlled | controlled_write | passed×1, partial×2, cli_missing×1 | passed×4, failed×1, partial×3 | dependency IDs + executor health + task detail |
| agent-automation | 管理定时/API任务和渠道规则并检查实际触发 | automation task: disabled → active → disabled/archived；run: triggered → terminal；event另验 | L3/controlled | permission_sensitive | passed×8, failed×2, partial×3, unverified×1 | passed×12, failed×2, partial×1, blocked×1 | task/run/result/events + trigger_type；渠道事件另证 |
<!-- capability-status:end -->

## 查询和维护

从 Skill 根目录执行：

```sh
python3 scripts/capability_status.py query sql-query --scenario async-zero-row-download
python3 scripts/capability_status.py query workflow-lifecycle --scenario deploy-volume-defaults --channel candidate
python3 scripts/capability_status.py validate
python3 scripts/capability_status.py render
python3 scripts/capability_status.py render --check
```

查询默认渠道为 `regular`。只给能力 ID 用于发现，返回 `unverified` 和 `needs_scope_match=true`；
未匹配到指定场景或条件也如此。先用返回的场景、条件、来源和限制核对本次目标，再判断是否适用。
即使返回 `passed`，也只针对所返回条件；条件不同需另记场景，不能借其它成功替代。

字段、筛选器、分页、画布控件和重复子树是输入或观察状态，不单独成为能力。
能力 ID 只在业务对象或完成合同根本变化时调整。目录、[菜单映射](capability-inventory.md)和
[CLI 覆盖矩阵](coverage-matrix.md)都是同一 JSON 的派生视图；不要分别手改状态。
来源、日期、历史替代规则见[能力模型](capability-model.md)。
