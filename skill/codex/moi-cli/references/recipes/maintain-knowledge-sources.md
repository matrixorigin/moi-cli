# Recipe：追加、复用和维护知识来源

## 追加资料和复用已处理文件

上传新文件使用 `upload-model-local-file`，已有库用 `append-sources`，不重建原库。复用 Catalog 原始文件可把以下数组交给 `create-with-sources` 或 `append-sources`：

```json
[{"SourceType":"catalog_file","FileID":"实际原始文件ID","VolumeID":123}]
```

`123` 替换为文件真实 Volume。不能凭文件名选择处理后 Markdown 代替原始文件。表来源数组为 `[{"SourceType":"catalog_table","TableID":实际数字ID}]`，先用真实表结构与内容确认来源。数组使用 CLI Go 字段名，不照抄 HTTP 字段。

保存 model/source/job 返回。按原 source 检查 jobs、sources；用户已明确要求准备本库时，按[准备 Recipe](create-and-prepare-knowledge-base.md)推进需要的 reconcile，不能把空 jobs 当成功。source succeeded、enabled/effective 可用只证明准备状态；复用文件还需分段可读及新会话引用，表来源需真实 schema/只读 SQL 工具结果和 `sql_table` 正式来源，必要时独立 CLI SQL 对照。

服务返回 `vector_reuse:...` 可作为服务报告的向量复用证据；没有独立执行关联时，不据此保证完全没有解析执行。当前 `workflow list-file-executions` 没有 file-id 筛选参数。

跨库复用使用短、合法且唯一的知识库名称，并在创建前发现重名。长名称会拼入内部
`operation_id`，已有超过 128 字符导致复用失败的产品缺陷；短名称是已验证的适用条件，
不是缺陷修复。失败时先保留原 model/source/job 和错误，不要重复 append 或重建多个长名称库。
2026-11-15 使用短名称 `s048r` 复用同一 Catalog 原文件和表来源，两个来源均 succeeded，
文档 3 段及 ETag 可读；独立新会话回答 53 箱、银杉-053、7 辆车、orders 6 笔及总额 1500，
正式引用命中当前 model/source/file/chunk 和真实 SQL 结果。该通过不覆盖长名称缺陷，
也不证明任意文档规模或普通成员权限。

准备完成后的深度核对至少包括：model 详情的 file_ids/vector_table、source 列表、每个 source
的文档分段和 ETag，以及源 job 是否已进入终态。source 列表为空或 job 列表为空不能
单独证明“没有处理”或“已完成”；以 source 的 `ingest_status`、文档可读性和独立问答
引用为准。

## 跨库删源影响

同一原文件在两个库里可有不同source ID，而chunk ID可能相同；检索及正式source_refs同时核对
model/source/file/chunk（返回可以是chunk_id或chunk_ids），不能只用chunk ID判断归属。
若要验证删除某库来源不影响另一库，先保存两库精确source、文档正文/ETag、原文件字节/hash；
在B新会话取得正式引用后，仅删除已授权A中的source，再独立读B、另开新会话问答并核对正式引用。
删除source不是删除原Catalog文件，清理两个隔离库后仍应确认原文件及Volume保持不变。

本次同一已解析原文件进入两个隔离库，删A源后A读取404、B两段及ETag不变，新会话仍正式引用B；
原文件197字节/hash保持，测试库删除404。只证明跨库影响这一分支；不能据此承诺来源替换原子性、
并发一致性、向量物理清除或零新增解析。本次没有明确vector_reuse回执，不自行补上该结论。

## 新版原文件替代旧知识来源

适用：用户明确要求将指定知识库的某份原文件换成新版。此路径使用新 file/source 身份，不是原位覆盖同一个 file ID；同名文件不会自动替换。先保存旧 model/source/file/Volume ID、旧分段与本地原文件哈希，明确只移除目标旧 source。

1. 上传用户选定的新版字节，取得新 file ID，挂载到已确认 Volume；回读 file-info，下载核对新文件字节/哈希。
2. 对**原知识库** `append-sources`，数组只含新文件的 `SourceType:catalog_file`、真实 `FileID/VolumeID`。不要重建库，也不要按相同文件名认定 source 相同。
3. 查询本次 source jobs；有已授权且实际需要推进的 job 时 reconcile，随后等待同 source。`reconcile_required` 可能在 running 时仍为 true，不能每轮轮询都再次推进同一个正在执行的 job；先看 job 状态与已有 operation。
4. 新 source succeeded 后，读取其分段、原 file/case 血缘，核对新版唯一事实。此时旧、新可能同时存在；这是多步替代，不是原子切换，尚未移除旧来源前不要把知识范围报告为仅新版。
5. 在已授权范围内，对**精确旧 source ID**执行 `semantic-model delete-source --model-id "$model_id" --source-id "$old_source_id"`。失败或未知时回读，不重复删除或按名称扩大目标。
6. `list-sources` 应只保留预期新版与其它原来源；旧 source 的 `get-source-document` 返回明确404/ErrNotFound。旧 Catalog 原文件另用 file-info/download 检查仍存在、字节未变。
7. 新建独立问答会话，用未预填答案的问题核对新版事实、联合计算及正式 source_refs；其 model/source/file/chunk 都应指向新版。另问旧版独有事实，问题不包含该事实答案；不能检索到时应说明无法确定，不能把新版事实当旧版回答。

本轮虚构样例：同名库存文件31→49箱，新旧 file/source ID不同；准备新版后移除旧源，来源总数3→2，新会话补给量57→39箱、运输5→4次，正式引用随新版改变。只证明此知识库的来源替代与检索范围更新。`delete-source` 不等于删除 Catalog 原文件或物理抹除所有共享向量；跨库复用影响、同一 file ID 字节覆盖、强制重解析与向量单路刷新仍需分别验收。

若还要证明 Workflow 每次运行读取新知识，继续[运行时检索 Recipe](run-runtime-knowledge-workflow.md)，不能把本地读出的新事实静态写进 DSL 冒充动态刷新。

## 文档分段的 CAS 维护

先读取：

```sh
moi-cli -o -no-interactive -workspace-id "$workspace_id" semantic-model get-source-document --model-id "$model_id" --source-id "$source_id"
```

保存 ETag、稳定 segment ID、正文与启用状态。source ID 不是 file ID；不按 chunk_index 或数组顺序定位。每次变更使用当前 `if_match`，成功后保存返回/回读的新 ETag。

| 操作 | JSON 内容 | 额外参数 |
| --- | --- | --- |
| 新增 | `{"if_match":"当前ETag","Level":"chunk","Content":"授权正文"}` | 无 segment-id |
| 编辑 | `{"if_match":"当前ETag","Content":"授权新正文"}` | 原 segment-id |
| 停用/恢复 | `{"if_match":"当前ETag"}` | 原 segment-id，显式 enabled=false/true |
| 删除 | `{"if_match":"当前ETag"}` | 原 segment-id；用户已授权删除该段 |

```sh
moi-cli -o -no-interactive -workspace-id "$workspace_id" semantic-model create-segment --model-id "$model_id" --source-id "$source_id" --json-file "$request_file"
moi-cli -o -no-interactive -workspace-id "$workspace_id" semantic-model update-segment --model-id "$model_id" --source-id "$source_id" --segment-id "$segment_id" --json-file "$request_file"
moi-cli -o -no-interactive -workspace-id "$workspace_id" semantic-model update-segment-enabled --model-id "$model_id" --source-id "$source_id" --segment-id "$segment_id" --enabled=false --json-file "$request_file"
moi-cli -o -no-interactive -workspace-id "$workspace_id" semantic-model delete-segment --model-id "$model_id" --source-id "$source_id" --segment-id "$segment_id" --json-file "$request_file"
```

恢复用相同启停命令的 `--enabled=true`。更新用 CLI schema 的 `Content`、`OCRText`、`ImageDescription`、`Level`，不自行增加 HTTP 字段。变更可能重算 embedding；表来源不支持这些文档分段接口，不能删除最后一段。

完成证据为变更回执、独立回读及独立新会话问答。新增/编辑/恢复后应回答当前内容并正式引用正确 model/file/source/chunk；停用/删除后新问题不能继续答出被移除的唯一事实或引用该段。问题不带预期答案，避免会话记忆掩盖检索状态。

`ErrConflict` 时读取最新状态，保留冲突，不静默覆盖。响应未知先读原文档，不重复新增。ETag 是状态摘要，恢复相同内容/状态可能返回旧值，不要求全局唯一。当前生成类型的非 optional bool 可能省略 false，按已确认的类型合同解释；仍结合回执与引用核验，不泛化成任意缺字段=false。插入/删除可能改变 chunk_index，不宣称字节级还原。

## 语义条目导出再导入

适用于用户要维护/复用知识规则条目；不能代替文档来源和分段迁移。
先用 `schema semantic-model create-entry` 核对 CLI 输入形状，再确认业务 domain/kind：
`Domain` 只用已确认的 `rag` 或 `nl2sql`，不能填写任意业务项目名。
本轮误填 domain 时服务返回 `semantic entry kind is invalid`，改成 `rag` 后同种
`business_rule` 条目创建/修改/迁移成功；不能凭该错误把业务规则能力判为不可用。

已验证的 CLI 输入示例（替换成用户授权的规则）：

```json
{"Domain":"rag","Kind":"business_rule","Key":"transport_rule","Tables":[],"Spec":{"rule":"每车容量9箱，不足一车仍按一车计。"}}
```

1. `create-entry --model-id --json-file` 保存返回的条目 ID；`update-entry --entry-id` 后用
   `list-entries` 独立核对 domain/kind/key/spec，不用 mutation 的 updated=true 代替内容核验。
2. `export --model-id` 返回 model 与 entries。只取 entries 的 domain/kind/key/tables/spec，
   映射成 CLI 输入 `Domain/Kind/Key/Tables/Spec` 数组，移除服务端 ID、时间和 sql_preview。
   不能把整个 export 对象直接交给 `import`，也不能把原条目 ID 当作目标 ID。
3. 对已授权目标库执行 `import --model-id --json-file` 一次，再分别 `list-entries`、`export`，
   比较条目内容和归属，确认目标有独立 ID。未验重复 import 的幂等/覆盖，超时先回读，不能盲重导。
4. 若用户要求删除原条目，只对原库的精确 entry ID 执行 `delete-entry`；回读原库和目标库，
   确认删除范围。不可把测试清理当作通用任务中的默认删除。

本轮两个空隔离库中，一条 rag/business_rule 规则8→9后往返，目标内容一致、ID不同；
删除原条目不影响目标。四个测试库（含输入纠错时创建的两个）及其 Catalog 数据库均清理后回读404。
该往返证据只证明条目管理/迁移；后续rag/business_rule运行消费限定实测见下文，仍未覆盖其它kind、整库备份或跨工作区迁移。

## 当前 CLI 边界

- 常规0.1.8的 `update-source-governance` 仅暴露 model/source/workspace ID，没有 enabled/expires_at/tags 输入或 typed JSON。不能空调用冒充整份文档启停/过期完成，也不能猜 `--json` 绕过。此分支为 `cli_missing`。
- 常规0.1.8的 `get-source-document` 未暴露 cursor/limit，当前默认窗口20段。检查 segment_status 的 total/next_cursor；有下一页时不宣称读全，未出现在首屏不证明删除。
- 图片产物字段是服务端只读，不能手写 ID。图片维护后的问答按[图片读取与正式引用](query-knowledge-and-check-citations.md#图片读取与正式引用)核验，包含实际图片分析及独立下载核图；开启图片索引、图纸检索命中或文字问答通过都不能代替该验收。

当前 dev 二进制的 `update-source-governance` 已暴露 `--enabled=false/true`、`--expires-at`、
`--clear-expires-at`、`--force-enabled-after-expiry` 和 `--tags`；`--tags '[]'` 是明确清空，
省略则保持原值。先读原 enabled/expiry/force/tags，只传本次要求改变的字段，再回读保存值和
`effective_enabled`。过期后仅 enabled=true 不足以恢复有效状态；只有用户明确要求过期后继续使用才设置
force。治理写入不需要重新上传、append 或重解析，但新会话检索/正式引用仍需另验。

2026-11-14 QA 的 dev 实测已完成同一 source 的加标签+停用、独立会话拒答、恢复+清空标签、
另一独立会话正式引用。停用后 enabled/effective_enabled 在 proto JSON 中省略，恢复后为 true；
分段和 ETag随治理状态变化，恢复后 ETag回到原值。清空标签后的列表/document 将 tags 省略，
经 `jq` 读取为 null，而不是字面 `[]`；可据此确认空集合投影，但不能声称响应包含空数组字段。
上述 QA 子场景未覆盖 expires_at 和 force-enabled-after-expiry；生产完整链如下。

另有 2026-11-14 生产 `governance-*` 直接回执覆盖上述 QA 之外的完整治理链：同一 source
先写入 `codex-prod-governance` 标签，设置有效期后自动失效，强制启用后恢复检索，再清除
过期限制并恢复正常使用；启停、恢复、有效期、强制启用和标签都有保存值及独立问答对照。
这些证据可同时支撑数据视角的来源维护和 Agent 视角的文件信息维护，不能只报告启停而漏记
标签和有效期。各功能仍按实际字段与问答结果验收，不把 source 标签误写成知识库整体标签。

当前 dev 的 `get-source-document` 使用可重复 `--query-param name=value`，没有专用
`--page-size` / `--cursor` flag。按该产品操作支持的 `page_size`、`cursor` 逐页读取，保留原 cursor
字符串，不把数字外观的 cursor 转成数值。合并时要求 total 稳定、segment ID 唯一、cursor 不重复、
并集与完整大页一致；越界空页不单独证明来源为空。help 暴露参数不等于本轮线上治理或分页已通过，
稳定版与 dev 的结果也不得互相借用。

## 业务规则的真实问答消费

保存规则写入请求与问答请求时用不同文件名，不覆盖历史请求快照。条目管理成功后，用未预填答案的独立新会话核对实际`search_rag_chunks`产物的`semantic_entries`（id/model_id/domain/kind/key/rule）、最终任务和独立计算，不从保存回执推定运行消费。`spec.files`为空的rag规则为库内全局规则，具体作用域仍需按正式返回核对。

2026-11-01文档只含47箱/琥珀仓，不含容量或口令；独立rag/business_rule规定8箱及唯一口令，新会话检索到条目、答6车及正确口令。同一条目改13箱和新口令，保持文档分段与ETag不变，新会话检索到新版条目、答4车及新口令，均completed。写入请求文件曾被问答请求复用覆盖，因此规则写入正文原始快照缺失；消费结论依据真实写入响应、独立回读和任务工具产物，不声称原请求完好。

语义条目不是文档chunk。不得拼`semantic_entry:<id>`或`file_id=semantic-entry`冒充rag_chunk；正式select_final_sources会拒绝unknown rag chunk source。本轮原任务纠正为真实文件引用后完成。条目消费由真实检索产物证明，文件source_refs只覆盖实际原文事实，不能为规则本身背书。

## 标准问答、文件匹配及删除后的消费

`rag/standard_qa`的Spec使用`question`、`answer`及可选`files`，业务规则使用`rule`及可选`files`；沿当前schema保留外层`Domain/Kind/Key/Tables/Spec`。`Spec.files`保存真实原始file ID，不用文件名、source或chunk ID。创建后回读只是管理验收，运行消费仍须实际semantic_entries和最终答案。

范围对照用同库A/B文件各自条目，Agent真实`search_rag_chunks`使用其正式`file_ids`输入，核对chunks、semantic_entries和正式source_refs均在目标文件范围。这个工具字段不是新增CLI flag；只在提示里说“看A”不能证明已限范围。2026-11-01 A检索命中自己的standard_qa口令及7箱/车规则，41箱至少6车；B仅命中自己的5箱规则，22箱至少5车且拒答A口令。问题与原文未预填条目独有事实，四个原任务最终completed。

精确删除A的两个条目并回读后，保持原文/分段/ETag不变，用独立新会话检索：被删条目不再出现，仍能正式引用41箱及仓库，但拒答旧口令、容量和车数，也不串入B规则。不能拿旧会话记忆验收删除，也不能承诺历史会话、物理向量擦除、普通成员权限或恶意越权隔离。该样例省略usage_mode，不外推所有kind/模式。

请求正文逐次独立保存、不得复用覆盖；摘要可以补算但要如实注明时间，不能冒充发送前留存。task ID从原SSE/账本程序化取得，读错ID的404属于输入错误；删除进程尚在运行时的回读不是最终清理判定。

## 文档级版本冲突与两类条目更新

2026-11-01两请求拿同一ETag并发改同段，客户端实际重叠约1.95秒，只有一条成功，另一条409；保留的正文和新会话正式引用均为赢家67箱/紫砂仓。旧ETag再写仍409，不覆盖现文。两个不同分段拿同一文档ETag并发修改也仅一成功，说明本次保护按文档版本；不是按段互不影响的锁。冲突后重读整文档及新ETag，保留两方意图，由业务决定是否重组请求，不自动拿新ETag覆盖别人结果。

`rag/business_rule`与`rag/standard_qa`原ID更新后，独立回读两条Spec和文档/ETag；文档未变，新会话真实semantic_entries采用13箱与新校验词、67箱算6车，旧已完成任务仍保持53箱/7容量的历史结果。条目更新和文档修改是两条状态链，管理响应不代替运行消费。

停用唯一箱数/仓库事实段后，新会话检索及正式refs不含该段，拒答箱数/仓库/车数，但仍可采用独立有效条目的容量和校验词。恢复后文档/ETag回到停用前；不以资料不足推定整库/条目均不可用，不声明旧会话记忆或向量物理擦除。

## 来源追加、重复请求与替换窗口

当前正式`append-sources`与`delete-source`是分开的来源动作，没有已确认的一步原子replace、CAS或请求幂等键输入。要替换资料时，先把允许的新旧共存范围、生效规则及旧源移除条件说清楚；先准备新版并核正文，再精确删除旧source，最后用无file过滤的独立新会话验整库检索和正式引用。没有该合同不等于后端永远做不到；不能承诺跨这些步骤的原子切换。

2026-11-01小样本：不同新file双请求并发追加均保留；同一新file双请求返回同source且只新增一个file/source。已存在同file再次追加也保持source/job身份、正文及ETag，但来源列表被置pending、job标记reconcile_required=true，同时文档详情仍succeeded。因此“未重复新增”不代表无副作用no-op，也不能只凭旧文档可读宣布来源ready。按真实job准备状态统一reconcile并复查，不为让状态变绿重新append。

新旧四源共存时，独立问答按正文明确修订日期取47箱、另站26箱，11箱/车不可混装得5+3=8车，仍能回答旧源独有历史内容。旧source双请求删除一成功一404，剩余三源正文/ETag保持；删后整库新会话不带file_ids/文件名过滤，实际检索和正式refs均只含三源，旧独有内容拒答。四份Catalog原文件删源前后全文hash相同，说明删KB来源与删除原文件分开。不能据此声称物理向量清除、旧会话遗忘、任意并发规模或来源切换全程一致性。

## 条目迁移的重复、冲突与批次拒绝

2026-11-02两空隔离库的条目管理对照：首次跨库导入两条rag规则/标准问答成功，目标有独立entry ID，源条目不变；完全相同正文再次导入返回409 `ErrConflict`，目标两条的ID、Spec和时间均保持。不能把重复导入当幂等成功，或声称import会覆盖同key正文。

现有key换Spec，以及一批内两条新记录使用相同key，均明确409且原目标不变。一个有效条目与一个非法kind同批导入，两种顺序均400 `ErrParamInvalid`，后续完整list/export未出现任何新条目。空数组import返回`ErrEntriesRequired`，独立export仍空。这些负例只证明本次小批次明确拒绝、未观察到部分落地，不推广成所有输入和并发情况下的事务原子保证。修改已有条目应按真实entry ID使用正式update-entry并核对，不用import重放或擅自删旧条目消除冲突。

本轮只验管理/迁移，没有检索、模型或引用任务，不能把条目保存成功说成运行消费通过。两库删除后get/export/list-entries均404，自动创建的两个DB和四个空Volume也分别404；不宣称历史/物理数据擦除。

## 来源分页

正式 `semantic-model list-sources-page` 已在现有知识库上按 `page`/`page-size` 逐页读取：第1页和第2页
各返回2条且 `total=4`，`page-size=50`一次返回4条；超出范围的第999页返回total和分页参数但省略items。
分页验收必须按 `total` 合并各页并核对 source_id、model_id、file/resource ID 和状态；缺少items的越界页不等于
知识库为空，也不等于源被删除。只有在合并范围完整后再检查source jobs、分段和ETag。

## NL2SQL 语义条目的运行时核验

条目创建、导出和 SQL 预览不是完整运行证据。为字段说明、表说明、指标、关系、约束、动态
查询和标准问答分别准备不含答案的问题，保存真实 task/session；检查实际生成的 SQL、执行
结果和最终答案，表来源还要核对正式 `sql_table` 引用。约束条目要确认它确实进入生成 SQL，
例如限额或过滤条件，而不是只在 `list-entries` 中存在。

2026-11-17 的隔离订单/地区样例已覆盖 7 类条目：字段和表说明影响取数，指标与关系完成
汇总/JOIN，约束注入查询，动态查询和标准问答返回答案。条目 kind 与原型页面分类不必
机械一一对应，验收以真实业务作用为准；其它数据源、复杂 dialect 和普通成员权限仍需另验。
