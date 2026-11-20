# Agent Team / Ingress 扩展

MOI内插件的发现、正式引用、安装前置和实际工具结果按[插件Recipe](../recipes/manage-plugin-capabilities.md)执行。
插件MCP调用与OAuth、渠道事件分别验收，不整组判通或判不通。

本机 `moi-cli agent --help` 暴露 team、team session、team message 和 ingress 命令，但当前
生成的 `references/commands/agent.md` 尚未列出完整参数。执行前对目标命令运行
`moi-cli agent COMMAND --help`；已核实的Team业务JSON与终态合同见[Team Recipe](../recipes/configure-agent-team-ingress.md)，
可在用户授权范围内多步完成，不因缺生成schema停在单步；未知字段仍不得猜测。

## Team

命令分为：

- `create-team`、`list-teams`、`get-team`、`update-team`、`delete-team`
- `list-team-members`、`add-team-member`、`remove-team-member`
- `list-team-sessions`、`create-team-session`、`rename-team-session`
- `submit-team-message`、`list-team-messages`

Team、成员、会话和消息ID逐步从返回结果取得。创建、更新、删除、成员变更和消息提交
都是写入或外部副作用，按已授权目标与Recipe推进并保留原始返回。单轮、消息防重和显式前文续轮已有实测，
自动历史接续失败另记，不能承诺自动记忆；多步执行也不能扩大到未获授权的成员或外部渠道变更。

## Ingress

命令分为：

- `list-ingress-capabilities`
- `list-ingress-connections`、`create-ingress-connection`、`get-ingress-connection`
- `update-ingress-connection`、`delete-ingress-connection`、`verify-ingress-connection`
- `resolve-we-com-qr-binding`

连接配置可能包含凭证、OAuth 或外部渠道信息，不把 JSON、token、连接地址或二维码绑定数据
写入日志和报告。`verify` 只证明连接检查结果，不等于渠道事件已经到达 Agent。

## 分别判断支持范围

- Team的已核实命令按Recipe执行；生成schema/reference缺口、自动历史接续、Ingress真实事件分别记录，不能整组一概停工或一概判通过。
- 不把 UI 的团队/自动化控件直接当成 CLI 成功。
- CLI source仍应将完整参数和终态合同纳入skill export；对尚无可审计输入/终态的Ingress分支保持partial，不凭UI控件猜字段。
