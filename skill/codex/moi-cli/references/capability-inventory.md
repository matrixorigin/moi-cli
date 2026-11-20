# MOI 有限能力目录

本目录把 24 个产品菜单映射到 21 类业务目标，用于发现入口；不代表已穷尽业务语义或完成在线验证。
跨菜单归并见 [Canonical 能力目录](capability-catalog.md)，具体合同和场景状态均以[能力卡片](capability-cards.json)为源。

层级 `L0` 表示只读发现，`L1` 表示单域组织形态，`L2` 表示跨域组合，`L3` 表示受控治理；
层级不是验收状态或授权。下表的 `regular` 状态只是关联场景的计数，不能升级为菜单或能力全通过。
候选构建结果另见[覆盖矩阵](coverage-matrix.md)，不合并到常规结果。

<!-- capability-status:begin -->
来源：`capability-cards.json`（SHA256 `2619e43c50d1c1373c34d7e2fbddd1623896c9a3633857147fb11e135e7aca02`）。此段由 `scripts/capability_status.py render` 生成，修改JSON后重生成；手写说明留在标记外。

| 产品菜单 | 业务闭环 | 层级 | capability IDs | 关联能力regular场景计数 | 入口 | 限制 |
|---|---|---|---|---|---|---|
| 概览 | 固定 workspace，按用户目标进入数据、Workflow、SQL、Catalog、计算或 Agent 闭环 | L0 | workspace-context, catalog-discovery, sql-query, workflow-lifecycle, connector-lifecycle, compute-lifecycle, dashboard-analytics | passed×32, failed×7, partial×3, blocked×2, cli_missing×4 | workspace list、各域 list/tree/get、`current-principal-*` | 概览本身不是业务终态；必须归并到目标闭环 |
| 连接器 | 发现连接器、校验配置、创建或更新连接 | L1 | connector-lifecycle | blocked×1, cli_missing×1 | connector list/get/create/validate、服务端详情 | 凭证不能进入对话；不同 source type 输入差异大 |
| 导入任务 | 本地/结构化 source → staging → import task → task run → 目标表核验 | L2 | source-staging, data-ingestion | passed×7, failed×2, partial×2 | local-upload、create-task、run-task-now、task-runs、catalog/sql 核验 | 每文件独立配置的定向恢复已验；单配置多文件仍整组重放；旧task-files入口缺输入 |
| 导出任务 | 确认源和格式 → export task → task state → 产物下载/目标核验 | L2 | data-export | passed×1, blocked×1 | create-export-task、state/info/files、文件哈希或小预览 | 当前 Recipe 仍有下载边界和覆盖确认待完善 |
| 工作流 | 发现/创建/更新版本 → 手动或 Volume 批处理 → 单次/文件级状态 → 结果/产物详情 | L2 | workflow-lifecycle, job-recovery | passed×8, failed×3, cli_missing×1 | workflow list/get/create-version/update/run/list-runs/get-run/result、batch-retry-failed-files、file lineage | workflow execution、batch task、file execution 严格区分；空 Volume 不提交 |
| SQL 编辑器 | 固定表/query 输入 → execute → describe → result/download | L1 | sql-query | passed×14, failed×4, partial×1 | sql execute、describe-query、query-result、download | DDL/DML 风险升级；结果集不整体进上下文 |
| 数据看板 | 看板/chart 对象 → chart run → result/download | L2 | dashboard-analytics | passed×8, failed×1, partial×1 | dashboard list/get/create/create-chart/run-chart/chart-result | 单表钻取已验，JOIN钻取400；合法零行结果/下载500；UI外观和告警未验 |
| Catalog | Catalog/database/table/volume/file 发现、详情、Volume 非空预检和产物确认 | L1 | catalog-discovery | passed×2, partial×1, cli_missing×2 | catalog tree/list-databases/database-children/table-info/volume-info/files/file-data-asset | workspace/runtime file与Catalog ID不可互换；未知嵌套子树发现仍受限 |
| 计算资源 | compute 发现、创建、runtime、暂停/恢复、删除前置检查 | L3 | compute-lifecycle | passed×1, blocked×1 | compute list/get/runtime/preflight-delete/delete | 删除和计费/运行影响必须人工锁定范围 |
| 数据分享 | 发布对象 → 目标 workspace 订阅 → 目标侧 catalog/sql 可见 | L3 | data-share | passed×1, failed×1, partial×1 | data-share publishes/create-publish/subscriptions/subscribe + 目标侧查询 | 跨 workspace 权限与目标选择不能猜 |
| 知识库 | Catalog/Workflow file or table source → knowledge resource → source job → index → 检索与引用 | L2 | knowledge-preparation | passed×8, failed×2, partial×1, cli_missing×1 | semantic-model list/create-with-sources/upload/list-source-jobs/list-sources/get + lineage | 按 file_id 追溯上游；解析、ingest、index、检索、引用分开验收 |
| SQL 历史 | 当前主体 SQL 历史查询、详情和 profile 审计 | L0 | audit-observation | passed×1, failed×1, partial×2, cli_missing×1 | query-history/overview/query-detail/query-profile/download | 详情/profile缺start/end输入；不证明重新执行 |
| 作业 | workflow/dashboard/agent task 失败诊断、取消、重试、暂停/恢复 | L2 | job-recovery | passed×2, failed×1 | list-runs/get-run/result、task 详情、audit 或错误字段 | 不同 execution/task ID 不互换 |
| 操作日志 | 当前主体 audit log 和 overview 审计 | L0 | audit-observation | passed×1, failed×1, partial×2, cli_missing×1 | workspace-member audit-logs、account/agent 相关日志 | 日志可见不代替权限或业务完成 |
| 用户管理 | 成员发现、邀请、状态更新、删除和变更审计 | L3 | member-lifecycle | blocked×1 | manage-workspace-members Recipe、member/audit details | 邀请、禁用、删除必须明确 user ID 和影响 |
| 角色权限 | 角色/binding/grant 发现、受控变更、服务端授权核验 | L3 | iam-authorization | passed×1, failed×2, partial×1, blocked×1 | review-and-change-iam Recipe、roles/permission view/current capabilities | Core/PDP fail closed，不本地推断 allow |
| 智能体 / 对话 | agent/workspace chat/A2A 提交、task 状态和结果核验 | L2 | agent-conversation | passed×2, failed×2 | agent send/a2a/get-task，model workspace-chat-completion | 会话上下文、工具权限和流式结果需保留 ID |
| 智能体 / 成果 | workspace file/runtime artifact 发现、下载和关联对象核验 | L1 | source-staging, agent-conversation | passed×3, failed×2, partial×1 | catalog-runtime-files、file-runtime | file_path不保目录；rename/update/session筛选缺输入；删除与元数据保留分开核验 |
| 智能体 / 智能体 | agent package 构建/验证/加载后，资源绑定和运行验证 | L2 | agent-package-resource | passed×2, failed×1 | agent init/build/verify/load/resource-get、执行核验 | package 成功不等于 agent 已安装/可运行 |
| 智能体 / 技能库 | agent skill package 导入、文件下载和内容核验 | L1 | agent-skill-tool | passed×6, failed×2, blocked×1, cli_missing×1 | agent skill file/package commands | UI 技能市场与 CLI 包操作非一一对应 |
| 智能体 / 工具库 | plugin listing、安装/OAuth/连接核验 | L3 | agent-skill-tool | passed×6, failed×2, blocked×1, cli_missing×1 | plugin list/install/connection/verify 命令 | OAuth、凭证和外部连接必须由用户安全输入 |
| 智能体 / 知识库 | agent 绑定 knowledge resource后，问答引用/source refs 核验 | L2 | knowledge-preparation, agent-conversation | passed×10, failed×4, partial×1, cli_missing×1 | semantic-model、agent resource binding、send/A2A | 绑定动作和问答验收需要在线闭环 |
| 智能体 / 运行环境 | compute、model、provider backend→ agent executor/运行任务核验 | L3 | agent-runtime | passed×1, partial×2, cli_missing×1 | configure-agent-runtime Recipe、backend/router/executor-health | 依赖状态不等于 agent 业务成功 |
| 智能体 / 自动化 | 定时/API任务 → run/result；渠道规则 → 连接/事件核验 | L3 | agent-automation | passed×8, failed×2, partial×3, unverified×1 | 定时/API Recipe、agent resource-*、callback/channel get/test | Cron及固定指令PAT API已验；动态参数/schema失败，外部事件另验 |
<!-- capability-status:end -->

## 使用规则

1. 复杂任务先落到一条或多条业务目标，再读取对应 domain/recipe，按具体场景/条件查询状态。
2. 同一菜单中的闭环独立保存 ID、状态和证据；菜单能打开不代表任务完成。
3. 业务失败、部分完成、缺前置、缺 CLI 输入、未验分别记录；风险和授权仍按[能力边界合同](contracts/capability-boundary.md)。
4. 新接口先加入发现面；必须补齐对象、状态、业务证据和恢复合同后，才可记录相应场景的结果。
5. CLI 与 UI 语义不一致时保留实际限制；不使用 UI 证据或其它条件的成功覆盖原失败。

状态定义、日期及来源语义见[能力模型](capability-model.md)。表格由同一 JSON 生成，手写说明只放在生成标记外。
