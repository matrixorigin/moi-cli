# 开始任务前检查环境

在首次远程操作前运行一次；CLI、身份、endpoint、workspace 或认证通道变化后重新检查。
同一任务中条件未变时复用检查结果，不逐条命令重复检查。纯离线读文档不要求联网。

## 本机检查

从实际加载的 `moi-cli` Skill 根目录执行：

```text
python3 scripts/preflight.py
```

检查 Skill 版本和 references 指纹、CLI 可执行性/版本/必需全局选项、Astra helper 路径及
非敏感登录状态。输出为 JSON；`local_ready` 只表示本机检查通过，线上层仍为 `not_checked`。
Python 3 本身缺失时，在终端报告缺少依赖，不继续执行；解压包的 `scripts/doctor.sh`
会先检查 Python，再调用同一份 helper，并额外核对包版本。

程序路径按实际执行方式确定：

- `moi-cli`：使用 `MOI_DEV_CLI_BIN`，否则使用 `$HOME/.local/bin/moi-cli-dev`。
  不悄悄改用 PATH 中另一个 CLI；需要自定义位置时显式设置 `MOI_DEV_CLI_BIN`。
- Astra：正式 CLI 使用 `MOI_ASTRA_BINARY`，否则在 PATH 查找 `astra`。
  `ASTRA_BIN` 只是旧准备脚本使用的变量，单独设置它不能改变 CLI 的认证 helper。
  两者不一致时报 `astra_helper_path_mismatch`；由操作者核对后，在当前终端设置
  `export MOI_ASTRA_BINARY="$ASTRA_BIN"`，后续检查和业务命令沿用相同环境。

预检查不自动改 PATH、配置或登录态，不调用 credential 导出入口，也不读取凭据文件。
正式 CLI 请求仍由 Astra 按自身协议管理现有凭据；不能把正常请求解释成重新登录。

## 只读联网检查

endpoint 必须是当前会话已经确认的目标地址，不从含密钥的配置文件推测，也不默认连接生产。
workspace 必须使用当前会话已选择并发现的真实 ID；缺少时可以只发现工作区列表，
但报告仍保留 `workspace_required_no_auto_selection`，不能自动选择第一项。

```text
python3 scripts/preflight.py --online --endpoint '<已确认的 Product API 地址>' --workspace-id '<已选择的真实 workspace ID>'
```

普通 native 任务默认 `--auth native`。确实需要自动化 API 时，改用 `--auth pat`；
需要同时核验两条通道才用 `--auth both`。不接收 PAT 参数，不从聊天收集 key。
PAT-only 检查不要求 Astra 登录；两条通道互不替代，一条失败不阻止另一条被检查。

联网检查仅调用正式 CLI 的 `workspace list` 和目标工作区的
`iam current-principal-capabilities`，显式传 endpoint、认证通道和 workspace。
先核对工作区是否在真实列表中、是否处于可用状态，再检查权限详情读取入口。
不使用 `workspace get` 的本地句柄结果代替真实远程访问证明。

默认每个远程调用最多 90 秒，为 Astra 的有界凭据刷新留出时间；不要为提速随意缩短
真实 native 请求的超时。超时返回未知原因，不自动重试、重新登录或切换身份。

## 如何读结果

| 结果 | 人话解释和处理 |
| --- | --- |
| `local_ready` | 本机工具检查通过，线上可用性还没验证 |
| `ready_for_read_checks` | 所选认证、工作区发现及权限详情读取入口通过；继续按目标领域核对业务权限 |
| `needs_attention` | 查看各层及 `checks` 中的原因，不能把整件事归成“没登录” |
| `endpoint_required` | 缺少已确认的目标地址，先补地址；没有发起远程请求 |
| `workspace_required_no_auto_selection` | 尚未选择工作区，不能进行工作区内操作 |
| `login_required_or_expired` | 本机 Astra 状态提示登录需要处理；只有确实失效才让操作者本机登录 |
| `native_credential_unavailable` | CLI 没拿到 native 凭据；先核对 helper 路径及状态，不能仅凭这一项断言密码失效 |
| `pat_missing` | PAT 通道缺少配置，重新 Astra 登录不能补齐 |
| `authentication_rejected` / `permission_denied` | 分别为请求认证被拒绝、请求权限被拒绝；不自动增加权限或切换账号 |
| `role_configuration_invalid` / `role_configuration_conflict` | 当前角色配置不合法或冲突；按正式权限入口核对角色，不把它当登录失效，也不自动改角色 |
| `workspace_not_in_visible_list` | 所选 ID 未出现在本次列表中，核对环境、身份和目标；不直接断言对象不存在 |
| `cli_contract_missing` | 当前 CLI 缺少所需命令或参数，不能猜参数或绕过 CLI |
| `cli_response_decode_failed` | CLI 无法解析服务端响应；认证可能已通过，不能把它说成登录失效 |
| `network_or_tls_failed` / `timeout` | 连通或超时问题，尚不能判定业务能力是否可用 |
| `cli_call_failed` | 原因尚未确认，按领域故障排查，不能自行编造根因 |

各层分别保留 `native_api`、`pat_api`、两条通道的 workspace 与 permissions 状态。
若工作区发现通过、权限详情解码失败，报告“能发现目标工作区，权限详情未确认”；
不要否定已证实的读取能力，也不要由此声称写入、运行、分享或 IAM 修改已获准。
只有缺口会影响下一步时才暂停对应步骤，其余步骤遵循原任务授权和领域验收合同。

检查结果只输出路径、版本、数量、客户端观察分类和层级状态，不输出原始响应、账号资料、
token、Cookie 或权限载荷。分类不代替服务端业务错误原因；需要进一步调查时沿用领域证据流程。
PAT 可读性不证明其与 Astra 为同一账号，也不独立证明自动化 API 实际运行成功。
