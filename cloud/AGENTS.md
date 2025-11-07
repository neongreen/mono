# AGENTS.md — cloud

Guidelines for agents working in `cloud/` (OpenTofu + Spacelift) with `fnox`.

## Secrets & Execution (fnox)

- Always run commands that require secrets via `fnox exec -- …`.
- Do not export secrets into your shell.
- Execute tools with secrets injected:

```bash
# Spacelift CLI
fnox exec -- spacectl whoami
fnox exec -- spacectl stack list

# Local tofu (validation only; no remote backend contact)
cd cloud/terraform/01-hcloud
fnox exec -- tofu init -backend=false -input=false
fnox exec -- tofu validate
```
