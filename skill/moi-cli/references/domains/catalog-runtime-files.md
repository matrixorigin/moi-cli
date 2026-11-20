# Catalog Runtime / Workspace File 扩展

`catalog --help` 暴露以下扩展命令；`file list` 的生成 reference 已在
[文件命令](../commands/file.md)提供，其他命令仍逐项核对当前二进制 help：

- `list-workspace-files`（`file list` 的 alias）
- `upload-runtime-artifact`
- `download-workspace-file`
- `analyze-runtime-image`

这些命令与普通 `catalog upload`、Connector staging、Knowledge source 和 Agent workspace
file 的目标不同。执行前运行对应命令的 `--help`，确认输入字段和运行上下文；不能把 runtime
artifact 当作普通 Catalog Volume 文件。

通用边界：

- `upload-runtime-artifact` 与 task-scoped `file upload --publish-artifact` 共用正式上传/发布owner；需核验原task的artifact关联和下载内容。
- `download-workspace-file` 必须使用真实 workspace file ID，并确认本地输出路径不会覆盖用户文件。
- `analyze-runtime-image` 需要明确的图片和分析目标，返回分析结果不等于图片已进入 Catalog 或 Knowledge。
- `list-workspace-files` 只用于发现当前 workspace 可见的 file placement，不替代 Volume 文件列表。

Volume 内的文件/文件夹枚举改用 `catalog files --volume-id`；根目录和子目录分别使用各自 canonical
Volume ID，当前 QA 已验空或省略 `--parent-id` 列出其直接条目。不要用根 Volume 加非空子 folder
parent 代替，本次该组合返回 400。名称、扩展名、file IDs、`file/folder` 类型、页号分页和排序属于
该 Volume 列表的输入。`file list` 仍用于 workspace placement 的 token 分页与跨 Volume 引用核对，
两套分页、作用域和 ID 不可混用。详见 [Catalog 文件与文件夹筛选](catalog.md#volume-文件与文件夹筛选)。

## 分页和精确文件筛选

`file list --page-size N` 返回 `next_page_token` 时，下一页把它原样交给 `--page-token`，
保持 workspace、name/extensions/file-ids 等筛选条件不变。不自行解码、构造 token 或假定首屏等于全部。
2026-11-01 两页各两条，取得四个不同 placement；另有两页各一条的对照。
这是 workspace placement 分页通过，不是常规 `catalog files` 已补上 Volume 筛选。

已知物理 file ID 可用 `--file-ids` 精确筛选，再检查每项的 `volume_id` 和 placement `id`。
同 file ID 在不同 Volume 可返回多条；`total` 此时是 placement 数，不保证是去重物理文件数。

命令文档存在不等于业务范围已全验：未知目录的完整枚举、任意 Volume 筛选及任务内授权
仍各有前置条件；只按已经核验的 placement 集合恢复，不能把未完整发现的对象当作不存在。

2026-10-31 已补一条只读闭环：`file list --page-size 20` 发现真实 workspace file 后，
用返回的 `file_id=builtin_sampledata_05` 执行 `catalog download-workspace-file`，本地
3648 字节 PDF 的 MD5 与列表返回的 `ac48ca3175189a60b31854350e096a10` 一致。该证据只
覆盖“发现 → 下载 → 哈希核对”；这一条单独只覆盖下载。另一次历史Agent工具任务已实际通过 `file upload --publish-artifact`
上传并发布runtime artifact，原task.artifacts和workspace file关联一致，正式下载2003字节、MD5/SHA256与上传回执相同。
两种上传命令共用owner，这不是另一个尚未验证的上传生命周期；不同成果类型、版本治理或删除仍需另验。

任务内上传依赖运行时注入的短期授权。桌面Codex常规身份缺该上下文时，不复制sandbox凭证或proof到桌面端，
也不把身份范围拒绝判成上传失败。按[工具与文件交付](../recipes/agent-tools-and-deliver-files.md)
从同task实际上传回执取得file_id/artifact_id，再核对task.artifacts、workspace file.source_task_id、
正式下载及内容/字节/hash。模型仅声称“已交付”或上传返回ID均不充分。

## 桌面附件、任务成果与删除后的历史

2026-11-02区分了两条正式路径：桌面`agent upload-workspace-file`创建可下载文件，
本次同名两文件有不同file_id，但按名称/ID的`file list`均没有placement；这不等于任务成果发布。
任务内`file upload --publish-artifact`则需核task.artifacts及实际source_task_id/source_session_id、
file_id/artifact_id和下载哈希。`agent get-task`使用正式发现的完整agent code，缩写404不能直接判任务失败。

同任务不同sandbox目录、不同任务上传相同basename，本次产生不同文件/成果ID及正确来源，
但workspace file_path仅basename，没有保留目录命名空间。名称和file_path都不作唯一键；
要目录保留、文件改名或原ID更新时先核正式输入并提示缺口。当前也没有已确认session/context列表筛选，
应按返回项核来源；这些证据不能代替跨身份权限隔离测试。

本次精确前缀+page-size=1按原token遍历三页取得三处placement，扩展名和file ID筛选另有对照。
删除时锁定实际file_id和其placement的volume_id，不能把数字placement id当file_id。
删除一份任务文件后，file-info/下载404且该placement消失，同名兄弟仍可读；原completed task
全文及旧artifact引用保持。交付历史引用前应再次确认当前可下载性，文件清理不等于清理任务历史。

无placement桌面附件不带volume删除，本次400且文件仍可下载。只有用户已授权清理自己的这些隔离文件时，
本次才用`create-file`将精确file_id作为user_upload挂到自建隔离Volume后删除，并核文件及库/Volume404。
这不是生产默认补救步骤；不得自动为用户历史文件新建Volume或变更挂载。生命周期条件不足时保留现状并报告。
