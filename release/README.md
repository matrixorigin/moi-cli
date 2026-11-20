# MOI / Astra 配套客户端交付

独立 moi-cli 的下载、安装与源码构建见 [CLI README](../README_zh-CN.md#环境要求与安装)。以下内容面向维护 MOI / Astra 配套安装包的发布人员。

## 发布与更新

只更新 Astra／astra-edge 而不升级 moi-cli 时，设置 `moi_mode=reuse`、`moi_ref=<已有 bundle ID>`、`moi_manifest_sha256=<该 manifest 的 SHA256>`。`moi_version` 不再填写，由原 manifest 推导；`moi_source_sha` 保留原源码提交，不记为当前发布器提交。旧 moi-cli 归档（含内嵌 Skill）、版本、URL、大小和 SHA256 原样复用，不编译、不重新打包；仍用新版 Astra 与旧 moi-cli 做真实 Skill 和更新协议探针。默认 `moi_mode=source` 保持原有构建行为。两种复用独立：`astra_mode=reuse` 复用 Astra，`moi_mode=reuse` 复用 moi-cli。

MatrixFlow 源仓中的 `Release MOI Client Components` 工作流（`.github/workflows/release-moi-cli.yml`）支持手动触发。固定目标为杭州区公开读的 `astra-suite` bucket；不修改 ACL，不允许匿名写。默认 dry-run，上传仅允许 `matrixorigin/matrixflow` 的 `dev`，使用 repository Actions Secrets，不设人工审批。

`moi_mode=source` 时用 `moi_source_ref` 选择 moi-cli 源码，默认 `dev`，也可填写 `5.2.2-dev` 等上游分支、tag 或 commit。流水线在并行构建前将它解析为一个完整 SHA，所有平台的源码检查、Go 版本选择和编译都使用该提交，清单的 `moi_source_sha` 记录此来源。发布脚本仍由运行流水线的 `dev` 提交提供；不要把 Actions 的运行分支改为源码分支来发布。`moi_ref` 仍只用于复用已发布 bundle，与源码选择分开。

完整配置、权限、Secrets、JSON 清单、目录、发布步骤与测试见 [publishing.md](publishing.md)。MOI 默认直接从 Astra 指定源码构建 astra／astra-edge，无需等待其完整 GitHub Release；也保留已有 Release 下载模式。只更新 moi-cli 时选择 `astra_mode=reuse`，用已有 OSS bundle 及其 manifest SHA256 复用原 Astra／Edge 归档，不重新编译或打包它们。MOI 私有仓库负责配套清单；官网下载入口不持有上传凭据。Astra 继续独立发布开源版本，原有 Memoria／Docker 方式不变。

本机发布包含 `astra`、`astra-edge`、`moi-cli` 三个组件，按组件与平台存储在 `components/`，配套清单位于 `bundles/`。`operation=build` 只发布 QA，bundle 版本以 `-qa` 结尾，渠道为 `channels/qa.json`。QA 验收后用 `operation=promote-environment` 复用同一批二进制生成无后缀的 Prod bundle，先发布固定脚本 `channels/stable/install.sh`，再推进 `channels/stable.json`；Prod 不重新编译、不重新上传组件。只更新 moi-cli 时选择 `astra_mode=reuse`，用已有 OSS bundle 及其 manifest SHA256 复用原 Astra／Edge 归档，不重新编译或打包它们。Skill 内嵌于 moi-cli 并随版本导出；不维护独立 Skill release。默认 `platforms=all` 发布四平台；QA 可选 `darwin-arm64` 或 `linux-arm64`。写入 stable 不能缩小已有平台集合，首次 Prod 必须是四个平台。首期清单采用普通 JSON，仅保留 HTTPS 与 SHA256 校验，不做发布签名。

发布需要配置 OSS AK/SK 两个 repository Secrets，以及非秘密服务地址 Variable `MOI_RELEASE_ASTRA_URL_QA` 和 `MOI_RELEASE_ASTRA_URL_PROD`。workflow 输入 `astra_url` 必须与对应环境的 Variable 完全一致。不需要 Environment 审批、签名密钥或公钥 Variable。关闭 dry-run 后自动上传 QA；勾选 `publish_channel` 只推进 `channels/qa.json`。组件版本只接受 `x.y.z`，不接受 `0.1.5-rc1` 这类预发布版本。所选平台的离线探针不能替代完整产品兼容验收。新版客户端已实现启动检查、按组件安装与 `astra update`，详见 [更新设计和测试](updating.md)。不自动迁移非受管的既有安装。

已是 Prod 语义的历史 bundle 可选择 `operation=promote` 续期 stable：只填该 bundle 版本和原上传摘要中的 manifest SHA256。它不能把 QA manifest 或带 `-qa` 的 bundle 写入 stable。QA channel 用 `operation=publish-channel` 对同一 QA bundle 和摘要续期，不重新构建。QA 到 Prod 的转换只能走 `promote-environment`。stable 已存在时不能更换 Prod Astra 地址；已安装客户端要换地址，用新的安装目录重装。默认 dry-run 只匿名读取 OSS；关闭后按操作写入对应渠道，无需人工审批。

MOI 托管 astra-edge 使用固定版本 Runner 镜像更新，不在 Pod 内检查客户端 stable 或替换自身 binary；镜像发布、排空和部署不属于本 workflow。

## 现有本地配套工具

新三组件安装入口是发布时生成的 `bundles/<version>/install.sh`（生成器 `channel_installer.py`）。QA installer 读取 `channels/qa.json`，Prod installer 读取 `channels/stable.json`；Prod 晋升后还有相同字节的固定脚本 `channels/stable/install.sh`。它引导校验后的 moi-cli，调用共享更新器安装，再复用 `client_setup.sh` 配置用户级 PATH 与配套启动脚本。正式 Astra HTTPS 地址由对应环境 allowlist 写入清单与启动脚本，不含凭据。执行 `astra update check` / `astra update` / `astra update --rollback`；本地 Edge 需先退出，托管 Edge 不参与此流程。官网公开入口 `https://get.matrixorigin.cn/astra` 的跳转仍由 `deploy_workflow` 的国内官网 Nginx 管理；首次改指固定脚本须等它已公开发布并核验，本仓库不改跳转。引导程序里的 `moi-cli` 不是最终安装版本，最终版本由该脚本写死的 channel 决定。约束见 [公开安装入口](publishing.md#公开安装入口)。

更新器检测到活跃进程返回 `CLI_UPDATE_BUSY`；中断留下事务时普通业务返回 `CLI_UPDATE_RECOVERY_REQUIRED`，重新执行 `astra update` 会先恢复，再尝试更新。升级不自动改写 shell profile、认证或业务配置。

后台检查占用与运行占用分开：显式更新最多等检查锁 4 秒，仍忙则报 `CLI_UPDATE_CHECK_BUSY`，稍后重试即可。跨兼容合同的发布由 `update check` 返回 `reinstall_required` 并在启动时提示，不会静默停留也不会强行升级。新装须使用独立安装目录和 Skill 搜索目录。原址重装的归属检查、缺失受管 Skill 副本的有限恢复及错误码见 [恢复边界](updating.md#本地文件缺失的恢复边界)。

以下低层打包及安装探针用于发布兼容性检查；面向用户的入口是上述三组件安装器。

## 构建与打包

维护配套安装包时，在 MatrixFlow 完整源码检出的根目录执行 `make -C sdk build-cli-platforms`，生成 Linux/macOS × arm64/amd64 的 moi-cli。Astra 复用其 `.github/workflows/release-binaries.yml` 中的四种原生目标构建方法；不要把交叉编译成功当作目标平台运行通过。

`bundle.py --help` 列出必需参数：Astra binary、moi-cli binary、Astra LICENSE、配套版本号、输出目录。打包在实际目标平台运行两端的版本和新认证接口探针，生成包含两端精确版本的 `bundle.txt`、归档与 SHA256。输出名已存在会失败，不覆盖已有版本。

打包探针不是浏览器或完整产品验收；发布前仍要在每个目标平台执行统一登录与生命周期矩阵。部署的正式 Astra URL、发布地址和正式版本由部署方明确提供，本工具不会猜测。

## 安装

在包含 CLI `go.mod` 的目录执行 `sh release/install.sh --help` 查看配套包安装参数。必须指定 `--archive` 本地归档或 `--url` HTTPS 地址，并提供可信渠道取得的 `--sha256`。默认目录为 `~/.local/share/moi/cli`，也可 `--dir` 指定专用绝对路径。

安装器验证归档摘要、成员白名单、当前 OS/架构、两端版本、认证接口及离线 Skill 导出。包协议为 `moi-astra-native-v2`；旧开发预览包不能冒充新包。每个归档进入独立不可覆盖目录，通过一次 `current` symlink rename 激活配套 binary；旧版本保留。不修改既有登录、配置或 shell 文件。

Skill 默认安装到 `ASTRA_LOCAL_STATE_ROOT/skills/moi-cli`（若配置），否则 `~/.astra/skills/moi-cli`；用户无需在项目内安装。`--skill-dir` 可显式指定已配置的搜索目录。Astra 拒绝真实路径越出搜索根的 Skill，故不能从该目录链接到 binary 目录。安装器在搜索根内创建 `.moi-cli-releases/<包摘要>`，并将 `moi-cli` 指向其中的不可变副本。owner 文件绑定安装前缀；已有用户自定义 moi-cli Skill、其他安装前缀或被修改的托管副本均拒绝覆盖，其他 Skill 不受影响。

激活顺序是 Skill 指针、binary `current` 指针，各自原子 rename。捕获到激活失败或普通中断时恢复旧 Skill；这不是跨目录的原子事务。在两次 rename 之间 SIGKILL/断电可能留下版本暂时不一致，应检查并重新运行同一可信包完成安装；不要在升级期间启动新任务。不可变副本先在临时目录复制完毕再发布，避免半拷贝占用正式摘要目录。

成员白名单排序显式使用 `LC_ALL=C`，不依赖客户的 locale，也不改变客户终端的语言设置。串联下载、安装、激活的示例使用 `&&`：前一步失败就停止，不能把 PATH 中旧 binary 的输出当作本次安装成功。

将安装目录的 `current` 加入 PATH。首次使用 `astra --api-url` 显式选择受信任的 MOI Astra origin 并 `login`；后续 `astra login` 和 moi-cli 复用已保存环境。也可在首登前设置 `ASTRA_API_URL`。不得在不可信项目配置中设置该入口。

安装进程遭 SIGKILL 时可能留下 `.install-lock`，应先确认原进程已退出并检查该安装目录，再移除这个空锁目录；不得通过删除整个安装目录恢复。旧版本没有自动删除策略。

## 验证

跨仓真实加载检查使用 `check_native_skill.py --astra PATH --moi PATH --astra-license PATH`。显式提供当前宿主可执行的真实 binary；脚本实际打包、安装并执行 `astra skill list --source local`，断言 Skill 位于允许的搜索根且为 local/active。全过程使用临时安装、状态和认证目录，完成后自动清理，不登录、不调用模型、不修改用户 profile。此检查与用替身验证失败路径的安装器单测分开。

低层安装回归使用显式 `--preview-state` 将 Skill 放入测试临时目录；此选项不启用任何服务、不写入固定服务地址。正式 Skill 从 binary 导出，不单独下载不匹配的最新 Skill。

在 MatrixFlow 完整源码检出的根目录执行 `make -C sdk cli-release-test`，覆盖摘要失败、归档路径穿越拒绝、原子升级、重复安装、激活失败后恢复、旧版本保留、用户文件保护、输入不可覆盖、本机 HTTP 显式许可与禁止重定向。低层安装器覆盖 `C`、`en_US.UTF-8`、`zh_CN.UTF-8` 环境。

## 联调隔离边界

个人客户旅程的下载预览、固定本机端口和服务启动/数据准备脚本不属于产品交付，不接入默认 build/test/start 或发布 workflow。正式 `channel_installer.py` 只接受 HTTPS 渠道和显式提供的 Astra HTTPS 地址，不生成个人预览环境。

安装器回归在临时 HOME、认证/状态目录和 shell 配置中运行，不继承个人登录目录；仅使用合成 binary 或测试显式提供的 binary，不登录、不创建业务账号、不赠额。跨仓真实联调需要单独显式执行，不属于 `make test` 的副作用。
