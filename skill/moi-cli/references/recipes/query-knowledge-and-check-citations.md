# Recipe：知识库问答与引用核验

适用：用户要求根据指定知识库回答问题，并提供可核对的来源。文档上传、分段可读与
检索回答是不同验收项；先完成[知识库准备](create-and-prepare-knowledge-base.md)。

## 明确范围并发送

1. 保存真实 workspace、model、source 和 file ID；确认源已入库、启用、未过期。
2. 按[模型领域](../domains/model.md)选择适用于 `knowledge_qa` 的模型，生产环境优先
   `deepseek-v4.1-flash`。`model list` 是通用发现入口，不能证明模型在该通道可调用；
   调试时才做有界的最小通道检查，避免盲目轮换模型。推理等级按模型领域中记录的
   入口差异与实际验证结果处理，不能假设所有 A2A 入口都支持同一个 max 参数。
3. 用 `schema conversation create-ui-session` 确认输入后，创建独立会话并保存返回的 `id`。
4. 用 `schema agent a2a-stream` 核对请求。请求文件使用 `agent_code: explore`、
   `jsonrpc: "2.0"`、唯一 `id`、`method: "message/stream"`；`params.message` 使用
   `kind: message`、`role: user`、唯一 `messageId` 和 `parts: [{kind: text, text: 问题}]`。
   模型通过 `params.model` 明确提供。不要把验收预期答案写进问题。
5. `params.metadata` 绑定返回的会话 ID（字符串）和当前范围：`session_id`、`workspace_id`、
   `semantic_model_ids`（真实 model ID 数组）。`scope_metadata` 中的 `semantic_model_ids`
   与 `semantic_model_names` 使用对应 ID/名称字符串；`scope` 包含同一 workspace、session
   和 `scope_metadata`。这些是 A2A 运行输入，不是 CLI source JSON 的 `SourceType` 等字段。

多库输入沿正式前端`buildScopeMetadata`形状：`params.metadata.semantic_model_ids`是数值ID数组，
两处`scope_metadata.semantic_model_ids`是同顺序的逗号分隔字符串（例如`"50001,50002"`），
名称同样逗号分隔；用本次已发现ID/名称填写，不把数组直接塞进字符串字段。更换范围时使用独立新会话做对照。

```text
moi-cli -o -no-interactive -workspace-id "$workspace_id" conversation create-ui-session --json-file "$session_request"
moi-cli schema agent a2a-stream --validate-file "$question_request"
moi-cli -o -no-interactive -workspace-id "$workspace_id" agent a2a-stream --json-file "$question_request" --output "$stream_path"
```

CLI 的 JSON 结构校验只验证外层形状，`params` 仍须符合真实 A2A 协议。已有会话必须复用
已确认 ID，不能从会话列表首项取值。

## 读取结果并核对引用

- `a2a-stream` 退出码 0 只说明流已写出；流内仍可能是 failed。解析 SSE 的 `data:` JSON，
  关联同一 `taskId`，要求明确的最终 `status.state=completed`。结束前断流属于结果未知，
  使用 `agent get-task` 查原任务，不重新发送问题。返回若为 `{ "json": "JSON-RPC字符串" }`，须再解析一层后读取 task 的真实终态与 artifacts。
- 终态失败时保存 `metadata.reason_display`、`traceId`、task/context ID 和错误原文，停止
  宣称引用验收通过。`executor-health` 可达不能证明模型调用成功。
- 若存在 `artifact-update` 的 `knowledge.answer`，读取回答文本和其 metadata 的 `source_refs`。
  Astra 回答也可能在 `status-update` 的文本 parts 中；检查成功的 `select_final_sources`
  工具结果，并区分选源前的推理/说明与选源后的答案。不只匹配整份流中是否出现答案关键词。
- 要求答案满足用户问题，并且实际来源记录命中本次目标 file/source ID。模型说“已引用”、
  问题中出现文件名、流中出现某个 ID 均不能代替正式来源记录。
- 文档分段引用可使用单个 `chunk_id` 或多个 `chunk_ids`；按实际形式关联同一知识库、
  文件和分段正文，再检查正文能否支持答案，不能只识别单个 ID 字段。
- 来源还可能是合法候选全集。若任务要求每句精准引用，必须核对引用内容能支持该句，
  不能把候选集合当作模型实际用到的精确来源。
- `select_final_sources` 的 `semantic_model_id` 是整数，不从 CLI 的字符串 model ID 直接传成字符串。
  2026-11-14 QA 的一次首次选源因 `/sources/*/semantic_model_id` type 被拒绝，运行时纠正后另一次
  `ok=true`、`accepted=true`，最终 source_refs 与这次成功选源一致。验收要分别保留失败 call 和成功 call，
  不能把首次失败抹掉，也不能只凭较早的检索 artifact 判最终通过。

验收至少保留：原问题、终态、正式答案、来源记录、目标文件匹配结果。文档已入库但问答
服务失败时，只报告准备通过、问答未通过；不重新上传、建库、改模型路由或绕过权限来制造通过。

深度验收时再加一层来源闭环：对本次范围内的每个 source 读取 `get-source-document`，
核对 segment 正文、稳定 `segment_id`、`etag`、enabled/effective_enabled 和 `ingest_status`；
随后用独立新会话提问。最终 `knowledge.answer.metadata.source_refs` 必须同时命中当前
`semantic_model_id`、`source_row_id`/source、file 和支持答案的 chunk。观察到
`search_rag_chunks`、`select_final_sources` 只证明当前运行时 RAG 路径被调用，不足以推断底层
fulltext/vector 分支、并发替换原子性或向量物理清理已经通过。

## 任务内检索、证据与 cite

这一路径只用于平台已经准备好的 live task。`semantic-model find-files/search/read-source/
list-passages/read-document/search-images/read-evidence/cite` 除产品身份外还需要当前任务的
authorization proof；普通 native Astra 终端不自动具备该证明。不要复制任务证明、伪造 task scope，
也不要把外部 CLI 的参数校验成功写成检索通过。没有准入时保留已完成的准备结果，让用户在选中该
知识库的新任务中提问。

在已准入任务中按“发现文件 → 检索 → 完整读取 → 引用”推进：

1. `find-files --query` 只用于定位当前任务可见资料；保存返回的 model/file/volume 身份，不按文件名猜 ID。
2. `search --keywords ... --max-hits 8 --output "$search_result"` 保存真实检索结果。flag 默认
   `before=2`、`after=1`；使用 typed JSON 时显式给需要的 `max_hits` 和上下文边界。`--fulltext-only`
   是用户明确要求的检索模式，不作为向量失败后的静默 fallback。
3. 从结果读取 enabled passages。`read-source` 或 `list-passages --from "$result" --index N`
   沿真实来源读取；后续页只使用前次返回的 `source_page`/cursor 和同一 model/file 身份，
   不与 keywords/max_hits/before/after 混用。`read-document` 用于 parsed Markdown 的精确 file/cursor
   读取，不能把分段 ID 当 markdown file ID。
4. 对将用于答案的每条证据完成必要读取；图片结论还需 `search-images` / `read-evidence` 的真实图片证据，
   文本命中不证明已经看图。保留 evidence index 与正文、页码、图片和 source 身份的对应。
5. 用保存结果生成正式 cite，例如 `cite --from "$evidence_result" --indices 0 --request-id "$request_id"`；
   多个 `--from` 与各自 `--indices` 按真实结果配对。request ID 只用于完全相同请求的replay/核对，
   不把不同证据选择伪装成同一请求。

`cite` 成功证明当前任务为选定证据生成了引用结果，不等于 Agent 已回答，也不等于最终答案自动采用它。
需要验收问答时仍须完成上文 A2A 流程，核对终态、答案事实、`search_rag_chunks` /
`select_final_sources` 和 `knowledge.answer.metadata.source_refs`。正式 source_refs 必须能回到当前
model/source/file/chunk 或 visual hit；不得从本地 cite 结果拼造 A2A source_refs，也不得反过来用答案
文本声称运行时 cite 已通过。

## 跨库联合问答与范围对照

先独立读取每库各source文档，按`model + source + file + chunk`建立预期事实表，再发不含预期答案的问题。
2026-11-01 两库四文档联合问答返回两站各自的库存/规则，分别算得补给27箱运输3次、39箱运输4次；
completed任务的正式source_refs精确覆盖四份当前文档及支持计算的分段，未混用两站规则。
只绑定第一库的新会话问第二库独有事实时，实际检索后说明资料不足，检索结果和正式引用均只来自第一库，
未披露第二库的库存与核验标记。验收拒答时同时检查检索产物和引用，不能只看回答说“没有使用其他库”。
这个样例证明选定范围筛选和有限的多库联合引用，不证明普通成员权限隔离、恶意越权防护、任意多库规模
或来源并发替换原子性。检索返回的routes可作为运行元数据记录，不能据聚合`fulltext/vector_l2`标签
声称两个底层分支已分别验收。流30秒超时后只回读原task，不重新提问。

## 同一会话的多轮追问

多轮验收不能只新建两个会话各问一次。第一轮创建 UI session，保存返回的 session ID、
A2A `contextId` 和 task ID；第二轮同时复用同一个 `session_id` 和
`params.message.contextId`，使用新的 `id` 与 `messageId`。问题必须引用上一轮事实，
例如第一轮只问待运箱数和每车容量，第二轮要求“根据刚才两个数字计算车数”，不能在第二个
问题里重新给出答案。

每轮分别核对 `status.state=completed`、答案和 `knowledge.answer.metadata.source_refs`，再用
`conversation list-ui-messages --json '{"session_id":真实ID}'` 回读同一 session 中的两轮
user/assistant 消息。`list-ui-messages` 当前只有 JSON 输入，没有独立 `--session-id` flag。

2026-11-17 的生产隔离复测中，第一轮回答 53 箱、每车 8 箱；第二轮沿用同一 context 算出
7 辆，两个任务均 completed，正式引用均命中当前 model/file/chunk。会话历史包含两条用户
消息和两条助手消息，用户消息保存同一 `semantic_model_ids=[130002]`。结合既有双库联合
问答证据，单库多轮、多库联合及来源引用均有 CLI 证据；页面选择按钮和数量徽标仍属于 UI。
测试结束后删除本次 session 并确认 get 返回 404，不删除历史用户会话。

## 图片读取与正式引用

图片问题需要分别确认“检索到图片”和“实际看图”。`semantic-model search-drawings`
支持当前 help 中的 `--knowledge-base-id`、`--query-image`、`--top-k`；用本次授权图片作查询时，
保存真实返回的 source file、object、image/page image ID、页码、bbox 和检索 channel。
相似图片命中只证明检索，不代替 Agent 看图与正式引用。文本查图没有命中也不证明图片不存在，
可在原范围用文件标题等已知线索作有界检索，不编造标识。

### 多图联合向量化与相似度矩阵

最多 6 张图的验收先为每张图保存真实 file/source ID，完成来源和图片索引准备，再逐张作为
`--query-image` 调用 `search-drawings`。`--top-k` 要覆盖预期的候选图块数量；响应会按
`source_file_id` 聚合成 drawings，不能把返回图块数当成不同文件数。

建立 6×6 矩阵时，每次查询记录全部 target file、score 和 rank，并断言：6 次查询各返回
6 个不同来源、36 个单元齐全、查询图自己的 `source_file_id` 排名第一。相邻帧是否比其它
文档更相似可作为样本断言，但不能凭分数高低编造语义。2026-11-17 的 6 张图均满足上述
条件；文本查图 `WORD-7421` 返回 0 被单独保留，没有混入图片向量通过结论。

准备阶段仍需逐源核对持久 document 的 succeeded、分段和 ETag。若 sources/job 已成功而
document 仍 pending，按准备 Recipe 做一次有依据的终态 reconcile，再回读；不要循环推进。

独立下载实际返回的图块和页图，打开核对正文、图中数字、箭头/位置、页码及裁剪区域：

```text
moi-cli -o -no-interactive -workspace-id "$workspace_id" semantic-model preview-artifact --model-id "$model_id" --file-id "$image_file_id" --output "$image_path"
moi-cli -o -no-interactive -workspace-id "$workspace_id" semantic-model preview-artifact --model-id "$model_id" --file-id "$page_image_file_id" --output "$page_path"
```

保存下载字节数、格式、尺寸和哈希；服务可能重绘或压缩，视觉对应不等于原文件字节相同。
普通文本 OCR 可能只有页眉页脚，不能据此补出图片中的事实。

问答中让 Agent 先加载真实 `read_file` 工具 schema，并使用检索返回的真实图片 ID。
当前工具默认 `max_bytes=65536`（64 KiB），允许上限 `1048576`；图像必须完整读取，
超过默认值时根据实际文件大小显式传足够的 `max_bytes`，且不超过当前 schema 上限。
例如已确认图片小于 1 MiB 时，合法参数形状为：

```json
{"file_id":"检索返回的图片ID","max_bytes":1048576,"image_question":"请读取图中标注和箭头，只报告实际可见内容。"}
```

`volume_id` 是可选的规范根 Volume，不能把知识库 raw Volume 或血缘资产的 Volume 自动当作
派生图片的规范根。2026-10-31 隔离实测中，携带 `volume_id` 的读取返回 JSON-RPC `-32602`；
相同图片保留合法 `max_bytes`、按 schema 省略可选 `volume_id` 后成功。2026-11-14/15 补充核对发现，
该派生页图没有原文件 Volume 的直接 placement，旧调用的参数归属不匹配，不应再写成 `volume_id`
功能通用失效。原文件在其规范根 Volume `130012` 中有直接 placement 时，携带同一 file/Volume 的
`read_file` 已成功；这只证明原文件的正确归属路径，不证明派生页图也能沿用该 Volume。
仅在当前 schema 允许且文件身份/授权边界明确时可省略，由正式服务解析文件归属，不能删必填字段、
切换身份或将省略参数当成普遍的权限绕过方法。

图片初始化、检索命中和实际读图必须分别通过。2026-11-15 的 `preserve_embedding=true` 样例同时保存文本/
图片 embedding，来源 succeeded，2 张图片 ready。第一次独立问答没有读出标注；第二次显式合法 `max_bytes`
并在错误 Volume 被拒绝后省略可选参数，实际读出 `MOSS-64 / 38 L`、`V-8 / CLOSED` 和
`SKY-27 → MOSS-64 → AMBER-91`，正式 `visual_hit` 指向当前 model、原文件和第 1 页页图，独立下载
页图核对一致。两次结果都保留：第一次不完整不能提前算通过，第二次也不能只凭模型自述，必须有真实
工具产物、正式引用和独立核图。

必须看到成功工具结果的 `encoding=image_analysis`、`image_analysis.service_slot=img_parse`
及实际分析文本，再核对最终答案是否与图片一致。模型自述“已看图”不算证据；图片分析使用
专门服务槽位，不能把主问答模型可调用等同于该槽位可用。

要求 `select_final_sources` 成功，正式答案的 `source_refs` 包含 `type=visual_hit`、真实源文件、
model/source、`image_file_id` 或 `page_image_file_id`、页码（及返回的 object/visual_refs）。
将这些字段与检索结果、独立下载页图逐一关联。带图片 ID 的拒答只证明选源成功，不证明图片
事实问答通过。

流式 CLI 可能在 30 秒超时而原任务仍完成。保存已收到的真实 `taskId`，用
`agent get-task --agent-code explore --task-id "$task_id"` 回读原任务及 artifacts；
不要据超时重发。保留首次失败及后续有界纠正，公开的 `-32602` 不足以证明执行器未绑定、
模型不支持图片或文件无权限，根因仍需同一 trace/task 的服务端日志。

## 规则条目和冲突资料的证据范围

rag/business_rule的运行消费按[规则维护Recipe](maintain-knowledge-sources.md#业务规则的真实问答消费)核验。`semantic_entries`与文档chunk/file是不同身份，禁止伪造正式引用；文件来源只能证明其实际覆盖的事实，不能把规则推论说成原文引述。

冲突资料要检索双方并分别核对真实source/file/chunk和正式source_refs。没有权威、更正、时序等可判定依据时明确不能定案，不任意挑一个数值。本轮同批待复核签收记录19/23箱，回答明确不一致、无法确认唯一数值，两个正式引用分别对应两份来源。只证明这类限定冲突，不代表全部冲突推理已覆盖。

本轮三次A2A流均30秒超时，沿原task回读最终completed，没有重发或另开成功任务掩盖原任务。

## 有明确生效与替代依据的冲突

资料明确给出批准、生效、到期和替代关系时，先检索双方，按用户目标日期选择适用规程并说明另一版不适用的原因；两份正式refs都须匹配真实model/source/file/chunk。2026-11-01虚构运输规程修订01已到期，修订02自11月1日生效并明确替代旧版，独立问答选11箱/车、57箱至少6车，正式引用覆盖两版支持分段。批准/时效来自文档本身，未作外部认证；不能把较新的文件名、更大数值或模型偏好当权威。没有这些依据的冲突继续明确不能定案。
