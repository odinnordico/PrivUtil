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
cd web && npm test -- --grep "pattern"

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
Browser (React + nice-grpc-web) → HTTP server (gRPC-Web wrapper) → gRPC handlers → Go business logic
```

**Key layers:**
- `cmd/privutil/main.go` — CLI flags, initializes gRPC server and HTTP server
- `internal/server/server.go` — HTTP server that wraps gRPC-Web, serves embedded React assets with SPA fallback
- `internal/api/grpc_server.go` — registers all gRPC service implementations
- `internal/api/*_handlers.go` — domain-grouped handler files: `data_handlers.go`, `text_handlers.go`, `encoding_handlers.go`, `gen_handlers.go`, `security_handlers.go`, `system_handlers.go`
- `proto/privutil.proto` — single proto file defining all ~29 RPC methods; generated Go code is committed
- `web/src/` — React frontend; `web/src/proto/` contains generated TypeScript proto bindings (also committed)

## Adding a New Tool

1. Add the RPC method and request/response messages to `proto/privutil.proto`
2. Run `make proto` to regenerate `proto/privutil.pb.go`, `proto/privutil_grpc.pb.go`, and `web/src/proto/`
3. Implement the handler in the appropriate `internal/api/*_handlers.go` file and register it in `grpc_server.go`
4. Add the React component under `web/src/components/` and wire it into the router in `web/src/App.tsx`

See `wiki/Adding-New-Features.md` for a detailed walkthrough.

## Configuration

The server accepts flags and environment variables:

| Flag | Env | Default |
|------|-----|---------|
| `-port` | `PORT` | `8090` |
| `-host` | `HOST` | `localhost` |
| `-log-level` | `LOG_LEVEL` | `info` |

## Commit Convention

This project uses [Conventional Commits](https://www.conventionalcommits.org/): `feat:`, `fix:`, `refactor:`, `chore:`, `docs:`, `test:`.
