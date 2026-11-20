# Workspace 领域

## 发现

账号级 workspace 列表：

```sh
moi-cli -o -no-interactive workspace list
```

选定 workspace 后详情查询：

```sh
moi-cli -o -no-interactive -workspace-id "$workspace_id" workspace get --workspace-id "$workspace_id"
```

`workspace list` 返回的真实 `id` 是后续 workspace-scoped 操作的唯一来源。不要默认选择第一项，
不要因空列表创建 workspace。`workspace create/ensure/start-creation/delete/update` 都是写入或
生命周期操作，必须有明确名称、目标和用户授权。

## 核验

workspace 详情只能证明资源可见和请求成功；不能证明 Catalog、SQL、Connector、Workflow 或
Compute 已经可用。继续执行领域任务前，使用对应领域的 list/tree/metadata 做最小健康检查。
