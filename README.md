# t/ Torvecode

Torvecode is Torve AI's lightweight coding agent for the terminal. The Go TUI includes streaming chat, reviewed file edits, shell tools, sessions, optional MCP and optional LSP. Models, quota and billing always come from the connected Torve AI account.

## Install and connect

```sh
npm install -g @torveai/cli
torve auth login
torve
```

Use an existing Torve AI API key with `torve auth login --api-key`. For CI or a temporary shell:

```sh
TORVE_API_KEY=sk-trv-... torve run "review this repository"
```

Other release channels:

```sh
brew install torveai/tap/torvecode
go install github.com/jisunahamed/torvecode/cmd/torve@latest
```

## Commands

```text
torve                         interactive session
torve auth login              website device authorization
torve auth login --api-key    validate and securely store an API key
torve auth status             validate the active credential
torve auth logout             revoke/remove a saved session
torve models                  list available models and tool capability
torve run "PROMPT"            non-interactive; tools denied by default
torve run --allow-tools "..." non-interactive with tools enabled
torve doctor                  diagnose installation and connectivity
```

Set `TORVE_API_URL` and `TORVE_WEB_URL` for local/staging development. `NO_COLOR` disables color. Torvecode reads `.torvecode.json` and stores project history in `.torvecode`; it never imports old OpenCode configuration automatically.

Documentation: [user guide](docs/USER_GUIDE.md), [authentication](docs/AUTHENTICATION.md), [development](docs/DEVELOPMENT.md), [release](docs/RELEASE.md), [troubleshooting](docs/TROUBLESHOOTING.md), [maintenance and privacy](docs/MAINTENANCE_PRIVACY.md), and [upstream provenance](UPSTREAM.md).

Torvecode retains the upstream MIT license in [LICENSE](LICENSE).
