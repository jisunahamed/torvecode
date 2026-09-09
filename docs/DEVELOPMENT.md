# Development guide

Requirements: Go 1.24+, Node.js 24+, pnpm 11+, PostgreSQL/Supabase and the services documented by the Torve API workspace.

```sh
pnpm install
pnpm db:generate
pnpm db:migrate:deploy
pnpm dev

go test ./...
go run ./cmd/torve doctor
```

For local integration set `TORVE_API_URL=http://127.0.0.1:3100` and `TORVE_WEB_URL=http://localhost:3000`. Never use production credentials in tests. Enable `cliToolUse` only after a model's tool-call behavior is verified.
