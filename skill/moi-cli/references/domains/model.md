# Model 领域

## 生产环境模型选择

用户在 2026-10-30 确认，MOI 生产页面当前提供 `deepseek-v4.1-flash`、`glm-5.3`
和 `kimi-k3`。在生产环境执行问答、分析和复杂任务时优先使用 `deepseek-v4.1-flash`；
按用户最新选择，默认不指定推理等级，包括复杂任务。不发送 `reasoning_effort`、
`astra_reasoning_effort` 或 `llm_reasoning_effort`，也不以空字符串或 null 代替省略。
用户明确指定其它模型或等级时，按其要求执行。这是 CLI 任务的默认选择，不代表已修改
服务端配置；本次已验证的路径可用，不保证后续所有业务任务都成功。

候选模型会变化，以上记录不是固定白名单。2026-10-31 的只读 `model list` 返回 12 个系统模型；
它把多个系统模型都标成 `system_default=true`，因此不能单凭该字段证明唯一默认路由。
本 Skill 仍按用户明确的生产策略选择 DeepSeek，但实际通道可用性必须按目标入口复核。
通用 `model list`、智能体对话与知识库问答的
模型范围可能不同：知识库 Explore 使用 `knowledge_qa`，普通智能体对话使用 `agent_chat`。
通用列表出现某模型，不证明该模型在目标通道可调用。本机 CLI 0.1.8 的 `model list`
没有 `--service-slot`，不能编造该参数，也不能把未过滤列表当作知识库专用列表。

需要排查通道时，可以对明确 workspace/model/input 做一次最小调用，保留实际结果：

```text
moi-cli -o -no-interactive -workspace-id "$workspace_id" model workspace-chat-completion --service-slot knowledge_qa --json-file "$model_check_request"
```

直接模型调用必须选择正确的服务通道。`agent_chat` 和 `knowledge_qa` 不能省略；本次省略
`--service-slot` 的最小请求对 DeepSeek、GLM、Kimi 均返回 500 `chat completion stream failed`，
而补上对应通道后三个模型均能返回正确答案。这个失败是调用输入/路由缺失，不是三个模型同时不可用。
只有明确指定等级时，这类直接模型调用的 JSON 才设置 `reasoning_effort`；历史补测中
`reasoning_effort: "max"` 请求成功，当前默认仍省略该字段。
A2A 的等级字段依入口和运行时而异：Astra 智能体客户端使用
`params.metadata.astra_reasoning_effort`，知识库页面当前使用 `llm_reasoning_effort`。
不要把直接调用字段或不同入口的字段互换，也不能仅凭任意 metadata 被接受就说等级生效。

2026-10-30 已用三个模型 × 未指定/low/high/max × 五种路径完成 60 项生产 CLI 检查。
以下为当日结果，不作为未来版本固定限制：

| 路径 | deepseek-v4.1-flash | glm-5.3 | kimi-k3 |
| --- | --- | --- | --- |
| 直接模型调用，agent_chat | 四种设置都返回正确答案 | 四种设置都返回正确答案 | 四种设置都返回正确答案 |
| 直接模型调用，knowledge_qa | 四种设置都返回正确答案 | 四种设置都返回正确答案 | 四种设置均无可用后端 |
| 普通智能体 A2A，astra_reasoning_effort | 未指定成功；low/high/max 内部错误 | 未指定成功；low/high/max 内部错误 | 未指定成功；low/high/max 内部错误 |
| 知识库 A2A，页面的 llm_reasoning_effort | 四种请求均正确回答并引用；等级未确认 | 四种请求均正确回答并引用；等级未确认 | 四种请求均在模型阶段失败 |
| 知识库 A2A，astra_reasoning_effort | 未指定成功；low/high/max 内部错误 | 未指定成功；low/high/max 内部错误 | 未指定时模型失败；low/high/max 内部错误 |

Astra 显式等级失败均为 `-32603 / INTERNAL`，没有返回任务 ID，底层原因需要服务端
日志确认。不能把模型接口接受 `reasoning_effort` 当成整个智能体等级链路可用；也不能
把页面 `llm_reasoning_effort` 请求成功当成等级已生效，当前本地 Astra 源码不消费该字段。
用户明确要求某等级时，保留其选择并报告不支持或失败，不能静默去掉参数。
用户未明确指定等级时，沿用上述默认，并如实记录等级未指定。
记录请求使用的等级；响应未回显等级时，不声称已独立证明上游实际执行了该等级。

只读模型目录还观察到：`glm-5.2`、`glm-5.3-flash`、`kimi-k2.7-code` 当前也列在
Genesis 系统模型中，并返回 `low/high/max`；`qwen3.7-plus` 返回
`none/minimal/low/medium/high/xhigh`；`qwen3.8-27b`、`qwen3.8-flash`、
`qwen3.8-max` 返回 `low/medium/xhigh`；`qwen3-vl-plus` 与 `qwen3.7-max` 未返回等级字段。
这些是发现结果，不是实际推理成功证明；没有目标任务时不要擅自 probe 或把缺失字段猜成支持。

workspace LLM router 和 Provider router 的只读配置当前可回读：策略为 `ROUND_ROBIN`，
健康检查间隔 30 秒，最多重试 2 次。LLM、embedding、parser backend 列表成功但为空，
只能说明没有可列出的自有 backend，不能推断系统模型不存在。`list-endpoints` 相关 help
把 backend ID 标成可选，但实际省略时分别返回 `ErrIDRequired` 或 400；没有真实 backend ID
时记录为 CLI/服务输入合同缺口，不猜 ID。

2026-10-31 的只读运行环境复核还确认 `executor-health` 返回 `astra_reachable=true`，模型目录和
router 健康信息可读；`provider-backend list` 返回空对象。这里的“空 backend”是目录结果，不是
模型调用失败，也不是可以直接写入后端配置的理由。要判断某个入口是否真的可用，仍需在明确
workspace、model、service slot 和输入后做一次最小调用，并回读任务或结果。运行环境配置写入、
网络白名单、Runner/Sandbox 依赖和实际网络策略仍未通过常规 CLI 证明。

GLM 的部分知识库流在 CLI 的 30 秒等待结束前未返回终态，读取原任务后均已完成，答案和
引用通过核验。此时先读原 task，不延长超时掩盖问题，不重复发送问题。

最小调用成功只证明该次模型请求可用；完整任务仍需核验工具结果、任务终态和产物。
通道返回无可用后端时保留原错，不擅自改路由、换凭证或重新登录。

## 发现与运行

只读发现：

```text
moi-cli -o -no-interactive -workspace-id "$workspace_id" model list
moi-cli -o -no-interactive -workspace-id "$workspace_id" model list-workspace-llm-models
moi-cli -o -no-interactive -workspace-id "$workspace_id" model list-workspace-embedding-endpoints --backend-id "$backend_id"
```

`workspace-chat-completion` 和 embedding 调用是有额度、模型和输入副作用的运行请求；执行前
必须明确模型、服务 slot、输入和 workspace。后端/endpoint 创建、更新、状态切换属于配置
写入，不能用 probe 成功替代运行成功。

## 模型类型筛选的证据边界

当前正式筛选参数是 `model list --query-param type=...`，不是 `model_type=...`；后者在本次
调用中被忽略并返回未筛选的 chat 列表。`type=chat` 返回 11 个，`type=embedding` 返回 3 个，
非法类型返回 400。不要把参数名猜成响应字段名。

2026-11-17 还观察到 `type=ocr`、`type=vision`、`type=reasoning` 都把同一批 11 个模型改成
对应 `model_type` 返回，`type=rerank` 没有返回模型。这些回执只能证明接口响应形状，不能证明
相应类型真实可用或十类调试页签全部存在；同名模型按请求重标类型的行为需要服务 owner 核准
正式合同。模型目录、真实类型调用和页面页签是三层不同证据，不能互相替代。

## 采样、Seed 与开放 JSON 字段

先运行 `moi-cli schema model workspace-chat-completion`，把字段分成“typed 正式输入”和
`additionalProperties` 允许的开放输入。当前 typed schema 明确列出 `max_tokens`、`temperature`
和 `top_p`，没有列出 `seed`、`frequency_penalty`、`presence_penalty`。后三者被请求接受不等于
上游已经执行，不能用 HTTP 200 或回答生成成功证明参数生效。

2026-11-17 对 DeepSeek 的复核中，`max_tokens=1` 返回 `finish_reason=length` 且
`completion_tokens=1`，可证明上限生效。相同 Seed、相同消息、`temperature=0` 的两次独立请求
分别返回 `P12_SEexport_SEED_OK` 和 `P12_SEED_OK`，没有形成确定重放。由于 Seed 不在 typed
schema 中，这个结果保留为“开放字段未证明生效，需要正式合同或服务端确认”，不直接归类为
产品缺陷，也不把 Max Tokens 的成功外推成 Seed 成功。

采样参数验收要针对具体效果设计：Max Tokens 看终态和 token 数；Seed 看同一完整输入的多次
完整输出；频率惩罚不能仅凭回答是否命中某个标记推断概率分布。模型有随机性时，记录实际
输入、模型、服务 slot 和完整输出，不把单次文本差异写成稳定结论。
