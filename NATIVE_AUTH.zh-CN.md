# 使用 Astra 的 MOI 统一登录

默认使用 Astra 原生用户登录；既有 PAT 用法需要显式选择 `--auth pat`。需要安装支持本认证合同的配套 Astra / moi-cli；关闭 Astra 对话进程后，moi-cli 仍可使用这次登录。

## 使用方式

先通过 `astra login` 登录受信任的 MOI Astra 服务；首次地址由安装/部署配置提供，或显式使用 Astra 的全局 `--api-url` / `ASTRA_API_URL`。不要使用未经确认的服务地址。UC 浏览器页负责 GitHub / Google / Email 登录，CLI 不收集这些提供方的密码。

```bash
astra login
astra auth status --json
moi-cli --auth moi -o workspace list
astra auth workspace
moi-cli --auth moi -o catalog list
astra logout
```

workspace 初始化是后台任务，未选择或尚未就绪不代表登录失败。`astra auth workspace` 显式选择后，工作区操作仍由服务端 Core IAM 验权；本地选择不授予权限。

工作区角色仍由 `moi-cli config set-role <workspace-id> <role-id>` / `unset-role` 管理，保存在 `~/.moi-cli.json`，PAT 和统一登录共用同一套角色配置。命令开始时校验并冻结全部角色映射；配置非法时不发送请求，也不静默改用默认角色。统一登录的服务地址、账号和默认工作区来自 Astra，不使用旧配置中的 endpoint、PAT 或默认工作区。显式传入的 `--endpoint` 必须与登录环境一致，否则拒绝请求，不会将登录凭据发往其他地址。

## 参数和环境

| 配置 | 行为 |
| --- | --- |
| `--auth auto`（默认） | 使用 Astra 统一登录；尚未登录时提示登录，不自动回退旧 PAT |
| `--auth moi` | 明确要求统一登录；缺少 Astra、退出、凭据损坏或过期时失败，不回退 PAT |
| `--auth pat` | 显式使用既有 PAT 模式 |
| `MOI_AUTH_DIR` | 两端共同使用的认证目录，必须为绝对路径；默认 `~/.moi` |
| `MOI_ASTRA_BINARY` | 受信任的配套 Astra binary 绝对路径；未设置时从 PATH 查找 `astra` |

统一状态保存在私有 `auth.json`，moi-cli **不自行读取 Token 内容或刷新**；它通过继承的私有管道向短命 Astra helper 取一次请求所需的授权。凭据不进入进程参数、标准输出或错误输出。刷新、原子保存、账号切换和撤销只有 Astra 一个 owner。

命令启动后冻结账号、环境和登录 generation；中途切换账号不会让同一操作悄悄访问另一个账号。不要把手动指定的其他 endpoint、旧 PAT 和新登录混用。退出或删除登录状态后，`auto` 也不会重新启用遗留 PAT。

即便已有 PAT，未安装 Astra 时 `auto` 仍会提示安装 Astra；确定要使用 PAT 时显式运行 `moi-cli --auth pat ...`。

认证 helper 每次请求最多等待 25 秒；随后业务 HTTP 请求有独立的 30 秒预算（含响应体读取）。父调用取消仍会同时终止等待。每页请求仍检查登录 generation，不缓存可跨退出或切换账号使用的授权。

## Astra 或其他 agent 调用

让 agent 进程继承相同的 PATH、`MOI_AUTH_DIR` 和可选的 `MOI_ASTRA_BINARY`，它即可执行普通 moi-cli 命令。不要把 auth.json、Token 或 Cookie 写入 prompt、Skill、环境变量或项目文件。Astra 的 Genesis 凭据留在服务端；Codex 等第三方 agent 的模型服务不由这条链路提供。

本改动没有新增 MCP Server。macOS/Linux × arm64/amd64 的构建、Skill 交付与配套安装工具见 [release/README.md](release/README.md)。个人隔离环境的跨项目联调脚本不包含在产品交付中；正式协议与安装回归测试保留在各自模块。
