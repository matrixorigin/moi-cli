# Recipe：工作流运行时检索知识库

适用：每次运行都要读取当前知识、计算并给出正式引用。已验证的正式路径是“绑定知识库的 Agent → API 自动化生成的 Workflow → Astra 运行时 → 知识工具”。它不等于任意手写 DSL 都能直接调用 RAG。先按[自动化 Recipe](configure-agent-automation.md)确认权限、PAT、指令和收尾范围。

## 准备真正可运行的 Agent

用 `agent resource-get --resource agent-builder --tail-path /resources` 发现知识库资源，先核对当前版本的正式路由；如果该路径不可用，不猜资源 ID。Catalog file、Knowledge model 与 Agent 绑定分别记账，不按同名推测 ID。

Agent 的正式 Knowledge binding 使用 `binding.knowledge_base_refs`：

```json
{"binding":{"knowledge_base_refs":[{"id":"真实知识库资源ID","workspace_id":"真实workspace","kind":"knowledge_base"}]}}
```

创建/变更限已授权的隔离 Agent 或用户指定对象。普通 `resource-post agents` 保存成功不代表 Astra 可运行；Astra 还需要已发布、runnable 的不可变版本和 active runtime binding。不能只改 provider 就 invoke。

若需要创建包，使用官方 `agent init → verify --source → build → verify --file → load-agent-package`，详细包操作由可用的 `agent-package` Skill 负责。最小 `agent.yaml` 使用 init 生成的结构，修改 `moiResourceRefs.knowledgeBases`：

```yaml
moiResourceRefs:
  tools: []
  skills: []
  knowledgeBases:
    - name: 目标工作区中已确认的知识库精确名称
  catalogFiles: []
  channelBindings: []
```

这里的 Knowledge ref **不支持 `workspaceID`**，按加载目标 workspace 的精确名称解析；Agent 资源 binding 则使用真实 ID，不混用两种格式。包中不写 endpoint、凭据、runtime binding 或签名。跨工作区名称解析未经本流程验收。

2026-10-31 实测：先将本次隔离 Agent 的 `runtime` 正式 PATCH 为 `{"provider":"astra","profile":"default"}`，再加载其包，正式加载流程创建 runnable 版本并注册运行绑定。该步骤会采用产品的 Momo foundation；不能未经核对就迁移已有业务 Agent。回读：

```sh
moi-cli -o -no-interactive -workspace-id "$workspace_id" agent resource-get --resource agents --tail-path "/$agent_id/versions"
moi-cli -o -no-interactive -workspace-id "$workspace_id" agent resource-get --resource agents --tail-path "/$agent_id"
```

核对 `lineage.default_version`、该版本 runnable、知识绑定、provider/profile 和实际加载诊断。包加载后默认版本已建立时不要重复发布；未就绪则停在诊断阶段。Knowledge 工具由正式运行时根据绑定注入，不需要伪造工具或复制内部 `agent_task.prepare_run` / `agent_task.invoke` 节点到普通 Workflow。

## 同一自动化的两轮运行

按自动化 Recipe 创建 disabled API 任务，指定真实 Agent/owner、DeepSeek、不传 effort、固定检索指令。启用后从返回取得真实 ServiceName，`-auth pat` 调用 `workflow invoke-dynamic-service`。本流程用空 `Payload`，不掩盖另已记录的动态 payload/schema 缺陷。

指令只写任务、范围和引用要求，不预填库存、阈值、计算结果或旧标记。仅在用户已授权具体知识库、旧来源及新版文件的替代验收时，才做下面两轮对照；普通查询或自动化创建授权不包含上传和删除来源。只需读取当前知识时完成一轮即可。

第一次运行完成后，按[来源维护](maintain-knowledge-sources.md#新版原文件替代旧知识来源)完成原文件来源替代。随后用**同一 automation、同一指令、同一包版本、同一空 Payload**单次再触发，保存新的 run。两轮应独立检索，并使用各轮当时有效来源。

## 核验实际检索与引用

`agent-automation-runs/<run>/result` 的 succeeded/final_text 还不够：

1. 用 `agent resource-get --resource agent-runtime-tasks` 发现该 Agent 的真实 task；不能按名称或最新顺序认定关联。读取候选 task 的 manifest，按 `body.automation_run_id` 精确等于本轮 run ID 确认。
2. 用 `agent resource-get --resource agent-runtime-manifests --tail-path /真实manifest_id` 和 `agent resource-get --resource agent-runtime-tasks --tail-path /真实task_id/events` 保存 manifest/events。列表窗口无法覆盖时保留限制，不在 tail-path 拼查询参数绕过校验。
3. manifest 要显示正确的 KB 与绑定文件范围，并检查实际工具列表和调用。Momo foundation 可能扩充工具，包的 tools:[] 不保证运行时仅有检索工具；必须确认实际调用未越过本次授权。events 应有 `search_rag_chunks` 的实际调用及结果、`select_final_sources` 的 accepted 结果，以及 `knowledge.answer` artifact 的正式 `metadata.source_refs`。
4. 将引用的 model/source/file/chunk 与本轮文档分段逐项匹配。替代后不得继续引用旧来源，检索证据也不能混入旧文件。只看回答文字中的“已检索”不算证据。
5. 原 task completed、automation run succeeded，答案及计算正确，未知内容不编造。按授权停用测试自动化，并回读运行总数；不能用归档掩盖额外执行。

Astra events 的实际工具字段为 `payload.tool_id/arguments/output`，artifact 在 `payload.artifact`；按实际返回解析，不套用另一 provider 的嵌套结构。

当前自动化 HTTP DTO 未公开 generated Workflow app/execution ID，events 也裁掉内部 Data。可以据正式提交/准备事件与 run→manifest→task→检索→引用证明这条链路，但不能猜出 Workflow ID 后声称已直接读取节点 trace。

## 失败如何判断

2026-10-31 同样的 DeepSeek 在 `default/default` 自动化首轮模型调用零输出失败，错误 `MODEL_EXECUTION_FAILED`，没有工具调用；补齐正式包发布及 Astra 绑定后成功。知识绑定和工具注入原本已成功，因此首个失败不能归为“知识库检索失败”，也不是缺登录。实际更底层模型错误未由当前 CLI 暴露，不能断言唯一根因是某个 service slot。

本次环境可见的 `moi:llm.output.generate` v3 schema 未提供 Knowledge/tools 输入；Python WorkItem 的 execution SDK 也不等于 Product SDK。不要把令牌塞进 DSL、复制内部运行授权或用 raw HTTP 拼一个未经产品支持的 RAG 节点。需要任意 DSL 内直接检索时，仍应记录正式节点/输入入口的缺口。
