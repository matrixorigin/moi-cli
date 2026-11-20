---
name: moi-cli
description: Use release-matched moi-cli procedures for MOI workspace, Catalog, Connector, SQL, Workflow, Knowledge, Dashboard, Data Share, IAM and Agent tasks in prepared MOI conversations and native CLI operation.
metadata:
  skill-version: 5.2.3
  package: moi-cli
---

# MOI product tasks

## MOI 原生执行适配

本 Skill 以最新 dev 的能力目录、执行合同、恢复规则和领域 Recipe 为内容基线，名称与仓库产品版本统一为 `moi-cli` / `5.2.3`。

在 prepared MOI conversation 中，先读 workspace `README.md`、`context.json` 和[运行时合同](references/runtime.md)，使用平台注入的 `bin/moi-cli`、task-scoped 身份和 workspace。调用形式为 `moi-cli -o DOMAIN COMMAND`，全局 `-o` 位于 domain 前。

不要读取本机登录或配置文件，不输出凭证，不修改 HOME，不执行 `astra login`，也不覆盖平台注入的 endpoint、PAT、Role 或 workspace。文档中的 `MOI_DEV_CLI_BIN`、`~/.local/bin/moi-cli-dev` 和本机 preflight 仅适用于独立 Codex，不用于 prepared MOI 运行态。

## 统一执行合同

首次远程操作前，prepared MOI 按[运行时合同](references/runtime.md)核对任务上下文和注入身份；
独立 Codex 才按[环境预检查](references/contracts/environment-preflight.md)运行 `scripts/preflight.py`。
两种模式都要区分 native/PAT、目标工作区及权限详情的可用状态，未检查的通道不能说已通过。
后续命令沿用已确认的 `-endpoint`、认证通道及程序路径；终端工具每次启动独立进程时，
每次显式传入相同环境变量，不能假定上一次 shell 的 `export` 会保留。
独立模式的版本基线和动态服务字段差异见[兼容合同](references/contracts/dev-compatibility.md)。

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

## 原生任务入口

任务边界、Workspace 与 ID、相关流程、文件副作用和完成标准统一见[原生任务入口](references/contracts/native-task-entry.md)。内容沿用最新 dev 基线，不删减其执行要求。
