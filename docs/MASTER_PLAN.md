# Torvecode implementation map

The Go CLI resolves a website session or `TORVE_API_KEY`, requests the authenticated `/cli/models` catalog, registers only Torve models, and sends OpenAI Chat Completions or Anthropic Messages to the Torve gateway. Existing quota, package, billing and request logging remain authoritative.

The web/API workspace adds RFC 8628-style device authorization, short-lived gateway credentials, rotating refresh tokens, connected-device management and Torvecode-specific pages. Releases build `torve` for Windows, macOS and Linux on x64/ARM64; npm installs the matching binary package.

Release acceptance requires cross-platform install/login/smoke tests, gateway regression tests, secret-log scans, a compressed artifact below 50 MB, startup p95 below one second and idle memory below 150 MB on the documented runner.
