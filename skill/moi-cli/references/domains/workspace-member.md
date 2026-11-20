# Workspace Member 领域

只在用户明确指定 workspace、用户和动作时执行成员变更。只读发现：

```text
moi-cli -o -no-interactive -workspace-id "$workspace_id" workspace-member members
moi-cli -o -no-interactive -workspace-id "$workspace_id" workspace-member member --user-id "$user_id"
moi-cli -o -no-interactive -workspace-id "$workspace_id" workspace-member current-principal
moi-cli -o -no-interactive -workspace-id "$workspace_id" workspace-member audit-logs
```

`invite`、`update-status`、`delete`、用户标签变更和角色相关操作都具有权限或成员影响。
接受邀请、禁用、删除前必须确认具体 user ID 和影响范围；审计日志只能证明记录可见，不能
替代权限判断。
