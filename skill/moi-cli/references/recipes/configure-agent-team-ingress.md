# Recipe：Agent Team 与 Ingress

Team 已有独立命令，generated reference 未完整导出，但可结合本机 help 和真实 Product SDK 合同执行。
typed schema 缺失应明确记录，不能因此无视已核实的正式业务 JSON，也不能凭 UI 名称猜字段。

## Team、成员和会话

| 对象 | 输入/命令 | 核验 |
| --- | --- | --- |
| Team创建 | create-team，JSON `{name,description,members:[{agent_workspace_id,agent_id}]}` | Team id、version、membership_version |
| Team修改 | update-team，JSON `{name,description,expected_version}` | 使用 team.version |
| 成员变更 | add-team-member/remove-team-member，team-id、agent-workspace-id、agent-id、expected-version | 使用 membership_version，回读成员集合 |
| 建会话 | create-team-session，team-id，JSON `{name}` | 保存 context_id 和 session.version |
| 会话改名 | rename-team-session，team-id、context-id，JSON `{name,expected_version}` | 使用 session.version；以 list-team-sessions 回读 |
| 发消息 | submit-team-message，team-id、context-id，以下JSON | task completed、list-team-messages 与答案断言 |

```json
{
  "message_id":"每轮唯一ID", "idempotency_key":"每轮唯一键",
  "parts":[{"kind":"text","text":"本次要求"}],
  "requested_member":{"agent_workspace_id":"实际owner","agent_id":"实际AgentID"}
}
```

模型/effort 不是当前 Team 消息字段；先核对成员自身模型和实际任务事件，不在开放 JSON 里悄悄增加参数。
多成员自动选择可能另有模型调用。超时后查原消息和task；精确重放须保持相同 message_id、idempotency_key
和内容，不能改键后重试。业务防重通过应以同一task及没有重复回合证明。

2026-10-30 原会话和全新隔离会话均出现同 context 第二轮未取得前轮437；显式带入已核验前文能计算448。
正式runtime回读确认两轮均completed/committed、conversation_input=true、同成员版本和DeepSeek，
provider反馈两轮session_turn均为1，未显示截断。当前公开请求不要求客户端history/provider session，
没有用户漏传必需参数的证据；共享会话由服务端维护，禁止伪造managed metadata。
本次生产路径自动接续实测失败，不能泛化为整个Team不支持多轮。公开读模型不含实际provider请求和历史，
精确根因须沿首轮共享session写入、次轮冻结/派发、provider历史加载三处核查，不能仅凭结果中的session ID变化定根因。
复杂工作可由 Codex 阶段账本带入最少确认事实，此方式的成功单独报告，不冒充自动历史继承通过。

系统 Agent 的 owner 需正确解析，任务读取见[工具与文件交付](agent-tools-and-deliver-files.md)。

## 版本、归属与同键异文

Team配置使用`version`，Agent成员集合使用独立`membership_version`，会话改名使用`session.version`，不能互换。2026-11-02隔离样例：改配置只推进配置版本，增删成员只推进成员版本；各自旧expected_version均409 `channel_target_version_conflict`且新状态保持。跨Team误配另一个真实context，改名和消息都404 `channel_target_not_found`，没有新会话或回合。此为同工作区两个Team的归属对照，不代表跨账号/普通成员权限矩阵。

精确消息重放必须保持message_id、idempotency_key和完整正文。该样例相同两键但改parts正文仍exit0返回原task，没有显式内容冲突reason；原1 user+1 assistant、正文及task完整保持，新意图没有执行。这未满足“同键异文显式拒绝”的验收目标；也不能因无409就推断新正文已被采用，或据此断言违反已公布的后端合同。

调用者发送前保存请求正文与SHA，并把两键绑定到这一次确定意图；发现本地正文变了，不能复用旧键冒充重试。已授权的新意图另立新请求，旧请求结果先核清；不要在响应未知时换键重发原意图。业务防重、内容一致性检测和自动历史接续分别判断，精确重放通过不修复后两者。

删除Team时使用当前真实`version`作为`expected-version`。2026-11-15隔离复测中，省略该字段返回400 `channel_target_invalid`，传入创建时的`version=1`后删除成功。删除后get仍返回`status=deleted`而不是404，成员列表仍能读到原关系，session入口转`channel_target_unavailable`。按真实生命周期状态确认停止服务，不反复delete追求404，也不称物理purge/零元数据残留。

## Ingress 与外部渠道

1. 先读取 `agent --help` 以及目标子命令的 `--help`，确认本机版本确实暴露该命令；不能根据 UI 字段猜 JSON。
2. 用 list/get 命令发现真实 team、member、session、connection 和 agent ID，保存 workspace 绑定关系。
3. 对 create/update/member/message/connection/verify 等写操作，只有在用户明确给出目标和动作后才继续；涉及 OAuth、token、二维码或外部渠道时，凭证只能由用户在本机安全输入。
4. 如果命令没有可审计输入合同、返回 ID 或状态查询，执行前标记为 `partial`，并把原始帮助和服务端返回作为限制证据。
5. 提交后查询 team/session/message/connection 详情；`verify-ingress-connection` 只证明连接检查，不证明真实事件已经到达 Agent。
6. 超时或断连时先查原对象，不能重复创建 team、session 或 ingress connection；未知状态保持未知并报告。

Team的创建/会话/单轮/防重、自动多轮、Ingress连接检查、真实渠道事件分别记录，不能用其中一项覆盖整组。
