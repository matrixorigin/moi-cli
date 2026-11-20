# Recipe：创建看板并运行图表

1. `dashboard list --json '{"page_size":10}'` 发现目标看板，或确认用户要求创建新看板。
2. 发现目标 database/table，并使用 `schema dashboard create`、`create-chart` 校验 JSON。
3. 执行创建并保存 dashboard/chart ID。
4. `run-chart` 后轮询或查询 `chart-result`，核对状态、列和值。
5. 按请求下载结果或做 drill-through；不把 chart accepted 当作结果完成。
6. 删除、更新和刷新先核对会话已有授权；对象和范围已获准时继续执行，仅缺授权才确认，不在普通结果核验中自动清理。

看板创建 JSON 使用 `sources:[{"table_id":实际ID}]`；图表配置使用 `name`、`chart_type`、`sql_text`、
`chart_config`、`refresh_interval`、`alert_enabled`。当前 `chart_config` 是 JSON **字符串**，
柱状图例如 `"{\"x\":\"region\",\"y\":\"total_amount\"}"`，不是直接嵌套对象。
`xField/yField`曾能保存且SQL能执行，但不符合本次钻取配置合同，不能用保存成功证明交互配置正确。
SQL 中数据库和表名来自实际资产详情；区域汇总的指标须先有独立 SQL 基准。
本次手动图表 refresh_interval 空字符串由服务规范为 manual，不以此承诺定时刷新通过。

保存 run-chart 的 execution_id，chart-result 返回的 execution_id 必须一致且 status=completed，
再核对 result.columns/result 中的实际指标。下载见[导出与核对](export-and-verify.md)。
图表 SQL 数据通过不代表图表外观、联动钻取、告警和自动刷新都通过。

## 手动刷新与钻取

用户已授权的刷新用 `refresh --dashboard-id`，再查询目标图表的 `chart-result`，保存新 execution_id、
completed与实际行值。图表下载另核对CSV内容；执行ID不同的快照不能混为同一次执行。

钻取前读取当前图表配置。柱状图的点维度 `role` 是 `x`，不是列名 `region`：

```text
moi-cli -o -no-interactive -workspace-id "$workspace_id" dashboard drill-through --dashboard-id "$dashboard_id" --chart-id "$chart_id" --json '{"dimensions":[{"role":"x","value":"east"}]}'
moi-cli -o -no-interactive -workspace-id "$workspace_id" dashboard download-drill-through --dashboard-id "$dashboard_id" --chart-id "$chart_id" --json '{"dimensions":[{"role":"x","value":"east"}]}' --output "$detail_csv"
```

`east`只是合成样例，使用用户选定的真实点。核对生成SQL的筛选条件、visible_total、明细行和下载内容。
2026-10-31样例查回两行400和200，CSV一致。其它图形各有维度要求，不套用bar合同。
`ErrDataDashboardDrillThroughUnsupported`也可能由配置或维度错误触发，先复核输入，不能直接判功能不支持。

更新图表前读取当前完整配置并使用当前schema。本次只传chart_config被拒为SQLInvalid；带原name、
chart_type、sql_text、layout_config、refresh_interval及新chart_config后成功。不把update当作已证明支持任意局部PATCH。
告警接收方和真实渠道事件另验，不因为刷新成功就自动发送消息。

## 图表类型与尺寸矩阵

通用图表验收按当前 `create-chart` schema 覆盖 `metric`、`table`、`line`、`bar`、`pie`、
`area`、`combo`、`stacked_bar`（响应投影为 stacked）和 `scatter`。每种类型分别执行
`run-chart`/`chart-result`，核对 completed、真实列和行；只保存配置不算通过。

尺寸矩阵使用同一真实数据源，对每种类型分别保存 `layout_config` 的 `third`、`half`、
`full`，再读取 dashboard 逐项核对 chart_type 和尺寸。2026-11-17 生产隔离验收的
9 类图表均有真实结果，27 个类型×尺寸组合全部回读一致。这个矩阵证明配置与查询结果，
不证明 SQL 编辑器或看板页面的实际渲染、颜色和交互；这些仍归 UI 验收。
