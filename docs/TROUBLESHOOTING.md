# Troubleshooting

- `torve` not found: reopen the shell and verify npm's global bin directory is on `PATH`.
- Browser did not open: copy the printed verification URL and code into any signed-in browser.
- Secure storage unavailable: install/unlock Secret Service on Linux, or use `TORVE_API_KEY` for the current process.
- Session expired: run `torve auth login` again. Refresh-token reuse deliberately revokes the device.
- No coding models: ask the Torve operator to grant and mark a Chat Completions or Messages model as CLI tool capable.
- Quota/balance errors: view package and usage in the Torve AI dashboard.
- Proxy/TLS errors: configure the operating system's trusted proxy variables; `torve doctor` reports the failing endpoint without printing credentials.
