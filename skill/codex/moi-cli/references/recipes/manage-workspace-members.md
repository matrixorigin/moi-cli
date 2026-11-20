# Recipe：受控管理 Workspace 成员

这个流程用于邀请、更新状态、删除成员或变更成员相关角色。成员变更会影响访问权限，
必须先锁定 workspace、目标 user ID、动作和影响范围。

1. 用 `workspace-member members` 列出目标 workspace 的成员；用 `member --user-id` 获取目标详情。
2. 用 `workspace-member current-principal` 和 `iam current-principal-capabilities` 确认当前主体具备执行该动作的能力。
3. 如果动作涉及角色，先用 `iam roles`、`role-permission-view` 和现有 binding 查询保存当前状态；不要从角色名称猜 ID。
4. 对 invite、update-status、delete、user-tags 或 role binding 读取对应 `schema`，用 `--validate-file` 校验完整输入。
5. 在真正写入前复述 workspace ID、user ID、动作和影响范围；只有用户明确要求该动作时才提交。
6. 写入后重新查询成员详情、角色 binding 和必要的 audit log，确认服务端状态、版本和实际生效结果。
7. 若返回 pending、冲突、权限拒绝或未知状态，停止后续变更，先查询原成员和审计记录；不能重复邀请或重复绑定。

邀请成功只证明邀请对象已创建；禁用、删除或角色变更必须以成员详情和权限核验作为完成证据。
审计日志用于证明服务端记录，不替代 Core/PDP 授权结论。
