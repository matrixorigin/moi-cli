# 风险与确认

| 操作 | 默认等级 | 执行前要求 |
|---|---|---|
| list/get/detail/schema/metadata/config show | 只读 | 可直接执行 |
| upload、import、Workflow run、Knowledge prepare、chart run | 写入或异步 | 目标、输入和 workspace 明确 |
| update、pause、resume、retry、cancel | 受控写入 | 具体对象、动作语义和恢复影响明确 |
| delete、truncate、overwrite、publish、subscribe | 高风险 | 具体对象、影响范围和用户确认 |
| invite、role binding、IAM grant、跨 workspace | 权限敏感 | 用户、角色、资源和生命周期明确；遵循服务端授权 |

Skill 不替用户选择第一个 workspace、资源、角色或目标。缺少必需事实时只询问缺失事实，
不请求重复确认已完成的只读发现。任何批准都不允许跳过 Product API、服务端授权或终态核验。
