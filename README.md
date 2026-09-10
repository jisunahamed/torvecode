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
torve models                  list available models, tools and image input
torve skills create NAME      create a project SKILL.md template
torve skills add NAME FILE    add an existing local SKILL.md
torve skills list             list project and personal skills
torve documents status        check the optional MarkItDown integration
torve documents convert FILE  convert a document to Markdown
torve run "PROMPT"            non-interactive; tools denied by default
torve run --allow-tools "..." non-interactive with tools enabled
torve doctor                  diagnose installation and connectivity
```

In the interactive UI, type `/` to search commands. Use `/image` or `Ctrl+F` to attach an image file. `Ctrl+V` automatically pastes clipboard text or attaches a copied image.

The status bar shows the latest request's input and output tokens, effective context usage and cost when catalog pricing is available. Torvecode automatically compacts long sessions before the account's TPM budget becomes too small for a useful response; `/compact` remains available for manual control. Reasoning-capable OpenAI-compatible models show a live `Thinking...` state and accept streamed `reasoning_content` or `reasoning` events.

PDF, Word, PowerPoint, Excel and other document inputs use Microsoft's optional [MarkItDown](https://github.com/microsoft/markitdown) converter. Install Python 3.10+ and then install it locally:

```sh
pip install "markitdown[pdf,docx,pptx,xlsx,xls,outlook]"
```

Torvecode reports plan-specific RPM, TPM, concurrency, monthly quota and account-balance errors directly. It does not retry a request that the current plan cannot admit.

Set `TORVE_API_URL` and `TORVE_WEB_URL` for local/staging development. `NO_COLOR` disables color. Torvecode reads `.torvecode.json` and stores project history in `.torvecode`; it never imports old OpenCode configuration automatically.

Project skills are stored in `.torvecode/skills/NAME/SKILL.md`. Add `--global` to a skills command to use your personal Torvecode skills directory.

Documentation: [user guide](docs/USER_GUIDE.md), [authentication](docs/AUTHENTICATION.md), [development](docs/DEVELOPMENT.md), [release](docs/RELEASE.md), [troubleshooting](docs/TROUBLESHOOTING.md), [maintenance and privacy](docs/MAINTENANCE_PRIVACY.md), and [upstream provenance](UPSTREAM.md).

Torvecode retains the upstream MIT license in [LICENSE](LICENSE).
