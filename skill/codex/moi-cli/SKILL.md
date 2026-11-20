---
name: moi-cli
description: Use the local dev moi-cli for MOI workspace, Catalog, Connector, SQL, Workflow, Knowledge, Dashboard, Data Share, IAM and Agent tasks, with the latest verified dev procedures and recovery contracts.
metadata:
  skill-version: 5.2.3
  package: codex-moi-cli
---

# MOI product tasks

## Codex 执行适配

在 Codex 中通过本地终端工具执行旁路 dev CLI。CLI 路径使用
`MOI_DEV_CLI_BIN`，未设置时为 `$HOME/.local/bin/moi-cli-dev`；调用时使用
`${MOI_DEV_CLI_BIN:-$HOME/.local/bin/moi-cli-dev}`。每次终端调用只执行一个
独立的 CLI 命令。本 Skill 统一使用 `$moi-cli` 调用名，并明确使用上述 dev CLI；
不要回退到旧稳定 CLI 或旧 Skill 代替本版能力。

`moi-cli` 通过现有身份调用 MOI Product API。不要替换成 `core-cli`，不要直接访问
数据库，不要读取凭证文件，不要输出 token、密码或 Cookie，也不要在聊天中索要这些
凭证。native 登录确实失效时才让用户在本机执行 `astra login`；自动化 API 的 PAT
前置与登录态不同，按[自动化认证边界](references/domains/agent-automation.md)判断，不能见401就要求重登。

本 Skill 以产品清单中的 24 个当前菜单为范围基线。产品清单中的字段、分页控件、重复实例和
预留入口用于发现范围，不等于每条界面记录都有一个 CLI 命令。以业务能力、CLI 命令
和端到端 Recipe 组织能力；纳入范围不等于每项都已在线验收。

## 统一执行合同

首次远程操作前，按[环境预检查](references/contracts/environment-preflight.md)运行
`scripts/preflight.py`，区分本机、native/PAT、目标工作区及权限详情的可用状态。
条件未变时复用结果；不自动登录、修改配置或选择工作区，未检查的通道不能说已通过。
后续命令沿用已确认的 `-endpoint`、认证通道及程序路径；终端工具每次启动独立进程时，
每次显式传入相同环境变量，不能假定上一次 shell 的 `export` 会保留。
dev 的版本基线和动态服务字段差异见[兼容合同](references/contracts/dev-compatibility.md)。

执行前先按[能力边界合同](references/contracts/capability-boundary.md#执行前范围判断)判断用户目标是否在当前能力和已验证场景内；超出覆盖范围时，先明确提示操作者具体缺口，再按其授权决定是否进行隔离探索。菜单或命令存在不等于场景已覆盖。

所有任务按“发现 → 执行 → 状态 → 核验”推进，具体规则见：

- [执行合同](references/contracts/execution.md)
- [异步状态](references/contracts/async-state.md)
- [失败恢复](references/contracts/failure-recovery.md)
- [证据记录](references/contracts/evidence.md)
- [部分通过的补测与收口](references/contracts/partial-coverage.md)
- [风险与确认](references/contracts/risk-and-approval.md)
- [能力边界](references/contracts/capability-boundary.md)
- [文件输入与冻结](references/contracts/cli-file-inputs.md)
- [CLI 正式合同缺口](references/contracts/cli-contract-gaps.md)

执行前用 `scripts/capability_status.py query` 查询[唯一场景状态源](references/capability-catalog.md#查询和维护)，
按条件和常规/候选 CLI 渠道匹配；静态支持形态不能当成已验收成功。复杂任务使用
[本地任务记录与接续](references/recipes/record-and-resume-task.md)，保存范围、依赖和逐次证据。
Catalog→Workflow→Knowledge、Workflow 候选创建修改和 Agent 双版本任务，优先从
[复杂任务模板](references/recipes/use-task-blueprints.md)建立计划；每次继续前用
`scripts/task_blueprint.py recover` 做只读防重判断，未知提交不得重发。

多步任务必须保存前一步返回的真实 ID，再把它作为后一步唯一输入。不能按名称、列表
顺序或模型猜测 workspace、资源、任务、查询、Workflow 或执行 ID。

三条主链路的默认深测要求：

- Workflow：Volume/批处理返回的父 dispatch、文件 execution 和普通 one-shot execution 必须分开保存；
  `list-runs` 后逐个读取 `get-run`、`get-run-result`，并在结果可用时核对 `case_result`、节点状态和 trace。
  父 dispatch 不能直接当 execution 查询；父 ID 查询 404 时先按 `dispatch_job_id` 展开子 execution，再判断是否失败。
- Catalog：交给 Workflow 或 Knowledge 前先读 `volume-info`，再用真实 file ID 做 `file-info`/预览或下载；
  文件数量、文件详情、任务结果和 SQL/产物必须相互印证。输入文件没有派生 artifact 时，
  `catalog-file-lineage-overview` 的 NotFound 只表示没有输出血缘，不能单独判成 Catalog 失败。
- Knowledge：准备完成后仍要读取 `list-sources`、每个目标 source 的文档/分段和 ETag；
  新会话问答必须同时核对最终状态、答案事实、检索工具和 `knowledge.answer` 的正式 `source_refs`，
  且引用要匹配当前 model/source/file/chunk。只看到 ingest succeeded 或模型说“已引用”都不算问答通过。

## 任务边界

- 查询、列表、详情、schema、metadata 和本地 `config show` 属于只读操作，可以直接执行。
- 上传、导入、Workflow 运行、Knowledge 变更、Dashboard/Data Share 写入和 Agent 资源变更，只有用户明确给出目标和动作时才执行；目标不完整时先补齐 workspace、资源 ID、文件或输入。
- 删除、覆盖、发布、取消、重试、成员、IAM、跨工作区和凭证相关操作必须锁定具体对象、影响范围和确认语义；不能自动恢复或替用户选择对象。
- 命令存在不代表环境、权限、后端输入和业务终态都可用；遇到缺口时保留 `partial` 或 `cli_missing`，不要猜参数或切换低层接口。

## Workspace 与 ID

列出账号可见 workspace：

```sh
"${MOI_DEV_CLI_BIN:-$HOME/.local/bin/moi-cli-dev}" -o -no-interactive workspace list
```

workspace-scoped 命令显式携带 `-workspace-id ID`。未明确指定时可复用当前会话已确认的
默认 workspace；若 CLI 报告缺少 workspace，先列出可见列表并请求选择。不要选择第一项，
不要创建 workaround workspace。

全局选项放在 domain 前：

```sh
"${MOI_DEV_CLI_BIN:-$HOME/.local/bin/moi-cli-dev}" -o -no-interactive -workspace-id ID DOMAIN COMMAND ...
```

## 相关流程入口

复杂任务先读 [业务目标编排](references/contracts/task-composition.md)，明确结果、输入输出引用和恢复点。
生产环境涉及模型调用时，先读[模型选择与推理等级](references/domains/model.md)：优先使用
`deepseek-v4.1-flash`，默认不传推理等级参数，并核对目标功能对应的模型通道；用户明确指定时遵循其选择。

先读与用户目标最接近的领域文档，再读需要跨领域组合的 Recipe：

- [完整 domain index](references/index.md)
- [覆盖矩阵](references/coverage-matrix.md)
- [Catalog 与 SQL](references/domains/catalog.md)
- [看板运行、失败恢复与结果核验](references/domains/dashboard.md)
- [能力模型](references/capability-model.md)
- [Canonical 能力目录](references/capability-catalog.md)
- [有限能力目录](references/capability-inventory.md)
- [能力边界](references/contracts/capability-boundary.md)
- [Connector 与导入导出](references/domains/connector.md)
- [Workflow 与作业](references/domains/workflow.md)
- [SQL 查询闭环](references/recipes/run-sql-and-fetch-result.md)
- [SQL 工作簿与版本](references/recipes/manage-sql-workbook.md)
- [文件导入到表](references/recipes/import-file-to-table.md)
- [Workflow 运行与恢复](references/recipes/create-and-run-workflow.md)
- [解析节点参数与结果核验](references/recipes/verify-parser-workitems.md)
- [作业失败、版本与定向恢复](references/recipes/inspect-and-recover-workflow-job.md)
- [知识库准备](references/recipes/create-and-prepare-knowledge-base.md)
- [知识库问答与引用](references/recipes/query-knowledge-and-check-citations.md)
- [知识来源复用与分段维护](references/recipes/maintain-knowledge-sources.md)
- [数据下载与内容核对](references/recipes/export-and-verify.md)
- [CLI 结果的本地筛选、排序和分页](references/recipes/query-cli-results-locally.md)
- [数据、知识与报告组合](references/recipes/compose-data-knowledge-report.md)
- [Catalog → Workflow → Knowledge 端到端血缘](references/recipes/catalog-workflow-knowledge-pipeline.md)
- [工作流运行时检索知识库](references/recipes/run-runtime-knowledge-workflow.md)
- [智能体工具与文件交付](references/recipes/agent-tools-and-deliver-files.md)
- [Agent 包构建、版本与实际运行](references/recipes/build-load-and-use-agent-package.md)
- [托管 Skill 脚本与冻结版本](references/recipes/use-script-skill-attachments.md)
- [定时/API 自动化](references/recipes/configure-agent-automation.md)
- [计算资源生命周期](references/recipes/manage-compute-lifecycle.md)
- [实测范围与保留缺口](references/verified-capabilities.md)
- [Agent Team / Ingress 扩展](references/domains/agent-extended.md)
- [Catalog Runtime / Workspace File](references/domains/catalog-runtime-files.md)
- [沙箱 file](references/domains/file-runtime.md)
- [操作日志](references/domains/operation-logs.md)
- [Agent 自动化](references/domains/agent-automation.md)
- [成员管理 Recipe](references/recipes/manage-workspace-members.md)
- [IAM Recipe](references/recipes/review-and-change-iam.md)
- [Agent Team / Ingress Recipe](references/recipes/configure-agent-team-ingress.md)
- [插件发现、安装与实际调用](references/recipes/manage-plugin-capabilities.md)
- [复杂任务模板与安全恢复](references/recipes/use-task-blueprints.md)

CLI 自带的命令 reference 是命令可用性的真源，通过已安装 CLI 离线读取：

```sh
"${MOI_DEV_CLI_BIN:-$HOME/.local/bin/moi-cli-dev}" skill read --path references/commands/DOMAIN.md
```

使用 `moi-cli schema DOMAIN COMMAND` 检查 typed JSON 输入，用 `--validate-file PATH` 做结构
校验。校验通过不代表资源存在、权限满足或业务输入有效。

## 文件与副作用

- 只上传用户选定的本地文件；文件路径不会自动成为上传授权。
- 不把整文件或 Base64 放进模型上下文，只读取必要的小预览。
- 文件上传、Catalog 挂载、Connector 导入、Knowledge source 和 Agent workspace file 是不同目标，必须先确认目标。
- CLI 超时、连接断开或返回未知状态时先查询原对象；不能因为没有收到结果就再次提交写操作。
- 创建后的响应解码失败也可能已经产生对象。先按本域正式发现入口找回并核实真实 ID；找不到就保持未知，不重复创建或把新对象当原结果。
- 不通过 raw HTTP、数据库、core-cli、复制凭证或切换身份绕过 Product API 和权限边界。

## 完成标准

面向用户的进度、验收记录和最终报告遵循[报告表达](references/contracts/reporting.md)：先说实际做成了什么、哪里没通过、接下来做什么；技术 ID 和命令放到证据附录。

完成必须以 CLI 返回、后续详情/结果查询、文件校验或目标资源查询为证据。提交成功、按钮
存在、列表出现或模型自述完成都不足以证明业务终态。

报告实际命令产生的对象、状态、ID、证据、未完成步骤和限制。异步任务在有限预算内仍
pending 时报告 pending，不伪装成成功。保留服务端 `code`、`reason`、`msg` 或 `message`，
不自行改写根因。

每次版本迭代还必须按[迭代报告规范](references/contracts/iteration-reporting.md)同步交付横向、
纵向、原版对比和总体进展报告，以及 `report-manifest.json`。三份主报告、24 个产品入口、
真实场景、失败与限制缺一都不能把本次迭代称为完整验收。
当前 5.2.3 的[总体进展报告](references/iterations/5.2.3/总体进展报告.md)、
[横向纵向功能报告](references/iterations/5.2.3/横向纵向功能报告.md)和
[原版对比报告](references/iterations/5.2.3/原版与dev版功能对比报告.md)已随包提供；
[失败原因分类](references/iterations/5.2.3/失败原因分类.md)逐项说明现有失败、部分通过、
缺条件和 CLI 缺口，恢复方法不会改写原失败路径；
125 条生产场景结果也可直接在包内复查。
功能清单 5.0 的补测另见[本轮 6 项结果](references/iterations/5.2.3/6项待补测结果.md)
和[最新 141 项 CLI 工作统计](references/iterations/5.2.3/CLI适合项-最新统计报告.md)；
308 项完整清单状态、141 项当前工作范围与原 dev 场景库分别统计，不能混用分母。
本版的[复杂任务模板与恢复合同](references/iterations/5.2.3/复杂任务模板与恢复合同.md)
记录任务模板、防重和文件参数的实际测试，不把工具改进冒充新增产品功能通过。
