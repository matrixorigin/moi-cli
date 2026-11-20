# Recipe：发布并订阅 Data Share

1. 用 `source-tables` 和 `publishes` 确认来源数据库、表范围和现有发布。
2. 只在用户明确指定目标 workspace、发布名和 table scope 时构造 `create-publish`。
3. 保存 publish ID；重新查询 `publishes` 确认服务端状态。
4. 在目标 workspace 或用户指定的订阅上下文执行 `subscriptions`/`subscribe`。
5. 查询订阅状态，并用目标侧 Catalog/SQL 核验可见性。
6. 发布、订阅、取消订阅和删除都保留对象 ID；失败或权限拒绝时停止，不切换 workspace 绕过。

## 表身份与两侧核验

`source-tables --source-database-id ID` 返回发布所用的 MatrixOne rel ID；`create-publish --table-ids` 使用这些值，
不是 Catalog table_id。selected 只含用户批准的表；all 会包含后续新增表，不能为省事改 all。
source_database_id 还必须绑定源 workspace，不脱离租户单独使用。

```sh
moi-cli -o -no-interactive -workspace-id "$source_workspace_id" data-share create-publish --name "$publish_name" --source-database-id "$database_id" --table-scope-mode selected --table-ids "$rel_ids" --target-workspace-ids "$target_workspace_id"
moi-cli -o -no-interactive -workspace-id "$target_workspace_id" data-share subscriptions
moi-cli -o -no-interactive -workspace-id "$target_workspace_id" data-share subscribe --subscription-id "$subscription_id" --sub-name "$subscription_name"
```

目标 workspace 必须由用户选择或会话中已确认，使用同一合法身份进入目标上下文不等于切换身份。
发布返回 id 是 publication ID；目标列表的 id 是 subscription ID，二者不能互换。
subscribe 返回 target_database_id，目标 Catalog tree 应有 is_sub=true、唯一订阅库和选定表。
重新在目标侧执行有界 SQL，与源侧同范围结果比对；同时核对未选表没有被投影。
能读不证明只读写保护已经验收；未经授权不尝试目标写入。取消订阅/撤回发布是独立生命周期，
不在读完后擅自撤销用户要保留的分享。

## 已有发布与隔离验收

先检查同一 `source_database_id` 是否已有发布。同源数据库只能有一份发布；新名称也会返回
`409 / ErrDuplicatePublish / This object has already been published`。保留原失败与回读结果，
不要自动撤旧发布或复制数据来避开限制。

用户已明确授权“只复制指定合成数据、创建隔离源库表、向指定接收方测试”时，可以新建隔离库表。
复制后先逐行比对获准数据，再用新库的 `source-tables` 取得新 rel_id；新发布使用 `selected`，
仅包含这张新表。原表和原分享保持原状。没有这项授权时报告前置冲突，不能把普通发布授权
解释成允许复制和分享新对象。

## 解除订阅与删除发布

只对已授权撤销的精确对象执行以下步骤；每个命令仍按当前会话的 endpoint、认证方式和对应
workspace 执行，全局选项放在 domain 前。

```sh
moi-cli -o -no-interactive -workspace-id "$target_workspace_id" data-share unsubscribe --subscription-id "$subscription_id"
moi-cli -o -no-interactive -workspace-id "$target_workspace_id" data-share subscriptions
moi-cli -o -no-interactive -workspace-id "$target_workspace_id" catalog tree
moi-cli -o -no-interactive -workspace-id "$target_workspace_id" catalog table-info --table-id "$target_table_id"
moi-cli -o -no-interactive -workspace-id "$target_workspace_id" sql run --db "$subscription_name" --query 'SELECT COUNT(*) FROM approved_table' --limit 10
moi-cli -o -no-interactive -workspace-id "$source_workspace_id" data-share delete-publish --publish-id "$publish_id"
moi-cli -o -no-interactive -workspace-id "$source_workspace_id" data-share publishes
moi-cli -o -no-interactive -workspace-id "$target_workspace_id" data-share subscriptions
```

SQL 示例中的 `approved_table` 替换为已批准、且此前查询成功的真实表名。退订后应同时看到：
订阅状态为 `unsubscribed`、订阅库从目录消失、此前有效的表 ID 返回 not-found/404、新 SQL
请求无法读取该订阅库。当前实测sql run返回 `Unknown database`/404；保留服务端真实错误，
不要只凭 unsubscribe 的 `ok` 判定撤销完成。

删除发布后，源列表不应再有本次 publication，目标列表不应再有本次 invitation/subscription。
若有需要保留的其他分享，比较操作前后精确 ID 和范围，并在目标侧做小查询确认未受影响。
记录本次源库表是否保留，不能把删除发布说成删除了源数据。

这条流程验收的是“先解除订阅，再删除发布”。2026-10-31另测直接删除活跃发布：返回503，
两侧关系已消失，但接收方订阅库仍在且能读原六行。这是已确认的部分变更与清理失败，不能说撤销成功。
本次只对获准清理的隔离订阅库完成后续删除并回读目录和SQL；真实用户对象遇此情况应保留影响证据并交产品owner，
不要自动删库、重建关系或反复删发布。复用该失败源库再建发布也有500记录，不能推荐它作为恢复路径。

另一份独立readonly分享已分别测试目标 `INSERT`、`UPDATE`、`DELETE`，三者都被服务拒绝并返回
`cannot insert/update/delete from public table`。每次拒绝后都核对源、目标，逐行和六行总额1500
保持一致，再按退订→删除发布收尾。错误状态500本身不代表写入已发生，必须核对数据；本次新建
的隔离分享已清理，原分享未改动。

## 源数据变化、重订阅与范围更新

2026-11-02在获准的两工作区，用自己的隔离selected发布核对：源端INSERT、UPDATE、DELETE后，
接收侧每次新SQL与源端逐格一致；未选的已有表和发布后新增表都不可读。拒绝实际为500/table not found，
不能写成正确返回404，也不能只看非零退出就认为范围已安全收敛；应同时核可读表的真实数据和关系范围。
这是本次小表的实时读取证据，不是跨工作区复制快照或传播时延保证。

同接收库名订阅另一份自己的隔离发布返回409/ErrNameConflict，原订阅的数据不变。
不要自动退订、删库或改名解决冲突；应先明确用户要保留的对象和命名。

退订后可在授权范围内对同subscription ID重新subscribe，但必须重新读取target_database_id、Catalog tree
及表ID。本轮同名重订阅的数据库和表ID均变化，旧ID继续404；新目标包含退订期间的源新增行。
不能沿用之前保存的目标ID，也不能把旧入口失效当成重订阅失败。

变更已发布的表集合前先核`update-publish --help`。当前0.1.8只接publish/workspace ID，
没有table scope/table IDs等正式变更输入，schema也不可用；不能把调用成功当成表集合已修改。
说明“当前CLI未建立范围更新方法”，不要据此断言后端做不到，不猜参数或删除重建用户发布。

当前 dev 已能正式执行 `update-publish`。2026-11-15 隔离发布只含 `approved_orders` 时，approved 可读、
未选表不存在；加入 `later_table` 后它变为可读，移除后再次不存在，approved 始终可读。每次都用接收方
真实 SQL 和 selected rel_id 集合核对，不能只信更新回执。该路径补齐了 dev 的范围更新能力，但未选或移除表
仍返回500/table does not exist，不能据此宣称规范403/404权限语义已改善。
