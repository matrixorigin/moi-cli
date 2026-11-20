# moi-cli

[English](README.md)

`moi-cli` 是 MatrixOne Intelligence（MOI）5.x AI Studio 的命令行工具，适用于交互操作与脚本自动化。它通过 Product Go SDK 调用产品 API，沿用服务端的工作区和角色权限，覆盖工作区、Catalog 和文件、连接器与数据传输、工作流、知识库、SQL、智能体、模型、计算资源、任务监测和数据共享。

## 环境要求与安装

- **[GitHub Release v5.2.3](https://github.com/matrixorigin/moi-cli/releases/tag/v5.2.3)。** 确认版本与部署匹配后，下载对应系统和架构的 `moi-cli-<版本号>-<系统>-<架构>.tar.gz` 及 `SHA256SUMS`。系统为 `linux` 或 `darwin`（macOS），架构为 `amd64` 或 `arm64`；例如 macOS Apple 芯片使用 `moi-cli-<版本号>-darwin-arm64.tar.gz`。用 `SHA256SUMS` 中对应文件的 SHA-256 值校验下载的归档，解压后将 `moi-cli` 放入 `PATH`。文件名中的版本号不带 Release 标签的 `v` 前缀。
- **旧版离线包。** 参照[官方安装和更新指南](https://docs.matrixorigin.cn/moi/zh/5.0/developer/quickstart/downloads/cli.html)，或向管理员获取与 MOI 部署、操作系统和 CPU 架构匹配的版本。官网目前提供 macOS arm64 的 `0.1.0` 离线包、SHA-256 校验值和安装步骤。
- **由 Astra 管理的安装。** 使用部署方提供的 Astra / MOI 配套安装器。安装后，可运行 `astra update check` 检查更新，再用 `astra update` 更新。手工安装的用户可用匹配的新版本替换二进制。
- **从源码构建。** 需要 Go 1.25.7 或更高版本。在此仓库根目录使用随附依赖构建：

```sh
go build -trimpath -buildvcs=false -ldflags '-X main.version=5.2.3' -o moi-cli ./cmd/moi-cli
./moi-cli -version
```

以上构建命令仅用于对应 v5.2.3 的源码。安装后，验证 `PATH` 中的二进制应显示 `moi-cli version 5.2.3`：

```sh
moi-cli -version
moi-cli --help
```

## 认证与最小示例

**个人访问令牌（PAT）。** 在 MOI 控制台创建 AI Studio PAT 并传入原始值，无需添加 `Bearer` 前缀。服务地址须包含完整路径前缀。公有云使用下例地址，私有部署使用管理员提供的地址。请在受保护的终端配置令牌，命令参数可能出现在终端历史或进程列表中。

```sh
moi-cli config set-endpoint 'https://api.moi.matrixorigin.cn/v5'
moi-cli config set-api-key 'YOUR_PERSONAL_ACCESS_TOKEN'
moi-cli -auth pat -o -no-interactive workspace list
```

这条只读命令以 JSON 返回当前账号可见的工作区。选取一个 ID 并保存，再查询该工作区的资源：

```sh
moi-cli config set-workspace 'YOUR_WORKSPACE_ID'
moi-cli -auth pat -o -no-interactive workflow list
```

PAT 模式默认使用 `~/.moi-cli.json`。可用 `MOI_CLI_CONFIG_FILE` 指定规范的绝对路径。配置以 `0600` 权限写入，文件中含有令牌原文，请勿分享。`moi-cli config show` 只检查本地配置，令牌以掩码显示。

使用 `config set-role <workspace-id> <role-id>` 选择已绑定的角色，使用 `config unset-role <workspace-id>` 恢复服务端的 Default Role。Role ID 必须是规范的正十进制 `int64`。任一已保存的映射无效都会阻止业务请求。用 `config show` 定位并修复。详见[配置与身份认证](https://docs.matrixorigin.cn/moi/zh/5.0/developer/cli/configuration-authentication.html)。

**Astra 统一登录。** 安装匹配的 Astra 客户端后，先登录可信的目标环境。普通本地命令默认使用 `-auth auto`，采用 Astra 的服务地址、身份和工作区，即使尚未登录也不会自动改用已保存的 PAT。显式传入的端点必须与登录环境一致。受管智能体沙箱提供 `MOI_RUNTIME_AUTHORIZATION` 时，CLI 使用注入的任务凭据。

```sh
astra login
moi-cli workspace list
```

## 自动化与命令查询

将全局选项放在命令名前。`-o` 请求结构化命令输出 JSON，`-no-interactive` 关闭交互提示。读取输出前请检查退出状态，成功为零，失败为非零。全局 `--request-timeout` 限制单次 HTTP 请求及响应体读取。对异步操作，保留任务或运行 ID 并查询最终状态。`agent a2a --no-wait` 可在受理时返回。超时或状态不明时先查询原任务，再决定是否重试。详见[非交互模式与自动化](https://docs.matrixorigin.cn/moi/zh/5.0/developer/cli/json-output-automation.html)。

帮助、类型化输入说明和内嵌 Skill 导出均可离线使用：

```sh
moi-cli workflow list --help
moi-cli schema connector create-task
moi-cli skill export --dir ./moi-cli-skill
```

`schema <domain> <command> --validate-file <path>` 只离线校验支持的 JSON 输入结构。服务端仍会校验业务参数、资源及权限。`skill export` 的目标目录须尚未存在。

在支持 `--file-path` 的上传命令中，请明确选择普通文件，为含空格的路径加引号。该参数与 `--json` / `--json-file` 互斥，总大小默认上限为 64 MiB，可用 `--max-upload-bytes` 调整。

智能体包的 `agent init`、`agent verify`、`agent build` 和 `agent unpack` 也可在本地执行。通过服务加载或导出智能体包需要认证。

## 版本兼容性与许可证

`moi-cli` 5.2.3 面向 MOI 5.2.3 的 AI Studio API（`/v5`）。使用其他 MOI 版本时，请按部署说明选择匹配的制品。旧离线包的能力覆盖可能不同。 `core-cli` 和旧版 Catalog SDK 面向 Catalog 服务。

MOI CLI 采用 [Apache License 2.0](LICENSE)。依赖和独立分发的组件保留各自的许可证与版权声明。

## 文档

- [AI Studio CLI 快速开始](https://docs.matrixorigin.cn/moi/zh/5.0/developer/quickstart/ai-studio/cli.html)
- [命令参考](https://docs.matrixorigin.cn/moi/zh/5.0/developer/cli/command-reference.html)
- [使用指南](USAGE.zh-CN.md)和[Astra 认证](NATIVE_AUTH.zh-CN.md)
