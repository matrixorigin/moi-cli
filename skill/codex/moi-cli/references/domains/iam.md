# IAM 领域

只读检查：

```sh
moi-cli -o -no-interactive -workspace-id "$workspace_id" iam current-principal-roles
moi-cli -o -no-interactive -workspace-id "$workspace_id" iam current-principal-capabilities
moi-cli -o -no-interactive -workspace-id "$workspace_id" iam roles
moi-cli -o -no-interactive -workspace-id "$workspace_id" iam permission-schema
moi-cli -o -no-interactive -workspace-id "$workspace_id" iam role-permission-view --role-id "$role_id"
```

创建、更新、删除角色、继承关系、对象 grants、data privileges 和 user-role bindings 都是
权限敏感写入。typed JSON 先用 schema/validate-file；涉及 CAS 的 binding 必须使用返回的
当前版本，不能拼接版本或绕过服务端授权。Core/PDP 拒绝时 fail closed，不在 Skill 中加入
本地 allow fallback。
