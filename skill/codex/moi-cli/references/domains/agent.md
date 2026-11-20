# Agent 领域

## 发现与运行

```sh
moi-cli -o -no-interactive -workspace-id "$workspace_id" agent list
moi-cli -o -no-interactive -workspace-id "$workspace_id" agent agent-card --agent-id "$agent_id"
moi-cli -o -no-interactive -workspace-id "$workspace_id" agent executor-health
moi-cli -o -no-interactive -workspace-id "$workspace_id" agent get-task --agent-id "$agent_id" --task-id "$task_id"
```

对话、A2A、资源、插件、技能、反馈和 workspace file 是不同对象。使用 `agent` reference
确认每个命令的输入，不把 UI 名称当作 agent ID。`send`、`a2a`、`workspace-mcp` 和插件
连接可能产生异步任务或外部副作用，必须保存返回 ID 并查询状态。

包操作 `init/build/verify/unpack/load-agent-package` 是本地或受控资源操作；验证通过不等于
已安装、已绑定或可运行。team/ingress 等命令若未出现在 references 中，标记为 CLI 文档缺口。
本地包、包内自定义技能、加载后的自动绑定和真实使用闭环见
[构建、加载并使用 Agent 包](../recipes/build-load-and-use-agent-package.md)。

工具与正式文件闭环见[交付Recipe](../recipes/agent-tools-and-deliver-files.md)，Team见[协作Recipe](../recipes/configure-agent-team-ingress.md)。
模型计算、实际工具执行、文件产物、知识引用、包加载是独立验收条件。

## 绑定资源的正式入口

可用资源通过 `agent resource-get --resource agent-builder --tail-path /resources` 发现；
读取 agents 的 `/AGENT_ID/bindings?agent_workspace_id=OWNER` 取得真实绑定。配置可能含敏感内容，
只投影本次所需ID/类型/版本，不整包写报告。更新走 `resource-patch --resource agents --tail-path /AGENT_ID/bindings`，
按真实合同使用 `binding_ops:[{op:"add",kind:"tool",ref:{id,workspace_id,kind:"tool"}}]` 等增量操作。
整类 skill_refs/tool_refs/knowledge_base_refs 数组可能替换原绑定，先保存原状态并确认范围。

Agent knowledge-bases资源ID不自动等于 semantic-model ID，需从正式资源发现或桥接结果取得真实引用。
模型专用 `/AGENT_ID/bindings/model` 接收 `{"model":{"default_model":"deepseek-v4.1-flash"}}`，
默认省略 params_override/effort。仅在用户授权的目标Agent上变更，不改系统MOMO来凑验收。
绑定保存后仍需本次真实工具调用或带来源问答；知识库列表出现、包verify成功均不能替代。

技能验收要从原任务记录确认实际 `skill` 调用及正确技能正文，并核对最终回答。测试提示只提供
技能名和输入，不把技能的转换规则、秘密测试标记或预期答案再写给 Agent，否则无法证明它用过
绑定技能。独立 `/skills/{id}/execute` 当前会保存状态但不交付 assistant 回答，不能仅凭其
`completed` 代替对话中的技能使用和结果验收。

## 带附件的托管Skill

这是独立于 `.moiagent` 包内声明式技能的ZIP导入链。ZIP根放SKILL.md，规则等支持文件放references/；
只打包本次允许的文件，不含凭据。先inspect核对清单、大小及hash，再正式import：

```sh
moi-cli -o -no-interactive -workspace-id "$workspace_id" agent inspect-skill-package --file-path "$skill_zip"
moi-cli -o -no-interactive -workspace-id "$workspace_id" agent import-skill-package --file-path "$skill_zip"
moi-cli -o -no-interactive -workspace-id "$workspace_id" agent get-skill-file --skill-id "$skill_id" --file-path "$reference_path" --output "$local_copy"
```

当前import无typed schema不等于不能用；已有正式文件入口可直接导入，不能因schema命令缺失误判能力缺失。
保存Value.skill.id/version及逐文件digest；下载附件后与源字节核对。按上面的binding_ops增量绑定，
不替换原有技能。A2A只给技能名和合成输入，真正规则放附件，避免提示泄漏答案。

运行时由skills/read给出当前task冻结Skill文件版本及正式download-skill-files指令；按这条指令在该task沙箱中物化文件，
读取真实附件后处理输入。宿主Codex不拼装runtime身份，也不把别的会话的冻结版本输入拿来调用。
成功要有同task的附件下载、实际读取支持文件、符合附件规则的结果与completed，不仅是读到SKILL.md正文。
2026-10-31样例完整通过，源附件与独立下载字节一致；测试Skill已解绑并删除，Skill与附件详情404，原绑定保留。

收尾仅对已授权的测试资源进行：先用binding_ops的remove解绑精确Skill引用，再用resource-delete删除该Skill，
回读原Agent绑定及Skill/附件状态。此前无脚本样例不单独证明脚本及更新冻结；脚本与独立A/B切换的限定实测见下文，同ID附件原位更新仍缺入口。

## 默认版本与附件更新边界

切Agent默认版本可能重建包声明的绑定；先切默认，再基于最新回读状态增量绑定托管Skill，独立回读后发任务并核对manifest。独立A/B附件包切换下的在途冻结及真实脚本执行已通过，详见[脚本附件Recipe](../recipes/use-script-skill-attachments.md#独立附件包切换的在途冻结)。import每次新建Skill ID，元数据更新无files/content，不能说支持同ID附件原位更新。

## 技能市场与自然语言创建

技能市场的搜索、分类筛选和分页分别保存查询参数、返回条数及真实 Skill ID；内置技能绑定后
还要在实际 Agent 任务中核对技能调用和最终业务结果，列表里出现不代表可执行。自然语言创建
使用 `polish-skill-stream` 形成草稿，再用正式创建入口保存；绑定后通过新任务验证实际结果，
最后按精确引用解绑、删除并回读 404。

```sh
moi-cli -o -no-interactive -workspace-id "$workspace_id" agent resource-get --resource skills --tail-path '' --query-param search=MOI --query-param page_size=5
moi-cli -o -no-interactive -workspace-id "$workspace_id" agent resource-get --resource skills --tail-path '' --query-param category=design --query-param page_size=2 --query-param page=1
moi-cli -o -no-interactive -workspace-id "$workspace_id" agent polish-skill-stream --json-file "$draft_request" --output "$draft_sse"
moi-cli -o -no-interactive -workspace-id "$workspace_id" agent resource-post --resource skills --json-file "$polished_skill"
```

2026-11-17 的生产隔离复核中，搜索、design 分类和两页分页均返回预期对象；内置财务技能
完成金额审查并算出差额 20 元。自然语言草稿保存为独立 Skill 后，绑定 Agent 的新任务返回
唯一合成标记，随后解绑、删除且详情 404。这个结论只覆盖当次市场技能和合成自定义技能，
不等于同 ID 附件原位更新或任意外部插件都可用。

## 成果批量删除与文件边界

通用成果批量删除使用正式 `agent-artifacts` 资源入口，请求中列出真实 artifact ID；分别保存
`succeeded` 和带原因的 `failed`，不存在对象的 `not_found` 是跳过记录，不能计为删除成功。
删除后同时核对成果列表和 Catalog placement。物理文件元数据仍可能存在，成果/placement 删除
不等于物理擦除；没有真实 Volume 归属时不要猜 ID 调 `catalog delete-file`。

```sh
moi-cli -o -no-interactive -workspace-id "$workspace_id" agent resource-post --resource agent-artifacts --tail-path /batch-delete --json-file "$artifact_batch"
```

2026-11-17 的隔离复核中，两个真实成果删除成功，一个不存在对象返回 `not_found`，成果总数
和 placement 均归零；两个物理文件的 file-info 仍可读。普通成员或无权限成果的跳过原因
需要专用成员另验，不能用管理员删除成功代替。
