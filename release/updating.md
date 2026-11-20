# MOI 客户端更新设计与测试

## 范围与归属

MOI 管理的本机 Astra、astra-edge、moi-cli 与内嵌 Skill 共用一个更新器，owner 为 moi-cli 的 `internal/update`。`astra update` 委托同一不可变版本目录中的 moi-cli，不搜索 PATH。独立安装的 Astra/Memoria、服务端与 Kubernetes Runner 不加入公共更新渠道。首期自动检查、提示，由用户显式升级；不静默替换，不要求重新登录。

入口：`astra update check`、`astra update`、`astra update --rollback`。moi-cli 提供同语义入口供安装引导程序及独立使用者调用。首次安装入口引导一个经摘要验证的 moi-cli，再调用同一个更新器；不在 shell 中实现第二套更新逻辑。旧双 binary 开发预览不自动认领，需安装到新的专用目录。

交互安装/更新在 stderr 显示下载百分比（实际写入字节 / 清单大小），并显示校验、解包、Skill 检查和激活阶段；100% 仅表示下载结束，不代表安装完成。引导脚本通过 curl 进度条展示首次 moi-cli 下载。只有 stderr 为 TTY 且非 `TERM=dumb`、`MOI_AGENT_CALL=1`、CI 时启用；因此 `curl | sh` 仍可显示，重定向与 Agent 调用保持安静。后台 refresh 永远静默，stdout JSON 合同不变。首次从无进度条的旧更新器升级仍使用旧显示，安装新版后后续升级才有进度条。

## 文件与并发合同

首次安装及显式重跑安装器会准备认证目录：使用绝对路径 `MOI_AUTH_DIR`，未设置时为 `~/.moi`。缺失时创建为 0700；已存在且确为当前用户拥有的普通目录时，通过打开的目录描述符收紧到 0700，并复核路径身份。允许末尾斜杠，但拒绝点路径段、符号链接、其他用户目录、文件、空路径及根目录，返回 `CLI_UPDATE_AUTH_DIRECTORY`。不递归 chmod、不 chown、不读写或删除目录内的凭据；普通更新和后台检查不执行此步骤。该步骤失败时安装不能宣称成功，用户应检查指定目录后重新运行安装器。

安装器只准备安装进程实际读取到的目录；安装后修改 `MOI_AUTH_DIR` 不会自动准备新目录。Astra 登录前仍以当时的环境变量检查实际认证目录，其权限校验才是运行期保障。

- 专用 prefix 下：`installation.json`（渠道、平台、Skill 搜索根、`compatibility_reference`、Astra Server origin，无凭据）、`runtime.lock`、`check.lock`、`cache.json`、`state.json`、`transaction.json`、`releases/<manifest-sha256>/`、`current`。安装器把自身版本清单中的合同和服务地址写入该文件；stable 与任一安装绑定字段不同就拒绝激活，避免旧安装器把旧地址与新 binary 混装。升级本功能前已经激活的安装若两个新字段都不存在，仅从本地已验 current manifest 派生内存合同；不读取 mutable channel、不写迁移文件，半缺字段或没有 current 的安装拒绝。
- release 保存原始 manifest、组件归档、解包后的 binary、组件 LICENSE、导出的 Skill 和文件摘要。只复用摘要匹配的组件；相同版本不同内容拒绝。不会覆盖旧 release。
- 全局 Skill 根内 `.moi-client-releases` 保存该安装独占的不可变 Skill 副本；`moi-cli` 为受管指针。自定义同名 Skill 不覆盖。
- Astra、Edge、moi-cli 普通调用从实际 executable 定位安装，持有 `runtime.lock` 共享锁直到进程退出。更新持有独占锁；已有消费者时明确报 busy，不杀进程。启动在持锁后检查当前版本与事务状态，拒绝旧 executable 或半完成切换。
- 切换前持久化 transaction，依次切换 Skill、current、state，最后删除 transaction。中断后普通业务拒绝启动；`update` 先恢复到事务记录的旧版本，首次安装则恢复到未激活状态。不依赖 PID 文件或瞬时进程扫描。
- transaction 删除并同步目录才视为提交完成；即使 state 已写入，只要 journal 仍在，恢复仍回到旧版本，之后可以重新更新。

## 下载与信任

消费发布端 `moi-client-channel-v1` / `moi-client-bundle-v1`。固定 HTTPS 渠道及同源下载，不发送 UC/PAT/OSS 凭据、不跟随重定向；仅显式本地测试允许字面 `127.0.0.1` HTTP。先验证有效期、单调序号、manifest 原始字节 SHA256、平台及组件字段，再下载有变化的组件。归档验证大小、摘要、白名单、普通文件类型、权限与解压上限，验证完成后才执行版本、更新占用协议和 Skill 探针。

不做签名（已确认的产品范围）；HTTPS + 摘要不能抵御 bucket 写权限泄露。渠道兼容验收标识不是在线服务端能力探测。首期只在相同 `compatibility_reference` 内升级和回退，跨合同发布明确拒绝，需另行验收安装，不把文本标识当成自动证明。禁止渠道推进导致组件自动降级；显式回退只允许保留的上一版本、相同合同与已验证本地摘要。

## 提示与保护

共享检查缓存/锁，1 小时限频：启动时距离上次成功检查不足 1 小时就复用缓存，否则发起后台检查；不是运行期间每小时轮询。后台网络检查短超时，离线不阻塞业务；显式 check 的错误必须可见。机器输出、`--version`、帮助、认证 helper 不触发提示；`MOI_AGENT_CALL=1` 显式关闭 Agent 提示。绝不读写认证文件、模型 Key、业务配置或聊天状态。升级只管理明确归属本 prefix 的文件，无自动清理旧版本。

后台 refresh 不等待 `check.lock`，网络预算为 3 秒。显式 check/update 对该锁最多等待 4 秒，超过后返回 `CLI_UPDATE_CHECK_BUSY`，提示稍后重试；调用方取消或等待预算被调用方截止时间截断时返回 `CLI_UPDATE_CHECK_CANCELLED`。`runtime.lock` 仍立即拒绝运行占用，使用 `CLI_UPDATE_BUSY`，不把真实 I/O 错误伪装成占用。

跨 `compatibility_reference` 或 Astra Server origin：`update check` 成功返回 `status=reinstall_required`、`reason=CLI_UPDATE_COMPATIBILITY_CHANGE` 和安全操作提示；缓存保留同一 reason，但 `available=false`，含义是不可就地升级。Go/Rust 启动提示同时识别此状态；只展示校验后的版本标识与固定 reason 映射，不输出下载的任意文案。`update` 仍拒绝激活。用户应使用对应 stable 的官方版本化安装器和新目录；自定义 `--skill-dir` 必须是绝对路径且以 `/skills` 结尾，配套 launcher 自动将其父目录设置为 `ASTRA_LOCAL_STATE_ROOT`，无需用户再配置 Skill 搜索路径。保留旧安装和认证，不手改清单或共用旧安装独占的 Skill 槽位。

首次安装的 shell 配置发布也采用恢复合同：`setup.lock` 保存 owner PID，活跃 owner 不被抢占，进程已经退出留下的锁可安全回收；profile 在同一文件系统写入完整临时文件后原子替换，中断不会留下半段 PATH block。原 profile 在修改前仍备份到安装目录的 `setup/`。

### 本地文件缺失的恢复边界

以下是更新器 owner 内的显式恢复例外，不扩展为任意文件损坏时自动覆盖：

| 情况 | 行为 |
| --- | --- |
| 专用 prefix 被用户删除，原址重装 | 仅在新 prefix 无 current/state、releases 为空，且全局 Skill owner 精确匹配该 canonical prefix、指针目标属于该受管 store 时，删除残留 symlink 后重新安装。不会删除 store 或自定义 Skill。 |
| 只删 state 或 current，release 还存在 | 不视作首次安装，拒绝指针/状态不一致，不能借重装覆盖活跃安装。 |
| 全局 `.moi-client-releases/<hash>` 整个副本缺失 | 持独占运行锁，先校验本地 release 的 manifest、全量文件摘要和 binary 权限，再从其 `skills/moi-cli` 暂存复制、同步并原子发布。适用于当前版本、更新复用、回退目标和事务恢复目标。 |
| 副本部分缺失、内容变化、symlink 或普通文件占据目标 | 拒绝，不覆盖；源 release 损坏时也不重建。 |
| 只读完整性验证发现副本缺失 | 返回 `CLI_UPDATE_SKILL_COPY_MISSING`，不产生写入。 |

权限不足、磁盘满、必需安装文件缺失分别映射为 `CLI_UPDATE_LOCAL_PERMISSION`、`CLI_UPDATE_LOCAL_DISK_FULL`、`CLI_UPDATE_LOCAL_FILE_MISSING`，不泄露原始路径和 OS cause。未知 I/O 仍使用 `CLI_UPDATE_LOCAL_IO`。这些恢复不会恢复已被用户删除的认证、会话或业务数据，不建议以删目录作为日常维护手段。

首期仍不提供 `--prune`，也不降低全量哈希校验强度。不要手删 current/previous 指向的 release 或 Skill 副本来回收空间。受控回收入口和减少重复读取的优化单独跟进。

## 测试矩阵

| 分组 | 必须覆盖 |
| --- | --- |
| 清单 | 正常、坏 JSON/摘要、未知 schema、过期/未来时间、序号回退/同序号换内容、错误平台/组件/版本/URL、大小超限 |
| 传输 | HTTPS、显式 loopback、跨源/重定向拒绝、非 200、截断、超时；不携带认证头 |
| 归档 | 正常三组件、缺成员、重复、路径穿越、绝对路径、symlink/hardlink、目录、特权权限、解压超限 |
| 安装/更新 | 新装、无更新、仅 moi-cli 变化时复用 Astra/Edge、校验失败原版不变、同版本变内容拒绝、版本降级拒绝 |
| 恢复 | 两个指针各阶段中断、state 写入失败、事务残留拒绝业务、显式恢复、恢复重复执行、上一版本回退及损坏拒绝 |
| 手动删除 | 原址重装、缺失 Skill 全副本重建、无更新/更新/回退/事务恢复/复用目标；缺状态、错误 owner、外部指针、自定义目录、损坏源、部分副本和 symlink 拒绝 |
| 并发 | 多更新者、Astra/Edge/CLI/helper 共享占用、忙碌时拒绝更新、旧版本启动竞态、进程退出自动释放锁 |
| 隔离 | 非受管安装不变、自定义 Skill 不覆盖、目录 symlink 拒绝、认证/配置/工作区 sentinel 不变 |
| 提示 | 1h 限频（小时内、到期、超期、时钟回拨）、并发检查、离线、机器输出/Agent/helper 不污染、显式 check 不隐瞒失败 |
| 检查锁与终态 | 显式 check/update 等待后成功、后台不等待、超时、取消、非占用 I/O；跨合同 check/缓存/Go 与 Rust 提示可见但不激活 |
| 跨项目 | Go/Rust 同锁互斥、真实 binary 探针、四平台构建；真实本机两版本安装→升级→回退 |

测试仅使用临时目录和本机测试 HTTP，不写 OSS，不动用户正在使用的 r4 预览。后续记录实际通过项和未执行项，不能以 fake binary 的状态机测试替代真实 binary 验收。

## 2026-09-17 本地执行记录

- Go 更新器 17 个测试组及子用例通过，包含 race 检查；CLI 全包测试与 `go vet` 通过。
- `make -C sdk capability-alignment-test`、`cli-skill-check` 通过；moi-cli Linux/macOS × amd64/arm64 交叉编译通过（不等价于四平台原生运行）。
- Astra Core 安装占用的 3 个 Rust 测试通过；Core clippy（warnings denied）与仓库 Rust format check 通过。
- 发布回归 59 项通过；旧双 binary 安装器回归 16 项通过。
- 真实 macOS arm64：当前 Astra/Edge 与分别标识为 0.1.1、0.1.2 的两份 moi-cli，通过 `check_native_update.py` 完成下载引导、三组件安装、Astra 实际 Skill 发现、Astra 委托更新、仅 moi-cli 下载、Go/Rust/Python flock 互斥、旧 executable 拒绝、回退和合成认证目录保留。
- 真实测试发现并修复：渠道时间 `+00:00` 经本地 JSON 往返变为 `Z` 后，`time.Time` 结构比较错误拒绝同一渠道；现在按时间值比较，并有专项回归。测试合成认证目录也使用 native 认证要求的 0700 权限，不放宽生产权限检查。

本轮不登录真实账号、不调用模型；认证保留由合成 sentinel 和“不调用任何认证 API”的路径验证，不声称已覆盖真实 UC refresh。没有触发 GitHub Actions、OSS 上传、正式分发、其他三平台原生升级、真实断电或托管 Runner 镜像升级。恢复测试覆盖各事务阶段的持久化残留与重复恢复，不把它描述为真实断电实验。

重复真实验证：使用包含本次改动的宿主 Astra/Edge，构建两份不同版本号 moi-cli（`-ldflags=-X=main.version=0.1.1` / `0.1.2`，文件 basename 都为 `moi-cli`），向 `check_native_update.py` 传入 `--astra`、`--edge`、`--moi-before`、`--moi-after`、`--astra-license`。需要发布脚本已锁定的 Python 依赖环境。

当前 `check_native_update.py` 通过真实 `moi-cli update install --allow-loopback-http` 初始化临时安装，后续仍由真实 Astra 委托执行更新和回滚。该开关只用于本地回环验收；正式 `channel_installer.py` 继续拒绝 HTTP，脚本不再尝试为 HTTP 夹具生成正式安装器。HTTPS shell 引导的生成和执行由 `bootstrap_test.py` 单独覆盖，不能将本地回环验收记作真实公网 HTTPS 引导已通过。

### 同日审查修复回归

- Go 更新器现有 25 个顶层测试（原 17 + 新增 8，包含参数化子用例）；`go test -race ./internal/update ./cmd/moi-cli -count=1`、CLI 全包 `go test ./...`、目标包 `go vet` 通过。
- 新增覆盖同址重装、缺状态与外部归属拒绝、5 种缺 Skill 副本恢复路径、4 种损坏/占位拒绝、显式两种命令等待检查锁、后台不等待、超时/取消、跨合同提示与原版本保留、错误文案不泄露路径。
- Astra Core 4 项测试、warnings-denied clippy、Rust 格式检查通过；目标 binary 构建通过。macOS debug CLI 链接器报告 `__eh_frame` 过大的 unwind 性能警告，不是编译失败。
- SDK 能力对齐、内嵌 Skill 生成一致性、moi-cli 四平台交叉编译、59 项发布测试与 16 项旧安装器测试通过。
- 重新构建真实 macOS arm64 Astra/Edge 与两版 moi-cli 后，完整 `check_native_update.py` 通过：安装、Skill 发现、仅 moi-cli 更新、三 binary 同锁互斥、旧 executable 拒绝、回退、无版本变化时重建缺失 Skill 副本、保留全局归属的原址重装、跨合同发布可见但不激活、合成认证目录保持不变。
- 新增验收夹具的路径保护曾因 macOS `/var` → `/private/var` 别名误拒绝，已将临时根统一 canonicalize 后重跑通过，没有放宽生产路径检查。所有删除操作只针对本轮生成的临时测试安装和 Skill 副本。
- 磁盘满和权限不足仅验证公共错误分类，未实际耗尽磁盘或改变用户目录权限；四平台交叉编译不等于各平台原生验收。
