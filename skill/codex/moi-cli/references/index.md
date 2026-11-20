# CLI domain index

Generated references describe command availability, not environment readiness.
Use --help for defaults and schema for typed JSON.
Read a linked document without filesystem discovery using
`moi-cli skill read --path references/commands/DOMAIN.md` with the actual domain.
Command presence is not proof that every required SDK/backend input is exposed.
For example, Catalog file placement/listing currently has the gaps documented
in [files](files.md). Stop on a missing required input instead of guessing flags.

- [workspace](commands/workspace.md), [workspace-member](commands/workspace-member.md)
- [catalog](commands/catalog.md), [connector](commands/connector.md), [sql](commands/sql.md)
- [workflow](commands/workflow.md), [semantic-model](commands/semantic-model.md)
- [agent](commands/agent.md), [model](commands/model.md), [provider-backend](commands/provider-backend.md)
- [compute](commands/compute.md), [dashboard](commands/dashboard.md)
- [data-share](commands/data-share.md), [iam](commands/iam.md), [account](commands/account.md)
- [config](commands/config.md)

把有限功能组织为可编排能力时，先读 [能力模型](capability-model.md) 和 [Canonical 能力目录](capability-catalog.md)；命令 surface 只提供原料，
不直接定义覆盖完成度。

- [有限能力目录](capability-inventory.md)
- [Canonical 能力目录](capability-catalog.md)
- [能力卡片](capability-cards.json)（按 capability_id 查输入、输出、完成断言与恢复点）
- [能力边界合同](contracts/capability-boundary.md)
- [部分通过的补测与收口](contracts/partial-coverage.md)
- [业务目标编排合同](contracts/task-composition.md)
- [覆盖矩阵](coverage-matrix.md)

按实际目标继续：

- [SQL工作簿与版本](recipes/manage-sql-workbook.md)
- [下载与核对](recipes/export-and-verify.md)
- [CLI结果的本地筛选、排序和分页](recipes/query-cli-results-locally.md)
- [知识维护与复用](recipes/maintain-knowledge-sources.md)
- [Catalog → Workflow → Knowledge 端到端血缘](recipes/catalog-workflow-knowledge-pipeline.md)
- [工作流运行时检索知识库](recipes/run-runtime-knowledge-workflow.md)
- [作业失败、版本与定向恢复](recipes/inspect-and-recover-workflow-job.md)
- [解析节点参数与结果核验](recipes/verify-parser-workitems.md)
- [数据/知识/报告组合](recipes/compose-data-knowledge-report.md)
- [Agent 工具与文件](recipes/agent-tools-and-deliver-files.md)
- [托管 Skill 脚本与冻结版本](recipes/use-script-skill-attachments.md)
- [Agent 定时/API 任务](recipes/configure-agent-automation.md)
- [计算资源](recipes/manage-compute-lifecycle.md)
- [MOI插件能力](recipes/manage-plugin-capabilities.md)
- [实测范围](verified-capabilities.md)

高风险编排 Recipe：

- [受控管理 Workspace 成员](recipes/manage-workspace-members.md)
- [审查并变更 IAM](recipes/review-and-change-iam.md)
- [配置 Agent Team 或 Ingress](recipes/configure-agent-team-ingress.md)（当前为 partial）

Administrative and permission changes require explicit intent and normal server
authorization, never automatic recovery. MOI-hosted agent/skill resources are
separate from this local Skill. Do not expose credentials from configuration.
