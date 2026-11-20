# dev CLI 兼容合同

本 Skill 以最新 dev CLI 为执行基线，对外名称统一为 `moi-cli`；不要回退旧版 CLI 或旧 Skill 代替本版能力。

- dev 源码基线：`d5c5c3a71515364a519f4ca5ac9c63d0baf07f1d`。
- 开发二进制自报版本：`0.1.0-dev.d5c5c3a`，不是正式发布版。
- 默认可执行文件：`$HOME/.local/bin/moi-cli-dev`。
- 稳定版继续使用 `$HOME/.local/share/moi/client/bin/moi-cli` 和 `$moi-cli`。

## 动态服务字段

dev 的 `workflow invoke-dynamic-service` 和 `invoke-workflow-dynamic-service`
仍然存在，参数形式不变，但请求 JSON 使用小写：

```json
{"service_name":"实际服务名","payload":{}}
```

稳定版 0.1.8 使用 `ServiceName` / `Payload`。不要把原稳定版请求文件不加修改地交给 dev，
也不要把 dev 的小写请求文件交给稳定版。dev 导出的命令文档可能漏列这两条命令，
应以当前二进制的 `--help` 和真实结果为准。

## 已完成回归

- 原 Skill 使用的 133 个命令及参数均通过 dev `--help` 检查。
- 两条动态服务路由均返回 `COMPLETED`。
- Workflow 部署、运行、状态和结果核验通过。
- Catalog 上传、挂载、列表、详情和 SQL 查询通过。
- Knowledge 来源与索引通过，reconcile 后 `ingest_status=succeeded`。
- Dashboard、模型、Connector 和 Agent 只读检查通过。

这些结果只覆盖上述场景，不表示 dev 的全部新功能或所有历史场景均已验证。

## 0.4.2 新增能力与验收

- Workflow：新增 MOWL 新建/同 App 更新、默认值和运行值、整次与节点恢复、cron/Volume 配置恢复、
  dynamic service oneshot。实际 cron 到点和 Volume 新文件分发未验证。
- Catalog：新增 Volume 文件/文件夹筛选、分页、上传下载、CSV 入表和表/查询 CSV 导出；
  外部 Connector 导出未验证。
- Knowledge：新增 create-with-spec、来源 job reconcile、来源启停/标签治理和独立会话正式引用。
  `create-with-sources` 与带文本/图片 embedding 的 drawing spec 在本次 QA 返回 500，空 standard 的
  files 初始化返回 400，均保留为未通过，不能用 create-empty+append 的成功替代。

对应场景按 candidate 渠道登记在 `references/capability-cards.json`，并保留各场景条件、
源码版本、验证日期和外部证据位置。普通 stable regular 记录不因此升级为 dev 通过。
