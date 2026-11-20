# dashboard command reference

Generated from CLI source. Do not edit. Business requirements remain server-owned.
Use `moi-cli schema DOMAIN COMMAND` for typed JSON input and `--help` for defaults.

## dashboard chart-result

- `--chart-id`: chart-id
- `--dashboard-id`: dashboard-id
- `--query-param`: Product query parameter name=value, including at=<RFC3339Nano> (repeatable)
- `--workspace-id`: workspace-id

## dashboard connect-web-socket

- `--action`: ping、preview、create-chart 或 update-chart
- `--chart-id`: 图表 ID
- `--dashboard-id`: 看板 ID
- `--json`: 图表请求 JSON
- `--json-file`: 图表请求 JSON 文件
- `--message-id`: 本次连接中的消息 ID
- `--sql-text`: 预览 SQL
- `--timeout`: 连接、发送及读取的总期限
- `--workspace-id`: 工作区 ID，省略时读取配置

## dashboard create

- `--json`: json
- `--json-file`: json-file
- `--workspace-id`: workspace-id

## dashboard create-chart

- `--dashboard-id`: dashboard-id
- `--json`: Request JSON; inspect chart_config with: moi-cli schema dashboard create-chart --argument json.chart_config; layout_config with --argument json.layout_config
- `--json-file`: json-file
- `--workspace-id`: workspace-id

## dashboard delete

- `--dashboard-id`: dashboard-id
- `--workspace-id`: workspace-id

## dashboard delete-chart

- `--chart-id`: chart-id
- `--dashboard-id`: dashboard-id
- `--workspace-id`: workspace-id

## dashboard download-chart-result

- `--chart-id`: chart-id
- `--dashboard-id`: dashboard-id
- `--output`: output
- `--query-param`: Product query parameter name=value, including at=<RFC3339Nano> (repeatable)
- `--workspace-id`: workspace-id

## dashboard download-drill-through

- `--chart-id`: chart-id
- `--dashboard-id`: dashboard-id
- `--json`: json
- `--json-file`: json-file
- `--output`: output
- `--workspace-id`: workspace-id

## dashboard drill-through

- `--chart-id`: chart-id
- `--dashboard-id`: dashboard-id
- `--json`: json
- `--json-file`: json-file
- `--workspace-id`: workspace-id

## dashboard generate-sql

- `--dashboard-id`: dashboard-id
- `--question`: question
- `--title`: title
- `--workspace-id`: workspace-id

## dashboard get

- `--dashboard-id`: dashboard-id
- `--workspace-id`: workspace-id

## dashboard list

- `--json`: json
- `--json-file`: json-file
- `--workspace-id`: workspace-id

## dashboard preview-sql

- `--chart-id`: Optional chart ID for persisted SQL/database scope
- `--dashboard-id`: dashboard-id
- `--sql-text`: sql-text
- `--workspace-id`: workspace-id

## dashboard refresh

- `--dashboard-id`: dashboard-id
- `--workspace-id`: workspace-id

## dashboard run-chart

- `--chart-id`: chart-id
- `--dashboard-id`: dashboard-id
- `--workspace-id`: workspace-id

## dashboard task-execution-overview

- `--workspace-id`: Workspace ID (default: from config)

## dashboard update

- `--dashboard-id`: dashboard-id
- `--json`: json
- `--json-file`: json-file
- `--workspace-id`: workspace-id

## dashboard update-chart

- `--chart-id`: chart-id
- `--dashboard-id`: dashboard-id
- `--json`: Request JSON; inspect chart_config with: moi-cli schema dashboard update-chart --argument json.chart_config; layout_config with --argument json.layout_config
- `--json-file`: json-file
- `--workspace-id`: workspace-id

## dashboard update-chart-order

- `--dashboard-id`: dashboard-id
- `--json`: json
- `--json-file`: json-file
- `--workspace-id`: workspace-id

