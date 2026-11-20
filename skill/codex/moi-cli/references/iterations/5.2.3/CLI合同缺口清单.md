# CLI 正式合同缺口清单

> 版本说明：本报告完整保留 0.4.12 内容基线的验收结果；5.2.3 进行仓库化与命名统一，未把基线结果冒充本轮新增线上验收。

本清单由 `cli-contract-gaps.json` 生成，包含 **10 个整体缺合同项和 25 个部分缺合同项**。它说明哪些事情当前不能安全完成，不是功能失败率。

正式合同补齐前，只能保留现有证据并停止对应写入。不得用开放 JSON 猜字段，不得切到 raw HTTP、core-cli 或数据库直连，也不能拿相邻命令的成功冒充本功能。

查询机器清单：`python3 scripts/cli_contract.py gaps`。验证清单及文件参数：`python3 scripts/cli_contract.py validate`。

| 编号 | 缺口层级 | 当前已经做到 | 还缺什么正式合同 |
| --- | --- | --- | --- |
| AR06 成果批量下载 | 整体缺合同 | 仅发现batch-get-data-assets和batch-resolve-file-artifacts，未提供选择多个成果并打包ZIP的命令；download-workspace-file只接受单个file-id。未用本地ZIP冒充产品能力。 | 缺正式批量成果ZIP下载入口和输入合同。 |
| CN06 模型格式适配 | 整体缺合同 | 当前正式参数只有地址、模型、reasoning protocol、MIME类型和超时等，没有per-channel请求/响应JSON模板或JSONPath映射合同。 | 需要正式的adapter override输入；不能用reasoning-control-protocol代替格式适配。 |
| CP05 站点 / Region 切换 | 整体缺合同 | 当前只有workspace-id及规格字段，没有站点/Region切换输入，无法按站点查看不同价格或实例。 | 需要正式站点/Region参数及相应目录合同；不能猜测region字段。 |
| CR07 计算资源时长统计 | 整体缺合同 | 规格提供credit_per_hour，资源详情只有时间戳和配置，没有在线时长或实付金额。current-credit没有资源ID输入，本次还因native会话失效返回401。没有形成单资源账单核验。 | 需要正式的单资源在线时长与实付账单入口；401属于独立认证限制，不是判断该入口存在或不存在的依据。 |
| MM03 新建模型 | 整体缺合同 | 现有命令管理的是workspace LLM backend/endpoint，不是后台模型SKU的新建、编辑、删除或启停；没有对应正式管理合同。 | 需要后台模型SKU生命周期接口，不能把workspace backend配置当成模型SKU管理。 |
| MM04 编辑 / 删除模型 | 整体缺合同 | 现有命令管理的是workspace LLM backend/endpoint，不是后台模型SKU的新建、编辑、删除或启停；没有对应正式管理合同。 | 需要后台模型SKU生命周期接口，不能把workspace backend配置当成模型SKU管理。 |
| MM05 启用 / 禁用模型 | 整体缺合同 | 现有命令管理的是workspace LLM backend/endpoint，不是后台模型SKU的新建、编辑、删除或启停；没有对应正式管理合同。 | 需要后台模型SKU生命周期接口，不能把workspace backend配置当成模型SKU管理。 |
| RP16 告警权限 | 整体缺合同 | 当前148个action未提供alerting资源/action；搜索alert permission没有匹配管理合同。 | 告警权限需要先冻结资源/action和公开合同，不能把workflow/query_history等其它授权冒充告警权限。 |
| RP27 角色标签与筛选 | 整体缺合同 | resource-tag --type role返回400 RESOURCE_TAG_INVALID_ARGUMENT；搜索role tag labels没有匹配。 | 需要角色标签类型、赋值、筛选和分组的正式合同；当前只确认workflow标签类型存在。 |
| UM07 API Key | 整体缺合同 | 搜索返回matches=[]；当前只支持使用已配置PAT，未找到用户API Key生命周期管理合同。 | 需要公开API与CLI的用户API Key创建、查看、撤销接口及权限边界，当前不能猜参数或复用PAT代替。 |
| AK01 知识库列表、发现 | 部分缺合同 | 列表和本地名称筛选可用，标签目录可读；知识库整体标签及全部处理状态仍缺统一合同。 | 整体标签筛选、启用状态和完整处理状态组合。 |
| AK02 知识库创建与基础信息管理 | 部分缺合同 | 创建和来源管理可复用既有证据；整体标签及文件/表数量和处理状态组合仍未形成完整闭环。 | 知识库整体标签维护及完整状态组合。 |
| AK03 知识库文件与表管理 | 部分缺合同 | 文件管理多数子项已有证据，批量删除仍无正式入口。 | 批量删除及失败重试等完整管理链。 |
| AK05 文本分段管理 | 部分缺合同 | 分段读取、编辑、启停和ETag已有证据；搜索、召回优先排序和召回次数没有正式输入。 | 分段搜索、两种排序及召回次数的真实业务结果。 |
| CA01 对象树 | 部分缺合同 | catalog tree 返回catalog/database/table/volume；模型与算子可分别读取，但没有统一出现在该树中。 | catalog tree 未统一展示模型和算子类型。 |
| CA04 库详情 | 部分缺合同 | database-children返回表和Volume；模型与算子需要分别入口，未在同一库详情中统一列出。 | 库详情中的模型与算子统一枚举未覆盖。 |
| CH13 模型思考等级 | 部分缺合同 | 直接模型调用可按等级工作，但普通A2A显式等级存在内部错误；默认未指定最稳定，不能宣称所有入口max均生效。 | 不同模型、服务slot和A2A入口的等级支持不一致。 |
| CO02 创建连接器 | 部分缺合同 | 创建ID 517534000619，active、source_type=5、usage=[1,2]；validate返回valid=false，符合无效端点预期。 | 当前validate没有saved connector-id输入，只证明保存前配置校验；保存后连接测试、真实外部连通和页面取消未验。 |
| CP01 规格目录列表 | 部分缺合同 | 目录返回15项，覆盖query/task/gpu等类型，包含CPU、内存、credit_per_hour和启用状态；CC0为1核2GiB、6 credits/小时。 | 站点/Region价格切换、规格实例数和管理后台完整目录未验证。 |
| DD09 看板告警记录 | 部分缺合同 | 统一告警记录可读，但没有正式逐条/全部已读入口。 | 未读角标、已读动作和用户×记录存储；统一记录页需UI。 |
| DL02 载入任务详情与编辑 | 部分缺合同 | 周期任务创建、名称/调度/过滤修改成功；回填、预处理和实时水位线仍无正式输入。 | 回填、预处理、实时水位线及独立实时任务编辑入口。 |
| DL04 载入任务暂停 / 继续 | 部分缺合同 | pause/resume已在周期任务上验证；CLI命令盘点没有独立stop入口。 | 独立停止运行中任务的stop命令和终态。 |
| DS06 权限控制 | 部分缺合同 | 只读发布和文件下载能力存在；create-publish正式参数只有名称、源库、表范围和目标workspace，没有独立只读/可下载切换输入。 | 独立只读/可下载权限切换未提供正式CLI输入，不能用其它字段猜测。 |
| LG01 审计列表 | 部分缺合同 | 本轮时间倒序列表、page/page-size及method+status-code+result+时间范围组合筛选均返回正确记录。 | 操作人、模块、操作类型、来源、对象关键词和高敏开关未由当前flags覆盖；完整页面筛选仍需UI或补合同。 |
| MM07 模型状态指示 | 部分缺合同 | 路由只有策略、健康检查间隔和重试次数；模型目录不含正常/不稳定/异常/未知四态状态。 | 四态模型状态的实时指示和变化未验证。 |
| MO03 二级筛选器 | 部分缺合同 | model_type可分组，名称可区分DeepSeek/GLM/Kimi/Qwen系列。 | 没有独立的系列字段和完整十类×系列联动合同，只有chat一类。 |
| MO04 次级筛选 | 部分缺合同 | 可按backend_name筛选Genesis/system-llm；目录不提供上下文长度和四态状态字段。 | 厂商全量枚举、上下文长度和四种状态筛选未覆盖。 |
| MO05 模型详情抽屉 | 部分缺合同 | DeepSeek条目含model、backend_id/name、model_type、source、system_default和reasoning_efforts。 | 简介、参数、能力标签、定价和API示例未完整提供，不能宣称完整详情抽屉。 |
| MO09 状态实时显示 | 部分缺合同 | provider路由返回ROUND_ROBIN、30秒健康检查、最多2次重试；模型目录本身不含四态状态。 | 正常/不稳定/异常/未知四态的实时显示和变化过程未验证。 |
| OT10 CLI | 部分缺合同 | CLI已覆盖工作区、Catalog、SQL、导入导出、Workflow、Knowledge、IAM和Agent多条链路；P12仍有后台和模型管理缺口。 | 不能据此宣称与全部开放API一一对齐，后台Region、渠道和模型SKU管理仍缺合同或条件。 |
| SK02 技能详情与版本管理 | 部分缺合同 | 附件下载和独立新包切换已有证据；update-skill不存在，import-skill-package不接受目标ID，resource-put只有开放JSON，未形成同ID附件更新合同。 | 同ID附件原位更新及完整历史快照回退仍缺公开合同。 |
| SK04 技能管理与启停 | 部分缺合同 | 元数据和绑定可组织；同ID附件更新没有正式入口，不能猜resource-put字段。 | 统一启停、标签编辑和同ID附件更新未覆盖。 |
| WF24 视频解析节点 | 部分缺合同 | 视频音轨自动识别英语并完成转写；正式输入只有语言、去噪和切片，结果只有音频文档。 | 关键帧、独立字幕和场景检测没有正式配置或输出；长视频另需样本。 |
| WF27 清洗节点 | 部分缺合同 | n-gram和特殊字符过滤均验证，全部内容被过滤时明确失败；正常文本在阈值下保留。schema没有翻译输入。 | 通用语言翻译无正式入口；节点折叠标题结构及全过滤为空时的错误语义需产品确认。 |
| WF34 工作流分支 | 部分缺合同 | 版本、候选和节点恢复分支已有证据；当前只有分支列表/取消，没有完整切换、删除和AI效果对比入口。 | 实例级AI效果对比及分支切换、删除。 |

完整证据与禁止替代路径见 `references/cli-contract-gaps.json`。
