# Connector 领域

## 发现连接器和任务

```sh
moi-cli -o -no-interactive -workspace-id "$workspace_id" connector list
moi-cli -o -no-interactive -workspace-id "$workspace_id" connector get --connector-id "$connector_id"
moi-cli -o -no-interactive -workspace-id "$workspace_id" connector tasks --page 1 --page-size 20
moi-cli -o -no-interactive -workspace-id "$workspace_id" connector get-task --task-id "$task_id"
```

连接器类型、source/usage type 和目标必须来自用户选择或实际列表。创建连接器前可用
`connector validate`；不要把密码、AccessKey 或连接串写入对话、文件或日志。

## 文件导入

本地CSV可用connector upload一次上传并自动启动；或先local-upload暂存，再create-task。
按[文件导入Recipe](../recipes/import-file-to-table.md)选择一种模式，不能混用。暂存模式示例：

```sh
moi-cli -o -no-interactive -workspace-id "$workspace_id" connector local-upload --file-path "$file_path"
moi-cli schema connector create-task
moi-cli schema connector create-task --validate-file "$request_file"
moi-cli -o -no-interactive -workspace-id "$workspace_id" connector create-task --json-file "$request_file"
```

`local-upload` 返回的 `conn_file_ids`、`file_ids` 不可互换；创建任务使用 schema 明确的字段。
暂存后可用 `connector preview-file --conn-file-id` 按真实表头行、数据起始行、工作表和 CSV 方言
检查解析结果；预览只证明该暂存文件可按所给参数读取，不证明任务已创建、已启动或已入表。
需要核对暂存原始字节时使用 `connector file-stream --conn-file-id --output`，它与 Catalog
`download-file --volume-id --file-id` 的对象和 ID 不同。
创建后先查询同一task状态，仅在尚未启动且本次目标需要手动启动时使用run-task-now。
connector upload已经自动启动，不能再次run-task-now。创建任务不等于已运行或目标表已正确写入。使用 `task-runs`、`get-task` 和目标 Catalog/SQL
核验。失败文件可用 `retry-task-files`，但必须先取得LoadTaskFile ID。
旧版`task-files`缺task-id且实测500时，使用`task-runs`返回的workflow_app_id/workflow_execution_id调用
`workflow list-run-import-files`，核对import_context与原task/run后读取真实失败记录。
dev 当前 `task-files` 另有 repeatable `--query-param name=value`，分页/筛选名必须遵循正式 Product
API，不能自行猜成 task-id；该输入存在也不改变上述 LoadTaskFile 与 Catalog file ID 的边界。
重试按文件选择涉及的表，不保证裁剪同表全部输入；多表分组、已落地数据和恢复方法见
[导入Recipe](../recipes/import-file-to-table.md)。不因超时重复创建任务。

## 导出

导出使用 `schema connector create-export-task`，然后用 `export-task-state/info/files` 检查状态
和产物。返回 export ID 不等于已完成下载；覆盖本地文件前必须确认路径。
`export-task-files` 可用 `--statuses`、`--limit/offset` 缩小检查范围；没有产物时不能用空列表推断
任务成功。本机表/查询 CSV 下载走[下载Recipe](../recipes/export-and-verify.md)，不需要也不应
伪造成外部 Connector 导出任务。

## 现成连接与新建配置的边界

常规0.1.8的create/validate仅接name、source-type、usage-type，update仅接connector-id；
SDK虽有config options，CLI尚未接出配置。先看当前help；缺这些输入时不能说“给凭据就能建好”，
也不创建无配置空对象冒充外部连通。用户已配置且明确授权的连接可以走现有发现/读取路径。

外部导入需要真实源和可核对的数据范围；Connector导出还需要目标目录、覆盖策略及目标端内容校验。
缺这些条件保留“缺测试条件”，不归为接口失败；本机下载另走下载Recipe。
一批导入失败后必须先独立查目标表，不能假设好文件已成功或所有写入自动回滚。
本次好/坏CSV混合上传实际两份都失败且表0行，不构成“部分成功恢复”证据；
后续已用上述Workflow入口取得原任务的真实文件记录，原“没有ID入口”的判断已纠正。
仍不能拿Catalog file ID代替LoadTaskFile ID，也不能把同表合批和每表单文件的重试语义混用。

当前 dev 的 `config-file` 已能完成配置生命周期。2026-11-15 使用只含本地无效端点和虚构凭据的
隔离 S3 配置，通过 stdin 传入，完成 create/get/update/validate/delete，删除后 get 为404。两次 validate
均返回 `valid=false`，符合无效端点预期。stdin 用于避免把凭据写进进程参数；报告只保留脱敏摘要。
该结果证明创建、替换、验证和删除配置可用，不证明真实 S3 连通、导入或导出通过，也不能把无效端点
验证失败记为有效连接。

通用“编辑连接器”验收只需证明所选类型的连接参数确实被替换并回读一致；具体数据源类型由
各自功能点另行验收，不要求在一个通用编辑条目里重新穷举全部类型。2026-11-17 复核既有隔离
S3 回执：region、bucket 和名称均已替换，get 与更新输入一致，因此该通用编辑项可判通过。
真实 S3 连通仍是另一项，不影响这个判断，也不能由编辑成功反推。

通用“创建连接器”则不能把同一路径直接改绿：当前 `connector validate` 只有 config/name/
source-type/usage-type，没有 saved connector ID 输入；已有证据只证明保存前配置校验。
保存后的连接测试、真实外部连通和页面取消仍须分别验收，缺哪项就保留哪项。

删除还要验引用保护。2026-11-17 新建隔离 S3 配置和一个引用它的周期载入任务后，
`connector delete` 返回 409 `ErrConnectorInUse`，明确提示仍被运行或未完成任务使用；
随后 get 仍能读到原 connector，task 也仍引用同一 connector ID。先暂停并删除本次隔离任务，
再删除 connector 成功。结合此前“无引用时删除后 get 404”，通用删除功能的两条路径均有证据。
清理中直接删除活动任务曾返回 ErrServer，暂停后删除成功；这不是绕过引用保护，也不能删掉
其它用户任务来释放连接器。
