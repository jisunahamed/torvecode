# TorveCode

TorveCode is Torve AI's coding agent for the terminal. It can inspect a project, edit files, run commands, review diffs, resume sessions, use skills, and connect to optional MCP and LSP tools.

## Install

TorveCode ships as a small npm launcher plus a prebuilt native binary for Windows, macOS, and Linux on x64 and ARM64. Go or Bun is not required.

```powershell
npm install --global @torveai/cli@latest
torve --version
torve
```

If Windows reports `EBUSY` during an update, close every running TorveCode window and run the install command again.

## Update

Update from inside the CLI installation:

```powershell
torve update
```

`torve upgrade` is the equivalent long form. To reinstall the latest stable release directly:

```powershell
npm install --global @torveai/cli@latest
```

## Authentication

```powershell
torve auth login
```

Follow the browser prompt to connect your Torve AI account. The same login flow can store a Torve AI API key when that authentication method is available for the account.

## Common commands

```text
torve                         Start the interactive coding interface
torve run "fix the tests"     Run a non-interactive task
torve models                  List available Torve AI models
torve auth login              Connect your Torve AI account
torve auth status             Show the active authentication method
torve auth logout             Remove the local login
torve update                  Install the latest stable release
torve --help                  Show every command
```

Inside the interface, type `/` to browse commands, sessions, models, themes, skills, MCP connections, diffs, and settings. Press `Esc` to close dialogs or return to the previous screen. Use `Ctrl+C` to cancel active work and press it again to exit.

## Skills

TorveCode includes skill discovery and skill execution. Project skills can be stored with the project, while personal skills can be kept in the user configuration directory. Use `/skills` in the interface to browse available skills.

## Uninstall

```powershell
npm uninstall --global @torveai/cli
```

The source is licensed under the [MIT License](LICENSE). Third-party license and copyright notices remain preserved in the repository and release artifacts.
