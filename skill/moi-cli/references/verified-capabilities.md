# 按场景登记的实测范围与保留缺口

当前机器可查询的唯一状态源是 [capability-cards.json](capability-cards.json) 中的 `observations`。
当前发布版本是 Skill `0.4.9`；文中标注的旧版本只表示历史执行或历史沉淀版本，不能当作当前安装版本。
下表只展示尚未通过的有效场景记录；通过场景可用 `python3 scripts/capability_status.py query CAPABILITY_ID`
从 Skill 根目录发现，再按 `--scenario` 或 `--condition KEY=VALUE` 精确查询。默认渠道为 `regular`，候选必须显式选择。
不带场景/条件或没有匹配条件时返回 `unverified/needs_scope_match`，不能拿同一能力的任意成功当通行证。

本次新增 2026-11-14 生产环境 dev CLI `0.1.0-dev.d5c5c3a` 的 125 条原版场景对照结果，保留
Skill 0.3.20 历史摘要及后续 QA 记录。生产记录包含 83 条通过、16 条失败、14 条部分覆盖、11 条缺条件和
1 条 CLI 缺口；不用新增成功抵消旧失败，也不把缺条件写成通过。`validated_on=null` 表示该场景未进入完整业务执行
或原摘要没有逐项确认日期，记录日不是复测日。仓库外证据标记为 `external_not_bundled`，仅说明来源位置，
便携包校验不要求该文件存在。状态定义和历史替代规则见[能力模型](capability-model.md)。

<!-- capability-status:begin -->
来源：`capability-cards.json`（SHA256 `2619e43c50d1c1373c34d7e2fbddd1623896c9a3633857147fb11e135e7aca02`）。此段由 `scripts/capability_status.py render` 生成，修改JSON后重生成；手写说明留在标记外。

| capability_id | scenario_id | 渠道/版本 | 验证日期 | 状态 | 实际结果 | 限制/来源 |
|---|---|---|---|---|---|---|
| catalog-discovery | catalog-files-no-filter-rejection | regular/0.1.8 | 2026-11-02 | cli_missing | 正式help只提供workspace-id；不带Volume输入调用返回ErrParamInvalid/status400，无法得到可归属的文件清单。 | 这是CLI输入/接口合同缺口，不证明Catalog文件服务不可用；不能按返回错误自行猜参数或改用raw HTTP。；来源 boundary-interface-20261102:catalog-files-no-filter-rejection |
| catalog-discovery | unknown-nested-discovery | regular/0.1.8 | 2026-11-01 | partial | 已知嵌套清空可用，tree/children未枚举嵌套。 | 不把已知子树清理成功外推未知子树发现。；来源 history-0.3.20:已知嵌套文件夹清空 |
| catalog-discovery | volume-file-filter-input | regular/0.1.8 | 未逐项确认；记录日2026-11-02 | cli_missing | 常规CLI缺所需筛选输入。 | 候选成功不是全局CLI已升级。；来源 history-0.3.20:Catalog→Workflow→Knowledge 新批处理与问答 |
| sql-query | async-zero-row-download | candidate/0.1.0-dev.d5c5c3a | 2026-11-14 | failed | 原异步空结果读取和下载仍失败 | 非零NULL对照或内联成功不能覆盖原失败。；来源 dev-full-regression-20261114:sql-history-zero/summary.json |
| sql-query | async-zero-row-download | regular/0.1.8 | 2026-11-01 | failed | query-result和下载500 empty query result。 | 非零NULL对照或内联成功不能覆盖原失败。；来源 history-0.3.20:SQL NULL/空串/0与空结果对照 |
| sql-query | inline-constant-zero-row | candidate/0.1.0-dev.d5c5c3a | 2026-11-15 | failed | 纯常量零行 SELECT 在省略和显式数据库时均返回 405 ErrResponseInvalid，没有 statement ID。 | 只判定本次常量零行形式；真实表零行内联可执行，不能混为同一结果。；来源 skill-044-data-query:evidence/skill-0.4.4-coverage/data-query/calls/inline-zero-constant-with-db/response.json |
| sql-query | inline-zero-row-result-and-download | candidate/0.1.0-dev.d5c5c3a | 2026-11-15 | failed | 真实表内联零行返回 success 和三列 schema；query-result 与下载均 500，未生成 CSV。 | 内联执行成功只证明查询终态和列结构，不能证明空结果可读取或下载。；来源 skill-044-data-query:evidence/skill-0.4.4-coverage/data-query/calls/inline-zero-table-result/response.json |
| sql-query | sql-csv-null-empty-distinction | candidate/0.1.0-dev.d5c5c3a | 2026-11-14 | failed | 生产 CSV 中 NULL 与空串仍为空白；JSON 可区分 | 这是本样例CSV表达限制，不断言后端根因或所有CSV格式；需要区分时保留JSON Valid及列类型，不能宣称CSV无损往返。；来源 dev-full-regression-20261114:production-sql-fidelity/assertions.json |
| sql-query | sql-csv-null-empty-distinction | regular/0.1.8 | 2026-11-02 | failed | 同一查询JSON能区分NULL与空串，但CSV两个字段都导出为空白，不能无损恢复这项区别。 | 这是本样例CSV表达限制，不断言后端根因或所有CSV格式；需要区分时保留JSON Valid及列类型，不能宣称CSV无损往返。；来源 depth-sql-fidelity:sql-csv-null-empty-distinction |
| sql-query | sql-inline-decimal-constants-405 | candidate/0.1.0-dev.d5c5c3a | 2026-11-14 | failed | 生产内联仍为405，与原版相同；异步精确计算通过 | 该失败只针对本次入口与SQL；无抓包/部署commit，根因未知，不说小数运算整体不支持。；来源 dev-full-regression-20261114:production-sql-fidelity/durable-real-task/evidence/calculate-001/response.json; production-sql-fidelity/decimal-async-state.json |
| sql-query | sql-inline-decimal-constants-405 | regular/0.1.8 | 2026-11-02 | failed | sql run对纯DECIMAL减法与中文常量exit1，stderr405/ErrResponseInvalid，stdout空且无query或statement ID。 | 该失败只针对本次入口与SQL；无抓包/部署commit，根因未知，不说小数运算整体不支持。；来源 depth-sql-fidelity:sql-inline-decimal-constants-405 |
| sql-query | sql-reconciliation-sr-07 | candidate/0.1.0-dev.d5c5c3a | 2026-11-14 | failed | 生产仍为405，与原版相同；QA曾改善为通过，不能外推。独立UNION计数正确 | 原请求仍失败，内部原因未知；新查询成功不是修复原请求，也不推成所有标量子查询不支持。；来源 dev-full-regression-20261114:production-sql-reconciliation/final-counts.json; production-sql-reconciliation/final-counts-union.json; sql-reconciliation/final-counts.json |
| sql-query | sql-reconciliation-sr-07 | regular/0.1.8 | 2026-11-02 | failed | 包含四个标量COUNT子查询的只读请求返回405/ErrResponseInvalid，没有query ID；另立UNION ALL只读查询取得15/7/6/6正确计数。 | 原请求仍失败，内部原因未知；新查询成功不是修复原请求，也不推成所有标量子查询不支持。；来源 complex-sql-reconciliation:SR-07 |
| sql-query | workbook-version-governance | candidate/0.1.0-dev.d5c5c3a | 2026-11-15 | partial | 同ID重存仍可能覆盖，但保存前重读可识别顺序陈旧快照并另建版本，A=328、B=369。 | 真正并发仍需单一写入者或外部协调，不能把本地重读当成服务端CAS。；来源 skill-046-data-analytics:state.json#workbook_guard |
| sql-query | workbook-version-governance | regular/0.1.8 | 2026-11-02 | partial | 合法保存/执行及错归属拒绝有证据；finalize可被同版本save覆盖。 | 无已确认CAS；不能宣称不可变或通用并发合并。；来源 history-0.3.20:SQL工作簿 |
| source-staging | agent-file-lifecycle | regular/0.1.8 | 2026-11-02 | partial | 正式file/task来源与字节、分页、兄弟保留有证据。 | file_path不保目录；缺rename/update/session筛选；无placement删除另有生命周期边界。；来源 history-0.3.20:Agent文件 |
| data-ingestion | catalog-zero-byte-staged-create | candidate/0.1.0-dev.d5c5c3a | 2026-11-15 | partial | 0 字节 CSV 可 local-upload 暂存并取得 conn_file_id；schema 结构有效，但 create-task 返回 ErrServer/status=200。任务列表无目标，原表两行与 Volume 0 文件保持。 | 暂存成功不代表可建任务；schema 通过不代表业务输入有效；没有任务 ID，内部原因未知。；来源 skill-044-data-query:evidence/skill-0.4.4-coverage/data-query/calls/zero-create-task/response.json |
| data-ingestion | catalog-zero-byte-upload-error | candidate/0.1.0-dev.d5c5c3a | 2026-11-14 | partial | 仍ErrServer且status=200，目标两行和Volume两文件保持 | 不是成功空导入或明确400拒绝；没有任务失败行统计；内部原因和未返回ID的潜在暂存元数据未知；未重发。；来源 dev-full-regression-20261114:catalog-boundaries/cb03-upload.json |
| data-ingestion | catalog-zero-byte-upload-error | regular/0.1.8 | 2026-11-02 | partial | 实际上传失败：CLI exit1，ErrServer，错误正文status200；无task/file ID。SQL原2行全值不变，Volume仍2文件128B。 | 不是成功空导入或明确400拒绝；没有任务失败行统计；内部原因和未返回ID的潜在暂存元数据未知；未重发。；来源 depth-catalog-boundaries:CB-03 |
| data-ingestion | composite-key-update | candidate/0.1.0-dev.d5c5c3a | 2026-11-15 | failed | 2026-11-15 对原失败任务只选择原文件重试，写表仍报 unsupported DML: update primary key on duplicate；重试前后隔离表均为原 2 行且内容不变。 | 不是缺参数、登录或 Skill 编排问题；当前 dev 源码的修复逻辑尚未在当前生产原场景生效。；暂存表加显式 SQL 是替代业务路径，不是原自动更新恢复。；来源 skill-044-data-query:evidence/cli-dev-full-regression-20261114/data-assets/production/import-composite/composite-update-result.json, skill-048-import-retry:evidence/skill-048-import-retry/calls/composite-retry-result |
| data-ingestion | composite-key-update | regular/0.1.8 | 2026-11-01 | failed | 两次自动更新失败。 | 单键字段更新或显式SQL合并是不同路径。；来源 history-0.3.20:复合键自动更新两次失败 |
| data-ingestion | single-config-group-replay | candidate/0.1.0-dev.d5c5c3a | 2026-11-15 | failed | 2026-11-15 再次只选择 bad 文件 517534000258（业务键 300）重试，仍因未选择 good 文件的业务键 200 冲突失败；重试前后隔离表均只有 baseline 100 和 good 200，bad 300 未写入。 | 选择范围与实际 writer 输入不一致，不是调用者选错文件。；每文件独立 table-config 是规避方案，不是同配置选择性重试修复。；来源 skill-044-data-query:evidence/cli-dev-full-regression-20261114/data-assets/production/import-policies/group-retry-result.json, skill-048-import-retry:evidence/skill-048-import-retry/calls/group-retry-result |
| data-ingestion | single-config-group-replay | regular/0.1.8 | 2026-11-01 | failed | 重试仍出现失败/整组重放，成功文件范围隔离未通过。 | 独立table-config成功不覆盖此原路径。；来源 history-0.3.20:单配置多文件、已有数据时只retry失败文件 |
| data-export | external-connector-export | candidate/0.1.0-dev.d5c5c3a | 未逐项确认；记录日2026-11-14 | blocked | 没有真实隔离外部目标与目标端核验条件 | 本地下载不等于外部导出通过。；来源 dev-full-regression-20261114:production-scenario-results.json#external-connector-export |
| data-export | external-connector-export | regular/0.1.8 | 未逐项确认；记录日2026-11-02 | blocked | 缺隔离真实连接和目标端核验条件。 | 本地下载不等于外部导出通过。；来源 history-0.3.20:没有外部目标的Connector导出 |
| workflow-lifecycle | deploy-volume-defaults | regular/0.1.8 | 未逐项确认；记录日2026-11-02 | cli_missing | 旧常规CLI无default-values输入flags。 | 候选支持须先看当前binary help。；来源 history-0.3.20:Catalog→Workflow→Knowledge 新批处理与问答 |
| workflow-lifecycle | manual-publish-pointer | candidate/0.1.0-dev.d5c5c3a | 2026-11-14 | failed | v2已发布，但App仍指向v1且运行400，旧问题未修复 | 正式candidate替代成功不能称旧路径修复。；来源 dev-full-regression-20261114:state.json#manual_publish |
| workflow-lifecycle | manual-publish-pointer | regular/0.1.8 | 未逐项确认；记录日2026-11-02 | failed | 原部署指针问题未修复。 | 正式candidate替代成功不能称旧路径修复。；来源 history-0.3.20:同App正式候选更新并运行 |
| workflow-lifecycle | workflow-controls-expression-deploy500 | candidate/0.1.0-dev.d5c5c3a | 2026-11-14 | failed | 仍不支持该表达式；错误由500变为明确400，不计业务通过 | 对象是否曾创建仍未知；total51只返回20，正式list/apps help无分页/筛选输入；无法定位原ID，不能声称已精确清理。；没有服务端日志，不能确定是表达式、DSL编译、参数推导或其他原因；后续变体不证明原失败已修复。；来源 dev-full-regression-20261114:calls/wf-null-failure-deploy/response.json |
| workflow-lifecycle | workflow-controls-expression-deploy500 | regular/0.1.8 | 2026-11-02 | failed | deploy返回ErrServer/status500，无workflow/version ID；list/apps首20均未发现唯一名称，未重发原请求。 | 对象是否曾创建仍未知；total51只返回20，正式list/apps help无分页/筛选输入；无法定位原ID，不能声称已精确清理。；没有服务端日志，不能确定是表达式、DSL编译、参数推导或其他原因；后续变体不证明原失败已修复。；来源 depth-workflow-controls:workflow-controls-expression-deploy500 |
| job-recovery | node-plan-hash-idempotence | candidate/0.1.0-dev.d5c5c3a | 2026-11-14 | failed | 相同计划并发两次加顺序一次均创建不同执行，旧限制保持 | hash是合同来源保护不是幂等键；业务键方案不修复该路径。；来源 dev-full-regression-20261114:node-hash-result.json |
| job-recovery | node-plan-hash-idempotence | regular/0.1.8 | 2026-11-01 | failed | 双请求产生两分支各写文件，再重发再增文件。 | hash是合同来源保护不是幂等键；业务键方案不修复该路径。；来源 history-0.3.20:节点继续重复提交 |
| job-recovery | pause-resume-cancel | candidate/0.1.0-dev.d5c5c3a | 2026-11-14 | partial | 暂停/恢复/三种取消通过，但取消后恢复由原版409变为500，错误语义退化 | 不证明副作用回滚或任意取消任务可续跑。；来源 dev-full-regression-20261114:state.json#execution_controls + calls/controls-cancel-paused-resume-rejected/response.json |
| knowledge-preparation | document-governance-pagination | regular/0.1.8 | 未逐项确认；记录日2026-11-02 | cli_missing | 缺少正式flags。 | 不能把候选成功当常规CLI可用。；来源 history-0.3.20:整份文档启停/过期与大文档分段分页 |
| knowledge-preparation | duplicate-append-readiness | regular/0.1.8 | 2026-11-01 | partial | 只保一source但重置pending，旧分段仍可读。 | 旧内容可读不能直接算ready。；来源 history-0.3.20:Knowledge来源并发与移除 |
| knowledge-preparation | image-read-with-volume | candidate/0.1.0-dev.d5c5c3a | 2026-11-14 | failed | 派生页图沿用原文件 Volume 被拒绝；新证据确认该派生文件没有原 Volume 的直接 placement，属于输入归属不匹配。 | 保留该错误输入的失败记录，不把它解释为 volume_id 功能通用失效。；原文件在规范根 Volume 的成功读取见 image-read-with-canonical-volume；派生页图应按可选合同省略 volume_id。；来源 dev-full-regression-20261114:image-visual-result.json, skill-044-expansion-workflow-knowledge:state.json#image_volume_contract |
| knowledge-preparation | image-read-with-volume | regular/0.1.8 | 未逐项确认；记录日2026-11-02 | failed | 原读取失败保留。 | 准确内部根因未定，不用无volume成功覆盖。；来源 history-0.3.20:图片索引、实际看图和图片引用 |
| knowledge-preparation | knowledge-empty-create-response-decoder | regular/0.1.8 | 2026-11-02 | failed | create-empty两次、typed list及get均因semantic_model_kind字段解码失败。没有重提创建，沿正式agent-builder资源发现精确找回原model ID，export仍可读；空数组import明确ErrEntriesRequired，前后export无条目。替代发现可接续管理，不修复原命令失败。 | list-entries的{}未单独作为空库证明；结合export及首次导入后完整total=2/两条内容判断。；没有增加第二次A创建；B是不同已授权目标。；创建/读取失败不能因后续管理可用改判通过。；来源 depth-knowledge-migration:KM01 |
| dashboard-analytics | dashboard-relations-drel02 | candidate/0.1.0-dev.d5c5c3a | 2026-11-15 | partial | 原生 JOIN 钻取参数完整但仍400，request_id=b2032785-b9dc-4aa6-92cb-a90a75468bef；物理快照钻取和CSV可用。 | Snapshot is separately maintained and does not fix native JOIN drill or provide atomic synchronization.；来源 skill-046-data-analytics:state.json#dashboard_snapshot, skill-048-workflow-data:calls/join-dashboard-current + calls/join-drill-current + ../skill-046-data-analytics/calls/dash-snapshot-drill-success + ../skill-046-data-analytics/snapshot-east.csv |
| dashboard-analytics | dashboard-relations-drel02 | regular/0.1.8 | 2026-11-02 | partial | 两张JOIN图都能查询，但正确x维度的钻取均返回400不支持；A图的钻取下载同样拒绝。本地现有owner合同明确仅单物理表，不属于缺少输入，也没有用快照结果冒充JOIN钻取成功。 | 两张JOIN图的原生钻取业务目标未实现；不是CLI漏传x维度。；本地源码仅作合同线索，不证明线上部署版本；线上拒绝回执独立保存。；来源 complex-dashboard-relations:DREL02 |
| dashboard-analytics | zero-row-result-download | candidate/0.1.0-dev.d5c5c3a | 2026-11-15 | failed | 已有合法零行 execution completed，但 chart-result 与 CSV 下载失败；本轮未重复创建看板。 | 非零结果或执行 completed 不能覆盖空结果交付失败；at 是时间参数，不是 execution ID。；来源 skill-044-data-query:evidence/cli-dev-full-regression-20261114/query-analytics/dashboard-zero/summary.json |
| dashboard-analytics | zero-row-result-download | regular/0.1.8 | 2026-11-02 | failed | 结果读取及CSV500。 | 非零NULL/空串/0对照不修复原空结果。；来源 history-0.3.20:看板空结果 |
| data-share | active-publish-direct-delete | regular/0.1.8 | 未逐项确认；记录日2026-11-02 | failed | 503且关系消失后目标仍能读，残留已隔离清理。 | 先退订再删除成功不能修复直接撤回。；来源 history-0.3.20:分享退订/删除、写保护与直接撤回 |
| data-share | unselected-table-denial-semantics | candidate/0.1.0-dev.d5c5c3a | 2026-11-15 | partial | 未选和移除表确实不可读，但错误仍是500/table does not exist；加入表后恢复可读。 | 不能把物化范围生效说成规范403/404权限语义已经改善。；来源 skill-046-data-analytics:state.json#share_lifecycle |
| data-share | unselected-table-denial-semantics | regular/0.1.8 | 2026-11-02 | partial | 不可读，但实际错误500。 | 不能称正确404或完整拒绝语义。；来源 history-0.3.20:数据分享 |
| connector-lifecycle | connector-config-authoring | regular/0.1.8 | 未逐项确认；记录日2026-11-02 | cli_missing | 正式面缺config输入。 | 不是只缺用户凭证；不得转低层。；来源 boundary-connector:现成连接与新建配置的边界 |
| connector-lifecycle | external-connection-verification | candidate/0.1.0-dev.d5c5c3a | 未逐项确认；记录日2026-11-14 | blocked | 缺少新的隔离外部源与安全配置条件 | schema/命令存在不等于连接通过。；来源 dev-full-regression-20261114:production-scenario-results.json#external-connection-verification |
| connector-lifecycle | external-connection-verification | regular/0.1.8 | 未逐项确认；记录日2026-11-02 | blocked | 缺真实隔离连接目标和安全配置条件。 | schema/命令存在不等于连接通过。；来源 history-0.3.20:没有外部目标的Connector导出 |
| compute-lifecycle | loaded-suspend-resume | candidate/0.1.0-dev.d5c5c3a | 未逐项确认；记录日2026-11-14 | blocked | 未提交负载，未把删除当作真实暂停恢复 | 不是计算运行实际故障；不为测试强行计费。；来源 dev-full-regression-20261114:compute/state.json |
| compute-lifecycle | loaded-suspend-resume | regular/0.1.8 | 未逐项确认；记录日2026-11-02 | blocked | 已绑定候选部署冲突，未提交负载。 | 不是计算运行实际故障；不为测试强行计费。；来源 history-0.3.20:计算资源创建/查看/预检/删除 |
| audit-observation | known-request-audit | regular/0.1.8 | 2026-11-02 | partial | 审计时间/request_id可关联，领域删除后404独立确认。 | 默认20条，无正式完整分页；首屏无记录不等于未发生。；来源 history-0.3.20:操作日志 |
| audit-observation | sql-history-list-contract | regular/0.1.8 | 2026-11-02 | partial | query-history返回exit0，但正文只有limit=10，没有items、total或记录。不能把它当作完整历史列表成功。 | 服务端是否按主体、默认参数或响应解码丢字段未定；没有猜时间参数。；来源 boundary-interface-20261102:sql-history-list-contract |
| audit-observation | sql-history-time-range | regular/0.1.8 | 未逐项确认；记录日2026-11-02 | cli_missing | 自己statement也400，正式CLI缺时间输入。 | 不能猜时间或ID绕过。；来源 history-0.3.20:SQL历史详情与profile、导入失败文件列表 |
| audit-observation | sql-query-detail-profile-known-statement | regular/0.1.8 | 2026-11-02 | failed | query-detail和query-profile都exit1，ErrParamInvalid/status400；没有详情或profile结果。 | 该statement来自同一批隔离只读查询；错误根因、是否需要其它标识或权限仍未知；不称后端功能整体不存在。；来源 boundary-interface-20261102:sql-query-detail-profile-known-statement |
| member-lifecycle | test-member-mutations | candidate/0.1.0-dev.d5c5c3a | 未逐项确认；记录日2026-11-14 | blocked | 生产只有当前超级管理员，没有可安全执行邀请/启停的普通测试成员 | 只读当前主体不能代替成员生命周期验收。；来源 dev-full-regression-20261114:calls/members |
| member-lifecycle | test-member-mutations | regular/0.1.8 | 未逐项确认；记录日2026-11-02 | blocked | 缺具体隔离测试成员，未执行。 | 只读当前主体不能代替成员生命周期验收。；来源 history-0.3.20:实际成员邀请/绑定/禁用 |
| iam-authorization | full-permission-view-decode | regular/0.1.8 | 2026-11-02 | failed | CLI未知响应字段导致解码失败。 | 可读data面独立核验不修复完整视图。；来源 history-0.3.20:权限管理 |
| iam-authorization | inheritance-canonical-rule | regular/0.1.8 | 2026-11-02 | partial | 未形成正向完整business合同闭环。 | 首批canonical-only/keep拒绝不判通用规则产品故障；clear漂移另列failed。；来源 history-0.3.20:权限管理 |
| iam-authorization | ordinary-subject-data-enforcement | candidate/0.1.0-dev.d5c5c3a | 未逐项确认；记录日2026-11-14 | blocked | 没有普通测试成员，未伪造allow/deny身份 | admin查询不证明Role授权。；来源 dev-full-regression-20261114:calls/members |
| iam-authorization | ordinary-subject-data-enforcement | regular/0.1.8 | 未逐项确认；记录日2026-11-02 | blocked | 没有切换/绑定专用测试成员。 | admin查询不证明Role授权。；来源 history-0.3.20:权限管理 |
| iam-authorization | revoke-clear-permission-drift | regular/0.1.8 | 2026-11-02 | failed | 返回500，数据版本未变但独立数据权限回读missing_in_mo。 | 权限漂移实际失败，不能称安全回滚；内部根因未知。；来源 history-0.3.20:权限管理 |
| agent-conversation | team-automatic-history | candidate/0.1.0-dev.d5c5c3a | 2026-11-15 | failed | 第一轮记住437和第二轮询问加11的消息均committed/completed，但第二轮回答没有更早的一轮，未回答448；既有运行时核对显示第二轮provider session_turn仍为1。 | 显式携带最少已确认事实可完成后续计算，但只是阶段账本绕行，不能改记自动历史继承通过。；没有证据表明调用者还需补history、provider session或其他CLI字段；精确根因需产品owner沿共享会话保存、冻结派发和provider历史加载排查。；来源 dev-full-regression-20261114:team/receipts/message-history.json; team/receipts/task-history.json, skill-044-agent-governance:team/receipts/messages-history-01.json, skill-048-agent:../skill-044-agent-governance/team/receipts/messages-history-01.json + ../cli-dev-full-regression-20261114/agent-governance/production/team/receipts/message-history.json + installed Skill configure-agent-team-ingress.md |
| agent-conversation | team-automatic-history | regular/0.1.8 | 未逐项确认；记录日2026-11-02 | failed | 两轮completed但第二轮runtime turn为1，自动历史未继承。 | 显式前文成功是另一条件，不修复自动历史。；来源 history-0.3.20:Team创建/会话/单轮/消息幂等 |
| agent-conversation | team-changed-content-same-key | candidate/0.1.0-dev.d5c5c3a | 2026-11-15 | failed | 同键异文返回原task且消息总数保持2，新意图未执行；精确重放单独保持幂等。 | 不能把精确重放幂等当内容冲突保护。；来源 dev-full-regression-20261114:team/receipts/message-changed-same-key.json; team/receipts/messages-after-replay.json, skill-044-agent-governance:team/receipts/message-changed-same-key.json + team/receipts/messages-after-replay.json |
| agent-conversation | team-changed-content-same-key | regular/0.1.8 | 2026-11-02 | failed | 静默返回旧task、原消息与回合保持；新意图未执行。 | 不能把精确重放幂等当内容冲突保护。；来源 history-0.3.20:Team消息重放 |
| agent-conversation | artifact-batch-delete-state-sync | candidate/0.1.0-dev.d5c5c3a | 2026-11-17 | partial | 两个成果删除成功，不存在对象not_found跳过；成果和placement归零，物理文件元数据仍可读取。 | 普通成员/无权限成果缺专用测试成员；成果/placement删除不等于物理文件擦除。；来源 cli-143-pending6-20261117:calls/ar08-batch-delete + calls/ar08-artifacts-final + calls/ar08-files-after-delete + calls/ar08-file-info-a-after |
| agent-package-resource | go-large-export | regular/0.1.8 | 2026-11-01 | failed | 仅落盘11886439后超时，partial不是可安装包。 | Python完整导出或Go真实执行不覆盖此失败。；来源 history-0.3.20:Go大包导出 |
| agent-skill-tool | oauth-external-events | candidate/0.1.0-dev.d5c5c3a | 未逐项确认；记录日2026-11-14 | blocked | 没有专用OAuth/外部渠道测试条件 | 无账号MCP成功不外推OAuth/消息事件。；来源 dev-full-regression-20261114:calls/plugin-installations |
| agent-skill-tool | oauth-external-events | regular/0.1.8 | 未逐项确认；记录日2026-11-02 | blocked | 没有专用安全账号与事件输入。 | 无账号MCP成功不外推OAuth/消息事件。；来源 history-0.3.20:插件能力发现、无账号MCP调用与收尾 |
| agent-skill-tool | plugin-lifecycle-pl02 | candidate/0.1.0-dev.d5c5c3a | 2026-11-14 | failed | 重复安装由旧503变为409 PLUGIN_STATE_CONFLICT，仍未证明可判定的幂等重复语义 | 不是并发、丢包或未知结果重试；没有重提该失败请求。；未取得服务端日志，根因未知；不能凭reason认定外部Context7故障、账号缺失或幂等成功。；来源 dev-full-regression-20261114:calls/duplicate-install/response.json; calls/installations-after-duplicate-read/response.json |
| agent-skill-tool | plugin-lifecycle-pl02 | regular/0.1.8 | 2026-11-02 | failed | 已成功安装后，原不可变正文顺序再次提交返回503/code15/PLUGIN_STORE_UNAVAILABLE；随后列表只有本轮一份Context7，ID与v1不变。未新增重复安装，但没有给出明确可判定的重复安装语义。 | 不是并发、丢包或未知结果重试；没有重提该失败请求。；未取得服务端日志，根因未知；不能凭reason认定外部Context7故障、账号缺失或幂等成功。；来源 complex-plugin-lifecycle:PL02 |
| agent-skill-tool | plugin-skill-body-read | regular/0.1.8 | 未逐项确认；记录日2026-11-02 | failed | 仍404。 | Context7工具调用成功不覆盖Skill正文读取失败。；来源 history-0.3.20:插件能力发现、无账号MCP调用与收尾 |
| agent-skill-tool | same-id-attachment-update | candidate/0.1.0-dev.d5c5c3a | 2026-11-15 | cli_missing | dev CLI没有update-skill入口，import-skill-package不接受目标Skill ID；通用resource-put只有开放JSON，没有附件上传合同。当前公开API仅提供PATCH元数据和GET附件，ImportSkillPackage固定Version=1并CreateWithFiles，不是同ID附件更新。 | 独立导入新包再切换Agent绑定是替代流程，不是同ID原位更新。；不能用未验证的resource-put/resource-patch猜测附件字段或发送覆盖请求。；来源 dev-full-regression-20261114:agent-update-skill-help.txt, skill-044-agent-governance:offline/agent-help.txt + offline/import-skill-package-help.txt + offline/resource-put-help.txt, skill-048-agent:offline/agent-help.txt + offline/import-skill-package-help.txt + offline/resource-put-help.txt + matrixflow-fork/moi/pkg/service/agentresource/skill_metadata.go + matrixflow-fork/moi/pkg/api/server.go |
| agent-skill-tool | same-id-attachment-update | regular/0.1.8 | 未逐项确认；记录日2026-11-02 | cli_missing | 正式CLI未提供输入。 | 独立新包切换不是原位更新。；来源 history-0.3.20:带附件托管Skill |
| agent-runtime | agent-provider-backend-empty-directory | candidate/0.1.0-dev.d5c5c3a | 2026-11-14 | partial | 后端目录为空且Router可读，仍无真实自有backend做端到端验证 | 空backend目录不能解释为模型不可用，也不能证明唯一默认路由；系统模型目录和实际通道调用需独立核验。；来源 dev-full-regression-20261114:calls/provider-backends/response.json |
| agent-runtime | agent-provider-backend-empty-directory | regular/0.1.8 | 2026-11-02 | partial | provider-backend list正常退出但返回空对象，未得到可供后续get或endpoint查询的backend ID。 | 空backend目录不能解释为模型不可用，也不能证明唯一默认路由；系统模型目录和实际通道调用需独立核验。；来源 boundary-interface-20261102:agent-provider-backend-empty-directory |
| agent-runtime | model-channel-check | regular/0.1.8 | 未逐项确认；记录日2026-11-02 | partial | 部分模型/通道成功，但Kimi knowledge_qa503、省略service-slot500。 | 模型目录不证明唯一默认路由；不改聚合默认配置。；来源 history-0.3.20:直接模型调用 |
| agent-runtime | runtime-authoring-inputs | candidate/0.1.0-dev.d5c5c3a | 2026-11-15 | partial | Runner 管理生命周期和 Sandbox 配额保持可用，但 Sandbox create 返回500，未验证终端、网络、配额扩容或负载。 | 未创建 Sandbox，未连接终端，未验证网络、配额扩容、真实负载或自有 backend。；Connector 配置生命周期通过不代替外部连接或 Sandbox 执行环境验收。；来源 skill-047-runtime-connector:calls/sandbox-create + calls/sandbox-quota-preserve + calls/sandbox-quota-after-preserve |
| agent-runtime | runtime-authoring-inputs | regular/0.1.8 | 未逐项确认；记录日2026-11-02 | cli_missing | 存在正式CLI输入/作用域缺口。 | 健康探测失败不授权改路由或切backend。；来源 history-0.3.20:Runtime策略配置生命周期 |
| agent-runtime | sandbox-create-small-limited | candidate/0.1.0-dev.d5c5c3a | 2026-11-15 | failed | sandbox create 返回 SANDBOX_ERROR/500，request_id=84949a3b-d874-4ac1-a9ee-b5378f9a3131；list 为空，get 为404。 | 未重复创建；公开错误不足以区分执行后端、资源或其它内部原因。；没有实例、终端、网络连接或计算负载，不证明 Sandbox 创建可用。；来源 skill-047-runtime-connector:calls/sandbox-create + calls/sandbox-quota-after-preserve |
| agent-runtime | model-type-filter-contract | candidate/0.1.0-dev.d5c5c3a | 2026-11-17 | partial | chat返回11个、embedding返回3个、非法类型400；ocr/vision/reasoning均把同一批11个模型重标为请求类型，无法据此证明真实类型支持。 | 需要正式十类枚举、真实模型和调试结果；页面页签由UI验收，重标行为需服务owner确认。；来源 cli-143-pending6-20261117:calls/dg10-type-chat + calls/dg10-type-embedding + calls/dg10-type-ocr + calls/dg10-type-vision + calls/dg10-type-reasoning + calls/dg10-type-invalid |
| agent-automation | agent-callback-rules-empty-list | regular/0.1.8 | 2026-11-02 | partial | list-callback-rules正常退出并返回items=[]、total=0、limit=50、offset=0，当前工作区没有可供继续读取的callback rule。 | 空规则列表只证明当前查询范围没有规则；不证明OAuth、外部事件、渠道连接或回调投递已通过。；来源 boundary-interface-20261102:agent-callback-rules-empty-list |
| agent-automation | agent-channel-list-requires-id | regular/0.1.8 | 2026-11-02 | partial | 省略provider返回ErrIDRequired；按wecom、slack、feishu分别查询均exit0，返回items=[]、total=0、limit=50、offset=0。正式列表能力存在，但无provider时不能猜默认目标。 | 本工作区这三个provider当前均无实例，因此只证明筛选列表读取和空列表语义；没有外部事件或OAuth验收。；来源 boundary-interface-20261102:agent-channel-list-requires-id |
| agent-automation | automation-revisions-ar-04 | candidate/0.1.0-dev.d5c5c3a | 2026-11-15 | partial | v2 指令、cron、active和next_trigger_at正确，但200秒内无v2 run且当前无晚到记录；独立替换任务 cron succeeded。 | Do not generalize to every revised task; do not reactivate the archived object; replacement does not repair revision scheduling.；来源 skill-045-agent-automation:state.json + receipts/revision-confirm.json + receipts/runs-final.json, skill-048-workflow-data:calls/cron-revision-task-current + calls/cron-revision-runs-current + calls/cron-replacement-run-current + ../skill-045-agent-automation/receipts/activate-v2.json + ../skill-045-agent-automation/receipts/disable-final.json |
| agent-automation | automation-revisions-ar-04 | regular/0.1.8 | 2026-11-02 | partial | v2启用返回next_trigger_at=06:52:00Z，120秒窗口23次runs查询仍只有v1；06:52:31Z停用。后续原task/runs及归档后延迟查询仍无第二run，新版答案与运行归属未完成。 | 不是已确认永久不触发：到due后约31秒已停用，后续在disabled/archived条件下只读；服务内部原因未知，不再启用或手动替代。；来源 complex-automation-revisions:AR-04 |
| agent-automation | channel-events | candidate/0.1.0-dev.d5c5c3a | 未逐项确认；记录日2026-11-14 | blocked | 没有真实外部callback/channel/OAuth事件条件 | 只读规则或连接存在不证明事件成功。；来源 dev-full-regression-20261114:calls/callback-rules/response.json |
| agent-automation | channel-events | regular/0.1.8 | 未逐项确认；记录日2026-11-02 | unverified | 没有本轮真实事件验收。 | 只读规则或连接存在不证明事件成功。；来源 history-0.3.20:OAuth/渠道事件仍缺测试条件 |
| agent-automation | dynamic-api-schema-payload | candidate/0.1.0-dev.d5c5c3a | 2026-11-15 | failed | PAT认证和提交前schema检查均已通过，run确实包含payload.a=19和payload.b=23；但后续Workflow把automation_task_id、trigger_context、payload等整个外层对象按业务a/b schema校验，最终WORKFLOW_SUBMIT_FAILED/INPUT_VALIDATION_FAILED。 | 不是缺PAT、缺a/b或CLI字段错误；不能通过放宽业务schema、固定指令或重复提交冒充修复。；缺b的400预期拒绝是另一项已通过能力，不覆盖合法输入失败。；来源 skill-047-automation-api:schema-result.json, skill-048-agent:../skill-047-automation-api/schema-result.json + ../skill-047-automation-api/bodies/codex-skill047-schema-r3-create.json + ../skill-047-automation-api/bodies/codex-skill047-schema-r3-invoke.json |
| agent-automation | dynamic-api-schema-payload | regular/0.1.8 | 未逐项确认；记录日2026-11-02 | failed | 500并产生failed run，提交时schema包络错位。 | 缺字段400正确不证明合法输入执行成功。；来源 history-0.3.20:API输入校验和业务参数/JSON输出 |
| agent-automation | dynamic-api-structured-output | candidate/0.1.0-dev.d5c5c3a | 2026-11-15 | failed | 动态payload已保存，但模型回答未收到a/b，structured_output为空，最终OUTPUT_CONTRACT_VALIDATION_FAILED。当前源码的conversationTurnInput在没有显式message_parts时只把instruction_text放入模型消息，没有投影command.Payload，与生产现象一致。 | 本项失败可由动态参数未进入模型解释；现有证据不单独证明严格JSON输出校验器本身有独立缺陷。；空output_contract或固定指令返回527不能替代动态参数加严格JSON输出验收。；来源 skill-047-automation-api:structured-result.json, skill-048-agent:../skill-047-automation-api/structured-result.json + matrixflow-fork/moi/pkg/service/agentresource/automation_runtime_session_input.go + matrixflow-fork/moi/pkg/service/agentresource/automation_runtime_adapter.go |
| agent-automation | dynamic-api-structured-output | regular/0.1.8 | 未逐项确认；记录日2026-11-02 | failed | 未得到参数业务结果，structured_output为空/输出校验失败。 | 空output_contract passed不能覆盖严格JSON失败。；来源 history-0.3.20:API输入校验和业务参数/JSON输出 |
<!-- capability-status:end -->

## 2026-11-02 新增深测（0.3.22）

这轮在已确认的“测试”工作区使用正式0.1.8补验23个限定场景：18通过、4失败、1部分完成。通过数包含按预期拒绝坏输入的负例，不表示所有流程执行都成功；也不与历史功能单元相加计算覆盖率。逐项条件与来源以生成状态和唯一JSON为准。

- Catalog：BOM/CRLF/中文换行、纯表头零行导入及下载、显式列改序/多余列、缺映射列的上传拒绝已验。零字节上传异常仍为partial，无任务ID，未重传。
- Workflow：带类型左连接、报表筛选/排序/空集合，以及重复键和null失败后的上游结果留存已验；未据节点内部处理宣称通用if/foreach。另一个deploy500没有ID，对象是否创建仍未知，有限列表不能排除残留。
- Knowledge：两类语义条目跨库保内容换ID，重复/同key冲突和非法kind混批拒绝已验。create/list/get响应解码失败保留failed；正式builder资源发现与export能找回原对象继续管理，不能把它说成原命令已修复。未新增问答或引用覆盖。
- SQL：大整数/DECIMAL/微秒/前导0/多行文字逐格保持；CSV丢失NULL与空串区别保留failed。内联常量SQL405与另一次异步成功分别保留；真实只读任务接续不重放，原失败步骤仍未完成。

本轮可定位的测试资源已逐ID清理核对；Catalog文件引用归零后元数据仍可读，不承诺物理擦除。上述无ID部署500及拒绝上传的潜在暂存未知状态均未隐去。没有修改CLI、后端或提交PR。

## 历史实测记录（保留0.3.20正文）

以下原正文按字节保留，作为历史来源，不是另一份独立维护的当前状态表。后续复测在 JSON 新增观测，
只有同场景、条件、渠道的显式 `supersedes` 才能替代；替代记录及旧失败仍可通过 `query --history` 读取。
原文中的数值只属于当时的隔离样本，不外推当前对象、其它场景或任意规模。

# 实测范围与保留缺口

最近补验2026-11-02（保留2026-10-30至11-01结果），常规调用 moi-cli 0.1.8；四项新增输入（默认值、Catalog文件筛选分页、知识分段分页、整篇治理）由隔离源码构建（自报0.1.0）验证、未替换全局CLI，生产环境用户选定的测试工作区；只使用隔离虚构数据。
这是一次环境观察，不是长期可用性保证，也不把21张能力卡当作21项已通过功能。

| 业务结果 | 本次判断 | 边界 |
| --- | --- | --- |
| CSV新建/追加/覆盖与XLSX新建入表 | 通过限定格式 | CSV追加4→6行，覆盖只剩替换1行；XLSX两行17/29合计46，再append同文件变四行92，证明该append不自动内容去重；其它格式/策略另验 |
| 表/SQL/看板CSV下载 | 通过 | 原始字段、行和汇总已比对；不是Connector写外部系统 |
| 多节点Workflow与报告交付 | 通过 | 读表→区域聚合→判断→Markdown→下载；6笔1500，对照规则1200，超出300 |
| 知识规则与流程组合 | 通过限定场景 | Codex先通过A2A检索引用，再编入流程；不是运行时动态RAG |
| 同App正式候选更新并运行 | 通过限定路径 | 真实Workflow Agent候选→验收→deploy-candidate指定原App→新版448，旧版437历史保留；旧手工create/publish指针问题仍未修，get仍有session_id解码错误 |
| Workflow暂停/恢复/取消 | 通过限定场景 | paused保持后沿原execution恢复完成；triggered/running/paused取消均回读cancelled；不证明副作用回滚或trace全部收尾 |
| 受控失败后retry | 通过限定场景 | 整次新执行成功、原失败保留；同部署版本；不证明副作用节点安全重放 |
| 知识资料追加/跨库复用/表来源 | 通过限定场景 | 文件引用、表真实工具查询和来源已核对；同一原文件建两隔离库，删A源后B分段/ETag不变、新会话正式引用B，原文件hash保持；并发替换/物理向量清除未验 |
| 文档分段新增/编辑/停用/启用/删除 | 通过 | ETag冲突保护、独立回读、新会话引用随状态改变；2026-10-31隔离样例库存分段46→45→46，三次独立问答分别27箱/3次、28箱/4次、27箱/3次，均有两份正式引用。原分段ID/正文/启用状态已恢复；不代表原文件替换重入库或向量单路刷新 |
| 整份文档启停/过期与大文档分段分页 | 常规CLI缺输入；候选通过限定样例 | 未发布候选已验分段两页合并/ETag、整篇停用/过期的新会话拒答及恢复引用；常规0.1.8仍无这些flags，清空tags另有SDK/null合同缺口 |
| 图片索引、实际看图和图片引用 | 通过 | 一次真实准备、相似图检索、img_parse实际看图、正确答案和visual_hit第1页引用、独立下载核图；带volume_id读取失败仍保留、准确根因未定 |
| 看板运行/手动刷新/汇总与钻取下载 | 通过限定场景 | 汇总四行共1500；正确x/y配置及role:x后east明细两行400/200，CSV一致；布局/颜色配置独立回读。UI外观和告警未验，定时刷新另列 |
| 分享退订/删除、写保护与直接撤回 | 部分通过，直接撤回失败 | 先退订再删除通过；隔离只读分享的INSERT、UPDATE、DELETE均被拒，源目标仍六行合计1500；直接删活跃发布503，两侧关系消失但目标仍能读；隔离残留已清理，原分享不变 |
| Agent工具读表、发布文件、下载 | 通过 | 工具实读、原task completed、正式file及字节/hash一致；流超时由同task回读解决 |
| Team创建/会话/单轮/消息幂等 | 部分通过 | 创建、单轮和幂等通过；同 Team/context/member 两轮均 completed，但第二轮 runtime turn仍为1，自动历史不继承。显式带前文可用；根因复核指向 completion保存→ingress冻结→provider请求续传链路，不是CLI缺参数 |
| 自动化配置/归档、Cron真实触发 | 通过 | trigger_type=cron、结果437、最终仅一次运行；测试任务已禁用归档 |
| 自动化API固定指令触发 | 通过 | 用户PAT正式调用、trigger_type=api、结果527、一次运行并归档；native/runtime-proof限制仍在 |
| API输入校验和业务参数/JSON输出 | 部分通过，业务执行未通过 | 缺字段400且零run正确；合法输入仍在Workflow提交时schema错位失败。独立无input_schema任务也未得到业务参数结果，严格输出验证失败；不是再登录能解决 |
| 模型目录、路由和通道前置 | 只读通过，实际调用按通道区分 | 当前可发现12个系统模型；DeepSeek按用户策略优先且默认不传等级。模型列表不能单独证明唯一默认路由；LLM/embedding/parser router均为ROUND_ROBIN、30秒健康检查、最多重试2次；`executor-health` 返回 `astra_reachable=true`；backend列表为空不等于系统模型不可用；endpoint列表省略backend ID时暴露CLI/服务输入合同缺口 |
| 直接模型调用 | 通过限定场景 | `agent_chat` 下 DeepSeek、GLM、Kimi 的默认和 `max` 已通过；`knowledge_qa` 下 DeepSeek、GLM通过，Kimi当前503；省略 `service-slot` 三者均500。其它已发现模型只读列出，未据此宣称可调用 |
| 计算资源创建/查看/预检/删除 | 通过限定场景，负载仍未验收 | 早期IDLE样例88秒收尾；本轮复用已绑定候选导致部署冲突，未提交负载，不能归计算故障。资源406秒后删除，超5分钟授权106秒，费用未知；无活资源残留 |
| IAM临时Role与数据权限 | 配置/回读通过 | 生命周期/CAS/删除通过；SELECT补row_column.mode=keep后写入和独立数据权限回读通过；未验普通成员实际allow/deny |
| 实际成员邀请/绑定/禁用 | 未测 | 当前主体和权限只读已查；需具体测试成员，不能用真实成员凑测试 |
| Agent包、声明式Skill和Python可执行工具 | Python与Go纯计算限定通过 | 包内Skill实际读规则；标准库及vendor packaging 26.2工具实际执行，wrapper exit0、独立结果、实际vendor import路径、同task completed；此处导出字节一致仅指Python包，本轮Go大包导出超时不完整。空py.typed加载400与仅加换行后成功的对照保留。两个近似Go-only包一次load成功、一次`ErrTimeout`且未生成根资源，本轮另有Go真实工具执行通过，但加载稳定性未证明；C++/Rust缺Docker daemon，网络及原生依赖另验 |
| 带附件托管Skill | 通过限定场景 | 无脚本包及标准库Python脚本附件均已实际使用；脚本下载、真实执行及独立输出核对、同task completed、清理404。instruction运行中v1→v2更新后原task仍下载冻结v1；本轮独立附件包A/B切换时，新旧任务并行、分别运行冻结脚本1221/2333；同ID原位附件替换仍无CLI输入 |
| 插件能力发现、无账号MCP调用与收尾 | 通过限定场景 | 正式引用和system namespace补齐后描述/发布详情成功；Context7正式安装→实际query-docs有来源返回→同task completed→删除安装回读404，原两个插件不变；GitHubSkill正文正确引用/路径仍404，OAuth/渠道事件另验 |
| SQL工作簿与版本 | 通过 | 创建/草稿更新/保存/改名/回读已保存版不变/执行527/删除404 |
| 多文件坏数据后恢复 | 通过限定场景 | 原批两文件均失败、COUNT确认零行；修正输入新任务写入三行合计71；不是部分成功或原失败文件重试通过 |
| 多表部分成功后原任务定向恢复 | 数据恢复通过，元数据异常保留 | 每表一文件，首次A成功2行56、B主键冲突；清理隔离预置冲突后，只retry B的LoadTaskFile记录ID，新execution完成、B1行23、A仍2行56、旧失败保留，无重传。恢复后仍报TARGET_PATH_RELATION_INVALID；累计行数仅补跑1行，不能算全部3行。2026-11-01再复核同表样例：首批冲突后SQL只有seed；只选conflict文件重试却写入good/conflict两行合计33，证实同表分组重放边界，任务历史失败计数仍保留。其它格式/策略另验 |
| SQL历史详情与profile、导入失败文件列表 | SQL历史仍缺输入；导入定位已有正式入口 | 历史缺start/end，自己statement也400；task-files缺task-id且500，但task-runs→workflow list-run-import-files实测可得准确LoadTaskFile ID。原“无法定位失败文件”的判断已纠正；不混用其它ID |
| 操作日志、Agent只读状态与workspace file下载 | 只读通过 | audit overview/logs可读；默认页20条并含request ID，callback rules为空、executor health可达；真实workspace file下载PDF大小/MD5一致。runtime artifact历史上传/发布/下载已核实，与上文文件交付同一流程，不重复算新能力；外部Ingress事件未验 |
| Catalog→Workflow→Knowledge 新批处理与问答 | 两个虚构文件限定样例通过 | 2026-10-31：已知非空 Volume 仅提交一次，list-runs 按 dispatch_job_id 关联两条本轮 completed 子 execution；按本轮 case 匹配解析/索引/ZIP 血缘。两份 ZIP 下载、完整性及原文件 metadata/关键正文核验通过；KB 仍为两个原来源，当前分段可读；批处理前后独立问答均取得正确联合计算和两份正式引用，资料未提供的电话回答无法确定。run 把分发 ID 放进 execution_id，直接 get-run 返回404的合同问题仍在；catalog files 缺 Volume 筛选输入。后续本日第二轮已验证失败文件定向重试、新file来源替代及Automation内置Workflow动态检索，限定范围见下三项。 |
| Volume部分成功后仅重试失败文件 | 通过限定场景 | 原批1成功1失败，显式选择失败execution，新dispatch accepted_count=1并completed，原成功未重跑、原失败保留。同已完成request-id一次顺序及两次并发重放均返回原dispatch/replayed=true，执行集合仍为原3条。2026-11-01同一全新键双请求同时首发也只新增一个目标失败文件execution，两次延后列表均3→4，同版本completed且case_result正确；不推广并发规模或业务副作用 |
| 新版原文件替代旧Knowledge来源 | 通过新file身份路径 | 同名原文件31→49箱，新source准备/分段/血缘成功后移除旧source，来源3→2；旧source404，新会话39箱/4次且正式ref只用新版与规则；旧独有事实无法证实，旧Catalog原文件下载字节/哈希不变。不是同file ID覆盖或向量物理擦除 |
| Workflow运行时动态知识检索 | 通过Automation限定路径 | 正式已发布Agent包1.0.0 + astra/default + KB绑定；同automation版本/指令/包/空Payload，两轮实际search_rag_chunks→select_final_sources→knowledge.answer，答案57/5→39/4，正式refs切换新版。初次default provider在模型首轮零输出失败保留；任意手工DSL直接RAG、隐藏Workflow/execution ID查询仍缺正式入口 |

| SQL写入、结构、计划和单请求事务 | 通过限定样例 | 建库/表、INSERT/UPDATE/DELETE、ALTER、视图和EXPLAIN已实际核对；单请求ROLLBACK后2行135，COMMIT后3行146；隔离数据库删除后独立查为0。跨CLI连接事务/全部SQL方言未验 |
| Workflow模板往返 | 模板分支通过 | 模板创建→更新→独立读取→原DSL导出文件→deploy导入→真实输出48；隔离App/模板删除404。普通Workflow调度/API authoring仍缺常规CLI输入 |
| Workflow SQL副作用重试 | 部分通过 | 写入后受控失败，表仍1行77；同业务键upsert重试新execution完成且仍1行77，原failed和同版本保留。不是自动回滚或任意副作用防重 |
| 看板定时刷新 | 通过限定样例 | 5m配置后首次立即执行不计作周期验收，后续11:19 UTC不同cron execution完成；Workflow和chart-result同execution/statement，四区域600/100/250/550合计1500；改manual撤调度并删除404。失败重试/并发刷新、UI和告警未验 |
| Agent多版本默认切换与回滚 | 部分通过 | 同ID两版本runnable并存，load不自动切默认；显式default切v2后新会话及旧context新轮实际manifest=1.1.0并输出v2，回滚v1新会话实际用1.0.0；历史已完成任务保持原版。根Agent删后404但默认v1/lineage元数据仍留；本轮补在途任务切默认与绑定后的并行运行/冻结，两个不同版本并发load及三版实际工具已限定通过；同版本相同包及不同内容双请求的限定补验见下文，胜包有完整导出及真实工具证据；并发default仍未验 |
| Runtime策略配置生命周期 | 仅配置通过 | 本次隔离draft→元数据更新version2→archived/version3、独立回读；从未激活/绑定，不证明网络/预算实际执行。Sandbox/Runner根作用域及后端配置仍缺CLI入口/输入 |

| SQL NULL/空串/0与空结果对照 | NULL及内联空结果限定通过；异步取数仍失败 | 2026-11-01一行三列能区分NULL、空串、0；零行SELECT状态success但取原statement结果500 empty query result、下载也500；正式sql run另起新查询能返回成功列结构且result/total零值缺省，非空对照正常。不是恢复原查询或500转空数组 |
| 双库联合引用及单库范围对照 | 通过限定场景 | 两个新会话，双库四文档分别核算27箱/3次/蓝槐仓与39箱/4次/青禾仓，四份model/source/file/chunk正式引用匹配。只绑定前库询问后库独有事实时拒答，检索产物与引用只在前库；不代表普通成员权限或全部检索分支通过 |

| Catalog表发现、批量结构、预览与下载一致性 | 小表限定通过 | 2026-11-01：数据库子对象与table-leaves表ID一致；目标表3列6行，table-info/data/preview和CSV逐行核对，总额1500；multi-table-info包含目标。默认窗口不外推大表完整读取；volume-full-path本次仅名，不当完整父路径 |
| Catalog同文件两处挂载后移除单处引用 | 通过限定场景 | 两个Volume各1文件，file list精确file-ids返回同物理ID的两处placement；delete-file-ref只移除A后A0/B1，B原文件92字节/SHA256不变。Volume改名后ID/文件数不变；隔离文件placement、两个Volume和数据库已清理。未证明物理擦除或KB自动更新；空文件夹只验创建/删除回执，嵌套枚举未验 |
| Workflow节点级继续与合同保护 | 只读及限定写入节点通过 | get-run-node与plan取得真实node execution/hash；continue_from_node生成独立rerun/branch/new execution，succeeded/completed，输出与原结果相同，trace仅check_gate/end；原执行与版本保留。错误hash返回409且执行集合不变；本轮另验普通写节点重放2行34、业务主键保护1行17、写后失败从写后继续1行33；overrides、任意副作用和输入修改未验 |
| 知识来源分页和历史候选分类 | 观察通过，范围有差异 | 一个库两页各一来源/末页为空；另一库正式来源2个，分页3项含legacy_unbound旧文件候选，无source_id；check-source-existence只返回正式新版。未backfill/reconcile。不能拿分页total承诺当前可检索来源数 |
| Knowledge语义业务规则条目往返 | 管理迁移限定通过 | rag/business_rule新增→修改8为9→独立回读→export提取条目→另一库import→再export内容一致、ID不同；删原条目不影响目标。错误domain纠正后通过，不归因登录；四个隔离库及Catalog数据库回读404。不是整库备份/文档向量迁移；本轮另有rag/business_rule真实消费8→13箱、6→4车对照通过，正式文件引用不代表规则引用 |
| Workspace file token分页 | 两页限定通过 | 原next_page_token用于下一页，两页各2条得到4处不同placement；已知file-ids筛选能发现同文件多处挂载。不是任意Volume枚举，不将placement总数当物理文件数 |

实际执行时重新发现对象和状态；这里的数值只是验收样例，不能复制成用户业务输入。历史失败保留，
只有同目标、同范围的复测成功才改变判断。命令存在、Schema通过、配置保存和模型口述均不构成业务通过。

仍需补齐的主要分支：外部Connector真实导入/导出、普通测试成员allow/deny、有负载计算启停与资源账单、
SQL历史所需时间范围、同一配置多文件恢复的范围隔离（本轮重试仍failed）、更多主键/格式/冲突策略、任意手写DSL内直接动态知识检索、同file ID原位覆盖、
看板UI外观/告警、Go线上稳定加载与大包完整导出、C++/Rust构建执行、原生依赖/网络工具、Agent更大规模同版本竞争/并发default切换、同ID Skill附件原位更新、OAuth及渠道/Ingress真实事件。在途任务版本与独立A/B附件绑定冻结、Go纯计算执行已有限定实测，不再列为完全未验。
常规0.1.8的整篇治理/分页仍缺输入，本地候选已有成功证据但未发布。
外部Connector新建配置、模型后端/路由配置和Sandbox/Runner网络管理有真实CLI输入或作用域缺口，不能全归为缺用户信息。
已复现失败的API业务输入/严格输出、Team自动历史、活跃分享直接撤回、图片携volume读取、异步空SQL取数/下载和插件Skill正文读取，
不能写成“只缺用户信息”。具体限制按对应Recipe；显式带前文只证明显式上下文可用，不代表Team自动记忆已修复。

## 2026-11-01 后续深入补测

| 场景 | 判定 | 限定结论与保留问题 |
| --- | --- | --- |
| 同表已有数据、同批部分成功后定向恢复 | 通过限定配置 | 每文件独立table-config指向同表，只retry失败文件；baseline/good/bad最终3行38，成功数据未重复；单配置多文件仍可能整组重放，task路径异常保留 |
| CSV字段/NULL、分号及真实XLS | 通过限定文件 | 复杂字段/换行/中文正确，显式反斜杠N为NULL、空串保留；分号两行正确，真实XLS两行46；不外推全部格式策略 |
| Go工具真实执行 | 通过限定纯计算 | load超时后原根和版本已创建；同task真实Go main.go执行exit0、sum50及SHA256独立一致；加载稳定性仍有限 |
| Go大包导出 | 实际失败 | 22,043,039字节仅落盘11,886,439，超时.partial不是可安装包，不抹去历史Python完整导出成功 |
| 业务规则实际采用及修订 | 通过限定rag规则 | 文档与ETag不变，8→13箱规则使新会话6→4车并取新口令；真实semantic_entries证明消费，规则不能冒充文档正式引用；写入原请求正文快照缺失已登记 |
| 冲突资料双来源 | 通过限定冲突 | 同批待复核19/23箱均正式引用，答案明确无法定案；其它冲突形式未全验 |
| 节点写入与失败后恢复 | 通过对照 | 普通追加确会重复；唯一键保护不重复；失败不回滚，从写后节点继续未重写；不承诺外发补偿或全局幂等 |
| 在途版本及附件绑定切换 | 通过独立包切换 | 旧新任务并行且各自冻结Agent/Skill版本、完整脚本和真实输出；default切换后需补绑定；同ID文件更新仍缺入口 |

以上主要是既有审计项的子场景深化，不把证据断言数当功能覆盖率。原导入失败文件定位与定向重试已有正式常规CLI替代链完成业务闭环；旧task-files入口缺口仍保留，但不能再等同该业务能力不通。审计状态应按同一业务目标证据复核，历史快照另保留。

## 同日进一步补验（Skill 0.3.15）

| 场景 | 判定 | 已验证及保留边界 |
| --- | --- | --- |
| 单配置多文件、已有数据时只retry失败文件 | 恢复未通过 | 只选bad而writer实际含good/bad；bad22落地但good主键冲突使execution failed。文件列表不是写入范围，已落地不自动回滚 |
| 表行冲突策略 | 两种限定通过 | conflict1保旧11、两行skipped、整表3行38；conflict2更新为17、仍3行44。均为明确新任务，不是修改原task，处理2行不是新增2行 |
| 文本文件写后恢复 | 对照通过 | 原失败仍留128B；写后节点继续保原file/placement；写节点继续新增同名不同file，2文件256B且完整下载一致。不外推取消/外发/任意文件格式 |
| 标准问答及files范围 | 通过限定rag样例 | standard_qa真实消费；A/B条目按真实file_ids检索不串入，正式refs与条目消费分开，不代表IAM隔离 |
| 删除语义条目后新会话 | 通过 | 文档/ETag不变，旧容量口令不再检索/回答，原文事实仍能正式引用。不证明物理擦除或旧会话记忆消失 |
| 有明确时效替代的冲突 | 通过限定文档 | 根据目标日期选生效修订02，11箱/车、57箱至少6车，双来源正式引用。无依据冲突仍不能定案 |
| 不同新版本并发load | 通过小Python包 | 两请求重叠1.21秒、两版runnable/default不变；三版真实工具49/89/129正确，旧任务冻结/回滚历史保持。不证明同版本竞争或并发default CAS |
| 已知嵌套文件夹清空 | 通过限定子树 | clean A保A空/删B，只去两处引用，root原文件36B下载一致；tree/children仍未枚举嵌套，不证明未知子树发现 |
| SQL内联合法零行 | 部分可用 | sql run成功返回列结构及零行缺省字段；原异步query-result和下载仍500，不能称修复或无说明重执行 |
| Go大包导出调整入口 | 缺输入，原失败保留 | 正式CLI无timeout/续传/压缩参数，本轮只核help/本地owner及旧partial证据，未重复下载，不声称新成功 |

以上是场景深化，不能按本表行数/断言数换算全产品覆盖率。107项审计目录中的M06-03因正式内联路径已限定可用而改partial，原失败路径保留；其余深测不虚增原通过数量。

## 同日边界补验（Skill 0.3.16）

| 方向 | 本次可用范围 | 未通过及保留边界 |
| --- | --- | --- |
| 导入 | 复合键保旧、单键字段子集更新、显式暂存加非主键SQL合并、NULL/小数/日期及文件内重复 | 复合键自动更新两次失败；IGNORE不吞数值越界。替代方法是新任务/SQL，不能称原路径修复 |
| 工作流取消 | 写后取消保文件，取消后resume按合同拒绝；SQL只读合同流程可从写后节点另建分支 | 含eval流程缺继续合同；不是所有取消任务均可续跑，无自动副作用回滚 |
| ZIP交付 | 小型两来源正文/源ID/CRC正确，显式aggregate侧文件完整 | per-document抽取metadata丢失；4ZIP精确删引用后file-info仍留元数据。无范围GC清理保证 |
| 知识变更 | 同段/异段同文档ETag并发冲突保护、新内容正式引用、两类条目原ID更新、唯一事实段停用后缺信息拒答 | 文档级版本保护不等于独立段无冲突；条目不是文件chunk，不证明物理擦除或任意并发 |
| Agent同版本 | 两个相同新包首发请求只新增一版，完成后同包/改内容重传均拒绝，真实工具/回滚历史核对通过 | 此同包样例本身不证明不同内容竞争，后续不同包双请求结果见下节；任意并发规模及default CAS未验，bindings来源投影不同须核根refs/source_ref/manifest |

本轮是原能力的场景深化，既有107项审计目录不因此增加功能总数或全产品覆盖率。预期拒绝和失败识别的证据核对通过，不代表失败业务可用；初次创建未知结果和清理元数据残留均须在任务报告中保留。

## 引用、来源与恢复重复提交的进一步边界

2026-11-01补验，正式CLI仍0.1.8；以下只对应隔离小样本，包含预期拒绝和重复副作用识别，不升级为通用幂等或全覆盖：

| 业务结果 | 本次观察 | 操作边界 |
| --- | --- | --- |
| Catalog同名与引用并发 | 两同名不同file全文各自正确；同file顺序/已有引用双请求/空目标双请求不增重复placement；改名保持ID/字节 | 并发移除同引用一成功一404，保留处不变；不是内容去重、导入防重或物理擦除 |
| 节点继续重复提交 | 同计划双请求创建两分支各写文件，顺序再次提交再增一份；142B变四份568B；写后SQL继续不增文件 | hash校验来源合同而非幂等键；暂停计划本例拒绝，错误/跨上下文hash拒绝；无key时先查原分支，不能盲重发 |
| Knowledge来源并发与移除 | 不同file追加无丢失；同file双追加只一source；删旧source后新会话只检索/引用剩余三源且旧独有内容拒答 | 重复append会重置pending，旧文档可读不等于ready；新旧共存不是原子替换，原Catalog文件仍可读 |
| Agent同版不同内容首发 | 双请求一成功一409，最终digest/导出/实际工具423均为胜包；默认显式切换、回滚和历史保持 | 不确定内部冲突分支；不保证固定赢家、任意并发或default CAS，也不保证零元数据残留 |

## 2026-11-02 横向深化

以下来自正式CLI0.1.8隔离小样本；没有改后端/CLI，没有把原107审计目录扩算成全产品功能总数：

| 分支 | 新观察 | 边界 |
| --- | --- | --- |
| SQL工作簿 | 跨workbook版本ID拒绝；draft编辑与finalize保存路径不同，双update/双save最终各保留一份；选定版本SQL真实执行正确 | finalize可被同版本save覆盖，不能当不可变；无已确认CAS，未证明通用并发合并或物理清历史 |
| 看板来源恢复 | 坏SQL配置拒绝并保旧；源表失效真实failed，重建同名表仍旧ID，正式重绑新ID后新执行正确、旧失败保持 | 不直接改托管Workflow，不归因登录；并发run各产生独立执行，live只代表所带execution |
| 看板空结果 | 合法零行执行completed，结果读取与CSV仍500；另有非零NULL/空串/0的明确标记对照 | 空结果交付验收失败，非零对照不修复原空结果 |
| Team版本及归属 | 配置/成员/session各自版本旧值拒绝；同workspace跨Team/context错配拒绝 | 顺序旧版本对照，不代表任意并发或普通成员跨账号IAM |
| Team消息重放 | 精确同键正文重放返回原task且无重复回合；同键异文也静默返回旧task、原文不变 | 显式内容冲突拒绝未通过；发送前本地绑定请求SHA与键，新意图未执行；自动历史原失败仍在 |
| 操作日志 | 已知虚构请求可用安全投影关联到审计时间/request_id，领域删除后404独立确认 | 默认20条无正式分页/精确历史输入，首屏没记录不表示没发生；仍只读/部分覆盖 |

## 2026-11-02 横向扩面与生命周期深化（Skill 0.3.19）

仍用正式CLI0.1.8；以下新增限定场景单列，原107项审计是历史目录，未全部重跑或重算全产品覆盖率。

| 分支 | 已验证的实际结果 | 未通过与边界 |
| --- | --- | --- |
| 数据分享 | selected源端增删改与接收端逐格一致；未选/新增表不可读；同名冲突保旧；退订旧入口失效，重订阅新ID含期间新增数据 | 未选表拒绝实际500，不叫正确404；update-publish缺范围输入；历史活跃发布直接删除503及残留未修复 |
| 权限管理 | 零成员Role独立应用策略写入/陈旧版本拒绝、禁用状态保护和隔离清理有证据 | capabilities/view未知响应字段解码失败；统一编辑/继承/set没有正向闭环。正确clear撤权仍500，数据版本未变但权限漂移missing_in_mo；是新失败，不是只缺信息或无副作用回滚 |
| Agent文件 | 桌面附件与任务成果区分；同basename多成果有各自ID/来源/完整内容；原token三页及筛选、精确删除和同名兄弟保留 | file_path不保sandbox目录；缺正式rename/update/session筛选；无placement附件删除需明确生命周期，隔离挂载清理不作自动生产fallback；已完成task保旧成果引用 |
| CSV与查询结果 | 1007行全量CSV对100/1000展示窗口，特殊字符/精确小数、源修改前后及删源后旧statement保持均核对 | 请求limit不是已验证严格行数上限，SQL LIMIT另核；CSV不区分NULL/空串；1007行不外推任意规模或无限历史保留 |

没有外部目标的Connector导出、普通成员allow/deny、OAuth/渠道事件仍缺测试条件。
本轮权限管理出现的兼容/写入问题和先前空结果、Team、防重、导入、图片等失败分别保留，不用历史通过统计覆盖当前风险。

## 2026-11-02 复杂横向补验（Skill 0.3.20）

以下是本批限定子场景，未把107项历史目录或三条纵向整套重跑，不累计成全产品覆盖率。IAM完整视图解码失败及正确撤权500/missing_in_mo等前批失败仍保留，不能用新场景成功覆盖。

| 方向 | 新增真实证据 | 保留限制 |
| --- | --- | --- |
| 多表业务对账 | 三表先选最新事件/分侧汇总，逐单金额与独立Decimal计算一致；晚到旧版、退款、缺明细和孤儿来源可辨；结果按键替换两次不增行；特定事务失败后余额保持 | 标量COUNT子查询一次405原因未知，UNION ALL成功是另一查询；不是跨表原子同步或任意事务保证 |
| 多图与钻取 | 两JOIN图汇总/CSV正确；单物理快照图可钻取，源DML后旧快照与新JOIN有真实对照，显式同步后正确；重排/删一图保兄弟及源数据 | JOIN直接钻取400仍保留；快照不会随整板refresh自动同步；UI未测 |
| 自动化修订 | v1真实cron成功；disabled修订v2正确，归档/旧历史保留、归档后启用被拒 | v2限定120秒未观察到新run，due后约31秒停用、延后仍无，partial且原因未知；没有手动替代或重新启用 |
| 插件生命周期 | 无账号Context7安装，安装v1→v2及旧版本409保护，真实文档调用；卸载后新任务仅发现未安装态，原安装/绑定回基线 | 重复安装503实际失败；未验旧在途授权撤销、不同发布升级或OAuth；截断工具正文不算全文核验 |

本批隔离数据库/表/看板/安装已按真实ID清理，自动化归档；SQL查询、模型任务/会话和旧运行历史仍保留。404不等于物理擦除；没有改CLI/后端、没有PR，也未外发消息。
