# moi-cli

[简体中文](README_zh-CN.md)

`moi-cli` is the MatrixOne Intelligence (MOI) 5.x AI Studio command-line tool for interactive use and automation. It uses the Product Go SDK and applies the server's workspace and role permissions. It covers workspaces, Catalog and files, connectors and transfers, workflows, knowledge bases, SQL, agents, models, compute, task monitoring, and data sharing.

## Requirements and installation

- **[GitHub Release v5.2.3](https://github.com/matrixorigin/moi-cli/releases/tag/v5.2.3).** Confirm the version matches your deployment, then download `moi-cli-<version>-<os>-<arch>.tar.gz` for your system and its `SHA256SUMS`. The OS is `linux` or `darwin` (macOS), and the architecture is `amd64` or `arm64`; for example, Apple silicon Macs use `moi-cli-<version>-darwin-arm64.tar.gz`. Verify the downloaded archive against its SHA-256 entry in `SHA256SUMS`, extract it, and put `moi-cli` on your `PATH`. The filename version omits the release tag’s `v` prefix.
- **Astra-managed installation.** Use the paired Astra / MOI installer supplied for your deployment. For managed installations, check with `astra update check`, then update with `astra update`. For manual installs, replace the binary with a matching build.
- **Build from source.** Use Go 1.25.7 or newer. From the root of this repository, build with the bundled dependencies:

```sh
go build -trimpath -buildvcs=false -ldflags '-X main.version=5.2.3' -o moi-cli ./cmd/moi-cli
./moi-cli -version
```

Use the build command above with the v5.2.3 source. After installation, verify the binary on your `PATH` reports `moi-cli version 5.2.3`:

```sh
moi-cli -version
moi-cli --help
```

## Authentication and quickstart

**Personal access token (PAT).** Create an AI Studio PAT in the MOI console and pass the raw token without a `Bearer` prefix. Set the full endpoint including its path prefix. Use the hosted endpoint shown below or the endpoint supplied by your deployment administrator. Configure the token in a trusted shell, since command arguments can appear in shell history or process listings.

```sh
moi-cli config set-endpoint 'https://api.moi.matrixorigin.cn/v5'
moi-cli config set-api-key 'YOUR_PERSONAL_ACCESS_TOKEN'
moi-cli -auth pat -o -no-interactive workspace list
```

The read-only command returns JSON with the workspaces visible to the account. Choose an ID from the output and save it before querying resources in the workspace:

```sh
moi-cli config set-workspace 'YOUR_WORKSPACE_ID'
moi-cli -auth pat -o -no-interactive workflow list
```

PAT connections use `~/.moi-cli.json` by default. Use `MOI_CLI_CONFIG_FILE` for a separate absolute, clean path. The CLI writes the file with `0600` permissions and stores the token in plaintext. Keep the file private. `moi-cli config show` inspects local configuration and masks the token.

Use `config set-role <workspace-id> <role-id>` to select an assigned role and `config unset-role <workspace-id>` to restore the server's Default Role. Use a canonical positive decimal `int64` for the role ID. An invalid saved mapping blocks product requests. Find the mapping with `config show` and correct it. See [Configuration and authentication](https://docs.matrixorigin.cn/moi/en/5.0/developer/cli/configuration-authentication.html).

**Astra shared login.** With a matching Astra client, sign in to a trusted target environment first. For ordinary local commands, default `-auth auto` uses the Astra endpoint, identity, and workspace, with no saved PAT fallback even before sign-in. Any explicit endpoint must match the signed-in environment. In a managed agent sandbox, the CLI uses the injected task credentials when `MOI_RUNTIME_AUTHORIZATION` is present.

```sh
astra login
moi-cli workspace list
```

## Automation and command discovery

Place global options before the command. `-o` requests JSON output for structured commands and `-no-interactive` disables interactive prompts. Check the exit status (zero on success, nonzero on failure) before consuming output. Use the global `--request-timeout` flag to bound an HTTP request and response-body reads. For asynchronous operations, keep the task or run ID and check the final state. `agent a2a --no-wait` returns at admission. After a timeout or an unknown outcome, look up the original task before retrying. See [Non-interactive mode and automation](https://docs.matrixorigin.cn/moi/en/5.0/developer/cli/json-output-automation.html).

Help, typed input descriptions, and embedded Skill export run offline:

```sh
moi-cli workflow list --help
moi-cli schema connector create-task
moi-cli skill export --dir ./moi-cli-skill
```

`schema <domain> <command> --validate-file <path>` checks the structure of supported JSON inputs offline. The server still validates business inputs, resources, and permissions. `skill export` requires a destination directory that does not exist.

On upload commands supporting `--file-path`, select regular files explicitly and quote paths containing spaces. Use `--file-path` separately from `--json` or `--json-file`. The default total limit is 64 MiB, adjustable with `--max-upload-bytes`.

Agent package authoring commands `agent init`, `agent verify`, `agent build`, and `agent unpack` also run locally. Loading and exporting a package through the service require authentication.

## Compatibility and license

`moi-cli` 5.2.3 targets the AI Studio API (`/v5`) on MOI 5.2.3. For other MOI versions, use the build specified by your deployment documentation. `core-cli` and the legacy Catalog SDKs address the Catalog service.

MOI CLI is licensed under the [Apache License 2.0](LICENSE). Dependencies and separately distributed components retain their own licenses and copyright notices.

## Documentation

- [AI Studio CLI quickstart](https://docs.matrixorigin.cn/moi/en/5.0/developer/quickstart/ai-studio/cli.html)
- [Command reference](https://docs.matrixorigin.cn/moi/en/5.0/developer/cli/command-reference.html)
- [Usage guide (Chinese)](USAGE.zh-CN.md) and [Astra authentication (Chinese)](NATIVE_AUTH.zh-CN.md)
