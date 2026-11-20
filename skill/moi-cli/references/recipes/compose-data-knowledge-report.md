# Recipe：数据、知识规则、Workflow 与报告交付

适用：用户要求导入数据、查询知识库规则、做统计判断并交付文件。先读[目标编排](../contracts/task-composition.md)和[Catalog → Workflow → Knowledge 端到端血缘](catalog-workflow-knowledge-pipeline.md)。结果应包括统计、判断、来源和可下载文件，不能只返回流程成功。

## 先固定输入和恢复点

记录 workspace、文件指纹、导入策略、目标表引用、知识库/source/file、统计口径和交付 Volume。阶段账本保存已确认 ID 与状态；后续失败只继续未完成阶段，不重新上传或导入已成功文件。

1. 按[文件入表](import-file-to-table.md)完成基础和增量数据，独立 SQL 核对内容。
2. 按[问答引用](query-knowledge-and-check-citations.md)读取用户指定规则。问题不提供预期答案；文档规则核对正式 `knowledge.answer` 的 source_refs 与实际分段正文；独立语义规则核对真实检索产物semantic_entries的id/model/domain/kind/key/rule，按[规则维护](maintain-knowledge-sources.md#业务规则的真实问答消费)保存条目来源。保存结构化规则值及其证据，不能把语义条目伪装成文档chunk引用。
3. 若实际使用 Volume 输入，按跨域 Recipe 先做该 Volume 非空预检；然后部署隔离多节点 Workflow；明确单次 execution 与 Volume 批处理的 ID 类型，执行数据读取、聚合、计算、文件写入和结果返回。
4. 按真实 execution/file execution 查询到终态；下载本次报告 file_id，逐项比对原始数据、SQL、规则与文件正文；报告分别保留文档规则的真实file/source/chunk引用、语义条目的真实身份及检索证据；不编造规则条目的文件式正式引用。只有用户要求报告再入库时，才按报告 file_id 新增并核验 Knowledge source/job，不能把规则来源与报告产物混为一条引用。需要看板时再组合图表结果核对。

这是 Codex 通过 CLI 编排知识检索和 Workflow 两个阶段。将已经检索的规则值、引用编入 DSL，不代表 Workflow 内部自动调用了 RAG。若用户要求运行时每次重新检索，必须另有真实检索节点或 Agent 调用，并核验该执行内的工具/引用证据。

## 已验证的多节点输入合同

- `moi:data.table.read` v3：传真实 `table_ref`（kind、catalog_id、database_id、table_id 及回读名称），save `.table_ref`。
- `moi:data.sql.process` v1：`table_ref: "{{ .state.table_ref }}"`，SQL 用 `${source_table}` 指向该表；save `.rows`。先用本机 workitem schema 确认版本，不能搬旧节点字段。
- jq 节点用 `$state` 计算，例如从区域 rows 的 order_count、total_amount 求和；不要把结果当模板文本再次解析。
- `moi:catalog.sink.write` v3 使用 `target_ref: {"volume_id": 实际ID}`。写 JSON 用 `json`+`format: "json"`；写 Markdown 用 `text: "{{ .state.report }}"`+`format: "markdown"`。两种载荷互斥，JSON 不会自动渲染成 Markdown。
- sink save `.file_id`；最终结果同时返回统计、判断、原 source_refs 和 file_id，避免只回一句“已生成”。

`"{{ .state.object }}"` 可保留结构化对象，不能提前把它 stringify 成不可用文本。复杂 jq 先用本地样例计算，再部署；不要手工修改编译后 Mowl 转义去替代 DSL authoring。

## 版本和失败恢复

部署后的 workflow/version/execution 全部来自返回。现有 CLI 的版本发布限制见[Workflow](../domains/workflow.md)。不能用同名 deploy 冒充更新同一 App，也不能将新建流程的成功用来掩盖旧流程版本失败。

`retry-run` 是整次新跑，不是失败节点续跑；有文件写入、导入或外发节点时，先明确幂等和补偿，见[作业恢复](inspect-and-recover-workflow-job.md)。

## Workflow 内部知识检索的发现门槛

先检查当前 workspace 的完整 `workitem-catalog` 及相关节点 input schema。普通 LLM 生成节点
只收 prompt/text/model 时，不等于拥有知识库检索工具；不能编造 knowledge_base_id 或 tools 参数。
2026-10-31 样例可见的 51 个节点里未发现专门的知识检索节点，`moi:llm.output.generate` v3
也未暴露 KB/tool 输入。这只证明当前发现入口未提供直接路径，不证明后端永远做不到。
通用 HTTP 节点也不提供隐含身份或授权；不把个人 PAT 写进 DSL 或用它绕过 Product API。
可以继续使用已验证的 Codex 检索→Workflow 计算交付编排，并把“单次 Workflow 内动态检索”
单列为待接入，后续需要正式节点或可传递身份的 Agent 调用合同。
