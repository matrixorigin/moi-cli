# Recipe：Catalog → Workflow → Knowledge 端到端血缘

适用：通过 CLI 处理 Catalog 文件，再用知识库检索、联合推理和正式引用证明结果。先明确交付物、输入集合及副作用；历史成功不能代替本轮证据。

## 1. 确认实际输入

固定 workspace，从 Catalog 详情和本次上传/挂载收据取得原文件 ID。Volume 输入先查：

```text
moi-cli -o -no-interactive -workspace-id "$workspace_id" catalog volume-info --volume-id "$volume_id"
moi-cli -o -no-interactive -workspace-id "$workspace_id" catalog file-info --file-id "$file_id"
```

保存原文件集合、大小/哈希（可取得时）、Volume 数量和读取时间。明确为 0 或未就绪的实际输入不提交批处理；缺字段、空响应和不可用的列表不能当作 0。当前 `catalog files` 缺少 Volume 筛选输入，不能机械调用它证明清单完整。隔离新 Volume 可用本轮挂载收据与数量核对；未知现有 Volume 无法枚举时报告缺口，不能假定已掌握全部输入。

表输入按 `table-info` 和有界 SQL 另验结构、行数与关键值。表、Volume、文件 ID 不互换。

## 2. 选择准备路径

先读 Workflow 默认输入和版本，不把不同路径混为一谈：

- 从现有 Catalog 文件创建/追加 Knowledge sources：保存原始 `FileID`、`VolumeID`，按[知识库准备](create-and-prepare-knowledge-base.md)查询 source jobs；已授权准备且返回 `reconcile_required=true` 时推进本次 job。KB 自有 raw Volume 为空不妨碍显式原文件准备。
- 手动运行已配置的 Volume Workflow：预检其实际 `source_ref` 指向的 Volume。创建 KB 不自动将外部原文件复制进 KB raw；若默认输入不符合本次授权，先补齐配置或输入，不能盲目 run。
- 用户要求单独处理后将新文档入库：先核验新产物，再显式将该文档作为新 source。标准 RAG 的输出 ZIP/JSONL 只是派生产物，通常不需要再加为来源。

保存每步真实 workflow/version、原文件、model/source/job ID；job 完成后列表可能返回 `{}`，因此初始收据与后续 source 状态都要保留。

## 3. 提交一次，追踪本轮文件执行

```text
moi-cli -o -no-interactive -workspace-id "$workspace_id" workflow run --workflow-id "$workflow_id"
moi-cli -o -no-interactive -workspace-id "$workspace_id" workflow list-runs --workflow-id "$workflow_id"
```

Volume 运行返回的 `workflow_run.execution_id` **可能实际是分发任务 ID**，即使 `execution_mode=one_shot` 也是如此。保存原始字段，不直接认定对象类型。生产已验证的追踪方式：

1. 从 `list-runs` 的 `executions[]` 精确筛选 `dispatch_job_id == 本次 run 返回的 ID`。
2. 筛出的每条 `execution_id` 是真实子执行；保存其 `file_id`、case、version。
3. 核对子执行的文件集合等于运行前已确认集合，版本等于本次目标；按当前 help 处理分页，无法穷尽时保留观察限制。
4. 对每个真实子执行读状态、结果。无匹配时有界等待；达到预算则报告待确认，禁止重新提交。

```text
moi-cli -o -no-interactive -workspace-id "$workspace_id" workflow get-run --workflow-id "$workflow_id" --execution-id "$execution_id"
moi-cli -o -no-interactive -workspace-id "$workspace_id" workflow get-run-result --workflow-id "$workflow_id" --execution-id "$execution_id"
```

可用[只读本地检查脚本](../../scripts/inspect_batch_runs.py)检查已保存的原始 `list-runs` JSON 或含 `exit_code/response` 的收据：

```text
python3 "$skill_dir/scripts/inspect_batch_runs.py" "$runs_json" --dispatch-id "$dispatch_id" --expected-file-id "$first_file_id" --expected-file-id "$second_file_id" --version-id "$version_id"
```

`skill_dir` 为本次读取的 Skill 目录；每个已确认原文件各传一次 `--expected-file-id`。脚本不联网、不写 MOI，只按精确 ID 检查执行集合。退出码 0 仅证明该集合 completed；2 为输入/关联不符，3 为子执行失败，4 为待完成或列表范围不足。后续仍要核验业务结果、产物和引用。

直接查父分发 ID 返回 404 不等于文件失败；`list-file-executions` 返回 `{}` 也不能证明没有子执行。只有子执行终态与业务结果都通过才能报告本轮处理成功。失败重试在授权内只选择失败子执行，使用原分发 ID 的 `--source-dispatch-job-id` 和真实失败 ID；保留新旧执行，不把重试提交当作成功。

## 4. 核对产物与血缘

对原文件读取关联资产与血缘：

```text
moi-cli -o -no-interactive -workspace-id "$workspace_id" catalog file-data-asset --file-id "$file_id"
moi-cli -o -no-interactive -workspace-id "$workspace_id" workflow catalog-file-lineage-overview --file-id "$file_id"
```

在 derivations 中按本轮 `case_id` 和逻辑产物槽筛选 parsed docset、output file、index；不能选第一条或全局最近记录。同一原文件可同时保留历史和本轮派生产物。需要交付时用输出收据指定的真实 Volume 下载；资产登记的 Volume 不一定等于下载所属 Volume，不凭目录名称推测。

```text
moi-cli -o -no-interactive -workspace-id "$workspace_id" catalog download-file --file-id "$output_file_id" --volume-id "$output_volume_id" --output "$local_output"
```

本地检查字节数、哈希、ZIP 完整性及关键正文。文件名不是血缘证据；解析 metadata 中的原 file ID、本轮 case 的 derivation 和 execution 输出应能互相印证。

## 5. 检索、联合问答和缺失信息

```text
moi-cli -o -no-interactive -workspace-id "$workspace_id" semantic-model list-sources --model-id "$model_id"
moi-cli -o -no-interactive -workspace-id "$workspace_id" semantic-model get-source-document --model-id "$model_id" --source-id "$source_id"
```

按原文件匹配 source。分别证明 ingest、分段、索引/检索和问答，不能只凭 `succeeded` 或 source 数量推导全部通过。重复处理同一原文件时，检查来源是否重复、分段是否重复；稳定 chunk ID 不能单独证明索引更新。

按[引用核验](query-knowledge-and-check-citations.md)使用独立会话，DeepSeek 默认不传推理等级，问题不预填答案：

- 两份资料分别提供规则和现状，要求联合计算且正式引用两份来源；
- 加问资料没有提供的信息，要求明确无法确定；检索无命中不构成凭空补全的授权；
- 若 SSE 超时，取同一 task 的终态，不重发。`get-task` 可能返回 `json` 字符串信封，要再解析一层；
- 核对正式 `source_refs` 的 model/source/file/chunk 与当前分段正文，兼容实际 `chunk_id` 或 `chunk_ids` 字段。

报告分开说明本轮输入、子执行、产物、来源/分段、答案/引用和未通过处。只有运行时真正执行检索的 Workflow 才能称“Workflow 内动态 RAG”；Codex 先检索再把事实放入 DSL 属于另一种编排，不混报。

## 内容变化的深度补验

有授权的隔离样例可用[分段 CAS 维护](maintain-knowledge-sources.md#文档分段的-cas-维护)临时改变一个事实，
在新会话中用相同、未预填答案的问题检查检索正文、结果和正式引用是否一起改变；选择会触发
非整数向上取整等业务边界的数值。完成后以最新 ETag 恢复原正文/启用状态，独立回读，再以
另一新会话核验回答回到原值。保留原始→修订→恢复三轮证据，不用旧会话记忆充当刷新。
这只证明分段编辑后的检索可见性。原文件来源替代按[来源维护](maintain-knowledge-sources.md#新版原文件替代旧知识来源)执行；Workflow 内的动态检索按[运行时检索](run-runtime-knowledge-workflow.md)执行。两者2026-10-31均已在隔离样例中通过，新file身份替代与Automation生成的Workflow是各自限定路径；同file ID原位覆盖、任意DSL直接RAG、向量单路更新仍未证明。

## 小型多来源ZIP与抽取字段

交付ZIP时独立下载每个输出，检查CRC、完整文件列表、Markdown/parse正文、原file ID及额外字段。2026-11-01两份合成documents绑定真实源ID，两个ZIP正文43/29箱分别正确，不串源；但实际输入`documents[].metadata.extraction_result`未出现在parse JSON或抽取侧文件，即使execution completed也不能称抽取字段完整。

显式顶层`extraction_result`另一路已产生各包的`*_extract.json`，保存同一aggregate总计72及双源ID；这是聚合语义，不是逐文档缺失的修复，不能按各包file_name将72算成每源数量。无血缘登记节点时，不把包内源ID当作Catalog artifact血缘已登记。

精确删除后同时核placement、file-info、Volume及必要的血缘：本次7处引用清空、3原文本404，但4个ZIP元数据延后回读仍可见。删除请求成功不是元数据消失或物理擦除。GC只有全工作区批量/时间阈值、无精确file ID输入时，不以工作区GC补本次隔离清理。
