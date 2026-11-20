# 证据记录合同

多步任务使用[本地记录器与任务断点](../recipes/record-and-resume-task.md)保存证据；对话只保留摘要：

```yaml
workspace_id: ""
target:
  type: ""
  id: ""
operations:
  - command: ""
    purpose: ""
    returned_ids: {}
    status: ""
verification:
  command: ""
  result: ""
  evidence: ""
remaining: ""
```

证据至少包含 workspace、目标对象、实际命令的作用、返回 ID、最终状态和核验命令。必要时
记录 query/statement、workflow/execution、task/job、model/source/file 等关联 ID。

凭证（密码、token、Cookie、连接串、PAT、API key）不得写入证据或展示。原始私密业务内容
不得进入聊天、公开报告或共享仓库；本地输入副本仅在明确授权的私有目录按 Recipe 留存。
报告本地文件时只引用安全路径和必要的名称/大小/哈希。

变更请求发送前保存独立正文文件及摘要，后续问答、重放、配置不能复用覆盖该文件。回执中只有`--json-file`路径不等于保留了当时请求正文；发现快照丢失时如实登记，不能将后写文件或从响应重建的内容冒充原请求。表配置、绑定和规则变更均适用。

正式调用用 `scripts/record_cli.py` 或 `task_run.py execute/probe` 逐次记录。记录器将实际输入
冻结后，把副本交给 CLI，避免源路径后来变化导致证据不匹配。恢复和完成时重校哈希。
响应只保存按常见凭证模式脱敏的文本；哈希不代表原始输出或线上字节。业务输入副本与
外部下载产物的留存、隐私和校验边界，按 Recipe 执行，不把哈希当授权或不可篡改证明。
