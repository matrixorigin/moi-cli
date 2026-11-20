# moi-cli 使用与更新指南

本指南介绍 MOI AI Studio 的 `moi-cli`。使用前可运行 `moi-cli -version` 查看已安装版本，运行 `moi-cli --help` 查看该版本支持的命令。

## 1. 它是什么

`moi-cli` 是在本地终端运行的 MOI 产品客户端，适合交互式操作、脚本和自动化调用。它通过 Product Go SDK 访问 MOI 产品 API。使用远程 MOI 环境时，本机只需客户端和网络连接，无需部署整套 MOI 服务；从源码编译时另外需要 Go 工具链。

它支持查询和修改远端资源。创建、删除、运行任务等命令会实际作用于配置的 MOI 环境。

## 2. 当前功能

下表列出常用命令组。运行 `moi-cli --help` 查看完整列表。

| 命令组 | 主要用途 |
| --- | --- |
| `config` | 配置服务地址、PAT、默认 workspace 和各 workspace 使用的 Role |
| `workspace` | 创建、查询、列举、删除工作区，以及获取连接信息等 |
| `workspace-member` | 工作区成员管理 |
| `catalog` | Catalog、数据库、表、Volume、文件及目录树等资源操作 |
| `workflow` | 工作流部署、修改、运行、暂停、恢复和运行记录查询 |
| `sql` | 执行查询、获取结果、取消查询、查询历史、元数据和 workbook 管理 |
| `agent` | 调用 Agent、流式调用、任务查询和取消、Agent/Skill 包及 channel/callback 操作 |
| `account` | 获取账号数据库连接信息 |
| `semantic-model` | 知识库相关查询和管理 |
| `iam` | Role、绑定、策略和权限相关产品 API 操作 |
| `model` | 模型列表、LLM 后端设置和模型探测 |
| `provider-backend` | LLM、embedding、parser 后端及 endpoint/router 管理 |
| `connector` | 数据源、文件上传下载、导入任务和导出任务管理 |
| `compute` | 计算资源、规格、运行状态、指标、暂停和恢复等 |
| `dashboard` | 任务执行概览 |
| `data-share` | 数据共享相关产品 API 操作 |

命令是否可成功执行还取决于目标服务版本、环境配置及当前用户权限。CLI 中存在某个命令，不代表所有用户都可执行该操作。

查看具体参数：

```bash
# 工作目录：任意目录；moi-cli 已加入 PATH
moi-cli --help
moi-cli catalog --help
moi-cli sql execute-query --help
moi-cli agent send --help
```

支持 `-o` 输出 JSON、`-no-interactive` 关闭交互提示；复杂参数和流式输出的选项以各子命令帮助为准。当前没有独立的顶层 `file` 或 `volume` 命令，应从 `catalog`、`connector` 等命令组查找对应能力。

### 知识库来源参数与定时调度

`create-with-sources` 和 `append-sources` 的 `--json` 接受来源数组，字段使用 SDK 的 `SourceType`、`FileID` 等名称。`SourceType` 合法值为 `local_file`、`catalog_file`、`catalog_table`，没有 `file`。两条命令的 `--help` 和 `schema` 都提供字段组合与可直接使用的数组示例：

```bash
# 工作目录：任意目录；moi-cli 已配置目标环境和 workspace
moi-cli schema semantic-model append-sources
moi-cli semantic-model append-sources --help
moi-cli semantic-model upload-local-file --file-path ./manual.md
# 用上传结果中的 file_id 替换 uploaded-file-id
moi-cli semantic-model append-sources --model-id 1 \
  --json '[{"SourceType":"local_file","FileID":"uploaded-file-id","FileName":"manual.md"}]'
```

`local_file` 必须有 `FileID`、`FileName`；普通文档可省略 `UploadKind`，按 `unstructured` 处理。`structured` 导入必须提供 JSON 字符串形式的 `TableConfig`。`catalog_file` 引用已有 Catalog 文件，共享文件应同时提供所在 `VolumeID`；`catalog_table` 使用正整数 `TableID`。请求返回后继续用 `list-sources`、`list-source-jobs` 确认解析与索引结果。

已保存工作流可直接修改调度配置，无需重新提交候选定义。五字段顺序是分钟、小时、日、月、星期；六字段在最前面增加秒。省略时区时使用 `Asia/Shanghai`，可通过 `CRON_TZ=` 显式指定：

```bash
# 工作目录：任意目录；用实际工作流 ID 替换 workflow-id
moi-cli workflow update --workflow-id workflow-id --execution-mode cron \
  --cron-expression 'CRON_TZ=Asia/Shanghai 0 7 * * *'
moi-cli workflow get --workflow-id workflow-id
# 恢复手动运行
moi-cli workflow update --workflow-id workflow-id --execution-mode one_shot
```

上述表达式表示上海时间每天 07:00；对应六字段为 `CRON_TZ=Asia/Shanghai 0 0 7 * * *`。检查返回的 `execution_mode`、`cron_expression` 和 `trigger_summary` 确认保存状态。修改已接受候选的流程定义仍须使用候选审核与部署路径。

## 3. 如何工作

```text
本地命令和 ~/.moi-cli.json
          ↓
moi-cli：解析参数和配置
          ↓
Product Go SDK：构造产品 API 请求
          ↓ HTTP(S)
moi-backend / 产品网关：认证、服务编排
          ↓
Core 及对应业务服务：权限校验、资源操作或任务执行
          ↓
响应经 SDK 返回，在本地终端输出
```

每次启动业务命令时，CLI 默认通过 Astra 获取共享登录的 endpoint、Token 和 workspace，校验持久化 Role mapping，再创建 Product SDK 客户端。只有显式 `-auth pat` 时才读取旧配置中的连接信息。普通请求默认超时 30 秒，图片分析使用其命令预算；可在命令组之前显式指定正数 `--request-timeout`，该期限覆盖请求和响应体读取，统一用于共享登录、PAT 和沙箱任务身份，不替代服务端或外层工具的期限。

```bash
# 工作目录：任意目录；SQL 请求已保存到本地 JSON
moi-cli --request-timeout 2m semantic-model query --json-file query.json --output query-result.json
```

PAT（Personal Access Token，个人访问令牌）标识用户身份。workspace 指定操作的工作区；Role 配置表示希望使用的角色。角色是否有效、操作是否允许由服务端验证，客户端配置本身不会赋予权限。

CLI 本地只负责参数、配置、请求和结果呈现。SQL、工作流和 Agent 的业务执行依赖服务端；服务端部署新版本不会自动改写用户电脑里的 executable。

实现入口见 [main.go](cmd/moi-cli/main.go)、[命令实现](commands/) 和 [AI Studio SDK 快速开始](https://docs.matrixorigin.cn/moi/zh/5.0/developer/quickstart/ai-studio/sdk.html)。

## 4. 如何安装和更新

### 4.1 下载或从源码构建

请按 [README 的安装说明](README_zh-CN.md#环境要求与安装) 下载和校验目标版本，或使用对应版本的源码构建。运行预编译的二进制只需要匹配操作系统和 CPU 架构。从源码构建需要 Go 1.25.7 或更高版本。

安装或重新构建以后，将二进制放到 `PATH` 中的目录，再确认路径和版本。

```sh
command -v moi-cli
moi-cli -version
```

### 4.2 平台和架构

默认构建使用当前 Go 环境的目标平台和架构，一次生成一个 binary。可以通过 `GOOS`、`GOARCH` 指定交叉编译目标：

| 目标 | `GOOS` | `GOARCH` |
| --- | --- | --- |
| macOS Apple Silicon | `darwin` | `arm64` |
| macOS Intel | `darwin` | `amd64` |
| Linux ARM64 | `linux` | `arm64` |
| Linux x86-64 | `linux` | `amd64` |

需要交叉编译时，在 README 对应的构建命令前设置目标 `GOOS`、`GOARCH` 和 `CGO_ENABLED=0`。按平台使用不同的 `-o` 输出路径，避免覆盖上一份二进制。

### 4.3 更新已安装的 CLI

- **Astra 管理的安装**。运行 `astra update check` 检查更新，再运行 `astra update` 更新配套组件。
- **手动安装**。下载与部署匹配的新版本并校验，或按 README 重新构建，然后替换 `command -v moi-cli` 指向的二进制。更新源码或单独完成构建都不会替换已安装的副本。

更新后，再次运行 `moi-cli -version`，确认实际运行的版本与下载或构建的版本一致。CLI 配置独立保存在 `~/.moi-cli.json`，替换二进制时可以保留。

## 5. 用户首次配置

推荐运行 `astra login`，完成浏览器登录后直接运行 `moi-cli workspace list`；默认工作区由 `astra auth workspace` 选择。无需手动配置 endpoint 或 PAT，已有旧连接配置不会覆盖共享登录。详情见 [统一登录指南](NATIVE_AUTH.zh-CN.md)。

以下是显式 PAT 模式的配置方式：准备目标环境的产品 API 地址和该环境可用的 UC PAT，并在业务命令中添加 `-auth pat`。PAT 需要通过目标环境的账号体系取得。

### 5.1 设置服务地址和身份

以下占位值需要替换为真实值：

```bash
# 工作目录：任意目录；moi-cli 已加入 PATH
moi-cli config set-endpoint 'https://<MOI服务域名>/newmoi'
moi-cli config set-api-key '<你的UC个人访问令牌>'
moi-cli config show
```

endpoint 应包含环境实际的产品 API base path。未配置时默认是 `http://127.0.0.1:8050/newmoi`；使用远端 MOI 时必须设置远端地址。

配置文件包含 PAT 原文；新建文件时使用 `0600` 权限。`config show` 只显示固定掩码，不输出 PAT。命令行传入的 PAT 可能进入 shell 历史，应按凭据处理，避免分享配置文件或包含令牌的命令记录。

### 5.2 选择工作区

```bash
# 工作目录：任意目录
moi-cli -auth pat workspace list
moi-cli config set-workspace '<workspace-id>'
moi-cli -auth pat catalog list
```

从工作区列表取得要操作的 ID。保存后，需要 workspace 的命令会使用该默认值；也可以单次指定：

```bash
# 工作目录：任意目录
moi-cli -auth pat -workspace-id '<另一个workspace-id>' catalog list
```

### 5.3 可选：指定工作区角色

未配置 Role 时，使用服务端 Default Role 的选择和验证流程。需要指定角色时，保存该 workspace 的 Role ID：

```bash
# 工作目录：任意目录；role-id 是用户在该 workspace 中已绑定的角色 ID
moi-cli config set-role '<workspace-id>' '<role-id>'

# 移除显式选择，后续命令使用服务端 Default Role 流程
moi-cli config unset-role '<workspace-id>'
```

Role ID 必须是规范的正十进制 `int64`，不能填写角色显示名称。配置不等于授权，服务端仍会校验绑定和生命周期；显式角色无效时不会回退到其他角色。

每个 CLI 进程启动时读取一次 Role 配置快照，变更对之后启动的命令生效。当前没有全局 `-role-id` 临时覆盖参数。

任意一条持久化 Role mapping 格式非法，都会阻止业务命令发请求，并提示 `No request was sent.`。使用 `config show` 查看，再用 `set-role` 或 `unset-role` 修复。详情见 [CLI 配置与身份认证](https://docs.matrixorigin.cn/moi/zh/5.0/developer/cli/configuration-authentication.html#配置工作区角色映射)。

### 5.4 检查连接及脚本使用

```bash
# 工作目录：任意目录；已配置 endpoint、PAT 和 workspace
moi-cli config show
moi-cli -auth pat workspace list
moi-cli -auth pat catalog list
moi-cli -auth pat sql execute-query --query 'SELECT 1'

# JSON 输出并关闭交互提示
moi-cli -auth pat -o -no-interactive catalog list
```

`config show` 只读取本地配置，成功不代表服务端连接或认证已通过。后续命令才会访问服务端；SQL 示例还依赖目标环境的 SQL 服务和权限。

全局参数放在命令组前，例如 `moi-cli -o catalog list`。endpoint、PAT 和 workspace 支持全局参数及配置文件；当前入口没有通用的 `MOI_*` 环境变量配置读取逻辑。另一个实现细节是：显式传入默认 endpoint `http://127.0.0.1:8050/newmoi` 时，仍可能被配置文件的 endpoint 覆盖；要切回该本地地址，可以使用 `config set-endpoint` 保存。

## 6. 常见排查

| 现象 | 检查项 |
| --- | --- |
| 找不到 `moi-cli` | 安装目录是否在 PATH；用 `command -v moi-cli` 检查 |
| 更新代码后命令没变化 | 是否重新构建并安装；是否运行了另一个路径下的旧 binary |
| unknown command / 参数不认识 | 查看当前 binary 的帮助；用 `moi-cli -version` 确认所用版本 |
| 连接失败或请求路径不正确 | endpoint、产品 base path 和本机到服务端的网络连通性 |
| 提示缺少 api-key | 通过 `config set-api-key` 保存目标环境 PAT |
| 提示认证失败 | PAT 是否有效且适用于目标环境；结合服务端返回信息排查 |
| 权限拒绝 | workspace、用户成员身份和当前 Role；保存配置不会新增权限 |
| `No request was sent.` | 通过 `config show` 检查非法 Role mapping 并修复 |

本文基于代码和已有契约核对，未连接真实 MOI 环境执行上述业务示例。
