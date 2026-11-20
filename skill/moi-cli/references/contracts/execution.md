# 执行合同

每个 MOI CLI 任务都使用四个阶段。阶段之间只传递上一阶段取得的真实事实。

## 发现

1. 读取当前会话已有的 workspace 和目标事实。
2. 缺少 workspace 时运行 `workspace list`，缺少资源时运行对应 domain 的 list/tree。
3. 对用户选择的对象运行 get/detail/metadata，确认可执行状态、输入和权限。
4. 对 typed JSON 先运行 `moi-cli schema DOMAIN COMMAND`，再用真实 ID 构造请求文件。

发现阶段不得执行写入，不得用名称、列表顺序或模型生成 ID。

## 执行

1. 一次终端调用只执行一个独立 CLI 命令。
2. 所有 workspace-scoped 命令显式传 `-workspace-id`。
3. 写入命令只使用用户明确的目标、已确认的输入和前一步返回的 ID。
4. 保留完整的机器可读返回中的对象 ID 和服务端状态；不复制凭证字段。

## 状态

- 同步命令：检查退出码和返回状态。
- 异步命令：保存 query/task/execution/job/model/source ID，按领域文档有界轮询。
- 仅对文档定义的 pending 状态继续轮询；失败、拒绝、无权限和参数错误直接进入终态处理。
- 不把 accepted/submitted/created 解释成 completed、可查询或可检索。

## 核验

用后续 get/detail/result、目标资源查询、SQL、文件存在性/内容或目标端状态证明结果。
核验对象必须与本次操作返回的 ID 关联，不能从历史列表中猜测同名对象。
