# MOI 客户端 OSS 发布操作

入口：[release-moi-cli.yml](../../../.github/workflows/release-moi-cli.yml)（仓库路径 `.github/workflows/release-moi-cli.yml`）。GitHub Actions 中名称为 **Release MOI Client Components**，仅手动触发，不随代码合入发布，不创建 GitHub Release，不部署托管 Runner。

## 独立 QA 快照测试（无需 Astra Release 或版本号变更）

入口：[astra-qa-snapshot.yml](../../../.github/workflows/astra-qa-snapshot.yml)，Actions 名称为 **Build Astra QA Snapshot**。这是与下文正式组件发布并行的测试路径。正式的版本校验、`components/`、`bundles/`、`channels/qa.json` 和 `channels/stable.json` 均保持原合同。

- 从合入后的上游 `dev` 运行；填写 `astra_ref`（默认 `main`）与平台。resolve 一次固定 Astra 完整 SHA，读取 Cargo 版本，但不要求该版本尚未发布、不修改版本文件，也没有 `astra_version` 输入。快照 ID 为 `<Astra SHA 前 12 位>-<Actions run ID>-<resolve attempt>`。新 run 或重新运行全部 jobs 会生成新 ID；仅重试失败 job 时沿用 resolve 已固定的身份，避免与已成功的平台产物混用身份。
- `moi_mode=source` 编译 workflow 固定的 Matrixflow commit，保留 binary 的源码版本，无需填写新版本号。`moi_mode=reuse` 则填写已有 `moi_ref` 与 `moi_manifest_sha256`，保留该归档和内嵌 Skill 的原始字节。快照仍将这些 bytes 复制到独立快照目录，不写正式组件对象。
- 默认 `dry_run=true` 只构建验证。取消勾选后，只有 `matrixorigin/matrixflow` 的 `dev` 能上传；沿用发布 Secrets 和 `MOI_RELEASE_ASTRA_URL_QA` Variable，不能输入任意服务地址或 Prod 目标。
- 产物只写 `qa-snapshots/<snapshot-id>/<platform>/<component>.tar.gz`、`manifest.json` 和 `install.sh`。同 ID 不同内容仍拒绝覆盖；同一 Cargo 版本的不同 commit、不同构建互不冲突。不生成 channel，不提供快照直接晋升 Prod 的入口。
- 原生构建复用固定 Astra 源码的构建/TLS/执行门禁和真实 Skill 探针。新增 smoke 用已构建归档替代下载传输，实际运行快照安装器、三个真实 binary 的版本命令、`astra-qa login --help`、继承外部配置路径时的真实离线配置写入与 Astra Skill 发现；它不模拟真实 UC 登录，不代表线上全链路验收完成。
- 跨 job 的 Actions artifacts 仅作传输，保留 **1 天**；最终下载产物在 OSS。此入口不会做到完全零 artifacts 占用。已上传 OSS 快照没有自动清理策略，清理应另行安排。

首次正式上传前，凭据 owner 需要确认现有 RAM 用户拥有 `astra-suite/qa-snapshots/*` 的 GetObject/PutObject 权限及现有 bucket versioning 查询权限。若授权仍只覆盖正式目录，新路径会失败；工作流不会自行修改 RAM、bucket ACL 或版本控制。

上传 Summary 提供固定安装链接。运行方式：

```sh
curl --proto '=https' --tlsv1.2 -fsS '<Summary 中的快照 install.sh HTTPS 地址>' | sh
# 若 ~/.local/bin 尚未在 PATH，按安装器输出添加后执行：
astra-qa login
astra-qa
# 所选源码支持 Device Code 时：
astra-qa login --device-code
```

默认安装根目录为 `~/.local/share/moi/qa-client`。只创建 `astra-qa`、`astra-edge-qa`、`moi-cli-qa` 命令；不创建或覆盖无后缀的正式命令。自定义目录使用 `sh install.sh --dir /absolute/qa-directory`；`--no-modify-path` 不在 `~/.local/bin` 建链接，直接执行安装根目录下 `bin/astra-qa`。

每个快照有独立的认证、会话与 Skill。切换后首次需要重新登录；旧快照和其状态保留。QA 启动器设置 `ASTRA_LOCAL_STATE_ROOT`、`MOI_AUTH_DIR` 和配套 `MOI_ASTRA_BINARY`，并显式覆盖继承的 `ASTRA_SERVER_URL` 为 QA 地址；Astra 入口显式传入 QA API 地址。所有入口将 `MOI_CLI_CONFIG_FILE` 绑定到快照私有配置文件，避免继承正式配置路径，也覆盖 Astra 直接调用配套 moi-cli 的子进程。`moi-cli-qa` 的 HOME 仅对该子进程指向快照私有目录，避免读取或改写用户的正式 `~/.moi-cli.json`。Astra 本身保留真实 HOME，仍可按授权访问用户工作文件。配套子进程通过快照内的 PATH 使用同一套三个 binary，不跟随之后的快照切换。

重复安装同一个 ID 会明确拒绝，可继续用现有命令或换独立目录。同版本修复后重新运行 workflow 获得新快照；新快照完整安装后通过一个 QA `active` 指针切换命令，旧进程仍指向旧快照。不要删除正在运行的旧快照，也不要把正式安装目录传给快照安装器。异常退出留下安装锁时，先确认安装进程已退出再处理锁，不自动抢占。

快照属于 standalone 安装，`astra-qa update` 不受现有更新器管理。该路径用于候选客户端功能测试，不替代正式 bundle 安装、升级、回滚及最终发布产物的验收；也不构建/更新 Astra Server 或托管 Runner 镜像，不修改 QA 部署及自托管页面的下载配置。

## 当前范围

### 只升级 Astra，保留 moi-cli

使用 `operation=build`、`moi_mode=reuse`，并提供不可变的 `moi_ref`（来源 bundle ID）与 `moi_manifest_sha256`。`moi_version` 留空，resolve job 从摘要校验通过的 manifest 推导版本和原始 `moi_source_sha`；后续所有构建、组装、上传步骤使用同一份解析结果。可同时选择 `astra_mode=source` 并提供新版 Astra 的版本与 commit。

例如保留当前已发布的 moi-cli `0.1.7` 时，选择 `moi_ref=2026.09.28.1`、`moi_manifest_sha256=0ecf89741d1f8086a8248c995d9fde118bbdc5d13c3b9923e9933100fd982677`，Astra 选择 `astra_version=0.2.11` 与对应源码 commit。新 bundle ID 必须独立且以 `-qa` 结尾。不要把新代码的 `moi_version` 填成旧版本来模拟复用。

每个平台匿名下载旧 moi-cli 归档、校验大小／SHA256／成员白名单，运行版本、真实 Skill 发现和更新占用协议探针，再原样保存归档。来源不覆盖所选平台、摘要／版本／源码不匹配或探针失败时终止，不回退到源码构建。publisher 已有对象冲突预检会复用现有 OSS 对象，不重新上传 moi-cli。仅新的 Astra／Edge 与 bundle 元数据被上传；原来源 bundle 不变。

安装／更新仍消费完整的三个组件清单，moi-cli 同版本同摘要保持不变。QA 到 Prod 的晋级沿用原入口，不重新构建；兼容标识、地址约束和渠道降级检查不变。复用二进制不代表确认新版 Astra 与服务端业务协议兼容，仍需 QA 验收。

- 在 Linux/macOS × amd64/arm64 原生 runner 构建并运行 moi-cli，执行 CLI 单测、Skill 生成一致性及既有安装器回归。
- 默认 `source` 模式：将 `matrixorigin/Astra` 的分支、tag 或 commit 一次解析成固定 SHA，再在所选平台编译 astra／astra-edge（默认四个平台），不等待 Astra GitHub Release，不构建服务端镜像、Web 或 SDK 包。分配 native runner 前检查源码版本及上游构建步骤合同，编译后执行 TLS、版本、help 和真实 Astra Skill 发现检查。
- 保留 `release` 模式：从已发布的明确版本获取 astra／astra-edge，核验 GitHub Release asset 的 SHA256、大小、归档白名单及 binary 版本，执行同样的离线探针。不回退到 latest，不是只做交叉编译。
- 增加 `reuse` 模式：按已有 OSS bundle 和可信 manifest SHA256，直接复用 Astra／Edge 原始归档，不重新编译、不重新打包；仍在所选平台运行版本、help 和 Skill 探针。来源 bundle 必须包含全部所选平台，否则在分配 native runner 前失败。
- 生成三个组件各自的确定性归档、所选平台完整配套清单。Skill 内嵌在 moi-cli 中，不作为独立对象上传。不能把缺失的候选产物自动当成单平台发布。
- 上传到 `oss://astra-suite` 的 `components/`、`bundles/`。`operation=build` 只生成 QA bundle；勾选 `publish_channel` 时最后更新 `channels/qa.json`，不写 Prod 固定脚本或渠道。`promote-environment` 复用 QA 组件，只上传新的 Prod manifest／版本化 installer，再更新 `channels/stable/install.sh` 和 `channels/stable.json`。相同组件内容复用，已有同名不同内容拒绝覆盖。
- 候选构建与验证 job 不接触 OSS Secrets。只有上传／渠道推进 step 持有发布凭据，不在这些 step 执行下载的 binary。

客户端 `astra update` 和清单安装器见 [updating.md](updating.md)。本 workflow 额外生成不可变的 `bundles/<version>/install.sh`，其摘要和大小记录在 manifest 的 `installer` 字段中；晋级前校验该脚本匿名可下载。Prod 晋升时再复制到可变的 `channels/stable/install.sh`，不写 bucket 根目录的 `install.sh`。不修改旧双 binary 预览协议。QA 与 Prod 的 Astra 入口分别由 repository variable `MOI_RELEASE_ASTRA_URL_QA`、`MOI_RELEASE_ASTRA_URL_PROD` 提供。workflow 输入 `astra_url` 必须与对应 allowlist 完全一致，并写入清单的 `client_config.astra_url` 和无秘密启动脚本；未配置或不是合法 HTTPS origin 时，在构建前失败，不回退 localhost。组件版本只接受 `x.y.z`。QA bundle 版本必须以唯一的 `-qa` 结尾，Prod 版本由去掉该后缀得到。四平台离线探针不等价于浏览器登录、真实模型、运行中升级或完整服务端兼容验收；选择发布 stable 前需完成这些验收。

所有组件必须响应 `--moi-update-protocol` 为 `moi-client-update-v1`，证明包含进程占用入口；旧 Astra release/reuse 产物不满足时，构建验证会拒绝发布，需选择包含本次改动的新版本源码。不能用旧版本号重新覆盖已有组件。

`compatibility_reference` 是配套验收合同标识，应在兼容的连续发布间保持一致；客户端只允许与安装器内嵌合同和 Astra Server origin 均一致的版本就地更新/回退，不等同于自动探测服务端兼容性。版本化安装器把这两个值写入本地安装合同；若 stable 已切换合同或服务地址，旧安装器会在激活任何版本前失败，用户必须取得对应 stable 的新安装器。旧开发预览安装需要显式使用新专用目录，不能无锁接管已经运行的旧版客户端。

## 固定目标与权限

| 项目 | 值 |
| --- | --- |
| Bucket | `astra-suite` |
| 区域 | `cn-hangzhou` |
| OSS endpoint | `https://oss-cn-hangzhou.aliyuncs.com` |
| 下载前缀 | `https://astra-suite.oss-cn-hangzhou.aliyuncs.com` |
| 发布仓库／分支 | `matrixorigin/matrixflow` / `dev` |
| 人工审批／GitHub Environment | 不使用，启动时通过 `dry_run`／`publish_channel` 选择行为 |

沿用用户配置的公开读、不公开写 ACL；代码不执行任何 Bucket/Object ACL 修改，不公开写，不扫描、同步或删除 bucket 内容。用户给出的 OSS 控制台 URL 不是下载 URL。此次没有使用凭据访问控制台或 bucket，因此未验证线上 ACL、版本控制、网络或 RAM 授权状态。

发布专用 RAM 用户建议只给以下权限，不复用管理员 AK/SK：

```json
{
  "Version": "1",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": ["oss:GetBucketVersioning"],
      "Resource": ["acs:oss:*:*:astra-suite"]
    },
    {
      "Effect": "Allow",
      "Action": ["oss:GetObject", "oss:PutObject"],
      "Resource": [
        "acs:oss:*:*:astra-suite/components/*",
        "acs:oss:*:*:astra-suite/bundles/*",
        "acs:oss:*:*:astra-suite/channels/qa.json",
        "acs:oss:*:*:astra-suite/channels/stable.json",
        "acs:oss:*:*:astra-suite/channels/stable/install.sh"
      ]
    }
  ]
}
```

首次上传会检查 Bucket 从未启用版本控制。OSS 的 `x-oss-forbid-overwrite` 对 Enabled／Suspended 的版本控制不生效，故本实现拒绝这两种状态，不自动关闭版本控制或削弱不可覆盖语义。若实际 bucket 已开启过版本控制，需要在首次发布前调整存储方案，而不是直接绕过检查。[OSS PutObject 官方说明](https://www.alibabacloud.com/help/en/oss/developer-reference/putobject)

渠道只有这一条 workflow 作为写入方；构建发布和独立晋级共用固定 concurrency group 与发布序号，禁止相互取消。QA 与 stable 各自比较自己的 sequence。发布序号和推进前重读能拒绝过期作业，但不是 OSS 条件写 CAS；不能另开脚本或其他仓库同时写同一个 channel 文件。候选版本对象通过禁止覆盖请求头保护。预检先 HEAD 比较大小，已存在对象只做一次 SDK 流式摘要校验；所有冲突检查完成后才写入，新对象上传后只做一次匿名流式下载校验，不重复完整读取或仅信任 metadata。晋级再做匿名校验，不重新读取整个归档到内存。匿名 HTTP／GitHub 下载的 429、部分 5xx、超时及连接重置最多重试三次；摘要错误和权限错误不重试，PUT 的不确定失败也不盲目重试。

## 配置 GitHub Secrets

在 `matrixorigin/matrixflow` 的 Settings → Secrets and variables → Actions 配置以下两个 repository Secrets。不需要创建 GitHub Environment 或设置 required reviewers；流水线中间不等待人工审批。上传／推进 job 仍校验仓库与 `dev` 分支，fork 只能 dry-run。请限制仓库写入和 workflow 修改权限；手动触发权限及凭据控制承担发布授权边界。

| 类型 | 名称 | 内容 |
| --- | --- | --- |
| Secret | `MOI_RELEASE_OSS_ACCESS_KEY_ID` | 发布专用 RAM AccessKey ID |
| Secret | `MOI_RELEASE_OSS_ACCESS_KEY_SECRET` | 对应 AccessKey Secret |
| Variable（非 Secret） | `MOI_RELEASE_ASTRA_URL_QA` | QA Astra Server HTTPS origin，不含路径、查询参数或凭据 |
| Variable（非 Secret） | `MOI_RELEASE_ASTRA_URL_PROD` | Prod Astra Server HTTPS origin，不含路径、查询参数或凭据 |

首期仍仅需要上述两个 OSS Secrets，不配置发布签名密钥或公钥。另需上述非秘密服务地址 Variable；不得填本地预览 URL。AK/SK 直接在 GitHub UI 填入，不发到聊天中。dry-run 不需要 Secrets，但仍需明确服务地址配置。

安装器将 binary 和 Skill 交给同一个更新器安装到 `~/.local/share/moi/client`，随后在 `bin/` 创建配套启动脚本并准备 PATH。bash 同时处理交互和既有登录 profile；zsh 尊重 `ZDOTDIR`；fish 使用其 `conf.d` 与 fish 语法；sh/dash/ksh 使用 `.profile`。未知 shell 不改配置，打印可立即执行的绝对路径。shell profile 是符号链接时不跟随、不改写，跳过该文件的 PATH 写入并继续安装；若因此没有可复用的 PATH 目录，安装结束时打印绝对命令，并说明用 `--shell-profile` 指定普通文件。`--no-modify-path` 同时禁止 shell 文件修改和用户 bin 命令链接发布，仍创建安装目录内的 launcher；已有同名命令或不同内容启动文件拒绝覆盖。只有已在 PATH 中的用户级 `~/.local/bin` 或 `~/bin` 可发布命令 symlink，不写系统 bin，不使用 sudo。子进程不能修改父 shell 的 PATH：如果没有可复用的 PATH 目录，须打开新终端或使用打印的绝对命令。

当前终端尚不能找到新命令时，安装器会打印基于实际安装目录的可复制 PATH 设置命令：POSIX shell 使用 `export PATH=...:"$PATH"`，fish 使用 `set -gx PATH ... $PATH`。用户复制执行后即可运行 `astra login`，无需切换工作目录；也可打开新终端或使用绝对路径。未知 shell 不提供可能错误的 PATH 语法。

正式启动脚本仅指定公开 Astra 地址及配套 helper/PATH，不改 HOME、不创建认证文件、不隔离用户工作目录、不修改既有业务配置。UC 登录仍由 Astra 写入默认 `~/.moi/auth.json`（或用户明确设置的 `MOI_AUTH_DIR`）。本地预览继续使用单独的隔离目录。服务 origin 是安装绑定，日常升级不偷偷切换服务或账号；迁移到另一服务 origin 应使用独立安装并显式登录。

按已确认的首期范围，保留 HTTPS、SHA256 完整性校验、发布仓库／分支限制及历史版本禁止覆盖，不做发布签名，不设人工审批。SHA256 能发现下载内容与清单不一致，但不能防御拥有存储写权限的攻击者同时替换 binary 和清单；发布来源信任依赖 HTTPS 入口和发布权限控制。

## 发版步骤

workflow 必须先进入默认分支，才会出现标准的手动触发入口。合并代码本身不会启动发布，仍须手动运行。

1. 选择 `operation=build`。首次推荐选 `astra_mode=source`，填写已经推送到 `matrixorigin/Astra` 的分支／tag／commit（例如 `main`）；无需先创建 GitHub Release。填写与该源码 `Cargo.toml` 一致的 `astra_version`，不带 `v`。本机未提交／未推送代码不在构建范围内。若选 `release`，指定已发布版本，忽略 `astra_ref`；缺少 release、平台 asset 或 GitHub SHA256 时直接失败。[GitHub Release asset digest](https://docs.github.com/en/rest/releases/assets)
2. 选择 MOI `dev` 运行 workflow。MOI 固定到本次 `github.sha`；source 模式先解析一次 Astra SHA 并核验版本和上游构建步骤，所有所选平台的构建 job 都 checkout 该 SHA，发布清单和来源对象记录此 SHA。
3. 首次可保持 `dry_run=true` 验证构建和打包，不使用 OSS 凭据、不写 OSS；reuse 模式需要匿名读取已发布 OSS 产物。所选平台产物通过 GitHub Artifact 传递，保留 7 天；这不是公开下载入口。
4. 准备实际发布 QA 时，设 `dry_run=false`，`bundle_version` 使用 `2026.09.22.1-qa` 这种以唯一 `-qa` 结尾的版本，`astra_url` 填 QA allowlist 中的地址。只想存储候选产物，就保持 `publish_channel=false`；想让 QA 客户端看到更新，就设为 `true`。构建验证成功后自动上传不可变产物。`publish_channel` 只写 `channels/qa.json`。
5. 上传摘要会输出环境、Astra 地址、bundle 版本、manifest SHA256 和 channel。保存 bundle 版本和 manifest SHA256。

QA 验收通过后晋级 Prod：重新手动运行同一 workflow，选 `operation=promote-environment`，`bundle_version` 填 QA 版本（例如 `2026.09.22.1-qa`），`manifest_sha256` 填 QA 上传摘要，`astra_url` 填 Prod allowlist 地址。Prod 版本自动去掉末尾 `-qa`。该操作不编译、不上传组件，只生成 Prod manifest／installer 并写 `channels/stable.json`。当前 stable 已存在时，Prod 地址必须与现有 stable 完全一致，组件版本不能降低，平台集合不能缩小。首次 Prod 必须包含四个平台。

已是 Prod 语义的历史 bundle 续期：选 `operation=promote`，填写该 Prod bundle 版本和原 manifest SHA256。它不重组 manifest。QA manifest、带 `-qa` 的 bundle，以及无法按 Prod allowlist 判定环境的历史 bundle 会被拒绝。

QA channel 到期后续期：选 `operation=publish-channel`，填写原来的 QA bundle 版本和原 manifest SHA256。它不重新编译，也不改 manifest。不要用同一次 `operation=build` 重发同一个 bundle id：manifest 记录已固定的源码与组件来源，源码或组合变化后字节会变，不可变对象会冲突。更换 Prod Astra 地址没有发布入口；已安装的 Prod 客户端必须用新的安装目录重装，不能靠修改 `compatibility_reference`。

表单的条件必填以字段描述中的 `Required for ...` 为准：不能把仅 build 使用的字段标成全局必填，否则 promote 也会被迫填写无关参数。预检按所选模式一次列出全部缺失或格式错误，不启动构建、不读取发布凭据。

| 模式 | 必填字段（除带默认值的选择项） |
| --- | --- |
| build / source | QA `bundle_version`（以 `-qa` 结尾）、`moi_version`、`astra_version`、`astra_ref`（默认 main）、`astra_url`、`compatibility_reference`（默认 moi-native-v1） |
| build / release | QA `bundle_version`、`moi_version`、`astra_version`、`astra_url`、`compatibility_reference` |
| build / reuse | QA `bundle_version`、`moi_version`、`astra_ref`（改填已有 OSS bundle ID）、`manifest_sha256`、`astra_url`、`compatibility_reference` |
| promote-environment | QA `bundle_version`、原 QA Summary 中的 `manifest_sha256`、Prod `astra_url` |
| publish-channel | 已上传的 QA `bundle_version`、原 QA Summary 中的 `manifest_sha256` |
| promote | 已上传的 Prod `bundle_version`、原上传 Summary 中的 `manifest_sha256` |

所有 build 模式还必须预先设置仓库 **Variable** `MOI_RELEASE_ASTRA_URL_QA`，`promote-environment` 和 `promote` 需要 `MOI_RELEASE_ASTRA_URL_PROD`，包括 dry-run。它们是 Astra 登录／聊天 HTTPS origin，不是镜像名、OSS 下载地址或 Secret。`astra_url` 必须与对应 Variable 完全一致。实际写 OSS（`dry_run=false`）还需配置前述两个 AK/SK Secrets；只能从 MOI `dev` 运行。

| Input | 说明 |
| --- | --- |
| `operation` | `build` 构建并上传 QA；`publish-channel` 续期已上传的 QA bundle；`promote-environment` 把 QA bundle 零编译晋级为 Prod；`promote` 续期已是 Prod 的 bundle |
| `bundle_version` | QA 为 `2026.09.22.1-qa`；Prod 为去掉 `-qa` 后的 `2026.09.22.1`。同名不同清单拒绝覆盖 |
| `moi_version` | moi-cli 的版本，不带 `v`；通过 `-X main.version=...` 写入 binary。内容变化必须更换版本 |
| `moi_source_ref` | `moi_mode=source` 的上游 Matrixflow 分支/tag/commit，默认 `dev`。构建前固定完整 SHA，清单记录为 `moi_source_sha`；reuse 和晋升忽略此参数 |
| `astra_mode` | 默认 `source`，直接编译；`release` 使用已有 GitHub Release；`reuse` 使用已上传 OSS bundle 的 Astra／Edge |
| `astra_ref` | **matrixorigin/Astra 仓库**的分支、tag 或完整 commit，source 模式默认 `main`；reuse 必须改填已有 OSS bundle 版本；release/promote 忽略。MOI 源码由独立的 `moi_source_ref` 决定 |
| `astra_version` | 不带 `v`；source 匹配源码版本，release 匹配已发布版本；reuse 从原 manifest 读取，不用填写 |
| `platforms` | build 默认 `all`（四平台）；`darwin-arm64` 或 `linux-arm64` 只用于 QA。写入 stable 时不能少于当前 stable 的平台；首次 Prod 必须是四个平台 |
| `astra_url` | build 必须等于 `MOI_RELEASE_ASTRA_URL_QA`；`promote-environment` 必须等于 `MOI_RELEASE_ASTRA_URL_PROD`，并且在 stable 已存在时与当前 Prod 地址相同 |
| `manifest_sha256` | publish-channel、promote、promote-environment 或 reuse 必填：复制原上传 Actions Summary 的 **Manifest SHA256**（64 位小写十六进制）。source/release 构建留空 |
| `compatibility_reference` | build 必填，默认 `moi-native-v1`。首次晋级 Prod 时沿用 QA 标识；stable 已存在时，晋级生成的 Prod manifest 和安装器沿用当前 stable 的标识，不改变 QA 清单或组件二进制。此标识不证明新旧版本功能兼容，晋级前仍需验收旧 Prod 客户端升级。旧 `promote` 也不能更改 stable 的标识或 Astra 地址 |
| `dry_run` | 默认 true；只验证，不读取发布 Secrets、不上传；publish-channel、promote、reuse、promote-environment 会匿名读取 OSS。其中 `promote-environment` 会读取当前公开的 stable channel，按实际 Prod 兼容标识生成摘要并校验；正式发布仍会重新读取渠道，以当时的状态为准。dry-run 的 Summary 会标明没有写入 |
| `publish_channel` | build 默认 false；为 true 且非 dry-run 时，只把 `channels/qa.json` 指向本次 QA bundle |

只修复 moi-cli：选择 `operation=build`、`astra_mode=reuse`，提高 `moi_version` 与新 `bundle_version`，`astra_ref` 填原 bundle 版本，`manifest_sha256` 填其上传摘要。原 Astra／Edge 归档逐字节保留，只重新构建 moi-cli，并在所选平台验证整个组合；清单仍记录原始 source/release 来源，不把 reuse 当成新来源。不依赖 Rust／OpenSSL 的字节可复现性。再次使用 source 重编同一版本仍可能因不可复现而触发不可变冲突，因此不是只更新 moi-cli 的复用路径。

例如只发布 `5.2.2-dev` 的 moi-cli：在 Actions 中仍选择 **Use workflow from: dev**，设置 `moi_mode=source`、`moi_source_ref=5.2.2-dev` 和 `astra_mode=reuse`，并按上文填写旧 Astra bundle 与摘要。默认 `moi_source_ref=dev`。源码解析失败或 checkout 与已固定 SHA 不符时终止，不回退到发布器源码；编译前和打包前均检查源码提交及 tracked 文件状态。构建所需的 Go 版本、CLI 测试及内嵌 Skill 来自所选提交；解析、打包、校验、上传和晋升脚本始终来自运行流水线的提交。源码构建 job 不接收 OSS 发布凭据。`moi_mode=reuse` 不 checkout 或解析源码分支，继续原样保留旧 moi-cli 归档及其来源。QA 验收后晋升 Prod 的方式保持不变。

### 单人 ARM 测试

在 Apple Silicon Mac 上测试时选择 `platforms=darwin-arm64`，Linux ARM 选择 `linux-arm64`。只启动一个 native job，但仍执行全部适用的构建、TLS、版本、Skill 和归档检查；不新增渠道、权限或不同的安装协议。默认 `all` 不变。

单平台 bundle 可以上传到 QA，但不能作为首次 Prod，也不能把已有 stable 的平台集合缩小。恢复或增加平台必须发布覆盖当前 stable 全部平台的新 QA bundle，再执行 `promote-environment`。

QA 的 `bundles/<id>/install.sh` 读取 `channels/qa.json`，Prod installer 读取 `channels/stable.json`。上传成功不代表该 bundle 已可安装。先推进对应 channel，再执行上传 Summary 中的安装链接。仅运行 `astra update` 不会自动发现未推进 channel 的 bundle。公开名称 `https://get.matrixorigin.cn/astra` 与这份脚本的关系见下文。

## 公开安装入口

OSS 不在 bucket 根目录写 `latest` 或 `install.sh`。Prod 晋升和已发布 Prod bundle 的续期会把对应的、已验证的版本化脚本复制到固定地址 `channels/stable/install.sh`，设置 `Cache-Control: no-store`，匿名读回验证后才更新 `channels/stable.json`。QA 发布不会改这个地址。已安装客户端的当前版本由安装时写入的 channel 决定：Prod 是 `channels/stable.json`，QA 是 `channels/qa.json`。这些客户端执行 `astra update` 即可，不经过官网跳转。

新安装的公开名称是 `https://get.matrixorigin.cn/astra`。它不属于本仓库：matrixflow 发布 OSS 对象，但不持有这个域名，也不改跳转。跳转由 `deploy_workflow` 仓库的国内官网 Nginx 维护，配置在 `deploy_website_zh/conf/nginx.conf`。CLB 在 HTTPS:443 终止 TLS，再把 `get.matrixorigin.cn` 的 `/` 以 HTTP 转到该 Nginx 的 80 端口。`/astra` 返回 302，响应带 `Cache-Control: no-store`。发布器上线后，可以用已有 Prod bundle 的 `operation=promote` 填原版本及 manifest SHA256，零编译补齐固定脚本；也可以等待下一次 `promote-environment`。确认 `https://astra-suite.oss-cn-hangzhou.aliyuncs.com/channels/stable/install.sh` 返回 200，且摘要、大小与当前 Prod manifest 的 `installer` 字段一致后，才能把该 302 改指固定地址；在此之前继续使用已晋升 Prod 的版本化脚本。`/health` 返回 200，供 CLB 健康检查；其他路径返回 404。`deploy_website_zh_v2` 只部署新官网前端，不会更新这份配置。QA 的 `install.sh` 读取 `channels/qa.json`，不能作为这个公开地址。

首次切换到固定地址时，改 `deploy_workflow` 的上述 Nginx 配置并发布到官网服务器，然后执行 `docker exec nginx nginx -t && docker exec nginx nginx -s reload`。之后的 Prod 晋升不再需要按版本修改官网跳转。本仓库的 release workflow 不会代做这一步。

该 `install.sh` 里的 `moi-cli` 只是引导程序：下载到临时目录，校验摘要和大小后执行 `moi-cli update install`，脚本退出时删除。命令行写死 channel、`compatibility_reference` 和 `astra_url`。最终安装的 `astra`、`astra-edge`、`moi-cli` 都来自当时 channel 指向的清单，版本以清单为准，不是引导程序的版本。

发布器会拒绝把不同的 `compatibility_reference` 或 Astra 地址推进已有 stable；新用户安装时，固定脚本也随晋升更新到本次已验收的引导程序。最终装上的 bundle 仍取决于 `channels/stable.json`。固定脚本和 channel 是两个 OSS 对象，无法原子更新：若固定脚本写入或匿名校验失败，channel 保持原值；若随后更新 channel 失败，固定脚本可能已经更新，但 channel 仍指向旧版。检查两个公开对象并用同一发布参数重试（序号尚未推进时可复用），不要在不一致时切换官网跳转或宣布晋升成功。

单平台不改变不可变对象规则：相同版本仍必须对应相同源码和产物；这不是覆盖旧 `0.2.7` 对象的开关。也不承诺四倍加速，省下的是其他平台的 job 和等待时间，本机平台的 Rust 编译仍需完成。

source 不会自动改写 Astra 版本文件。如果新源码仍使用已经发布过的版本号，先在 Astra 使用其已有版本准备流程更新源码版本，再选对应提交；只改 workflow 输入会因版本不一致而失败。同一个 Astra 版本不能先用 source 后换 release 来源，或反过来；需要改版本。这个约束避免“同版本不同来源／产物”，不是要求维护者手工比较文件摘要。

构建适配器解析所选 commit 的 Astra `release-binaries.yml`，执行其客户端 target 安装、musl 依赖安装、Cargo 构建、joint-build Edge TLS probe、binary 执行及 Darwin session execution lease 测试步骤，而非维护另一套硬编码 Cargo 参数。Linux apt 源处理也使用上游脚本（含 ubuntu.sources／sources.list 分支）。resolve 阶段即检查每个所选平台的步骤合同，实际构建时再使用相同解析器读取固定 checkout。上游新增未识别的运行步骤、缺失门禁（包括 TLS）或不支持的表达式时直接失败，要求显式更新适配器，不静默跳过；缺少 TLS 步骤的旧源码不能通过当前 source 模式，已有 GitHub Release 仍可走 release 模式。仍使用源码固定的 Rust toolchain；不执行上游打包和完整发布控制器。缓存依赖下载和按平台／toolchain／Cargo.lock 分组的编译目录。冷缓存 Rust 编译仍可能较慢；180 分钟是 job 超时上限，不承诺具体构建时长；产品兼容验收仍需完成。

GitHub API 请求仅向 `api.github.com` 携带短时 `github.token`，禁止携带它跟随重定向。commit 查询使用 SHA 文本响应，不下载大型 patch JSON。下载 Release asset 可以重定向，但不携带 token；凭据仅在 resolve／fetch step 注入，执行下载 binary 的 step 不接收该 token。不需要新增 GitHub Secret。

## 清单与目录合同

```text
astra-suite/
├── components/
│   ├── astra/<version>/<platform>.tar.gz
│   ├── astra/<version>/source.json
│   ├── astra-edge/<version>/<platform>.tar.gz
│   └── moi-cli/<version>/<platform>.tar.gz
├── bundles/<bundle-version>/manifest.json
├── channels/qa.json
├── channels/stable.json
└── channels/stable/install.sh
```

`platform` 精确为 `darwin-amd64`、`darwin-arm64`、`linux-amd64`、`linux-arm64`。每个组件归档只含对应 binary；Astra／Edge 额外包含其 LICENSE。归档固定 gzip 时间戳、tar metadata 和成员顺序，相同组件内容可跨配套版本复用。macOS 上的 `astra` 与 `astra-edge` 在打包前用 `otool -L` 检查：加载命令里出现 `libssl` 或 `libcrypto` 即拒绝上传。可移植构建必须静态链接 OpenSSL，不能依赖构建机 Homebrew 的绝对路径；非 Mach-O 的测试替身不在此列。

两个文件均直接保存普通 JSON 对象，不使用签名信封或独立签名文件。发布端按 `json.dumps(sort_keys=True, separators=(",", ":"), ensure_ascii=True)` 序列化并追加一个 LF，以保证相同内容产生相同字节。`manifest_sha256` 对下载的 manifest 原始字节计算，客户端应先校验原始字节再解析 JSON，不重新序列化后比较摘要。

配套 manifest 的 schema 为 `moi-client-bundle-v1`。新清单包含 `release_environment`（`qa` 或 `prod`）、所选平台的三组件版本、对象 key、URL、大小、摘要、MOI commit、兼容验收标识及 Skill 归属。历史清单可以没有 `release_environment`；客户端忽略未知字段，因此旧清单仍可读取。允许集合精确为四个平台、仅 `darwin-arm64` 或仅 `linux-arm64`；空集合、其他子集及未知平台均拒绝。QA 可以使用单平台。写入 `channels/stable.json` 时，首次 Prod 必须是四个平台，之后不能少于当前 stable 已有平台。每个平台的 `astra_source.mode` 为 `source` 或 `release`：source 记录 repository／commit_sha；release 记录 repository／tag／release_id／asset_id／原始包摘要。所有已选平台须具有同一 commit 或 release 身份。组件版本只接受 `x.y.z`。晋级比较统一组件版本，单平台和四平台之间切换仍拒绝降级。

`components/astra/<version>/source.json` 是 astra／astra-edge 共用的不可变来源记录：source 保存 mode／repository／commit_sha；release 保存 mode／repository／tag／release_id。它不改变 binary 归档成员或客户端更新规则；不同来源即使碰巧生成相同 binary，也不能复用同一个版本。

渠道 schema 为 `moi-client-channel-v1`，包含 `bundle_version`、`manifest_url`、`manifest_sha256`、`sequence`、`issued_at`、`expires_at`。先上传并回读全部产物，再更新并匿名校验固定 Prod 脚本，最后用一次 PutObject 更新 stable 渠道。渠道和固定脚本设置 `Cache-Control: no-store`，版本化产物设置长期 immutable 缓存。

渠道 sequence 为 `GITHUB_RUN_NUMBER * 1000 + GITHUB_RUN_ATTEMPT`；QA 与 stable 各自比较。同一 workflow 的后续新 run 可更新对应渠道，旧 run 的重试不能覆盖更新的 run。不要删除／重建 workflow 导致序列重置后强行覆盖。写入任一 channel 前都会读取该 channel 当前引用的 manifest，只和这一份比组件版本，拒绝降级。QA 不和 stable 比，也不套用 Prod 的四平台、平台不能缩小和 Astra 地址不变规则。写入 stable 时额外拒绝更换 Astra 地址或缩小平台集合。提高 sequence 不能绕过这些检查。有效期为推进后 30 天。stable 到期前用 `operation=promote` 对同一 Prod bundle 和摘要续期。QA channel 到期前用 `operation=publish-channel` 对同一 QA bundle 和摘要续期，只更新 `channels/qa.json` 的序号和有效期。序号与有效期用于防止误操作和陈旧元数据，不提供签名级的抗篡改保证。

遇到上传中断，允许已经写入的不可变对象保留；重试验证并复用，不能删除目录后重发。正式渠道不会在组件或清单失败时推进。若 channel PutObject 已成功但回读失败，作业会报告失败，此时应先检查实际渠道，不能假定仍是旧版本。

## 测试与尚未执行的验证

本地离线验证（仓库根目录）：

```bash
release_test_dir=$(mktemp -d /tmp/moi-release-tests.XXXXXX)
python3 -m venv "$release_test_dir/venv"
"$release_test_dir/venv/bin/python" -m pip install --require-hashes -r sdk/cli/release/publish-requirements.txt
"$release_test_dir/venv/bin/python" sdk/cli/release/release_pipeline_test.py
actionlint .github/workflows/release-moi-cli.yml
make -C sdk cli-release-test cli-skill-check
```

发布依赖（OSS SDK、YAML 适配器及其依赖）固定版本并锁定哈希；锁文件由 `uv pip compile --generate-hashes --python-version 3.12 --universal` 生成，不能手改。测试使用临时文件、内存 OSS，不访问实际 bucket。覆盖四平台集合、归档与摘要安全、同版本冲突、部分上传、读回失败、过期作业拒绝、Secrets 隔离、源码身份、上游步骤漂移、跨 MOI commit 的独立晋级／续期、保留原 manifest 字节、缺失对象拒绝晋级、复用不可重复打包的源归档、流式传输计数和只读重试。

PR 的 SDK 门禁通过 `test-static -> cli-publish-test` 运行此套件，并安装 hash-locked OSS SDK，不跳过 SDK 用例。仅修改 `.github/workflows/release-moi-cli.yml` 也会触发 SDK 门禁。直接运行 `make -C sdk cli-publish-test` 使用隔离的 `sdk/.runtime/release-venv`，不依赖发布时才发现测试失败。

发布 workflow 的凭据隔离、固定源码与晋级依赖断言由 `workflow-check` 中的 `.github/scripts/gates_ci_tool_setup_test.py` 执行；发布脚本或 `sdk/Makefile` 变更同时选择 SDK 与 workflow 两个门禁。SDK 套件中的表单及平台矩阵断言读取 `.github/workflows/release-moi-cli.yml`；现有 Product SDK gate 的 sparse checkout 已包含 `.github/workflows`，不依赖 Git 私有远端补取。

该测试目标从 `https://pypi.org/simple` 安装锁定依赖，避免自托管 CI 默认镜像尚未同步锁定版本而失败；仅对本次安装显式指定来源，不修改用户或 runner 的全局 pip 配置，保留版本和哈希校验。无法访问该来源时明确失败，不降级依赖或跳过测试。

真实四平台 GitHub-hosted runner 的可用性／额度、指定 Astra 源码编译或 Release 的适配性、RAM 权限、Bucket 实际读写与下载均须执行 workflow 验证。构建可先 dry-run，不需要 OSS Secrets；实际上传再配置凭据。本地离线通过不替代这些外部验收。

托管 astra-edge 后续通过固定 digest 的 Runner 镜像更新；本 workflow 不创建镜像、不操作 Kubernetes、不修改 QA GitOps。
