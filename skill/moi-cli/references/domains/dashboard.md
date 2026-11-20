# Dashboard 领域

## 发现和创建

```text
moi-cli -o -no-interactive -workspace-id "$workspace_id" dashboard list --json '{"page_size":10}'
moi-cli -o -no-interactive -workspace-id "$workspace_id" dashboard get --dashboard-id "$dashboard_id"
moi-cli schema dashboard create
moi-cli schema dashboard create-chart
```

创建看板和图表前必须取得真实数据表、数据库和用户选择的字段；typed JSON 先执行 schema
和 `--validate-file`。

当前 CLI 的 list 需要 JSON 输入，不传可能返回 `parse --json: unexpected end of JSON input`。
用 `schema dashboard list` 检查 page_size、page_token、search。返回 `{}` 时保留观察，
未核实 response 合同前不解释为没有看板或产品故障。

## 执行和核验

```text
moi-cli -o -no-interactive -workspace-id "$workspace_id" dashboard run-chart --dashboard-id "$dashboard_id" --chart-id "$chart_id"
moi-cli -o -no-interactive -workspace-id "$workspace_id" dashboard chart-result --dashboard-id "$dashboard_id" --chart-id "$chart_id"
```

`run-chart` 接受不等于结果完成。用 `chart-result` 检查状态和数据；需要下载时再执行
`download-chart-result`。删除、更新和刷新属于写入或副作用。

## 定时刷新与布局

布局和x/y/颜色配置的回读只能证明保存成功，页面外观仍需真实UI验收。
`refresh_interval` 从当前schema/正式合同选择；本次验证5m，收尾改manual。
创建图表后可能马上有首条执行，首条completed不能证明下一次定时触发。

确认该图表确有正式托管Workflow关联后，读取其cron表达式、版本、scheduler_task_id和运行历史；
本次合同用chart ID作为托管Workflow ID，并已用实际get确认。不能仅凭两个ID长得相同推断关联。
只通过dashboard命令更新调度，不直接修改托管Workflow。

等待后续不同execution在计划时间完成；期间不调用run-chart或手动刷新。核对原调度版本、
本次execution结果和实际表格内容。必须在删除看板前保存结果、必要时下载CSV；删除可能同时移除托管结果查询入口。
初始执行、后续cron执行、live chart-result可能返回不同execution，按ID分别归属，不拿live结果充当定时内容。
结束切manual，回读scheduler_task_id已移除，再按授权删除隔离看板并回读404。
告警需要专用接收目标和真实触发条件；auto refresh通过不代表告警或UI外观通过。

## SQL校验、失效来源与恢复

先保存目标chart完整配置与关联托管Workflow，再修改一次并独立回读。2026-11-02坏SQL被`ErrDataDashboardSQLInvalid`/400拒绝，原配置完整保持；配置校验通过与实际源表仍有效是两件事。

本例删除隔离源表后run仍返回execution，但执行failed，live结果503。重建同名表会得到新table ID，图表仍绑定已失效旧ID，不能靠同名重建宣称恢复。按正式dashboard schema更新真实source ID并核SQL/数据库引用，新execution才completed且数据正确；不直接改托管Workflow，也不把此错误自动归因登录。保留旧failed执行与完整结果，不用新成功覆盖旧失败。

更新图表本身可能产生自动执行，手动run另有execution。一次状态/live结果读取只归属它实际返回的ID；不能把先前completed当成本次更新或刷新完成。保存创建/更新/run各回执，沿其真实execution查终态、statement及case_result，下载内容再与独立SQL事实核对。

## 并发刷新和合法空结果

本例两个run-chart请求在途重叠，各创建独立execution并completed；live结果只指其中一个。没有已确认请求幂等键/CAS时，不假设同图双刷新自动合并，也不能把live视为指定历史执行的稳定结果。重复意图由调用者协调，响应不明先查询原执行，避免盲重交。

本例合法`WHERE 1=0`的两次执行都completed，独立typed SQL对照也成功；但chart-result及CSV下载均500，未形成空CSV交付。必须区分“执行完成”和“空结果展示/下载失败”，不伪造空数组或空文件，也不说换成非零SQL后原问题修好。另一个非零对照用`IS NULL`标记列与空串/0分别核验并下载成功；CSV中的空白本身不足以区分NULL与空串。空值语义按字段和明确标记核对。查询历史执行的`at`是时间参数，不是execution ID入口，不能拿它代替真实执行标识继续取结果。

删除隔离看板前先保存必要历史/下载；删除后托管Workflow与结果入口可能404，不等于历史物理擦除。

## JOIN汇总、单表钻取与显式快照

图表汇总成功不证明钻取可用。本轮两张JOIN图的聚合及CSV与独立SQL一致，但正确role:x的JOIN钻取和下载仍400/ErrDataDashboardDrillThroughUnsupported；现有合同只支持单物理表钻取。保存原错误，不通过改参数把它说成已修复。

若业务允许额外物化且用户已授权，可另建JOIN明细的物理快照表，再绑定单表图：本轮钻取/CSV与独立JOIN明细逐行一致。快照是另一条有维护成本的路径；源增删改和维度改名后，整板refresh只让JOIN图读取新值，快照图仍读旧表。必须显式同步快照、独立核表，再运行快照图并核本次execution和钻取/CSV，才能交付当前明细。分步同步不是原子更新。

2026-11-15 的另一组生产快照再次验证：原生JOIN钻取仍400，物理快照钻取成功，CSV只含`open/east/100.00`，
并正确排除`closed/50.00`。这同时核对筛选条件和实际文件内容，不把快照成功当作原生JOIN已支持。

同一看板的各图分别核execution、statement和数据；创建自动执行与手动执行可能竞争live结果，按各自ID保存历史，不拼成同一次成功。本轮重排顺序及layout.order回读一致；删除一图后其余图仍能运行/下载且源数据保持。配置顺序不等于页面视觉验收，删除后历史入口404也不等于物理擦除。
