# Recipe：有界管理计算资源

先用 compute list/specs/get/runtime 发现当前资源、真实规格、enabled、CPU/内存和费用。字段缺失不解释为可用或免费，
也不为节省费用擅选不同规格。创建和真实负载可能计费，需已有明确范围；测试须约定停止时间并记录清理责任。

1. 在本次指定 workspace 创建唯一隔离资源，保存返回 ID；客户端结果未知先查原列表，不重复创建。
2. 回读 get/runtime。创建显示 IDLE、runtime 空，不证明实际启动、执行任务或产生零费用。
3. 仅按当前合法状态做 suspend/resume；IDLE 可能返回 `SUSPEND_FAILED: ... cannot be suspended`，
   不能将拒绝写成暂停通过，也不为消除这个错误强行提交负载。真实运行/暂停/恢复需要获准任务和独立运行证据。
4. 达到约定停止点优先清理本次资源并回读；当前CLI无独立stop命令，不能把删除说成有负载停止通过。清理前执行 `preflight-delete`，核对依赖和影响。非空引用或未知状态不盲删。
5. 删除后分别读 get、list、runtime。某些实现 get 保留 DELETED 墓碑，并非404；列表无该活资源、runtime无残留才形成完整清理证据。

2026-10-30 的 CC0 样例完成创建→IDLE回读→preflight→删除，创建到删除状态88秒，runtime始终为空。
这个结果不证明有负载的启动/暂停/恢复已通过，也不证明具体账单；实际费用需计费事实另证。

每条实际调用先用本机 help/schema 确认 flags。compute runtime 不等于 Agent executor 健康，Agent任务成功也不证明用了本次新计算资源。

2026-10-31负载补验没有完成：复用已绑定旧App的候选，部署请求发生归属/幂等冲突，未把真实负载提交到新资源。
不能将该500归为计算服务故障。候选更新需来自本次目标App的正式生成与验收结果，不以旧候选改compute参数拼接新部署。
本次资源存续406秒，比获准5分钟多106秒，已回读DELETED、list/runtime无活资源；新隔离App已删，原App版本未变。
实际费用未知，当前CLI未发现资源级账单入口。后续付费测试必须在停止点之前留出清理时间，到点优先收尾，不能继续排查占用资源。

## Runner 无连接生命周期

只需要验证 Runner 管理面时，可以不启动计算、不建立连接：依次执行 `runner create/get/list/refresh/delete`，
核对创建后的 pending 状态、详情、列表、刷新结果和最终空列表。2026-11-15 的隔离 Runner 已走通该路径。
refresh 返回的一次性连接材料必须脱敏，不能执行或展示 `connect_cmd`。此结果只证明创建、读取、刷新和删除，
不证明 Runner 终端、网络、配额或真实负载；Sandbox 创建、Sandbox 网络/配额和计算资源启停仍需各自授权与证据。

Sandbox 可先验证不扩容量的配置管理：`update-quota` 写入与当前一致的 `max_managed_count=3` 和
`auto_ephemeral=inherit`，再用 `get-quota` 回读，2026-11-15 已通过。它只证明保持值写入，不证明扩容。
同一轮 small/10GiB/limited 的 Sandbox 创建仍返回500，request ID为`84949a3b-d874-4ac1-a9ee-b5378f9a3131`，
list为空、get为404，未形成实例，也没有重复创建。公开错误不足以确定内部原因，不能用配额更新成功掩盖创建失败。

`account current-credit` 是账号总余额，specs.credit_per_hour是规格费率；当前未发现单资源实付账单入口。
其它工作可能同时耗费额度，不能把余额差值归给某个测试资源。付费范围用尽或超时后，
继续负载测试需新的明确时间/费用范围，不能把之前一次授权理解成无限次启用。
