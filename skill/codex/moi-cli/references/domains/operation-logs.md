# 操作日志领域

产品操作日志页面展示请求方法、状态码、请求/响应详情。当前 CLI 可用的相关入口分散在
`workspace-member audit-logs`、`audit-log-overview`、Account/Agent 任务详情和各领域的
get/detail 命令中，没有一个通用 CLI 操作日志查询模型。

```sh
moi-cli -o -no-interactive -workspace-id "$workspace_id" workspace-member audit-log-overview
moi-cli -o -no-interactive -workspace-id "$workspace_id" workspace-member audit-logs
```

这些命令只能证明返回的审计记录可见，不能保证覆盖 UI 上所有服务请求字段。需要核对某次
操作时优先使用该操作返回的 request/task/execution ID，再到对应 domain 查询；不要从同名
日志猜测关联关系。当前覆盖等级为 `read_only/partial`。

2026-10-31 的安全投影实测证明：`audit-log-overview` 与 `audit-logs` 可读，页面默认返回
20 条记录，能看到方法、路由、状态码、结果、耗时和 request ID；随后只读操作会继续写入
审计，所以 overview 总数会随读取时间变化（例如 1701、1751、1760 不是同一次快照）。
报告中不展开请求体、响应体、操作者名称、客户端 IP 和 User-Agent。审计记录只能帮助定位
某个已知 request/task/execution ID，不能代替业务结果、权限结论或完整分页治理；当前仍是
`read_only/partial`。

## 已知请求的关联与默认窗口限制

正式`audit-logs --help`当前只有workspace-id，没有已确认page、时间范围或request-id输入，typed schema缺失；不猜筛选参数。2026-11-02默认page1/page_size20且total明显大于20，不能用首屏判断任意历史请求存在与否。overview和列表不是同一时点快照，总数变化也不是本人本轮操作次数。

关联已知操作时优先使用原回执的request/task/execution ID。若列表记录未把资源ID放在path，本例仅在进程内用已知虚构workbook ID/名称对原body作精确等值匹配，持久化只保留匹配boolean及方法/路由/状态/request_id/时间；不保存请求/响应正文、操作者/IP/UA/headers或metadata。审计创建时间需与原调用区间对应，不能只按路由相同就认作同一请求。

本例已删除对象独立get报404并返回request_id，该ID未出现于当次20条窗口；这不证明请求没发生或全局没有审计记录。领域原结果仍是业务证据，审计所报status_code不等于HTTP抓包，也不替代业务终态。当前仍read_only/partial，未因此升级任意历史查询能力。

## dev 已接出的筛选与分页

当前 dev 的 `audit-logs --help` 已提供 `--method`、`--status-code`、`--result`、
`--start-time`、`--end-time`、`--page` 和 `--page-size`。2026-11-17 生产只读复核：
`POST + status-code=200 + result=success + 明确UTC时间范围` 精确返回本轮3条记录，
`page=1,page-size=3` 也只返回3条；记录可看到真实路由、状态、request ID及必要时的请求/
响应字段。旧的“只有workspace-id”描述仅适用于历史版本，不能再当作 dev 当前限制。

筛失败记录时先按状态码和时间查询；`result` 的枚举必须以真实返回值为准，不能把 `fail`
当成 `success` 的反义词猜用。空筛选结果不证明没有失败，也不证明参数生效。操作人、模块、
操作类型、来源、对象关键词、高敏开关、字段级红绿 diff 和关联跳转仍没有由上述 flags
全部覆盖，继续按[部分通过合同](../contracts/partial-coverage.md)区分 CLI 缺口与 UI 验收。
报告仍默认不展开请求体、响应体、IP、User-Agent 或操作者隐私。
