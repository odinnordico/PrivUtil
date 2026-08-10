# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# Build (must use make — Go binary embeds the React frontend)
make build          # frontend + backend
make build-web      # React only
make build-go       # Go only (requires web/dist to exist)
make run            # build and start server

# Test
make test           # all tests
make test-backend   # go test -v -tags=manual -cover ./...
make test-frontend  # cd web && npm test
make test-coverage  # HTML coverage reports

# Single Go test
go test -v -run TestFunctionName ./internal/api/

# Single frontend test
cd web && npm test -- -t "pattern"

# Lint
make lint           # all linters
make lint-backend   # go vet + go fmt
make lint-frontend  # eslint

# Regenerate protobuf (after editing proto/privutil.proto)
make proto
```

## Architecture

PrivUtil is a full-stack Go + React app served as a single self-contained binary. The frontend is built with Vite and embedded into the Go binary via `//go:embed`.

**Request flow:**
```
Browser (React + nice-grpc-web) → HTTP server (connect-go handler over h2c) → connect handlers → Go business logic
```

**Key layers:**
- `cmd/privutil/main.go` — CLI flags, initializes the connect service handler and HTTP server
- `internal/server/server.go` — HTTP server serving the connect handler over h2c, with a Host-header allowlist, allowlist-scoped CORS, and CSP/security-header middleware; serves embedded React assets with SPA fallback
- `internal/api/server.go` — the `Server` type and `NewServer` (functional options); `internal/api/connect_adapter.go` — adapts `*Server` to the generated connect handler interface (`NewConnectServer`)
- `internal/api/*_handlers.go` — domain-grouped handler files: `data_handlers.go`, `text_handlers.go`, `encoding_handlers.go`, `gen_handlers.go`, `security_handlers.go`, `crypto_handlers.go`, `datetime_handlers.go`, `math_handlers.go`, `media_handlers.go`, `network_handlers.go`, `spell_handlers.go`, `text_tools_handlers.go`, `token_handlers.go`, `webdevops_handlers.go`, `system_handlers.go`
- `internal/mcp/` — exposes ~70 pure, offline tools to local LLM agents over MCP (Model Context Protocol), served in-process as a Streamable HTTP handler mounted at `/mcp` (behind the Host-allowlist/security middleware, plus body/rate limits and per-handler panic recovery via `addTool`). Tools follow one pattern: a typed input struct + a call into `*api.Server`. Wired in `main.go` via `srv.SetMCPHandler(mcp.Handler(apiServer, Version))`
- `proto/privutil.proto` — single proto file defining all 78 RPC methods; generated Go code (`proto/privutil.pb.go`, `proto/protoconnect/privutil.connect.go`) is committed
- `web/src/` — React frontend; `web/src/proto/` contains generated TypeScript proto bindings (also committed)

## Adding a New Tool

1. Add the RPC method and request/response messages to `proto/privutil.proto`
2. Run `make proto` to regenerate `proto/privutil.pb.go`, `proto/protoconnect/privutil.connect.go`, and `web/src/proto/`
3. Implement the handler in the appropriate `internal/api/*_handlers.go` file and add the matching adapter method in `internal/api/connect_adapter.go`
4. Add the React component under `web/src/components/` and wire it into the router in `web/src/App.tsx`

See `wiki/Adding-New-Features.md` for a detailed walkthrough.

## Configuration

The server accepts flags and environment variables:

| Flag | Env | Default |
|------|-----|---------|
| `-port` | `PORT` | `8090` |
| `-host` | `HOST` | `127.0.0.1` (loopback; use `0.0.0.0` for all interfaces) |
| `-allowed-hosts` | `ALLOWED_HOSTS` | (empty) — extra allowed `Host` header values, comma-separated |
| `-custom-dict` | `CUSTOM_DICT` | `<user-config-dir>/privutil/custom-dictionary.txt` |
| `-mcp` | `MCP` | `true` — serve the in-process MCP endpoint at `/mcp` |
| `-log-level` | `LOG_LEVEL` | `info` (`debug` or `info` only) |

## Commit Convention

This project uses [Conventional Commits](https://www.conventionalcommits.org/): `feat:`, `fix:`, `refactor:`, `chore:`, `docs:`, `test:`.
