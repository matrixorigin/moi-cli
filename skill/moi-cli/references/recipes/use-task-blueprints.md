# 使用复杂任务模板与安全恢复

任务模板就是事先写好的步骤和验收清单，告诉你先做什么、后做什么、怎样算完成。它不是线上工作流，也不自动执行。

用途：把已验收的复杂流程变成可重复的任务计划。任务模板只初始化本地 `task_run` 日志，不调用产品、
不自动登录、不选择工作区、不提交写入，也不替用户确认删除。真正执行仍使用正式 `moi-cli` 和
`task_run.py execute/probe`。

## 选择并初始化任务模板

```text
python3 scripts/task_blueprint.py list
python3 scripts/task_blueprint.py validate
python3 scripts/task_blueprint.py show catalog-workflow-knowledge
```

当前任务模板：

| 任务模板 | 适用结果 | 步骤数 |
| --- | --- | ---: |
| `catalog-workflow-knowledge` | 非空 Volume → Workflow → Knowledge → 正式引用问答 | 4 |
| `workflow-candidate-lifecycle` | v1 候选创建与运行 → 自然语言修改 → v2 部署与复验 | 6 |
| `agent-version-artifacts` | Agent v1/v2、默认切换与回滚、成果、取消和导出 | 7 |

以下示例只创建本地计划。`run_dir` 必须不存在，父目录由操作者选择为私有位置；所有必需输入
必须显式给出，工具不会从名称或列表顺序补 ID。输入用于固定任务范围，不代表线上资源已经存在。

```text
python3 scripts/task_blueprint.py init catalog-workflow-knowledge \
  --run-dir "$run_dir" --workspace-id "$workspace_id" --endpoint "$endpoint" \
  --input volume_id=150001 --input workflow_id=真实工作流ID \
  --input model_id=140001 --input question=本次要核验的问题
```

初始化后，任务模板 ID、版本、输入、步骤依赖和清理边界都保存在私有 `run.json`。每个验收条件仍须
由真实 CLI 回执满足；不能因为计划存在就报告业务完成。

## 每次继续前先做只读恢复检查

```text
python3 scripts/task_blueprint.py recover --run-dir "$run_dir"
```

| 状态 | 人话含义 | 能否提交新的主调用 |
| --- | --- | --- |
| `ready_for_single_execution` | 只有一个依赖已满足的步骤可以继续 | 可以，仅限返回的步骤和一次明确授权命令 |
| `blocked_unknown_submission` | 之前的写入已提交或结果未知 | 不可以，只查原对象 |
| `acceptance_incomplete` | 已收到响应，但验收证据没完成 | 不可以，继续核对原回执 |
| `acceptance_failed` | 原验收已经失败 | 不可以，另建有明确授权的恢复步骤 |
| `waiting_dependencies` | 上游还没验收完 | 不可以 |
| `completed` | 当前计划全部验收完 | 不再执行；清理另需确认 |

恢复检查还会拒绝两个步骤共用一个主调用 ID。它不会把 `failed` 改成成功，不会修改原验收预期，
也不会把新 invocation ID 当成重试许可。`task_run.py resume/status` 仍是证据校验入口；本命令
只在其上给出下一步安全判断，不替代原工具。

## 执行与清理

只有 `ready_for_single_execution` 返回目标步骤后，才按对应领域 Recipe 准备一条正式命令，并用
新的 invocation ID 调用 `task_run.py execute`。如果返回 `blocked_unknown_submission`，有真实类型
引用时使用 `task_run.py probe` 查原对象；没有原对象 ID 就保持未决，不能重建同名资源。

任务模板的 `cleanup` 只是提醒需要清理的对象和核验方式，**不是删除授权**。删除、取消、撤销发布、
跨工作区操作仍需锁定具体对象并单独确认。成果/placement 与物理文件元数据分别核对。
