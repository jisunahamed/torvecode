# Release runbook

The source repository is `jisunahamed/torvecode`. Before release, confirm npm scope `@torveai` and `torveai/homebrew-tap` access. Run Go tests on Windows/macOS/Linux, API/web tests, typecheck, build, dependency/license scans and staging device-login tests.

Tag a beta such as `v0.1.0-beta.1`. GoReleaser creates six signed/checksummed archives. Copy each binary into its matching npm platform package, publish all platform packages, then publish `@torveai/cli` with the beta tag. Update the Brew formula from the release checksums. Promote the identical artifacts to stable; do not rebuild them.

Rollback by deprecating the npm version, restoring the prior Brew formula, disabling device issuance at the API deployment and revoking affected sessions. Existing customer API keys continue to work.
