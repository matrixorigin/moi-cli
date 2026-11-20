# Recipe：创建并准备 Knowledge Base

1. 确认 workspace、唯一知识库名称和用户选择的本地文档。名称同时是不可变 Catalog 数据库名，
   遵守小写字母、数字和下划线等当前命名合同；不要用换名掩盖结果未知的创建。
2. `semantic-model upload-local-file`，保存 `file_id`。若返回 `original_name` 则使用它；未返回时沿用本次明确上传的文件名，不猜远端路径。当前该上传入口没有 `model-id` 参数，按 help 做工作区上传，不凭记忆添加。
3. 用 `schema semantic-model create-with-sources` 构造并校验 source JSON。需要创建时声明
   embedding/drawing 配置时，改用 `create-with-spec`，不要混用两种 JSON 形状。
4. 执行选定的创建命令一次，保存真实 model/source/job ID；创建返回或解码结果未知时先发现原对象，不重建。
5. 用 `list-source-jobs` 按 `source_file_id` 匹配本次上传的 job；用 `list-sources` 匹配 source。
6. 对本次已获授权准备的知识库，确认源与上传一致后，按下方状态机决定是否 reconcile。
7. 用只读 `list-source-jobs`、`list-sources`、`get-source-document` 和关联 Workflow 执行检查进度。
   在有限预算内等待，pending 报告 pending，失败保留错误和 ID。
8. 要求本次 source 的列表和持久 document 状态均为 `ingest_status=succeeded`，并完整读取分段和 ETag。
   作业候选列表 `{}` 或空列表、工作流 completed、分段可读均不能单独代替全部准备条件。
9. 用户要求检索与引用时，继续[知识库问答与引用核验](query-knowledge-and-check-citations.md)。准备通过不能代替检索和回答通过。

创建成功或文件接受不能证明已经可检索。

## `create-with-spec` 的创建配置

先检查当前 schema/help。普通标准库的最小 spec 只含实际名称：

```json
{"name":"唯一小写知识库名"}
```

drawing 库按用户选择显式声明，例如关闭结构化表并保留版本：

```json
{"name":"唯一小写知识库名","drawing_structured_enabled":false,"drawing_number_policy":"keep_versions"}
```

显式 false 不是省略：上述例子仍是 drawing 库。`overwrite_current` 与 `keep_versions` 的选择会影响
同图号检索范围，不能为了通过测试擅自换策略。创建后 `get` 必须回读类型、开关和策略；需要结构化表时
还要核对真实表绑定，不能只看 HTTP 成功。

```text
moi-cli schema semantic-model create-with-spec --validate-file "$spec_file"
moi-cli -o -no-interactive -workspace-id "$workspace_id" semantic-model create-with-spec --json-file "$spec_file" --files-file "$files_file"
```

需要显式文件配置时才追加 `--files-file "$files_file"`；`--tables-file` 是另一输入。无论是否传 files，
都要 `get` 回读 `files.embedding_model`、维度和向量表。最小 spec 可能只创建元数据，不能由 model ID
推断 embedding 已初始化。`preserve_embedding:true` 仅用于已有完整、兼容文本和图片绑定的场景，并回读
模型、维度、向量表和投影开关；不要从问答模型名猜 embedding 配置。该入口不接收 source 数组，
因此空 model 创建成功不等于任何文档已上传或准备。有文档时按既定 `create-with-sources` 路径创建，
或在明确授权的后续步骤向原 model 追加。

2026-11-14 QA 实测：最小 standard spec 和 `drawing_structured_enabled:false` + `keep_versions`
均返回 model；后者回读 kind=drawing、policy=keep_versions、tables=[]，false 开关在 JSON 中省略。
两个 model 的 get/export 都没有 files。空 standard 库随后用正式 `update --files-file` 提交
`{"embedding_model":"bge-m3","file_ids":[]}` 被 400 `Knowledge base embedding configuration cannot be changed`
拒绝，回读仍为空。另一个带文本/图片 embedding 和投影开关的 drawing spec 返回 500且列表无对象。
这些结果要求把“类型/策略创建成功”“embedding 已初始化”“带来源准备成功”分开报告；不得把后续
create-empty 初始化的 `bge-m3` 和 append 成功补记为上述 spec 已通过。

需要创建时初始化 embedding 时，必须在创建命令中同时提供完整 `--files-file`，不能指望事后 `update`
补齐。2026-11-15 生产实测的标准库配置为：

```json
{"file_ids":[],"embedding_model":"bge-m3","embedding_dimension":1024}
```

drawing 本次同时初始化文本和图片的配置为：

```json
{
  "file_ids": [],
  "embedding_model": "bge-m3",
  "embedding_dimension": 1024,
  "image_embedding_model": "qwen3-vl-embedding",
  "image_embedding_dimension": 2560,
  "image_embedding_backend_id": -900002,
  "image_distance_metric": "cosine"
}
```

其中 backend 是当前生产已验证值，其它环境应先发现可用绑定，不能照抄。向量表名由服务端生成，
再用 get/export 回读。标准 model `110007` 已创建即生成 `kb045std.kb_110007_text_index`；
drawing model `110008` 同时生成文本和图片向量表，后续 append/reconcile 的来源 succeeded，4 个分段和 ETag
完整。旧失败与缺少完整维度等配置有关，但这些成功也不证明任意模型、维度、私有 backend、
`create-with-sources`、问答、正式引用或图片检索已经通过。

最小 spec 本身也已是可独立验收的元数据合同：2026-11-15 的标准 model `130007` 和 drawing model
`110009` 类型、策略均正确，get/export 均无 files。应把“最小 spec 元数据创建通过”和“未初始化 embedding”
同时写清，不再用同一个 partial 混淆两种合同。

`create-with-sources` 也已用同一完整配置单独验证，不能再只引用 `create-with-spec` 的成功替代它。
model `130009` 通过 `create-with-sources --files-file ... --image-index-enabled` 创建，随后追加、准备和
一次终态 reconcile 完成，来源 succeeded、4 个分段和 ETag 可读，待协调 job 清空。返回类型为 standard
并带图片索引，不要改写成另一种 drawing kind。旧 500 与缺少完整维度等配置有关，不是该命令不可用；
本次仍只覆盖当前生产配置和本地合成文档。

## reconcile 的两阶段推进与终态

首次启动和成功后的终态落账可以是两次不同的受控推进。每次都先冻结当前 sources、jobs、document
及关联 execution，再决定是否产生一个新写意图：

- `reconcile_required=true` 且目标 job 尚未启动：在原 model 上执行普通 reconcile，省略 retry 参数。
- job running：只读等待，不把仍出现的 `reconcile_required` 当成再次提交理由。
- job 已完成而 source/document 仍未持久成功：核对原 job/execution 和真实未就绪 source 后，可另冻一次
  终态落账 reconcile；提交后只读核验，不盲循环。
- 原成功 job 退出活动列表：只有保留了原完整成功回执，且当前空列表明确无待协调项、持久 document
  succeeded、分段完整时，才按已准备接续。`{}`、部分列表、陌生 job 或持续 reconcile_required 均不放行。
- failed source：先保存 `job_id`、`source_row_id`、错误和失败阶段。`--retry-source-row-id` 是另一个
  需要明确授权的重试动作，不属于普通 reconcile。

终态投影也必须独立核对。2026-11-15 出现来源列表 `succeeded`、持久文档 `pending`，且 job 虽 succeeded
仍 `reconcile_required=true` 的状态；仅做一次终态 reconcile 后，文档变为 succeeded、6 个分段和 ETag
完整，待协调 job 列表清空。该结果只证明终态投影恢复，不允许在 job running 时反复 reconcile，
也不能推广到 failed source。

任何 reconcile 超时、断连或未知都不得重发；先回原 model/source/job 查询。准备完成的严格判据仍是
当前 model/source/file 身份一致、持久来源 succeeded、完整文档分段和 ETag 可回读。后续检索与问答另验。

## 图片索引准备

使用用户授权的真实含图文档；隔离验收可生成虚构图文样本，预期答案保存在本地，问题中不带答案。
读取当前 `create-with-sources --help` 与 schema，支持时在本次创建命令增加 `--image-index-enabled`。
沿用上面的任务推进方式，不因等待图片准备而重复上传或建库。

完成时除源 `ingest_status=succeeded`、分段可读外，还检查 `get` 中本次文件的
`image_index_file_statuses.status=ready`、`indexed_images` 和整体 `image_index_status`；
保留实际图片索引模型、表及维度。`workflow workflow-templates` 可确认
`standard_rag_with_image_index` 模板存在，但模板存在不是本次执行证据。

通过原始 file ID 查询真实准备血缘：

```text
moi-cli -o -no-interactive -workspace-id "$workspace_id" catalog file-data-asset --file-id "$file_id"
moi-cli -o -no-interactive -workspace-id "$workspace_id" workflow catalog-file-lineage-overview --file-id "$file_id"
```

在血缘中核对相同源文件、`workflow_invocation` 的 model/image_index 输入、真实 producer
case 和执行终态；图片解析、图片索引节点也需完成。`case_id` 与 `execution_id` 不是可随意
互换的标识；不能把列表 `{}` 或以 case ID 查询执行返回 404 当成工作流未执行。
没有可核验的执行终态时保留准备验收缺口。随后继续图片问答和独立核图，不能仅凭 ready 收工。
