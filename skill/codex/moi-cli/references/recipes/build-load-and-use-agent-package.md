# Recipe：构建、加载并使用 Agent 包

适用于把本地 Agent 指令和包内自定义技能加载到已确认的 MOI workspace，并验证技能确实能用。
涉及包结构或自定义工具构建时先加载 `agent-package` Skill，使用当前 `moi-cli` 的正式格式 owner；
不要手写生成的工具 artifacts，也不把 endpoint、workspace 或凭据写进 `agent.yaml`。

## 本地包

以下命令在用户选定的工作目录执行；路径变量需事先设为本次隔离文件。`init` 和 `unpack`
的输出目录不能已存在，构建不擅自覆盖已有包。包 ID 与语义版本使用本次唯一值。

```sh
moi-cli agent init --output "$source_dir" --agent-id "$package_agent_id" --name "$agent_name" --version "$package_version" --model deepseek-v4.1-flash --instruction "$agent_instruction"
moi-cli agent verify --source "$source_dir"
moi-cli agent build --source "$source_dir" --output "$package_file"
moi-cli agent verify --file "$package_file"
```

`init` 的默认模型不一定是本次希望使用的模型，按模型领域规则显式选择 DeepSeek，默认不传
effort。需要包内技能时，在 verify/build 前加入 `custom/skills/<name>/skill.yaml`，按包格式
owner 的当前示例填写 instruction、routingSummary 和 requirements。技能正文必须去掉首尾
空白；YAML 多行正文可用 `body: |-`，避免 `body: |` 保留末尾换行导致校验失败。

技能可用一个可独立核对的合成规则，例如转换短码并输出只存在于技能正文中的固定标记。
Agent 指令只要求遵循绑定技能，不应同时复制该规则。这里不要求业务数据、外部连接或付费资源。

## 加载并确认绑定

加载会在目标 workspace 创建或更新包对应 Agent，并材料化包内资源，属于受控写入；只对
已授权 workspace 和隔离包操作。线上命令显式使用本次 endpoint、认证方式和 workspace。

```sh
moi-cli -o -no-interactive -workspace-id "$workspace_id" agent load-agent-package --file-path "$package_file"
moi-cli -o -no-interactive -workspace-id "$workspace_id" agent resource-get --resource agents --tail-path "/$agent_id"
moi-cli -o -no-interactive -workspace-id "$workspace_id" agent resource-get --resource skills --tail-path "/$skill_id"
```

从加载结果保存真实 `plan.agent_id`、版本、source_digest 和 version.status；从该 Agent 的
`binding.skill_refs` 保存真实材料化技能 ID、owner workspace 和 binding_source，不能按名称
推测生成 ID。应确认 Agent 可运行、技能 active、正文与本次包一致。产品可能继承 MOMO 基础能力，
这不表示本次修改了系统 Agent，也不证明继承工具已通过验收。

加载超时或 runtime registration 失败时先查原 ID、版本和绑定，不换 ID/版本重复建资源。

## 真实使用与结果

按 `schema agent a2a` 和当前领域合同构造 `message/send` 请求，指向该 Agent 的真实 ID 和
owner workspace，使用 DeepSeek、不传 effort。提示只给合成输入和“使用绑定技能”的要求；
不要给技能内部规则、专用标记或预期答案。

```sh
moi-cli -o -no-interactive -workspace-id "$workspace_id" agent a2a --json-file "$request_file"
moi-cli -o -no-interactive -workspace-id "$workspace_id" agent get-task --agent-id "$agent_id" --task-id "$task_id"
```

保存原 task/context ID，回读同一任务到终态。成功至少同时满足：实际 `moi.tool.call` 指向
本次技能、`moi.tool.result` 返回正确技能正文、最终回答符合正文规则。包加载、列表出现、
模型声称“已使用”和 completed 状态各自都不充分。系统 Agent 的 task 查询用其正式
`--agent-code` 合同；不要把系统 owner 自动替换为当前 workspace。

## 导出与保留状态

```sh
moi-cli -o -no-interactive -workspace-id "$workspace_id" agent export-agent-package --agent-id "$agent_id" --version "$package_version" --output "$exported_file"
moi-cli agent verify --file "$exported_file"
moi-cli agent unpack --file "$exported_file" --output "$unpacked_dir"
```

核对导出包的 Agent ID、版本和 digest，比较源指令、技能正文等可编辑文件；本次可重复合成包
已验证原包和导出包字节一致。一般任务仍需按实际包核对，不把该结果推断为任意包都可无损往返。
报告Agent/技能最终状态、是否保留供复验，以及本次确实调用过的工具范围。指令技能、可执行工具、
托管文件技能附件、知识库绑定或外部连接分别验收。

## 包内可执行工具

按agent-package的当前格式，在源包加入 `custom/tools/<name>/tool.py`；本次Python样例仅用标准库，
从stdin读取JSON、向stdout输出JSON，不读取环境、凭据或业务文件。每个工具目录只保留格式owner支持的唯一入口，
由 `agent build` 生成artifact，不手写生成文件。包verify后正式load，保存加载后实际tool ID，
回读其active、code和binding_source=agent_package，不能用源目录名猜材料化ID。

A2A只给输入和使用该工具的要求，不给源码专用标记、算法前缀或预期结果。允许正式moi_tool_search发现工具合同，
然后检查本次custom tool的真实call/result、输入和进程返回。2026-10-31样例runner为moi-agent-tool-wrapper、
language=python、entry_file=tool.py，exit_code=0、stderr为空；stdout内源码标记与独立本地hash核对一致，
同task回读completed，导出包与原包字节一致。不能用bash模拟该工具、仅有包verify或最终文本代替实际执行证据。

标准库Python和后续vendor第三方依赖样例分别验证；其它语言、依赖安装、网络或业务副作用另验。
产品可能将代码工具标为external_effect，不能因为样例纯计算就声称所有代码工具只读。

## 同一 Agent 多版本、默认切换与回滚

仅对指定目标和获准版本操作；验收用全新Agent ID，保留原包和每轮真实任务证据。
同ID加载新版本只增加runnable版本，**不会自动切换default**。先读正式versions资源和
lineage.default_version，再判断下一轮会用哪版；不能把load成功或版本号较大当默认已切换。

```sh
moi-cli -o -no-interactive -workspace-id "$workspace_id" agent resource-get --resource agents --tail-path "/$agent_id/versions"
moi-cli -o -no-interactive -workspace-id "$workspace_id" agent resource-post --resource agents --tail-path "/$agent_id/versions/$new_version/default" --json '{}'
```

切换后独立读回default和bindings；包声明可能重新材料化绑定，基于最新回读状态增量补齐获准Skill等绑定并再次回读，再在新会话实际运行，核对manifest.agent_version、冻结绑定与输出。
本次额外比较了旧会话下一轮：显式切默认后，同一旧context的新task采用新默认v2；
切回v1后新会话又用v1。此前已完成的v1任务回读仍是原版本和结果。
会话ID相同不等于未来每轮固定旧版本；已启动task的冻结快照与后续新task采用的默认版本分别判断。
不能用“只load新包未切default”的旧会话结果证明版本冻结。

2026-11-17 的另一组生产隔离复核把版本与任务生命周期一起验收：v1/v2 均为 runnable，
默认先切 v2 再回滚 v1；完成任务含两个正式 file artifact，取消任务按原 task 回读为
`canceled`，导出 v1 包通过 `agent verify --file`，两个成果下载正文正确。成果批量删除和
Agent 根对象删除均有回执，根对象最终 404。这里证明的是本次双版本和任务组合，不把根对象
删除外推为所有物理文件、版本元数据或历史记录均已清除。

回滚是把default指向仍runnable的旧版并读回，不是重复load同一已runnable版本。
先核对默认已离开目标版本，再按正式 `.../versions/{version}/disable` 禁用非默认版，
已disabled且非默认的版本才可删除。默认版的禁用/删除约束要保留，不能为清理绕开。根Agent删除不等于包版本级联删除：
本次根对象读取404后，versions仍保留默认v1及lineage元数据；正式owner禁止删除默认版本，
未发现清空lineage入口。回读并如实登记元数据残留，不把它说成活跃任务，也不伪称零残留。
此前验证顺序升级、默认切换及回滚；不同新版本和同版本相同包双请求的限定补验见下文，同版本不同内容双请求首发补验见下文；并发默认切换仍未验。
服务内部的expectedCurrent/CAS不等于CLI请求已支持显式预期版本参数，不编造CAS flag。

## 其它语言、第三方依赖和网络

正式包owner支持Go、Python、C++和Rust，支持打包不等于线上已实际运行。
Go需要本机工具链和目标平台产物；Python有requirements.txt时需完整vendor，运行时把vendor加入PYTHONPATH，
不会自动安装或保证任意第三方依赖。C++/Rust的Linux构建需要对应Docker builder；
当前Rust owner明确拒绝Cargo.toml/Cargo.lock，使用自包含main.rs及本地modules。
这些边界按当前owner/help复核，不因本机缺编译器宣称MOI不能运行，也不为凑验收擅自下载工具链。
网络调用还需获准目的地、只读/写入范围和独立响应核验；语言执行成功不证明出网已获准或任意地址可达。

第三方依赖验收要核对实际import位置与版本，而非只看requirements。2026-10-31的纯Python
`packaging 26.2`随包vendor，正式工具进程从其`vendor/packaging`路径加载，版本排序结果独立核对，
wrapper内层exit0、stderr为空、同task completed，导出包字节一致；没有在运行环境pip安装。
缺vendor的构建负例明确失败。这不证明带原生扩展或任意依赖版本都可用；保留来源、许可证及实际文件hash。

同一库原样vendor的空`py.typed`标记曾使正式load返回400：`Column 'content' cannot be null`。
确认根Agent/版本404后，仅给这个允许空白的静态类型标记加换行并重新build/verify/load，加载与真实运行通过；
Python源码字节不变，调整单独留证。这是特定输入的对照，不能自动改写任意空文件、删依赖或宣称后端已修复。

Go-only最小包的结果必须按每次真实加载分别记录。2026-10-31 两个几乎相同的约23MB Go-only
包中，一个 build/verify/load 成功并生成 `runnable` 版本，另一个正式 load 返回 `ErrTimeout`，
未生成 Agent 根资源。这个对照说明 Go 语言和最小包路径至少有一次能通，但线上加载稳定性尚未证明。
遇到超时先按原 agent/version 查询根资源和版本状态；没有资源就不要重复换 ID、换版本或把超时写成
“Go 不支持”。本地 build/verify 只能证明包结构和本机产物，不等于线上工具已执行。

Go、C++、Rust 和网络工具的结论要分别记：语言打包、Agent 加载、工具实际调用、依赖导入、网络
访问是五个独立断言。某一个断言通过不能替代其它四个；某次包超时也不能外推为语言或运行时一概不支持。
本机Docker daemon未启动时，C++/Rust的Linux构建保留“构建环境缺条件”，不拿Darwin产物代替Linux运行验收。

## Go真实执行、大包超时与在途版本

2026-11-01新增独立Go包约22MB，load返回ErrTimeout后，按原agent ID查询发现根及1.0.0 runnable已创建；未重传。沿原对象发一次任务后completed，实际工具language=go、entry_file=main.go、wrapper exit0，输入[17,-4,29,0,8]的sum50及SHA256与独立计算一致。核对原task、runtime manifest及同callId的真实tool result；模型复述或build/load成功不算运行证据。前轮“超时未生成根”与本轮“超时已创建”都是各自真实观察，不能概括为超时必然成功或失败。

导出同包仍超时，预期22,043,039字节而本地仅11,886,439字节。保留错误和`.partial`文件，不作为可安装包交付；长度/hash/包结构未完整核验前，文件存在不等于导出成功。Python包的历史完整导出结论不变。

默认版本切换可能重新按包声明材料化bindings。本次包`skills=[]`时，切default会移除之前追加的托管Skill；正确顺序是先切默认、再基于最新回读状态增量补绑定、独立回读bindings，随后发任务并核对runtime manifest。不要用旧快照覆盖并发变更。

本轮在旧task working时切默认与Skill绑定并提交新task：旧task仍Agent1.0.0/Skill A，新task为1.1.0/Skill B，两个任务时间实际重叠，结果独立核对；回滚后历史manifest及结果不变。这证明在途任务冻结和后续任务采用新默认，不证明两个并发load升级请求/CAS冲突。清理Agent根后另核对精确绑定tool；本轮根删除并未自动删除tool，默认版本/lineage仍有元数据残留。

基于最新状态增量绑定与回读不等于请求自带CAS。当前CLI未提供对应预期revision参数时，不承诺原子防并发覆盖，不编造revision/CAS flag。

## 两个不同版本并发加载的限定结果

2026-11-01同一隔离Agent的1.1.0/1.2.0小Python包请求实际重叠约1.21秒，均load成功且runnable，旧1.0.0保留；加载后default及根工具绑定仍是旧版。随后显式切1.2、1.1并真实执行，各版专属工具结果129/89，升级前已启动的旧任务在切默认后才实际调用旧工具，结果49。任务manifest、实际tool ID/source_ref、同callId进程返回与独立计算一致；回滚default1.0后完整历史不变。

验收并发时保存各请求起止时间并核版本集合/default/bindings，不能用顺序load冒充；客户端在途重叠不证明服务端内部事务并行。两次不同版本成功本身不证明同版本竞争、任意并发量、并发切default原子性或CAS冲突保证。包加载不自动完成业务升级；默认切换后仍须回读绑定与真实任务。

当前正式0.1.8的全局及load/export help没有请求超时、续传或压缩输入。`max-upload-bytes`仅限上传大小，外层进程timeout/等待轮数也不延长CLI内部预算。本地owner的普通请求预算30秒与此前大包超时现象相符，但不是线上部署commit证明。没有新增可用输入时不反复新建同样Go大包；保留完整导出失败及partial，不能以小Python包成功抹去。

## 同版本首发、完成后重传与真实绑定

2026-11-01两个相同新版本小Python包请求重叠6.09秒，均返回同一digest/loaded_at，最终只新增一版，默认及工具绑定保持。版本已runnable后，顺序重传相同包或同版本不同内容包均400重复键拒绝；版本/默认未变，显式切新版后的真实工具仍执行原源码（独立结果159/hash正确），回滚后历史task/manifest不变。并发首发成功不是通用重复load幂等，不证明不同内容抢同一未完成版本、任意并发量或并发default的原子性。重复键先查原版本状态，不盲换ID/版本制造重复资源。

本次根Agent的`binding.tool_refs`将包工具标为`agent_package`，而`bindings`投影把同一工具标为`workspace_binding`。沿根refs保存真实工具ID，再独立回读tool.source_ref和runtime manifest核对，不仅按投影来源标签过滤后就判断绑定丢失。删除根后仍分别核工具和版本；默认版/lineage可能保留元数据。

## 同版本不同内容的并发首发

2026-11-01同Agent、同一新1.1版本的A/B小Python包具有不同digest；两次CLI请求除包路径外完全一致，在途重叠约1.13秒。本例B成功，A明确409/code4 `agent package load mode conflicts with persisted Agent`，最终只新增一份runnable版本且digest为B，default仍旧版。保留原错误，不把它精确归因某个内部digest CAS分支或服务事务并行。

独立导出与B完整包字节一致；显式切新版后，真实工具call/result、source_ref和运行manifest均为B，结果423而不是A的159；回滚后旧/新任务及manifest完整保持。A/B完成后顺序重传均400重复键，版本/default/bindings不变。冲突后先查最终真实版本/digest及当前默认，不能把自己请求失败等同“没有任何新版本”，也不换ID重建或自动抢默认。

这个双请求结果不保证固定A/B赢家、任意包/规模或并发默认切换原子性。正式输入未确认default的expected-old-version/CAS参数时保留该缺口，不能靠双写成功冒充保护；原默认版本/lineage/会话任务历史元数据可能仍保留，删除根不等于物理清除。
