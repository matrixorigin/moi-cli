# MOI 产品菜单到 moi-cli 覆盖矩阵

24 个菜单入口的静态基线来自 `qa-results/qa-feature-inventory-20261025-r2/功能清单-排除预留入口.md`；
命令面基线为正式 `moi-cli 0.1.8`。状态同时保留历史实测与后续有日期的在线补测，
每条观测分别标明验证日期和来源；没有把全部历史场景重新跑过。下面的状态计数来自同一份[能力卡片](capability-cards.json)，不是菜单完成率。

命令存在、文档组合形态和线上业务结论分别记录。`regular` 与 `candidate` 独立；来源未确认的验证日期保持未知。
同一菜单可映射多项能力，同一场景也可出现在多个菜单；不能跨行相加推算覆盖率。

<!-- capability-status:begin -->
来源：`capability-cards.json`（SHA256 `2619e43c50d1c1373c34d7e2fbddd1623896c9a3633857147fb11e135e7aca02`）。此段由 `scripts/capability_status.py render` 生成，修改JSON后重生成；手写说明留在标记外。

| 产品菜单 | CLI domain | 关键命令/组合 | capability IDs | 关联能力regular场景计数 | 关联能力candidate场景计数 | 缺口 | 优先级 |
|---|---|---|---|---|---|---|---|
| 概览 | workspace/catalog/sql/workflow/connector/compute/dashboard | 各域 list/get 的组合入口 | workspace-context, catalog-discovery, sql-query, workflow-lifecycle, connector-lifecycle, compute-lifecycle, dashboard-analytics | passed×32, failed×7, partial×3, blocked×2, cli_missing×4 | passed×46, failed×9, partial×2, blocked×2 | 不是独立 API；需按用户目标拆解 | P1 |
| 连接器 | connector | list/get/create/validate、structured-source-*、local-upload、create-task、task-runs、get-task | connector-lifecycle | blocked×1, cli_missing×1 | passed×1, blocked×1 | 常规CLI新建/验证/更新缺config输入；外部读取还需有效隔离连接 | P1 |
| 导入任务 | connector | tasks/create-task/run-task-now/task-runs/get-task/retry-task-files | source-staging, data-ingestion | passed×7, failed×2, partial×2 | passed×9, failed×2, partial×2 | 任务与目标表核验需结合 catalog/sql | P1 |
| 导出任务 | connector | create-export-task/export-tasks/export-task-state/export-task-info/export-task-files | data-export | passed×1, blocked×1 | passed×2, blocked×1 | 下载和目标端核验路径需补齐 | P1 |
| 工作流 | workflow | list/get/create-version/update/run/get-run/get-run-result/versions；批处理用 list-runs（dispatch_job_id关联）/batch-retry-failed-files/file-lineage | workflow-lifecycle, job-recovery | passed×8, failed×3, cli_missing×1 | passed×14, failed×3, partial×1 | 单次 execution、Volume 批量任务、文件级 execution 必须分开记账 | P1 |
| SQL 编辑器 | sql | execute/describe-query/query-result/meta-* / workbook-* | sql-query | passed×14, failed×4, partial×1 | passed×16, failed×6, partial×1 | DDL/DML/查询计划的风险与终态分层 | P1 |
| 数据看板 | dashboard + sql | list/get/create/create-chart/run-chart/chart-result/preview-sql | dashboard-analytics | passed×8, failed×1, partial×1 | passed×9, failed×1, partial×1 | 指标/下载/钻取已验，布局仅配置回读；定时与UI/告警分别判断 | P2 |
| Catalog | catalog | tree/list/system-tree/list-databases/database-children/table-info/volume-info/files/file-data-asset | catalog-discovery | passed×2, partial×1, cli_missing×2 | passed×7 | Workflow/Knowledge 前的 Volume 非空预检和 file lineage 需强制执行 | P1 |
| 计算资源 | compute | list/get/create/runtime/suspend/resume/preflight-delete/delete | compute-lifecycle | passed×1, blocked×1 | passed×1, blocked×1 | 生命周期、计费影响和删除影响分析 | P2 |
| 数据分享 | data-share | publishes/create-publish/subscribe/subscriptions/unsubscribe | data-share | passed×1, failed×1, partial×1 | passed×3, partial×1 | 发布对象、目标 workspace、订阅终态核验 | P2 |
| 知识库 | semantic-model | list/create-with-sources/upload-local-file/list-source-jobs/list-sources/get/delete；配合 catalog/workflow lineage | knowledge-preparation | passed×8, failed×2, partial×1, cli_missing×1 | passed×20, failed×1 | 必须按 file_id 追溯 Workflow 产物，并分离 ingest/index/检索/引用 | P1 |
| SQL 历史 | sql | query-history/query-history-overview/query-detail/query-profile/download-query-result | audit-observation | passed×1, failed×1, partial×2, cli_missing×1 | passed×5 | UI 筛选语义与 CLI 参数需统一 | P2 |
| 作业 | workflow + dashboard | list-runs/get-run/get-run-result/retry-run/cancel-run/pause-run/resume-run | job-recovery | passed×2, failed×1 | passed×3, failed×1, partial×1 | workflow execution、task、node ID 不能混用 | P1 |
| 操作日志 | workspace-member/agent/account | audit-logs、audit-log-overview、相关详情命令 | audit-observation | passed×1, failed×1, partial×2, cli_missing×1 | passed×5 | 页面展示的统一日志字段与 CLI 返回需对齐 | P3 |
| 用户管理 | workspace-member + iam | members/member/invite/update/update-status/user-tags | member-lifecycle | blocked×1 | blocked×1 | 邀请、禁用、删除和角色绑定均需显式目标 | P2 |
| 角色权限 | iam | roles/role-permission-view/permission-schema/put-user-role-bindings/object grants | iam-authorization | passed×1, failed×2, partial×1, blocked×1 | passed×5, blocked×1 | CAS、lifecycle、PDP 结果和回滚语义需补齐 | P2 |
| 智能体 / 对话 | agent + model | list/resource-get/send/a2a、model workspace-chat-completion | agent-conversation | passed×2, failed×2 | passed×2, failed×2, partial×1 | 会话上下文、执行环境和结果产物核验 | P2 |
| 智能体 / 成果 | agent + catalog/file | resource/list、workspace file/artifact download | source-staging, agent-conversation | passed×3, failed×2, partial×1 | passed×4, failed×2, partial×1 | 正式workspace/runtime file入口已有reference；作用域、原task及完整下载另行核验 | P3 |
| 智能体 / 智能体 | agent | list/init/build/load/verify/resource-* | agent-package-resource | passed×2, failed×1 | passed×4 | package、资源绑定和运行验证需拆开 | P2 |
| 智能体 / 技能库 | agent | list/get-skill-file/download-skill-files/import-skill-package | agent-skill-tool | passed×6, failed×2, blocked×1, cli_missing×1 | passed×8, failed×1, blocked×1, cli_missing×1 | UI 技能市场能力与 CLI 包操作不是一一对应 | P2 |
| 智能体 / 工具库 | agent | list-plugin-*、install-plugin、plugin connection commands | agent-skill-tool | passed×6, failed×2, blocked×1, cli_missing×1 | passed×8, failed×1, blocked×1, cli_missing×1 | OAuth、连接、凭证存储和测试连接需专门流程 | P2 |
| 智能体 / 知识库 | semantic-model + agent | semantic-model list/get/source commands、agent resource binding | knowledge-preparation, agent-conversation | passed×10, failed×4, partial×1, cli_missing×1 | passed×22, failed×3, partial×1 | 知识库资源绑定和问答验证需闭环 | P1 |
| 智能体 / 运行环境 | compute + model + provider-backend + agent | compute list/create/runtime、backend/router、agent executor-health | agent-runtime | passed×1, partial×2, cli_missing×1 | passed×4, failed×1, partial×3 | 资源、模型、网络白名单和执行器状态的关联 | P2 |
| 智能体 / 自动化 | agent + workflow/model | agent resource-* tasks/runs/result/events、callback/channel、workflow invoke | agent-automation | passed×8, failed×2, partial×3, unverified×1 | passed×12, failed×2, partial×1, blocked×1 | Cron与固定指令PAT API通过；动态参数/schema失败，外部事件另验 | P2 |
<!-- capability-status:end -->

## 使用方式

先按菜单发现 CLI domain，再按[能力目录](capability-catalog.md)查询具体场景和条件。
计数不构成执行承诺；外部连接、普通成员授权、计费运行和渠道事件必须各自补齐前置及完成证据。
失败、缺输入和没有执行是不同事实，不通过低层接口填补 CLI 缺口。

修改 JSON 后在 Skill 根目录运行 `python3 scripts/capability_status.py render`；
`python3 scripts/capability_status.py render --check` 会拒绝陈旧视图，且不写文件。
