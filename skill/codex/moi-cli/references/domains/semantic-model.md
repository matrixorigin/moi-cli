# Knowledge / Semantic Model 领域

## 创建入口与配置归属

按用户目标选择创建入口，不用另一条命令绕过业务校验：

| 目标 | 入口 | 输入归属 |
| --- | --- | --- |
| 空知识库 | `create-empty` | `name`、可选 `description` 和 `image-index-enabled` |
| 从已知文件/表创建 | `create-with-sources` | source 数组交给 `--json-file`；`--files-file`、`--source-selections-file` 是不同输入 |
| 创建时声明 embedding/drawing 配置 | `create-with-spec` | snake_case spec 交给 `--json-file`；文件/表配置分别交给 `--files-file` / `--tables-file` |

先运行 `schema semantic-model create-with-spec` 和 `semantic-model create-with-spec --help`。
spec 使用 `name`、`preserve_embedding`、`text_send_dimensions`、`image_send_dimensions`、
`drawing_structured_enabled`、`drawing_number_policy`；不要把这些字段塞进 PascalCase source 数组。
`create-with-spec` 返回 model 本身，保存其真实 `id` 后继续 `get`、`list-sources` 和
`list-source-jobs`。创建命令存在、schema 可校验或 model 已返回都不证明文件已经准备完成。

drawing 字段只属于创建，不能用 `semantic-model update` 改写。两个 drawing 字段都省略时是标准库；
显式 `drawing_structured_enabled:false` 仍是 drawing 库，只是关闭结构化表；只给
`drawing_number_policy` 也表示 drawing 意图。当前策略为 `overwrite_current` 和 `keep_versions`，
省略策略时由 owner 使用默认策略，客户端不自行补写未知默认。创建后分别回读
`semantic_model_kind`、`drawing_structured_enabled`、`drawing_number_policy` 和真实表绑定。

`preserve_embedding:true` 表示保留调用方提供的权威 embedding 绑定，不是“随便指定模型并跳过校验”。
此时 `--files-file` 必须带完整、兼容的文本和图片绑定；模型、维度、向量表、backend、预处理、
距离度量和投影意图由 Knowledge owner 校验。普通创建不要伪造这些值；由服务解析当前工作区
authoritative slot/binding。但不传 `--files-file` 的最小 spec 也可能只创建 model 元数据；必须回读
`files.embedding_model`、维度和向量表，不能假定服务一定已初始化文件生命周期。
`text_send_dimensions` / `image_send_dimensions` 是投影请求开关，
不是维度数值，也不等于 `create-with-sources --image-index-enabled`。后者控制图片索引，不能替代
drawing 类型、图片 embedding 绑定或真实图片准备/问答验收。

若空 spec 的 `get/export` 都没有 files 配置，先停止 append，分别记录 source/job 为空和 embedding 未初始化。
不要把旧 SDK 的“update 补 files 初始化”当成当前服务必然支持的恢复：2026-11-14 QA 的 dev 实测中，
最小 standard/drawing spec 均无 files；对空 standard 库提交已确认的 `bge-m3`、空 file_ids 仍返回
400 `Knowledge base embedding configuration cannot be changed`，回读未变。该请求没有修复原库，也不证明
其它已发布初始化路径均不可用。文件知识库准备应使用已初始化的原生 create-empty/create-with-sources
生命周期，或先在创建时提交经验证的 files 配置；创建失败和配置缺失不得靠重复建库掩盖。

## 新建知识库

```sh
moi-cli -o -no-interactive -workspace-id "$workspace_id" semantic-model upload-local-file --file-path "$file_path"
moi-cli schema semantic-model create-with-sources
moi-cli schema semantic-model create-with-sources --validate-file "$sources_file"
moi-cli -o -no-interactive -workspace-id "$workspace_id" semantic-model create-with-sources --name "$name" --json-file "$sources_file"
```

上传返回 `file_id`，可能不返回 `original_name`；缺失时沿用本次明确上传的文件名。source JSON 使用 CLI 的 `SourceType: local_file`、
`FileID`、`FileName` 字段，不把本地路径当作 FileID。

## 准备终态

保存返回的 `model.id`、source ID 和 job ID，再查询：

```sh
moi-cli -o -no-interactive -workspace-id "$workspace_id" semantic-model list-source-jobs --model-id "$model_id"
moi-cli -o -no-interactive -workspace-id "$workspace_id" semantic-model list-sources --model-id "$model_id"
```

accepted 或 job 创建不等于解析、索引和可检索。已有知识库使用 `upload-model-local-file`、
`append-sources`，不能创建新库替代原库。entry/segment 变更和 source 删除是独立写入。

## 来源任务的受控推进

`list-source-jobs` 是当前准备候选视图，不是完整历史；`list-sources` 和
`get-source-document` 也是不同投影。每次先保存同 model/source/file/job 身份及原始响应，再判断：

1. source/job 已接受但尚未启动，且返回 `reconcile_required=true`：对已授权准备的原 model 执行一次
   `reconcile-source-jobs --model-id`，保存回执，再只读检查同一批对象。
2. 原 job 已运行或完成，但 source/document 尚未落终态：先读完整 sources、jobs 和 document。
   只有真实未就绪来源及原任务终态表明需要终态落账时，才另冻一次有限 reconcile 意图；
   不把“一生只能 reconcile 一次”当产品规则，也不在每轮轮询中盲写。
3. 已成功 source 的原 job 后续退出活动列表：保留原成功 job 回执，接受明确的空活动列表，
   以 document 中持久 `source.ingest_status=succeeded`、完整分段和 ETag 接续；空列表本身不代表成功。
4. `reconcile_required` 持续为 true、出现陌生 job、状态投影冲突或部分列表：保留差异并停止升级结论，
   不循环提交。前次 reconcile 超时、断连或结果未知时只回读原对象，不重发。

`--retry-source-row-id` 只用于已经检查失败、且用户明确授权重试的 source row；普通推进和轮询必须省略。
不得用重新上传、重复 append、replacement workflow 或本地检索实现代替原生 reconcile。

## Workflow/Catalog 血缘闭环

当知识源来自 Workflow 输出或 Catalog Volume 时，必须先保存上游 `workflow_id`、版本、真实
execution/file-execution（或服务端明确的 batch/file 任务）和产物 `file_id`，再核对
`catalog file-data-asset` 或 `workflow catalog-file-lineage-overview`。随后在知识库侧按同一
`file_id` 匹配 `source_id`、`source_job_id`、`ingest_status` 和索引状态。不能只凭文件名、
source 数量、旧知识库记录或 `vector_reuse` 建立血缘。

需要把以下终态分开记录：

1. Workflow/文件解析产出成功；
2. Knowledge source ingest 成功；
3. embedding 或 image index ready；
4. 检索命中本次 source；
5. 问答返回正式引用，且引用的 model/source/file/chunk 与本次资料一致。

任一层缺少本次 ID 关联时，最多报告到已证实的上一层。历史 execution、旧 source 或已有
索引不能证明当前新提交处理了文件。

列表返回数量与 total 分开记录。需要完整分页观察时使用
`list-sources-page --model-id ... --page ... --page-size ...`，但不能假定它与 `list-sources`
统计同一集合：本轮一个库的正式来源为2个，分页返回3项，第三项是
`governance_status=legacy_unbound`、`legacy_origin=lineage_register` 的历史候选，没有正式 source_id。
保存 `legacy_backfill_required` 和治理状态，分别报告已管理来源与未绑定候选，不能把候选视为可检索资料。
`check-source-existence --file-i-ds` 可核对指定 file ID 是否已是正式来源；本次只返回新版，未返回旧候选。
按页合并时保留身份与分类；完整性同时检查 total、页大小和末页，不能靠缺少 items 单独推断总数为零。
不要因发现历史候选就自动 backfill/reconcile，否则可能把用户已移除的旧资料再次纳入。
当前生产验收中 `list-source-jobs` 对已完成图片知识库返回过 `{}`，而同一 `model_id` 的
`list-sources` 已返回 `ingest_status=succeeded`，Catalog file lineage 也能返回本次 producer
case 和完整节点终态。此时 `{}` 只能标记“source-job 观察入口未给出记录”，不能声称没有准备任务，
也不能用它替代 source、lineage 或检索证据；若任务状态本身是本次必需的硬验收项，应保留该观测缺口。
`job_status=failed` 是失败证据；`reconcile_required=true` 不授权自动运行 reconcile 或 reembed。
用户已要求准备明确的本次知识库时，按[准备流程](../recipes/create-and-prepare-knowledge-base.md)
判断是否需要推进任务；检索和回答按[引用核验](../recipes/query-knowledge-and-check-citations.md)另验。

## 任务内检索与正式引用

`find-files`、`search`、`read-source`、`list-passages`、`read-document`、`search-images`、
`read-evidence` 和 `cite` 是带 task authorization proof 的运行时命令。native Astra 登录或 PAT
本身不授予这类任务范围；在普通外部终端不要把它当成独立检索 API，也不要复制任务证明。
在已准备会话中沿[检索与引用 Recipe](../recipes/query-knowledge-and-check-citations.md#任务内检索证据与-cite)
保存检索结果、完整读取证据并生成 `cite` 回执。`cite` 回执与 Agent 最终
`knowledge.answer.metadata.source_refs` 是两层证据，二者不能互相替代。

## 图片知识库

需要检索图中内容时，按[图片准备](../recipes/create-and-prepare-knowledge-base.md#图片索引准备)
确认图片索引和真实准备工作流完成，再按[看图与引用](../recipes/query-knowledge-and-check-citations.md#图片读取与正式引用)
核对检索命中、实际图片分析、正式引用及独立下载的页图。图片索引 ready、图纸相似检索命中、
或答案出现图片 ID，均不能单独证明模型读懂了图片。

## Catalog 原文件与 KB 自有 Volume

`catalog_file` source 保留原文件身份；创建 KB 不会自动把文件复制到其自有 raw Volume。
用户已授权准备的显式 sources 可通过 source-job/reconcile 路径处理原 Volume 文件，
不需要先把文件加入 KB raw。默认 Workflow 的 `source_ref` 可能指向空的 KB raw Volume，
直接 `workflow run` 是另一条输入路径，必须预检实际默认 Volume。
标准 RAG 的正式引用仍对应原始 file/source/chunk；解析 JSONL 和输出 ZIP 是派生产物，
不能为了串流程再次作为来源添加，否则会改变资料范围并可能产生重复来源。

## 业务规则条目与迁移

条目不等于文档分段；`semantic-model export/import` 的条目往返不代表原文件、embedding、
分段、图片或完整知识库备份已经迁移。操作方法及已测范围见
[知识来源维护中的条目往返](../recipes/maintain-knowledge-sources.md#语义条目导出再导入)。

## 创建失败与响应兼容

2026-11-14 QA 的 dev `create-with-sources` 对本次唯一名称和 local_file 返回 500，完整 model 列表
没有该名称；另一个带文本/图片 embedding 配置的 drawing spec 也返回 500且未出现在列表。两次均保留
request ID并停止原写意图，没有重发、换名制造成功或把后续 create-empty+append 的通过记到它们名下。
这组结果只判定本次输入未通过，不推断所有来源创建或 embedding 组合均失败。

2026-11-02常规0.1.8的`create-empty`、`list`、`get`在本次环境遇到`unknown field "semantic_model_kind"`解码错误。创建命令虽exit1，实际已创建知识库；不能再次create或换名掩盖原结果，也不要归因登录。

已验的正式发现路径是`agent resource-get --resource agent-builder --tail-path /resources`：用本次明确记录的唯一名称与workspace确认候选真实ID，再由`semantic-model export --model-id`返回的model.id/name独立核实归属。只在身份关联足够可靠时继续`list-entries`、export/import等实际可用的管理入口；typed模型详情仍失败，保留兼容缺口。既不猜资源ID，也不从开放tail-path猜其它路由或绕SDK/raw HTTP。若正式列表覆盖不了原对象，保持未知，不以未找到断言未创建。

该发现方法两次找回原创建对象并完成精确清理，但没有修复创建/列表/详情的解码故障。它也不保证其它业务、普通成员权限或所有分页场景可用。
