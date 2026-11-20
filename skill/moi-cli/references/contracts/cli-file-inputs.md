# 正式文件输入与证据冻结合同

`record_cli.py` 和 `task_run.py` 只冻结明确登记的本地文件参数，避免源文件后来变化导致回执与
实际输入不一致。 dev Skill 的默认程序是 `$HOME/.local/bin/moi-cli-dev`；`MOI_DEV_CLI_BIN` 或显式
`--cli-bin` 可指定另一已确认程序，不会回退到稳定版 CLI。

## 结构化 UTF-8 输入

以下文件按原字节复制，保留 JSON/SQL 后缀；结构化内容还会执行常见凭证模式拒绝：

| 用途 | 正式参数 |
| --- | --- |
| 通用 JSON/SQL/schema | `--json-file`、`--query-file`、`--validate-file` |
| Workflow 候选与图 | `--acceptance-file`、`--workflow-file` |
| Workflow 表单和值 | `--runtime-fields-file`、`--runtime-layout-file`、`--default-values-file`、`--values-file` |
| Workflow 恢复与触发 | `--config-override-file`、`--volume-trigger-file`、`--dynamic-service-file` |
| Knowledge 配置 | `--files-file`、`--tables-file`、`--source-selections-file` |

`--json` / `--query` 只转换为对应文件参数，不能把开放 JSON 当成缺少正式字段时的猜测通道。
`--dsl-yaml` 是内联 DSL 内容，不是文件路径；`--source` 是包源目录，当前不做递归目录快照。

## 原始字节输入与输出

- `--file`、`--file-path`、`--query-image` 按原始字节复制，扩展名保留；这些文件可能包含私密业务
  内容，上传和本地留存范围必须先明确。
- `--output`、`--output-file` 只登记为 CLI 管理的外部产物，目标已存在时拒绝覆盖；
  工具不声称捕获、脱敏或哈希保护这些外部产物。
- `--api-key-file`、`--runtime-token-file`、`--actor-assertion-file`、`--provisioner-token-file`
  明确拒绝经过通用记录器，避免把凭证复制进证据目录；按对应专用认证流程处理。

macOS 的 `/var`、`/tmp` 目录别名会先解析为真实目录，再逐段以 `NOFOLLOW` 打开。最终输入文件
仍必须是普通文件，符号链接、设备和管道不被接受。目录可移植性不等于允许最终文件使用链接。

可用 `python3 scripts/cli_contract.py file-inputs` 查看当前注册表，用
`python3 scripts/cli_contract.py validate` 对照当前 dev CLI help 和缺口清单。未登记且名称以
`-file` 或 `-file-path` 结尾的参数会在启动前拒绝，避免“命令执行了但输入没留证”。
