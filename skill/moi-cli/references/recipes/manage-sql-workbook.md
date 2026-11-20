# Recipe：SQL 工作簿和版本

适用：保存可复用SQL、区分草稿与已保存版本，并执行选定版本。SQL内容与执行权限仍按用户授权核实。

```text
moi-cli -o -no-interactive -workspace-id "$workspace_id" sql create-workbook --name "$name"
moi-cli -o -no-interactive -workspace-id "$workspace_id" sql workbook-versions --workbook-id "$workbook_id"
moi-cli -o -no-interactive -workspace-id "$workspace_id" sql create-workbook-version --workbook-id "$workbook_id" --sql-content "$sql_content"
moi-cli -o -no-interactive -workspace-id "$workspace_id" sql workbook-detail --workbook-id "$workbook_id" --version-id "$version_id"
moi-cli -o -no-interactive -workspace-id "$workspace_id" sql save-workbook-version --workbook-id "$workbook_id" --version-id "$version_id" --json-file "$sql_json_file"
```

创建工作簿会产生初始草稿，用真实版本列表发现其ID。detail同时需要workbook ID和version ID；
少传version-id是输入未齐，不能记作服务故障。create/update-workbook-version用sql-content；
save-workbook-version的JSON文件内容是一个JSON字符串，例如 `"SELECT 31 * 17 AS answer"`，不是SQL对象。

保存后独立回读，核对版本成为finalize且正文一致。finalize不是不可变或只读保证，重复保存的边界见下文。只有用户要求编辑草稿或已授权隔离验收时，
才用update-workbook-version修改指定草稿，再回读已保存版证明未一起改变。普通保存执行不修改其他草稿。
运行选定版本时从detail取实际SQL，再走
[execute → describe → result](run-sql-and-fetch-result.md)；保存并不自动执行。
改名用update-workbook，删除用delete-workbook；仅对用户授权的目标执行并回读确认。

2026-10-31隔离验收已证明创建、草稿修改、保存版本保持不变、改名、执行得527和删除后404。
这不证明共享工作簿、并发编辑、所有SQL类型或SQL历史详情均已通过。

## 已保存版本、并发与执行时点

2026-11-02隔离样例：`update-workbook-version`对已finalize版本返回404且正文保持；但对同一workbook/version再次`save-workbook-version`传不同JSON字符串，成功覆盖同ID正文，仍只一份finalize版本。因此不能把finalize当冻结快照，也不能把重复save视作创建新版。业务需要保留旧SQL时先读版本列表和完整detail，明确另建版本，再核旧版未被一起改变；没有覆盖授权时不对已保存ID重存。

同草稿两个不同正文并发update都成功，最终只保留其中一份；另一个草稿双save也都成功，最终只一份正文且finalize，没有合并。当前正式help未提供已确认revision/CAS参数，不猜字段，不用CLI返回先后推断服务端提交顺序或“最后写入一定胜出”。多方编辑需协调单一写入者；冲突/回执不明先读原ID与内容差异，不盲重存。

2026-11-15 生产隔离复测已实际完成“另建版本保留旧 SQL”：A 保存并执行 `SELECT 31 * 17 AS answer`
得到 527，另建 B 保存并执行 `SELECT 31 * 18 AS answer` 得到 558，A 的正文未被覆盖。需要历史版本时
使用这条正式路径；它仍不证明同 ID 不可变、服务端 CAS 或多人并发合并。

保存前还应比较当前正文与本地旧快照。另一组隔离验证中，两位写入者共享 `SELECT 41 * 7`，第一人保存
`SELECT 41 * 8` 后，第二人重读发现变化，因此没有提交旧保存，而是另建 B 保存 `SELECT 41 * 9`；最终
A=328、B=369。该“重读→发现陈旧→另建版本”能阻止顺序陈旧覆盖，但不是服务端 CAS；真正并发仍需
单一写入者或外部协调，不能把本地检查当成提交时原子保证。

跨工作簿的version ID不能混用，本例交叉detail均404。执行前从选定workbook/version重新读取实际SQL并保存本次请求；保存动作不自动执行，也不保证后续再读同ID仍是旧正文。将query/statement与本次SQL正文对应起来，核真实结果；请求limit与返回结果limit分别记录，不按默认返回窗口认定全量。删除工作簿后仍可能保留SQL执行历史，不称物理清除。
