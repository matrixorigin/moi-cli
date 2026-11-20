## 任务边界

- 查询、列表、详情、schema、metadata 和本地 `config show` 属于只读操作，可以直接执行。
- 上传、导入、Workflow 运行、Knowledge 变更、Dashboard/Data Share 写入和 Agent 资源变更，只有用户明确给出目标和动作时才执行；目标不完整时先补齐 workspace、资源 ID、文件或输入。
- 删除、覆盖、发布、取消、重试、成员、IAM、跨工作区和凭证相关操作必须锁定具体对象、影响范围和确认语义；不能自动恢复或替用户选择对象。
- 命令存在不代表环境、权限、后端输入和业务终态都可用；遇到缺口时保留 `partial` 或 `cli_missing`，不要猜参数或切换低层接口。

## Workspace 与 ID

列出账号可见 workspace：

```text
"${MOI_DEV_CLI_BIN:-$HOME/.local/bin/moi-cli-dev}" -o -no-interactive workspace list
```

workspace-scoped 命令显式携带 `-workspace-id ID`。未明确指定时可复用当前会话已确认的
默认 workspace；若 CLI 报告缺少 workspace，先列出可见列表并请求选择。不要选择第一项，
不要创建 workaround workspace。

全局选项放在 domain 前：

```text
"${MOI_DEV_CLI_BIN:-$HOME/.local/bin/moi-cli-dev}" -o -no-interactive -workspace-id ID DOMAIN COMMAND ...
```

## 相关流程入口

复杂任务先读 [业务目标编排](../contracts/task-composition.md)，明确结果、输入输出引用和恢复点。
生产环境涉及模型调用时，先读[模型选择与推理等级](../domains/model.md)：优先使用
`deepseek-v4.1-flash`，默认不传推理等级参数，并核对目标功能对应的模型通道；用户明确指定时遵循其选择。

先读与用户目标最接近的领域文档，再读需要跨领域组合的 Recipe：

- [完整 domain index](../index.md)
- [覆盖矩阵](../coverage-matrix.md)
- [Catalog 与 SQL](../domains/catalog.md)
- [看板运行、失败恢复与结果核验](../domains/dashboard.md)
- [能力模型](../capability-model.md)
- [Canonical 能力目录](../capability-catalog.md)
- [有限能力目录](../capability-inventory.md)
- [能力边界](../contracts/capability-boundary.md)
- [Connector 与导入导出](../domains/connector.md)
- [Workflow 与作业](../domains/workflow.md)
- [SQL 查询闭环](../recipes/run-sql-and-fetch-result.md)
- [SQL 工作簿与版本](../recipes/manage-sql-workbook.md)
- [文件导入到表](../recipes/import-file-to-table.md)
- [Workflow 运行与恢复](../recipes/create-and-run-workflow.md)
- [解析节点参数与结果核验](../recipes/verify-parser-workitems.md)
- [作业失败、版本与定向恢复](../recipes/inspect-and-recover-workflow-job.md)
- [知识库准备](../recipes/create-and-prepare-knowledge-base.md)
- [知识库问答与引用](../recipes/query-knowledge-and-check-citations.md)
- [知识来源复用与分段维护](../recipes/maintain-knowledge-sources.md)
- [数据下载与内容核对](../recipes/export-and-verify.md)
- [CLI 结果的本地筛选、排序和分页](../recipes/query-cli-results-locally.md)
- [数据、知识与报告组合](../recipes/compose-data-knowledge-report.md)
- [Catalog → Workflow → Knowledge 端到端血缘](../recipes/catalog-workflow-knowledge-pipeline.md)
- [工作流运行时检索知识库](../recipes/run-runtime-knowledge-workflow.md)
- [智能体工具与文件交付](../recipes/agent-tools-and-deliver-files.md)
- [Agent 包构建、版本与实际运行](../recipes/build-load-and-use-agent-package.md)
- [托管 Skill 脚本与冻结版本](../recipes/use-script-skill-attachments.md)
- [定时/API 自动化](../recipes/configure-agent-automation.md)
- [计算资源生命周期](../recipes/manage-compute-lifecycle.md)
- [实测范围与保留缺口](../verified-capabilities.md)
- [Agent Team / Ingress 扩展](../domains/agent-extended.md)
- [Catalog Runtime / Workspace File](../domains/catalog-runtime-files.md)
- [沙箱 file](../domains/file-runtime.md)
- [操作日志](../domains/operation-logs.md)
- [Agent 自动化](../domains/agent-automation.md)
- [成员管理 Recipe](../recipes/manage-workspace-members.md)
- [IAM Recipe](../recipes/review-and-change-iam.md)
- [Agent Team / Ingress Recipe](../recipes/configure-agent-team-ingress.md)
- [插件发现、安装与实际调用](../recipes/manage-plugin-capabilities.md)
- [复杂任务模板与安全恢复](../recipes/use-task-blueprints.md)

CLI 自带的命令 reference 是命令可用性的真源，通过已安装 CLI 离线读取：

```text
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

面向用户的进度、验收记录和最终报告遵循[报告表达](../contracts/reporting.md)：先说实际做成了什么、哪里没通过、接下来做什么；技术 ID 和命令放到证据附录。

完成必须以 CLI 返回、后续详情/结果查询、文件校验或目标资源查询为证据。提交成功、按钮
存在、列表出现或模型自述完成都不足以证明业务终态。

报告实际命令产生的对象、状态、ID、证据、未完成步骤和限制。异步任务在有限预算内仍
pending 时报告 pending，不伪装成成功。保留服务端 `code`、`reason`、`msg` 或 `message`，
不自行改写根因。

每次版本迭代还必须按[迭代报告规范](../contracts/iteration-reporting.md)同步交付横向、
纵向、原版对比和总体进展报告，以及 `report-manifest.json`。三份主报告、24 个产品入口、
真实场景、失败与限制缺一都不能把本次迭代称为完整验收。
当前 5.2.3 的[总体进展报告](../iterations/5.2.3/总体进展报告.md)、
[横向纵向功能报告](../iterations/5.2.3/横向纵向功能报告.md)和
[原版对比报告](../iterations/5.2.3/原版与dev版功能对比报告.md)已随包提供；
[失败原因分类](../iterations/5.2.3/失败原因分类.md)逐项说明现有失败、部分通过、
缺条件和 CLI 缺口，恢复方法不会改写原失败路径；
125 条生产场景结果也可直接在包内复查。
功能清单 5.0 的补测另见[本轮 6 项结果](../iterations/5.2.3/6项待补测结果.md)
和[最新 141 项 CLI 工作统计](../iterations/5.2.3/CLI适合项-最新统计报告.md)；
308 项完整清单状态、141 项当前工作范围与原 dev 场景库分别统计，不能混用分母。
本版的[复杂任务模板与恢复合同](../iterations/5.2.3/复杂任务模板与恢复合同.md)
记录任务模板、防重和文件参数的实际测试，不把工具改进冒充新增产品功能通过。
