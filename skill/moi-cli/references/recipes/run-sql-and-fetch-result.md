# Recipe：执行 SQL 并取得结果

开始先选择路径：当前有界只读查询可用`sql run`直接返回结果，合法零行已有限定实测，见
[内联路径与限制](#合法零行查询的内联路径)。需要保存异步query再分步取数时采用下面流程；
其中空结果读取和下载仍有500记录，不以新查询成功替代原statement结果。

1. 确认 workspace、数据库和用户要求的只读 SQL；若表未知，先执行 `sql meta-databases`、`meta-db-tables` 或 `meta-table-columns`。
2. 以 `sql execute --db ... --query ... --limit N` 执行有界查询，保存 `query_id`。
3. 用 `sql describe-query --query-id` 查询状态；pending 时在有限预算内轮询同一个 query ID。
4. status 为 success 且返回非空 `statement_id` 后，执行 `sql query-result --statement-id`。
5. 报告 columns、返回 rows、total 和 limit；不要把样例 rows 说成完整数据集。
6. 失败、权限错误或缺少 statement ID 时保留服务端错误并停止；不要重跑相同 SQL 修复结果。

结果行可能是 `result[][]` 的 `{String, Valid}` 单元格，按 columns 类型解释；不要把
`Valid=false` 的空值当作空字符串。分别报告请求 limit 与结果 limit，它们可能不同。

空结果与 NULL 不同。2026-11-01 同环境对照中，一行内的 NULL、空串和数字0分别返回
`{String:"",Valid:false}`、`{String:"",Valid:true}`、`{String:"0",Valid:true}`，可以区分；
而 `SELECT CAST(1 AS SIGNED) AS synthetic_id WHERE 1=0` 执行和状态均 success，
`query-result` 仍返回 HTTP 500 / `Error 20301 (HY000): invalid input: empty query result`。
遇到该错误保存原 query/statement 与 request_id，报告“查询成功、空结果读取失败”，
不能转成空数组、说成登录失效，或用另一次 COUNT 成功冒充原结果读取已修复。

DML/DDL、查询取消、下载结果和工作簿保存不由只读查询自动扩展；已有明确授权则在其范围内执行。

## 已授权的 DDL、DML 与事务

写入前固定本次数据库、表、预期改动和恢复方法。`execute` 仍先返回 query_id；用
`describe-query` 取 statement_id 和 success，不使用 `describe --statement-id query_id`。
建库/建表/ALTER/INSERT/UPDATE/DELETE 等没有数据行时，只核对执行终态和独立结构或数据，
不强行取空结果。`EXPLAIN` 则读取实际计划行并核对目标表及筛选条件。

本次已验证一条 execute 请求内的 `BEGIN; …; COMMIT` 和 `BEGIN; …; ROLLBACK`；
分别再发独立 SELECT 验证持久化或回滚。不能把 BEGIN、写入、COMMIT 分成三次CLI调用，
据此假设它们共享连接和事务。此样例不证明所有SQL语法、隔离等级或并发事务行为。

Workflow 跨节点失败不会自动回滚前一节点已提交的 SQL。重试前先查已落地内容；若获准重放，
使用业务上正确的唯一键和 upsert 或正式去重机制，再独立核对重复数与内容，不能只检查流程 completed。

## 合法零行查询的内联路径

需要直接取得有界只读结果时，正式`sql run --query ...`提供内联结果。2026-11-01对同一隔离表和零行谓词，它返回独立新query/statement、status=success及三列schema，result/total零值字段缺省；同表非空对照返回一行正确数据。此合法成功响应可按该合同解释为零行，不能把任意缺字段或错误响应视作空结果。

原`execute → describe-query → query-result`对另一原statement仍500 empty query result，`download-query-result`也500；没有修复这些路径。应在任务开始按目标选择内联或异步路径；已有查询失败时保存原错误，不静默重执行。只有用户目标允许重新执行只读查询且数据时点差异可接受时，才能明确另发内联查询；它不能取回旧查询的历史快照，更不能用于重跑DML、未知副作用SQL或未知状态操作。

2026-11-15 生产补测再次确认：真实表内联 `WHERE 1=0` 返回 success 和三列 schema，但同一 statement 的
`query-result` 与下载仍分别为 500，未生成 CSV。纯常量零行 SELECT 在省略和显式数据库时又返回 405
`ErrResponseInvalid`，没有 statement ID。真实表内联、纯常量内联、异步读取和下载是不同路径，不能互相借用成功。

## 多表对账、晚到版本与结果落表

先明确一行代表什么业务对象，再分别把多个一对多来源汇总到同一粒度后JOIN。不能把明细和支付事件直接相连后求和；本轮该错误写法把订单25.25算成50.50，SQL仍成功。用独立金额计算、主键集合、缺失和孤儿来源对照实际查询结果，不能只看success或总行数。

版本事件先按明确业务版本及稳定次序用ROW_NUMBER选当前记录，再汇总；晚到旧版本不因行ID更大而生效。退款、撤销、缺明细和缺付款分别定义规则；LEFT JOIN之外另查无主对象的来源，避免孤儿记录静默消失。

用户已授权结果落表时，先核真实查询行，再用业务唯一键和明确的字段替换语义upsert；区分替换与累加。重复落表后独立比对每行、金额和主键数量。本轮由Codex读取已核查询再分步写入，不能宣称跨表原子同步或通用幂等。源改动后的新查询、当前结果表和原statement历史分别保存；本轮改源/删源后原statement的JSON/CSV仍保持原值，不代表删除历史。

事务失败后独立读取原表。本轮单请求转账成功后余额8/32；另一单请求更新后故意重复主键报500，后查仍8/32。只证明该错误样例未持久化中间更新，不推广到断连、其它SQL错误或Workflow跨节点。

本轮含四个标量COUNT子查询的一次只读统计返回405/ErrResponseInvalid，未得到query ID，根因未知；另行明确编写UNION ALL统计成功。保留原失败，不机械重交、不归因登录，也不把替代查询说成原路径修复或所有标量子查询不支持。

## 精度、CSV表达与重复列名

2026-11-02新增有界常量对照：JSON中的`BIGINT`超2^53、`UNSIGNED BIGINT`最大值、负`DECIMAL`六位小数、VARCHAR前导0及中文emoji/引号/字段换行均逐格保持。读取时先保存`columns`类型和单元格`String/Valid`，大整数与DECIMAL不能先转IEEE浮点再格式化。DATETIME(6)本次返回微秒及`+0000 UTC`文字后缀；这不证明用户时区转换或所有时间类型的语义。

同一statement的CSV对非NULL内容保持逐位/逐格一致，但`NULL`与有效空串都变成空字段，**此CSV不能无损恢复二者区别**。需要保留该区别时交付JSON的`Valid`与列类型，不能凭空串推断NULL，也不宣称CSV可作为无损往返备份。用CSV解析器核对引号和内嵌换行，不按文本换行数算数据行。

两个相同SQL alias在本次响应中改名为`repeated_0`、`repeated_1`，CSV表头与返回`columns`一致，值和类型保持。只按本次实际metadata和列位置解释，不把原SQL alias当作固定输出字段，也不硬编码所有重复名都会按同一规律改写。

本次纯DECIMAL常量减法与中文常量的`sql run`返回405/`ErrResponseInvalid`、stdout空且没有query/statement ID；仅多一个末尾换行的同内容SQL，另一次明确异步`execute → describe-query → query-result`得到精确`0.000001`。这是入口差异的限定对照，原405未修复、根因未知。先检查CLI exit/transport、stderr和stdout结构，stdout空时不要直接解析后因本地异常重发；若目标允许另一条只读查询，再明确选异步路径，不能重跑未知写入。

真实接续验收也通过：本地任务读取原成功statement，按5项预先声明判据完成后，新进程`resume`不新增CLI调用。原405步骤保持未完成、没有用成功的另一任务覆盖。使用方法见[保存任务与接续](record-and-resume-task.md)。

## SQL历史、详情与profile的分层

`sql query-history-overview`是当前已验的只读统计入口；本轮返回`total=8/success=8`。稳定版 0.1.8 的
`sql query-history`曾只返回`limit=10`，已知statement的`query-detail`和`query-profile`也曾返回400；
这些 regular 失败继续保留。当前 dev 在 2026-11-14 生产复测中已能读取历史列表、按时间范围筛选，
并对已知statement取得详情和profile，不能再把旧稳定版失败描述成 dev 当前限制。两条 CLI 的状态和证据
必须分别报告，详情仍只能使用真实statement ID。

2026-11-17 dev 补测进一步验证 `--statement SELECT`、`--status Success`、
`--selected-fields statement_id,statement,status,duration,result_count`、`--order desc
--order-by request_at` 和 `--limit/--offset`。筛选结果、字段投影和排序均可读；字段名必须
来自真实响应，不能把旧猜测的 `result_rows` 当成当前 `result_count`。组合请求若返回 400，
应拆成单参数定位，保留原失败，不能反复提交同一个无效组合。

列表下载和精确 `statement-id` 筛选仍是独立子项。此前同一 statement 能由 `query-result`
取出数据，但精确筛选、`query-detail`、`query-profile` 返回400的记录继续保留；本轮通用筛选
成功不证明该指定 statement 已恢复，也不证明页面下载或通过。
