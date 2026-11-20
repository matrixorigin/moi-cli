# Recipe：文件导入到表

适用：用户明确要求把本地文件导入已选 workspace 的目标表或 Volume。隔离验收中，用户已授权的合成样例也属于明确输入；先记录样例内容、独立目标和预期结果。

## 本地 CSV 直接入表

`connector upload` 支持一次请求上传文件、关联 Volume、创建表和启动导入。先发现目标
database/Volume ID，保存本地表配置；不能把此模式与下面的暂存模式混用而重复提交。

```sh
moi-cli connector upload --help
moi-cli -o -no-interactive -workspace-id "$workspace_id" connector upload --file-path "$file_path" --volume-id "$volume_id" --table-config "$table_config_json"
```

多文件使用重复的`--file-path`参数，在一次upload中提交；各文件须符合本次表映射，且只提交一次。
表配置是 JSON **内容**，不是文件路径；在执行工具中把已保存文件读成一个参数，避免 shell 转义破坏内容。
新建 CSV 表所需的主要字段如下，ID 和列定义必须来自目标和实际文件：

```json
{
  "sheet_name": "orders.csv",
  "new_table": true,
  "database_id": 123,
  "isColumnName": true,
  "columnNameRow": 1,
  "rowStart": 2,
  "csv": {"separator": ",", "delimiter": "\"", "isEscape": false},
  "create_table": {
    "name": "orders",
    "tableColumn": [
      {"column": "order_id", "dataType": "INT", "col_num_in_file": 1},
      {"column": "amount", "dataType": "INT", "col_num_in_file": 2}
    ]
  }
}
```

示例中的 `123` 是占位符，不能直接执行。`columnNameRow`、`rowStart` 和 `col_num_in_file`
在此配置中从 1 开始；示例跳过第 1 行表头，从第 2 行导入。它们不同于 WorkItem 的 `start_row`。

返回 `acceptance_status`、逐文件 `results`、`file_ids` 和 `task_id`。`accepted` 只证明已接收；
此命令已启动任务，不能再调用 `run-task-now`。查询同一 `task_id` 的 `get-task` 与 `task-runs`，
保留真实 table ID 和运行 ID。任务与运行的数字状态枚举不同，不混用；结合 `latest_rows`
中的完整统计和后续 SQL 证明结果：行数、失败/跳过行数、逐行内容或合计都符合预期。

结构化小表的最小核验同时包含：

1. `table-info` 或 `sql meta-table-columns` 的字段名、顺序和类型；
2. `table-data`/`preview-table` 的返回窗口与 total，不能把预览条数当整表行数；
3. 有界 SQL 的 `COUNT(*)`、业务主键逐行值、精确 `SUM` 及异常行统计；
4. `download-table` 的 CSV 解析记录数、列名、字段和嵌入换行，不用物理行数代替记录数。

需要区分 NULL 与空字符串时，在 `csv.null_values` 只声明实际使用的 NULL 标记，并用 SQL 的
`IS NULL` 与 `= ''` 分别统计。普通 CSV 下载通常把二者都写成空字段，必须保留 typed JSON 的
有效性信息，不能从下载后的空单元格反推 NULL。完整执行顺序和导出互证见
[Catalog 文件与结构化表核验](verify-catalog-file-and-table.md)。

## 追加和覆盖已有表

先读目标表结构，固定字段映射、业务策略和写入前快照。`connector upload` 的现有表配置使用
`new_table:false`、真实 database_id/table_id、`existed_table_opts` 和 `existed_table`；不能把新建配置简单换表名。

```json
{
  "sheet_name":"orders-append.csv", "new_table":false,
  "database_id":123, "table_id":456,
  "isColumnName":true, "columnNameRow":1, "rowStart":2,
  "csv":{"separator":",","delimiter":"\"","isEscape":false},
  "existed_table_opts":{"method":"append","table_name":"orders"},
  "existed_table":[
    {"tableColumn":"order_id","column":"order_id","col_num_in_file":1},
    {"tableColumn":"amount","column":"amount","col_num_in_file":2}
  ]
}
```

数字均为占位符。覆盖仅在用户明确授权清空/替换该表后把 method 设为 `overwrite`。
追加需证明旧数据保持、新记录只进入一次；覆盖需证明旧记录已移除、替换文件完整进入。
结合本次 task 的 latest_rows read/succeeded/failed/skipped/complete 和目标逐行/汇总核验，不能把累计导入数当本次行数。
命令已自动启动任务，超时先查原 task，不额外 run 或重传。CSV 示例不证明其它源类型、去重策略和部分失败恢复均已通过。

## XLSX真实入表与重复追加

XLSX使用同一正式 `connector upload --file-path --volume-id --table-config` 路径。
先读取真实工作表名、列和行，配置里的sheet_name使用工作簿内部工作表名（本次为Sheet1），不是XLSX文件名；
新建表配置保留new_table/database_id、表头行、起始行和create_table.tableColumn映射，按实际列填写1起始的col_num_in_file，
不携CSV专用配置。先确认Volume与数据库为本次获准目标。

本次一次upload已启动任务，后续只查询原task/runs，读入两行且succeeded=2、failed/skipped=0、complete=true，
SQL逐行取回17和29合计46。随后对同表按已有append配置只追加一次相同XLSX；SQL查回四行17/17/29/29合计92。
此append样例没有自动按文件内容去重；不能把再次上传成功当作幂等，也不把它当成去重策略故障。
重试或结果未知时先读原任务和目标数据，不能直接重传。

JSON local-upload本次只证明暂存成功，两次返回不同文件ID，不证明JSON入表或去重。
已查的前端结构化文件入口接受CSV/XLS/XLSX；其它格式需正式领域合同，不能凭开放SourceConfig对象猜格式和映射。
反过来，schema开放也不能直接断言CLI缺能力，应先查现有正式upload调用配置；本次XLSX就是由该合同跑通。

## 先暂存，再创建导入任务

1. 读取用户给出的 workspace、文件路径、目标连接器/Volume、格式和表映射；缺少任一关键目标时停止并询问。
2. 用 `connector list`、`connector get`、Catalog tree/database-children 发现真实连接器、目标 Volume 或数据库/表 ID。
3. 执行 `connector local-upload --file-path`，保存 `conn_file_ids` 和 `file_ids`。
4. 读取 `schema connector create-task`，构造本地 request JSON；用 `--validate-file` 校验结构。
5. 执行 `connector create-task --json-file`，保存 `task_id`。
6. 先查任务是否已自动开始；只有目标操作明确要求且没有现有运行时，才执行 `run-task-now` 或其它已确认的启动命令。
7. 用 `task-runs`、`get-task` 检查运行终态；失败时保存错误和失败文件 ID，不重复创建任务。
8. 导入到表时用 Catalog table-info 或 SQL metadata/有界 SELECT 核验表和数据；导入到 Volume 时核对文件关联和可见性。

上传成功、任务创建成功或任务已启动都不能代替目标内容核验。

## 坏数据与恢复

失败先查原task与task-runs；需要实际错误时按返回的workflow_execution_id查get-run，不猜任务ID。
2026-10-31两份CSV中一份INT字段为非数字，整批两文件均失败；独立COUNT确认目标表0行。
修正文件后，重新发现原表ID，以`new_table:false`和已确认的同表映射新建一次append导入，三行逐行查回、合计71且无重复。
这是“改正输入后的新任务”通过，不是原文件重试或部分成功已通过，也不承诺所有失败都会回滚。
恢复前必须先确认已写入多少行及唯一键，不能把此样例的0行推广到其它任务。

常规CLI的`task-files`缺少task-id/分页过滤参数且实测返回500，但这不代表没有失败文件定位入口。
先`connector task-runs --task-id`，保存对应run的`workflow_app_id`、`workflow_execution_id`，再调用：

```sh
moi-cli -o -no-interactive -workspace-id "$workspace_id" workflow list-run-import-files --workflow-id "$workflow_id" --execution-id "$execution_id"
```

核对返回`import_context.task_id/run_id`确为原任务/运行，使用`files[]`中状态明确为failed的
LoadTaskFile `id`。这些ID即使偶尔与暂存ID数值相同，也不能推导可互换。
`retry-task-files --task-id ... --ids ...`使用精确失败记录ID；省略ids会选择全部失败文件，
不在用户只允许部分恢复时省略。缺少准确关联和已修复的失败原因时停止，不重传冒充原任务恢复。

## 多表部分成功与原任务定向恢复

需要不同文件独立恢复时，先确认现有任务的写入分组。正式upload的多表配置为
`{"multi_sheet":true,"tables":[...]}`，每个table保留真实database_id/table_id、映射与明确
`conn_file_ids`。先`local-upload`取得暂存ID；若配置已经指定这些ID，后续`connector upload`
可只带`--volume-id`与`--table-config`，不重复上传文件。upload会自动启动，保存原task并查询。

重试选择文件不等于按文件裁剪全部写入：当前合同按选中文件找到涉及的表，随后可能保留该表配置中的
整组`conn_file_ids`。同表多个文件时，要先核对整个表分组与已落地数据，不能承诺只重放某一文件。
`update-task`及`update-task-from-spec`不支持修改表配置/conn_file_ids；开放JSON不能绕过正式输入类型。

### 同表多文件失败的额外核验

同一张表一次提交多个文件时，不能把 `workflow list-run-import-files` 的文件状态直接当作每个文件
的成功/失败结论。2026-10-31 的隔离演练中，同表两个CSV遇到预置主键冲突，execution 为failed，
task 的 `load_results` 也显示失败，但文件列表把两个文件返回为同一状态。独立SQL只查到预置冲突行，
两个输入文件的数据都未落地；不能据文件顺序推断第一个已经写入，也不能把列表状态当作真实写入结果。

发生这类情况时必须同时保存并核对：

1. 原 execution 的终态和服务端错误；
2. 原 task 的 `load_results`、成功/失败计数和运行关联；
3. 目标表的实际 SQL 行、唯一键和写入前快照；
4. `list-run-import-files` 返回的 `import_context.task_id/run_id` 与原任务是否一致。

只有四者能够互相解释时，才可以决定是否定向重试。若同表含多个输入，先确认重试会不会带上该表的
整组 `conn_file_ids`；不能只凭单个文件状态选择重试，也不能在未知状态下重新上传。重试前保存失败
文件快照，因为重试后文件可能迁移到新的 run，旧 run 的列表会改变。恢复后再次用 SQL 和 execution
结果核对是否重复写入；`cumulative_rows` 只表示本次或当前 run 的统计时，不得当作整张表总行数。

隔离验收可用两个已建表、每表一文件：第一表正常写入，第二表预置明确的演练主键冲突。
先独立SQL证明第一表已写入，第二表仅保留预置冲突；精确修复本次演练冲突后，只retry第二文件的
真实LoadTaskFile ID。恢复后逐表核对数据，并核对第一文件没有新增运行或重复写入、原失败execution仍保留。
这套修复只适用于已授权的隔离演练；业务任务必须先制定针对真实故障的修复和副作用核验，不能照搬删行。

2026-10-31该两表样例已实际完成：首次1成功1失败，A表2行合计56、B表只有预置999；
修复后同task只补B文件，新的execution completed，A仍2行56，B仅新记录23，旧execution仍failed。
未再次上传文件或调用run-task-now。该样例恢复后仍有`TARGET_PATH_RELATION_INVALID`，需按每表真实ID核验，
不能忽略错误或把空table_path当无数据。`latest_rows`与`cumulative_rows`均为补跑1行，不等于两表共3行。
LoadTaskFile记录的run归属会迁移：重试前先保存原失败列表，之后原run的列表可能不再含该失败文件；
这不表示原来没有失败，原execution及已保存快照分别核对。其它格式/去重策略仍未完整验收。

2026-11-01 复核同表原始恢复证据：精确清理隔离预置冲突后，仅选择conflict文件的LoadTaskFile ID重试；
新的execution输入`retry_file_ids`与文件列表都只含该ID，但目标SQL最终同时查回good与conflict两行，金额11+22=33。
任务表配置仍含两个`conn_file_ids`，说明选择范围不等于文件内容的重放范围；同表分组可能一起处理。
最新运行读入/成功均2行、失败0，而task仍保留历史`failed_file_count=1`和旧失败原因，不能用累计摘要推翻
新execution与SQL，也不能隐藏旧失败。此样例首次整批未落地，不证明已有成功数据时重放不会重复写入。

## 已有成功数据的同表分文件恢复

同表整组`conn_file_ids`与“每文件独立配置、多个配置指向同一table_id”须分别判断。2026-11-01后者已验：表中先有baseline一行5，新批两个配置各含一个文件；good写11成功，bad因演练主键冲突失败。保存SQL、原execution、原文件列表后，精确修复本次冲突，仅retry bad的真实LoadTaskFile ID；没有重传、没有另建任务。新execution完成，原失败保留，最终baseline/good/bad三行5/11/22合计38，先成功数据未重复。

这个结果限于本次按文件独立配置的任务，不覆盖同一配置含多个文件的整组重放。恢复前读取真实table_config，不能只因table_id相同就合并或拆分既有业务配置。

失败状态可能晚于execution投影到task和文件记录。本次execution刚failed时，task仍运行、文件仍显示成功；延后回读才收敛为1成功1失败。先有限等待并核对原run/task/file关联、错误及SQL，再选失败ID；不依据首个文件列表立即重试。恢复后`TARGET_PATH_RELATION_INVALID`仍存在，累计行数仅1而整表3行；业务数据恢复通过与元数据异常分别记录。

## CSV边界及真实XLS

CSV的`csv.separator`、`delimiter`、`isEscape`和`null_values`按实际文件显式配置。复杂字段样例已验逗号、双引号、单元格换行、中文；`null_values`仅指定反斜杠N时，该标记写成NULL而空串仍为空串。JSON中的反斜杠正确编码，用SQL返回的Valid及`IS NULL`独立核验，不能从Catalog显示的NULL字符串推断。分号分隔的两行样例也已通过，不等于自动识别全部CSV方言。

真正OLE/BIFF的XLS已验两行17+29=46，failed/skipped为0，Catalog及SQL逐行一致。sheet_name必须用真实内部表名（本次LegacyData），不使用文件名，不带CSV专属配置。不能用改扩展名的XLSX代替，也不外推公式、合并单元格、多sheet和所有坏记录策略。upload的`--dedup`针对Volume文件身份，不证明表行去重。

## 同组重放与主键冲突策略

已有表配置的`conflict`是表行冲突策略，与Volume文件`--dedup`不同。正式上传合同支持：`0`普通INSERT（省略同0）、`1` INSERT IGNORE保留已有行、`2`重复键时更新非主键列。必须先确认真实主键、字段映射及用户要保留还是更新旧值；不为消除报错自动改成忽略或覆盖，也不把`2`当作清空整表的`overwrite`。无正确唯一键时，这些策略不能保证业务去重。

2026-11-01单配置含good_changed/bad两文件，已有baseline5、good旧11及演练冲突seed999，首次执行failed、输入新值均未写入。精确移除演练seed后，只retry bad的LoadTaskFile ID，新execution仍failed：bad22已实际入表，writer却也读取未选good并因其主键200冲突。真实`tables.write-table.runtime_input.file_ids`含两份物理文件，SQL最终三行38。**选中bad且bad入表，不等于整次恢复成功，也不等于未选文件没有执行。**首个文件列表可能暂报成功，延迟task/run才收敛为失败；后来新任务重用暂存文件会迁移文件run归属，必须保留重试前快照。

在同样两份暂存输入上，另建已授权的明确策略任务：`conflict:1`完成、两行均skipped，旧11保持、整表三行38；`conflict:2`完成、处理两行且把11更新为17，整表仍三行、合计44。处理/成功2行不是新增2行；分别核对latest_rows的read/succeeded/skipped/failed/complete及SQL。此处是新配置任务，不是修改原失败task的冲突策略或原任务无损续跑。只证明本次有主键的限定配置，不保证所有格式、其它约束或任意并发。

2026-11-15 对原生产任务再次只选择 bad 文件重试，仍因未选择 good 文件的主键 200 冲突失败；
重试前后表中都只有 baseline 100 和 good 200，bad 300 未写入。这是当前生产 writer 的
选择范围缺陷，不是 Skill 选错 ID。本地 dev 源码虽有过滤修复逻辑，但原任务尚未复验通过，
不得据源码或独立配置的成功改记原场景恢复，也不要继续盲重试。

节点详情以真实trace的flow/node名核对。此例job_summary复合路径返回404，而trace的`tables.write-table`能读到实际输入；不能凭一次错误node key判执行缺失，更不能猜工作项ID代替node key。

## 复合主键、字段子集与坏值的范围

单主键策略2成功不能外推复合主键。2026-11-01复合`(order_id,line_no)`的策略1保旧/新增通过（2读1成功1跳过）；策略2两次独立任务均报`unsupported DML: update primary key on duplicate`，前后SQL完全不变。第二次已校正CSV配置且用纯文本控制，不能归咎NULL/中文。SHOW COLUMNS确认两列PRI，真实writer无显式key_columns；没有捕获线上生成SQL/部署commit，不断言具体实现根因，不向开放table-config猜加未支持字段。

2026-11-15 已对原失败任务和原文件再做一次定向重试，仍在写表节点返回同一主键错误，
前后两行及保留字段完全不变。当前生产原场景仍失败；本地 dev 源码排除主键列的逻辑
不能作为线上通过证据。继续采用下面的暂存表加显式 SQL 时，必须明确这是另一业务路径。

业务明确允许另建暂存及SQL合并时，可在同一正式CLI内采用已验替代：创建隔离暂存表→正常导入→查询全部小样本，确认源复合键唯一及金额/日期/NULL→`sql execute`明确列出目标字段，使用`INSERT INTO ... SELECT ... ON DUPLICATE KEY UPDATE`，仅更新业务允许的非主键列→独立核对新增键、旧值和未映射字段。此为新任务/新SQL，不是原失败task续跑，不静默切换。样例3→4行、合计62.84，旧保留字段不变，SQL rows_affected=3不等于新增3行；不外推重复源键、大表或任意事务合同。

单主键已有表只映射三列时，策略2已验旧金额20.20→25.25，未映射正文/日期/保留值不变；新行未映射列按NULL/default生成。所有字段都须回读，不以同一策略推定各种约束均可用。

`null_values`应放在`csv.null_values`，顶层同名字段会在本次规范化配置中丢失；核对保存请求和task真实配置。`conflict:1`不等于忽略所有坏记录：DECIMAL越界导致本次两行batch失败，正常同行也未入表；改正金额后另建任务成功。文件内复合键重复样例3读2成功1跳过、本次保留第一条，不保证跨文件或并发固定赢家。

## 文件为空、编码与列位置

2026-11-02限定CSV样例已验UTF-8 BOM、CRLF、中文emoji、引号及字段内真实换行；VARCHAR前导0保持。按实际字节和字段配置解析，用独立SQL的文本/HEX及下载原文件核对，不能经浮点或表格软件自动转换后判断精度。

只有表头与零字节文件是不同场景：表头文件完成零行导入，read/succeeded/failed/skipped均为0，目标表独立COUNT为0，`catalog download-table`可下载纯表头。该结果不修复SQL异步空查询结果读取/下载的历史失败。零字节上传本轮exit1/ErrServer、错误包络内status=200，无任务或文件ID；已观察原表与Volume不变，但错误语义和未知ID问题保留为部分完成，不把它叫作正常拒绝或自动重传。

2026-11-15 补充暂存路径：0 字节 CSV 可以先经`local-upload`取得真实暂存 ID，schema结构校验也通过，
但`create-task`仍返回`ErrServer`、包络status=200，request ID为`da57d852-89a0-436c-99c4-167152db65fe`。
任务列表没有目标，原表两行和Volume 0文件保持。暂存成功只证明文件被接收，schema通过只证明结构有效；
没有任务ID时不得猜测任务已创建或重复提交。

列改序和多余列必须按正式file_position显式映射：本轮4/3/2映射正确，多余列未写入，原有两行保留、最终四行逐格一致。缺少已映射的第三列在upload阶段即400/ErrParamInvalid、没有任务ID，原数据不变；这不证明writer已经运行或目标NOT NULL约束已验。报告区分上传校验、执行写入及最终数据，不能只看其中一层。

清理后文件引用消失、Volume为空且资源404，不等于底层文件元数据或物理内容已擦除；本轮已知file-info仍可读取。无新ID的拒绝请求不能猜对象删除或为清理执行GC。

## 混合媒体上传与文件去重

本地上传的格式验收要同时检查任务终态、Catalog 扩展名和下载字节。2026-11-17 在同一隔离
Volume 上传 PNG、WAV、MP4，任务 `status=4`、`success_file_count=3`，Catalog 分别显示
`png/wav/mp4`；三个下载文件的 SHA-256 均与源文件完全一致。结合此前 Markdown 和无扩展名
样例，通用“文档、图片、音频、视频等混合本地上传”可判通过。只证明受管文件字节保存与下载，
不证明媒体内容解析、转写或图片索引已经执行。

`connector upload --dedup` 的 `by` 使用 `name` 或 `md5`，`strategy` 当前正式实现支持
`skip` 和 `overwrite`；省略 `by` 时默认 MD5。名称去重实测：

```sh
moi-cli -o -no-interactive -workspace-id "$workspace_id" connector upload \
  --name "$unique_task_name" --volume-id "$volume_id" --file-path "$file_path" \
  --dedup '{"by":["name"],"strategy":"skip"}' --path-regex '.*'
```

同名文件再次上传返回 `skipped by dedup`、`No files were accepted, so no import task was created.`，
没有 task ID；这是正确的防重复结果，不要为取得 task ID 再传一次。

同轮 MD5 对照未通过：原 WAV 与再次上传的 WAV 文件名不同、字节和 SHA-256 相同，
`{"by":["md5"],"strategy":"skip"}` 仍返回新 file/task ID 和 `success_file_count=1`；
另一次同内容上传也再次接受。下载新文件再次确认哈希一致，因此不能把该路径写成 MD5 去重
成功。当前保留为产品行为异常候选，根因仍需服务端确认；修复前不要重复制造重复文件。
名称去重成功不能掩盖这个 MD5 分支，也不能把文件 `--dedup` 与表行 `conflict` 混用。
