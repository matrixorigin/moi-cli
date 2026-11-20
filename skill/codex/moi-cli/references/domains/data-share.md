# Data Share 领域

## 发现

```sh
moi-cli -o -no-interactive -workspace-id "$workspace_id" data-share publishes
moi-cli -o -no-interactive -workspace-id "$workspace_id" data-share subscriptions
moi-cli -o -no-interactive -workspace-id "$workspace_id" data-share source-tables --source-database-id "$database_id"
```

发布必须先确认来源数据库、表范围和目标 workspace ID；不能从列表顺序选择目标。
`source-tables` 返回 MatrixOne 表 rel_id；发布用的 `--table-ids` 不是 Catalog table_id。

同一源数据库只能有一份发布。创建第二份会返回 `409 / ErrDuplicatePublish`，不能只换发布名
重试，也不能擅自拆库、复制数据或撤销旧分享。若用户明确授权用合成数据隔离验收，才可在
新库复制获准的数据，并重新发现新表 rel_id；这不扩大成分享其他表或整库的授权。

## 发布与订阅

`create-publish`、`update-publish`、`delete-publish`、`subscribe`、`unsubscribe` 都是权限
敏感写入。执行后重新查询 `publishes` 或 `subscriptions`，并在目标 workspace 用 Catalog/SQL
核验实际可见性。提交发布不等于目标侧可查询。

解除订阅后，核对目标目录、已知表 ID 和新 SQL 请求均不能再读取该订阅；删除发布后，
核对源发布列表和目标邀请列表。本轮已验收先退订再删发布；不能据此宣称直接删除活跃发布的
级联清理已通过。另在新的隔离只读分享上分别验证了目标 INSERT、UPDATE、DELETE 均被服务拒绝，
源和目标数据未变；完整步骤见[发布、订阅与撤销 Recipe](../recipes/publish-and-subscribe-data-share.md)。
