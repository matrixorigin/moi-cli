# Recipe：智能体实际调用工具并交付文件

1. 用 `agent list`、`agent-card` 固定真实 Agent 及 owner workspace；指定本次允许读取的数据和唯一交付文件。系统 Agent 的 owner 可能是 `system`，不能擅改为当前 workspace。
2. 读取 `schema agent a2a-stream`；JSON-RPC 请求使用 `method: "message/stream"`，`agent_id`、`agent_workspace_id`、`model` 和 params。生产默认 DeepSeek，不传 effort。`agent a2a` 与 `a2a-stream` 的 method 不可混用。
3. `params.message` 使用唯一 messageId、role=user 和 text parts；metadata 固定业务 workspace。要求实际读指定表、记录工具结果、生成文件并发布正式 artifact。提示词不提供预期答案。
4. 用 `agent a2a-stream --json-file REQUEST --output STREAM` 保存事件，从 task/status 返回保存真实 task ID。超时先查原任务，不重复发起。
5. 用 `agent get-task` 读取原任务。当前系统 Agent 查询应按已确认的系统 code 使用 `--agent-code`；不能将 owner 不同导致的 404 解释为原任务执行失败。
6. 检查工具调用和真实工具返回：查询应具有成功状态、query/statement ID、列和数据。模型口述“已经查过”不是证据。核对最终 task completed 与正式 file artifact。

## 文件从运行时到用户

Agent 活动沙箱内通过 `file upload --file PATH --publish-artifact` 发布；本机用户 CLI 不伪造 task runtime 身份。返回 file_id、artifact_id、size、md5/sha256；task artifact 的 file_id 应与实际文件相同。

用户侧可用：

```text
moi-cli -o -no-interactive -workspace-id "$workspace_id" catalog list-workspace-files --name "$file_name" --page-size 100
moi-cli -o -no-interactive -workspace-id "$workspace_id" catalog download-workspace-file --file-id "$file_id" --output "$output_path"
```

名称只用于筛选，按真实 file_id 选 placement，检查 source_task_id 与原 task 一致。下载后核对字节数、服务端 MD5/SHA256、正文及原始数据；placement 的数字 id 不是 file_id。仅创建沙箱文件、仅返回路径或链接，都不证明正式成果可下载。

需要多轮时显式保存已确认事实和资源 ID。当前 Team 自动历史继承有失败记录，见[Team](configure-agent-team-ingress.md)；续轮不能默认完整记住前文，也不能把显式带入前文的成功当成自动记忆通过。

## 桌面 workspace file 生命周期

桌面文件与任务内 `publish-artifact` 是不同来源，不能混用验收。对本次明确授权的隔离文件，可用
`agent upload-workspace-file` 上传，再用 `catalog download-workspace-file` 下载并比较 MD5/SHA256；
未挂载时先查 `file list`，挂载到指定隔离 Volume 后用 `catalog create-file` 建立 placement。
2026-11-15 的 115 字节样例已验证上传/下载摘要一致、placement 0→1、删除后下载404、placement=0、
Volume=0，隔离 Volume 和数据库清理后均404。删除应使用本次真实 file/placement 和明确范围，
不把“无 placement 删除”当成用户历史文件的默认补救；跨身份权限和任务产物发布仍需单独验收。
