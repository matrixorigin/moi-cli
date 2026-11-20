# Codex MOI CLI Skill

This directory is the standalone Codex owner. It is intentionally separate from
the prepared-MOI owner in `../moi-cli`.

Both use the latest verified `moi-cli-dev` content as their baseline, while the
repository-facing Skill name and release are unified as `moi-cli` / `5.2.3`.
The Codex edition preserves the proven `MOI_DEV_CLI_BIN` execution model and
`agents/openai.yaml`; it does not install or replace Astra, moi-cli, credentials
or workspaces.

## Install from GitHub

```text
请使用 skill-installer，以 --method git 稀疏检出方式安装这个 Skill：
https://github.com/<owner>/<repo>/tree/<ref>/sdk/cli/skill/codex/moi-cli
```

The path basename and frontmatter name are both `moi-cli`, so the official
installer creates `~/.codex/skills/moi-cli`. Use `$moi-cli` on the next turn or
in a new conversation. WorkBuddy is outside this change.

## Files

- `moi-cli/SKILL.md`: latest execution, capability, recovery and domain routing.
- `moi-cli/USAGE.md`: GitHub installation, dev CLI prerequisite and updates.
- `moi-cli/agents/openai.yaml`: Codex discovery metadata.
- `moi-cli/assets/task-blueprints/`: three optional complex-task plans.
- `moi-cli/scripts/`: preflight, capability, evidence and recovery helpers.
- `moi-cli/references/`: latest domain recipes, capability catalog and 5.2.3
  reports. Business results are identified as the 0.4.12 content baseline;
  private GitHub installation verification is recorded separately.
- `tests/`: offline identity, contract, preflight and report checks.

No ZIP, WorkBuddy adapter, raw evidence directory, credential, fixed workspace,
machine-local path or second copy of the prepared-MOI owner is introduced.

## Validate

Run from the repository root:

```bash
python3 -m unittest discover -s sdk/cli/skill/codex/tests -v
```

After pushing to a private ref, install that exact ref with the official
installer into a clean temporary `CODEX_HOME` and run the installed preflight.
Public installation is verified only after the private-repository check passes.
