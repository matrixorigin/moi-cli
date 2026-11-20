# Recipe：执行托管 Skill 脚本并核对冻结版本

这里的Skill是MOI里托管的资源，与本机Codex的moi-cli Skill版本独立。
包先用正式 `inspect-skill-package` 检查，再 `import-skill-package` 保存返回的Skill ID、version、
contains_scripts和文件哈希。绑定前保存指定Agent的完整bindings，只增量加入本次Skill引用；
清理时保留原tools/knowledge/plugins/其它skills，结合当前revision回读，不能用过时快照盖掉他人变更。

按[Agent包与运行](build-load-and-use-agent-package.md)的A2A合同提交获准任务，默认DeepSeek且不传effort。
提示给真实输入和执行要求，不预填算法结果。任务先读Skill，再依据工具返回的task-frozen下载指令取附件，
下载目标目录必须尚不存在。下载只取文件，不会自动执行脚本。
检查实际代码、入口和依赖后，在授权任务沙箱执行；不能从源码推算答案替代执行。
核对真实工具call/result、进程内层exit、stdout/stderr和截断情况，再独立验证业务输出。
外层shell最后echo可能返回0而前一命令已失败，不能仅看最外层exit_code。
标准库Python成功不证明任意语言、第三方依赖和网络访问都可用。

长任务可先用正式conversations资源创建本次独占的新会话，保存返回ID作为contextId，然后只发送一次。
没有明确独占时，后续/并发消息可能覆盖latest_task_id，不能把该字段直接认作原请求任务。
30秒请求超时不是任务终态：响应已给真实task ID时优先沿该ID查询；否则读该已知会话latest_task_id作为候选，
同时读取持久消息，把原请求、隔离Agent、context和消息中的task关联核对清楚，再get-task。
重复文本或并发消息使关联不唯一时保留未知，不拿最新任务冒充原任务。
若首次发送没固定会话，只能在明确隔离Agent的会话/消息中精确核对原请求全文和持久化task关联；
关联不唯一就报告未知，不按最近时间选任务或重新发送。
resource命令tail-path需为空或以/开始，不把单独查询字符串当tail-path；Value信封按实际返回读取。

验收冻结版本时，先确认原task正在working，再更新本次隔离Skill；记录更新完成时间和当前version，
之后核对同task实际下载时间、下载version、文件hash及实际执行结果。
本次instruction当前v2与已启动任务下载冻结v1可同时成立；附件包切换的补充验收见下文。不能将其扩大成同ID附件原位更新或所有资源版本治理。

## 独立附件包切换的在途冻结

`import-skill-package`每次创建独立Skill ID；当前常规CLI元数据更新无files/content输入。因此已验路径是导入独立新版包→切换Agent绑定，不是同一Skill ID原位替换附件。后一目标应明确提示当前入口缺口，不猜更新body。

2026-11-15 dev补测仍只有`import-skill-package`，没有`update-skill`专用入口；通用`resource-put`只提供开放JSON，未形成已核实的同ID附件更新合同，不能把它当成正式替代命令。独立新包切换与同ID原位更新必须分开报告。

公开 Product API 也只提供元数据 PATCH 和附件 GET；`ImportSkillPackage` 固定 `Version: 1` 并调用
`CreateWithFiles`，内部ID用于新建或幂等确认，不是替换既有附件。当前缺口不只在CLI，公开产品写合同也不存在，
必须由产品owner先提供同ID附件更新合同，Skill不能猜字段或发送覆盖请求。

2026-11-01旧task working时切Agent默认/绑定，新task在旧task完成前提交并完成。旧task在切换后才下载A，仍取得A完整脚本并实际输出1221；新task读取B输出2333。源ZIP、导入digest、正式get-skill-file、沙箱完整源码SHA256及真实执行一致，两个task均completed。需保存新旧Skill ID、版本、Agent版本、绑定回读、任务manifest及实际下载时间。

若包声明skills为空，切default后要重新增量绑定并独立回读，不能先绑定再切默认后直接运行。下载目标目录必须不存在；先mkdir会导致file exists，需明确失败后换新的不存在目录，保留原错误。同一工具外层exit0不覆盖内部下载/执行错误。
