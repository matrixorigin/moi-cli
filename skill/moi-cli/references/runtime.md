# Prepared MOI runtime

Use the injected `bin/moi-cli` and the trusted workspace `README.md` and `context.json`. Credentials and workspace are supplied by the task environment. Do not use the standalone Codex preflight or local login files.

List visible workspaces with the global JSON flag before the domain:

```sh
moi-cli -o -no-interactive workspace list
```

For workspace-scoped operations, use the confirmed workspace from the task context. Keep the latest dev execution, evidence, recovery and domain contracts, but resolve executable and authentication through the prepared runtime rather than a personal computer path.
