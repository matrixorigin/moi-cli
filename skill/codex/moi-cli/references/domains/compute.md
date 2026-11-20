# Compute 领域

## 发现

```sh
moi-cli -o -no-interactive -workspace-id "$workspace_id" compute list
moi-cli -o -no-interactive -workspace-id "$workspace_id" compute get --id "$compute_id"
moi-cli -o -no-interactive -workspace-id "$workspace_id" compute runtime --compute-resource-id "$compute_id"
```

创建需要用户明确名称、规格和资源边界，并从 `compute specs` 或已确认 spec ID 取得输入。

## 生命周期

`create`、`update`、`suspend`、`resume`、`retry`、`delete` 都是写入。删除前先运行
`preflight-delete`，检查关联 Workflow、作业和 owner；不要因为列表为空就删除或创建替代资源。

计算资源状态只能证明生命周期状态；Workflow/Agent 实际使用的资源仍需在目标运行中核验。
