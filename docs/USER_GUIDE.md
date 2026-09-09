# User guide

Install with `npm install -g @torveai/cli`, then run `torve auth login`. The CLI opens the Torve AI approval page and prints a code for remote/SSH terminals. Confirm the same code in the browser. `torve auth status` verifies the session.

For an API key, run `torve auth login --api-key`; input is hidden and the key is validated before saving. In CI set `TORVE_API_KEY` for the process. An explicitly set but invalid environment key never falls back to a saved credential.

Run `torve` in a repository. File writes, shell commands and MCP tools request permission in interactive sessions. `torve run` denies tools unless `--allow-tools` is supplied. Use `torve models`, `torve doctor`, and `torve auth logout` for model discovery, diagnostics and disconnection.

Upgrade through the same install channel. Uninstall npm builds with `npm uninstall -g @torveai/cli`; remove project history by deleting that project's `.torvecode` directory after reviewing it.
