# Catalog 领域

## 发现顺序

```text
moi-cli -o -no-interactive -workspace-id "$workspace_id" catalog tree
moi-cli -o -no-interactive -workspace-id "$workspace_id" catalog list-databases --catalog-id "$catalog_id"
moi-cli -o -no-interactive -workspace-id "$workspace_id" catalog database-children --database-id "$database_id"
moi-cli -o -no-interactive -workspace-id "$workspace_id" catalog files --volume-id "$volume_id" --parent-id "" --page 1 --page-size 50
```

表、Volume、文件必须从返回对象取得真实 ID，再查询详情：

```text
moi-cli -o -no-interactive -workspace-id "$workspace_id" catalog table-info --table-id "$table_id"
moi-cli -o -no-interactive -workspace-id "$workspace_id" catalog volume-info --volume-id "$volume_id"
```

ID 必须连同资源类型保存：同一个数字可能同时属于 database、table 和 Volume。
表路径用 `table-full-path` 回读；`volume-full-path` 在已测环境只返回 Volume 名，不能据此拼出父级路径。
需要父级关系时，从 tree/数据库子对象及已保存的创建回执核对，不能凭同名对象推断。

## 文件边界

`catalog upload` 只负责上传字节。把文件放进已确认 Volume 根目录是另一条操作：

```text
moi-cli -o -no-interactive -workspace-id "$workspace_id" catalog upload --file-path "$file_path"
moi-cli -o -no-interactive -workspace-id "$workspace_id" catalog create-file --volume-id "$volume_id" --file-ids "$file_id"
```

上传返回 `file_id` 不等于文件已经挂载、可预览、可查询或已进入知识库。`file-info`、
`preview-file`、`file-data-asset` 需要文件能够解析到 canonical Volume；缺少 Volume 时保留
上传收据，不重复上传或切换低层接口。

挂载后必须沿同一 `volume_id + file_id` 回读列表、详情和内容；下载路径使用本次任务的新文件，
不能覆盖用户已有文件：

```text
moi-cli -o -no-interactive -workspace-id "$workspace_id" catalog file-info --file-id "$file_id"
moi-cli -o -no-interactive -workspace-id "$workspace_id" catalog preview-file --volume-id "$volume_id" --file-id "$file_id"
moi-cli -o -no-interactive -workspace-id "$workspace_id" catalog download-file --volume-id "$volume_id" --file-id "$file_id" --output "$download_path"
```

小文本/CSV 用解析器比较内容与源文件，其它类型至少比较字节数和 SHA256。预览成功或下载命令
退出 0 不能单独证明挂载位置、完整字节或内容正确。完整的文件到表核验顺序见
[Catalog 文件与结构化表核验](../recipes/verify-catalog-file-and-table.md)。

## Volume 运行前预检

把 Volume 交给 Workflow 或 Knowledge 准备前，先用 `catalog volume-info --volume-id` 读取
服务端的文件数量、大小和可用操作，再用 dev CLI 的 `catalog files --volume-id` 枚举实际条目。
`volume-info` 不保证返回文件清单，不能只用汇总数量代替文件身份核验。
至少保存 `volume_id`、本次 `file_id` 列表、文件数量、大小/哈希（若返回）和读取时间。

## Volume 文件与文件夹筛选

`catalog files` 是 Volume 作用域的条目列表；`file list` 是 workspace placement 列表，二者不能
互换。根目录和子文件夹都使用各自返回的 canonical Volume ID；当前 QA 已验空或省略 `parent-id`
可以列出对应 Volume 的直接条目。筛选和分页时始终保留同一 `volume_id` 与 `parent_id`：

```text
moi-cli -o -no-interactive -workspace-id "$workspace_id" catalog files --volume-id "$volume_id" --parent-id "" --name "$name_fragment"
moi-cli -o -no-interactive -workspace-id "$workspace_id" catalog files --volume-id "$folder_id" --parent-id "" --extensions txt
moi-cli -o -no-interactive -workspace-id "$workspace_id" catalog files --volume-id "$volume_id" --parent-id "" --file-ids "$file_id"
moi-cli -o -no-interactive -workspace-id "$workspace_id" catalog files --volume-id "$volume_id" --parent-id "" --page 2 --page-size 1
```

2026-11-14 QA 实测中，`--volume-id 根ID --parent-id 子folderID` 和
`--volume-id 子folderID --parent-id 子folderID` 都返回 `ErrParamInvalid/status400`；改为子 folder
自己的 `volume-id` 加空或省略的 `parent-id` 后返回其中真实文件。folder 列表项的 `parent_id` 是
父子关系输出，不代表当前输入一定接受非空 parent 筛选；不要沿用未实测的描述猜组合。

`--file-type file` 与 `--file-type folder` 已在本次 QA 分别筛出根 CSV 和子 folder。当前 reference
还暴露 `--order-by` 和 `--order`，可排序字段与方向仍以当前服务端正式合同为准，不能凭名称猜测。
名称是子串筛选，不是唯一键；`file-ids` 可以缩小范围，仍须
核对返回的 placement、Volume、parent 和文件类型。文件夹条目存在不等于已知其后代，进入子文件夹
要用它自己的真实 Volume ID 再查。

`page/page-size` 是页号分页，不是 `file list` 的 token 分页。翻页时保持全部筛选与排序条件不变，
保存每页条数、total（若返回）和实际 file/folder ID；首屏非空、total 或 `volume-info` 数量都不能
单独证明已经遍历完整。筛不到时先区分“该条件下无匹配”与“Volume/文件夹为空”。

`volume-info` 汇总和当前目录条目列表的统计范围可能不同。2026-11-14 样例根 Volume 汇总为
3 个文件、230 字节（两个根 CSV 加一个子目录文本），导入前根目录列表 `total=2` 则是 folder 与
根 CSV，不含子目录文本。不能只比较两个数字是否相等；应明确是否递归、是否计 folder，并用真实
file ID 集合、每个子 Volume 列表和字节数解释差异。

- 实际所选 Volume 明确返回文件数量为 0 或明确未就绪时，停止批处理提交，并报告“输入为空/未就绪”；
  缺字段、空响应、不可用的列表入口不等于文件数为 0。
  不得把 Workflow 已提交、历史执行成功或知识库已有 source 说成当前处理成功。
- 文件数量和列表不一致时，先重新读取原 Volume；不能按名称猜文件，也不能自动选择另一个 Volume。
- 文件数量非零只证明有输入，不证明文件解析、Workflow、source ingest、embedding 或问答已完成；
  后续仍需沿真实 file ID 做文件级核验。
- 深度核对可按 `volume-info` → `file-info` → `preview-file`/`download-file` → Workflow/Knowledge
  结果的顺序保存证据。若输入文件没有派生产物，`workflow catalog-file-lineage-overview`
  返回 `output_artifact_not_found` 只表示没有输出 artifact，不应被当作输入文件读取失败。
- 输出文件进入 Knowledge 前，先用真实 file ID 下载或 `file-data-asset`/lineage 查询核对，
  再把该 file ID 交给 semantic-model。Catalog file、workspace file、runtime artifact、
  Knowledge source 是不同作用域，禁止互换 ID。

## 递归发现未知嵌套目录

需要从数据库根定位未知嵌套文件时，使用[递归发现脚本](../../scripts/catalog_walk.py)，不要只检查根目录后
就宣称整棵子树为空。脚本只通过 dev Product CLI 逐层读取 `database-children` 和每个子 Volume 的
`catalog files`，处理页号分页、循环引用、重复分页项以及深度/节点预算：

```text
python3 scripts/catalog_walk.py \
  --database-id 130010 --workspace-id d175f639-2846-3433-6e24-43b4a53bee6b \
  --endpoint https://moi.matrixorigin.cn/newmoi --page-size 1
```

输出 `complete=true` 只表示在上述正式入口和预算内遍历完成；预算被截断时脚本直接失败，不输出完整结论。
2026-11-15 生产真实链路 `130010 → 130016 → 130017 → 130018` 已返回三级目录及嵌套文件，
强制每页 1 条时仍完整发现两个目录和两份文件，ID 无重复。另一组已有对象
`150007 → 140007 → 140011 → 140013` 也递归成功，下载 64 字节的 SHA256 与原文件一致。
根 `database-children` 只返回根 Volume，每个子 folder 必须沿自己的 Volume ID 继续读取；
当前下载也要求显式提供对应 Volume，不能只传 file ID。

同一物理文件可以在多个 placement 出现，递归结果按目录位置保留，不按 file ID 全局去重。大目录、
权限受限分支或分页不一致时保持部分/失败，不能改写成“已发现全部文件”。

## 表和结果

创建、更新、删除、truncate、导入触发和文件下载都属于写入或副作用。执行前先确认目标和
影响范围。表结构优先用 `table-info` 或 SQL metadata 查询；不要凭名称拼接 table ID。

数据库表发现可用 `database-table-leaves`；`table-overview` 与 `workspace-database-tables`
可交叉核对表名和列。对已选表，比较 `table-info` 的结构、`table-data` 的字段/记录、
`preview-table` 与 `download-table` 的实际 CSV。多表场景可用 `multi-table-info --table-ids`，
按返回表身份匹配，不能依赖数组顺序。小表下载的逐行比对见[下载核验](../recipes/export-and-verify.md)。
默认 table-data 窗口不代表任意大表已读全；有 total/page 信息时保留截断范围。

## 同文件的多处引用

同一个 `file_id` 可以挂在多个 Volume，形成不同 placement。已知 file ID 时可用：

```text
moi-cli -o -no-interactive -workspace-id "$workspace_id" file list --file-ids "$file_id" --page-size 100
```

同时保存 placement 的 `id`、`volume_id` 和物理 `file_id`。同 file ID 的多条记录不是重复文件，
枚举 placement 时不能先按 file ID 去重。`file list` 的 scope 和分页规则见
[Workspace File](catalog-runtime-files.md)，不能替代任意 Volume 的完整枚举。

用户只要求从某个 Volume 移除引用时，锁定该 Volume 与 file ID，再执行：

```text
moi-cli -o -no-interactive -workspace-id "$workspace_id" catalog delete-file-ref --volume-id "$volume_id" --file-ids "$file_id"
```

回读目标 Volume 数量和该 file ID 的 placements；保留处再显式带其 volume_id 下载并核对 hash。
2026-11-01 隔离样例中两处各一文件，移除 A 后 A=0、B=1，B 下载字节/hash 不变。
只证明本次引用隔离；不能承诺物理字节擦除、所有关联权限或知识来源自动更新。
不要用 `delete-file`、`clean-folder`、删除 Volume 替代“只移除此处引用”。

空文件夹的创建/删除有本轮回执；旧版 `database-children` 曾没返回其嵌套项，不能把单次缺项解释为
创建失败。dev 当前 `update-folder` reference 已有 `--name` 和 `--description`；更新后仍须沿原
folder ID 用 `volume-info`、`catalog files --parent-id` 或其它正式详情入口回读，不能仅凭退出 0
声称改名完成。Volume 的 `update-volume --name` 同样需要保存并回读证据。

从现有 Catalog 文件创建 Knowledge source 不会自动把原文件复制进 KB 自有 raw Volume。
显式 source job 可以处理原 Volume 的文件；只有运行默认指向 KB raw Volume 的 Workflow 时，
才以该默认 Volume 做批处理输入预检。不能因 KB raw 为空而判定原 source 无法准备。

## 已知嵌套文件夹的清空

正式`create-folder --database-id --parent-volume-id --name`创建嵌套文件夹并返回ID；本地owner以CatalogVolume承载文件夹，线上已用`volume-info`和`create-file --volume-id --file-ids`核对其文件。保存完整创建父子关系，不能把其它资源同数值ID混用。

`clean-folder --folder-id`会清除该目录及其后代的文件引用，并删除后代目录，**保留目标目录本身为空**。这不是单文件移除命令。只有已授权整个已知隔离子树时才执行；未知业务子树枚举不全时先提示超出可确认范围，不凭顶层列表为空就清理。

2026-11-01已知root→A→B中，同一36字节file挂在三处。clean A后A仍可读且0文件，B为404，placements从3处变为root一处，ref_count3→1；root原placement ID及下载字节/SHA256保持。tree/database-children均未枚举A/B，因此本次证明已知子树清空与引用隔离，不证明能发现任意未知嵌套结构或物理字节擦除。首次volume-info超时沿原ID回读成功，未重建或重复挂载。

## 同名文件与重复挂载、移除

文件名及file_path不唯一。2026-11-01同一Volume两份中文/emoji同名文本分别83/82B、内容42/99箱，按真实file ID下载各自正确。将名称筛选用于发现，执行/交付/清理仍锁定file ID、placement ID和Volume ID；`file-info`只接file ID，`download-file`另可显式带Volume，不能混用旗标。

同file同Volume顺序重复挂载，以及已有引用、空目标各两个真实并发请求，本例都没有增加重复placement；空目标只新增一处，原两处ID保持，字节不变。它是同一物理file ID的引用保护，不是内容哈希去重、同名覆盖或导入防重，也不保证任意并发规模。

同一处引用双请求并发移除则一成功一404；独立回读证实仅目标引用移除，其它两处原引用和同名另一文件保持。404须结合原目标当前状态解释，不换Volume或改用物理删除凑成功。Volume改名后本例ID/placement与完整下载保持；不能据此承诺移动、原位更新或未知子树发现。

2026-11-02 的 `ErrParamInvalid/status400` 属于当时正式 0.1.8 只暴露 workspace-id 的旧输入合同，
不能继续套用到 dev。当前 dev `d5c5c3a` reference 已明确提供 Volume、parent、名称、扩展名、
file IDs、类型、页号分页和排序输入；应使用这些正式参数并回读结果，不猜未公开枚举、不切 raw HTTP。
旧版失败记录只用于版本追溯，不证明当前 dev 仍缺少 Volume 文件枚举。
