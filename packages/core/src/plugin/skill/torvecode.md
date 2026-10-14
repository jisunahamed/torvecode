# TorveCode

Use this guide for TorveCode configuration, commands, skills, models, tools, permissions, MCP servers, themes, keybinds, troubleshooting, and local development.

## Product identity

- The command is `torve` and the product name is TorveCode.
- Torve AI supplies the authenticated model catalog.
- Prefer the account's current Torve AI model and quota. Do not expose unrelated provider setup in user-facing guidance.

## Common commands

- `torve` starts the interactive coding interface.
- `torve run "task"` runs a non-interactive task.
- `torve models` lists available coding models.
- `torve auth login` connects an account or API key.
- `torve auth status` shows connected credentials.
- `torve auth logout` removes a saved credential.
- `torve update` installs the latest stable release.

## Interface

Type `/` to open commands. Skills, sessions, models, themes, MCP connections, diffs, and settings are available from the command palette. Press `Esc` to close the current dialog or return to the previous screen.

## Configuration

Keep project instructions and skills inside the project. Store personal configuration in the operating system's user configuration directory. Never print, log, or save credentials in project files or session history.

## Troubleshooting

Use `torve --help` to verify command availability and `torve models` to verify the authenticated model catalog. If a Windows npm update reports `EBUSY`, close every running TorveCode process and run `npm install --global @torveai/cli@latest` again.
