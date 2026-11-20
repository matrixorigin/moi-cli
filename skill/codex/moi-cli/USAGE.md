# MOI CLI Skill 5.2.3 使用指南（Codex / Mac）

本 Skill 以当前验证充分的 `moi-cli-dev` 为内容基线，但仓库、安装和调用名称统一为 `moi-cli`，版本统一为 `5.2.3`。稳定旧版不再作为内容来源。

## 安装

仓库公开或获得私有仓库读取权限后，把下面指令发给 Codex：

```text
请使用 skill-installer，以 --method git 稀疏检出方式安装这个 Skill：
https://github.com/<owner>/<repo>/tree/<ref>/sdk/cli/skill/codex/moi-cli
```

安装目录是 `~/.codex/skills/moi-cli`，下一轮或新对话即可使用 `$moi-cli`。GitHub 只提供 Skill，不提供 CLI、Astra、登录态、PAT 或 workspace。

## 当前执行程序

本版沿用最新 dev CLI 的执行方式：`MOI_DEV_CLI_BIN` 指向已确认的二进制，未设置时使用 `$HOME/.local/bin/moi-cli-dev`。不要回退到旧稳定 CLI 冒充本版能力。

安装后先发送：

```text
$moi-cli 先检查 Skill 是否为 5.2.3，验证三套任务模板和 CLI 合同，并做 PAT-only 本机预检查。只做只读检查，不要登录或执行线上写入。
```

需要联网时，再明确 QA 或生产的 endpoint、workspace 和认证通道。不要仅凭浏览器登录状态判断环境。

## 更新

调用名始终是 `$moi-cli`。官方安装器不会覆盖已有目录；更新前让 Codex 备份旧安装，再从同一 GitHub 路径重新安装。版本号不需要写进调用名。

## 内容与边界

`assets/task-blueprints/` 保存三套复杂任务模板，`scripts/` 保存预检查、合同和恢复工具，`references/` 保存最新领域 Recipe、能力目录和 5.2.3 报告。报告中的业务结果来自 0.4.12 内容基线，本轮私有 GitHub 安装验证单独记录，不能混为新增产品功能通过。

不上传或复制登录文件、PAT、Cookie、固定工作区、本机私有路径和原始测试证据目录。超出正式 CLI 合同时明确说明缺口，不使用 raw HTTP、core-cli 或数据库直连绕过。
