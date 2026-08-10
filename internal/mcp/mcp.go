// Package mcp exposes PrivUtil's offline utility tools to LLM agents over the
// Model Context Protocol. The server is served in-process by the existing HTTP
// server (no separate process or instance) via a Streamable HTTP handler.
package mcp

import (
	"context"
	"log"
	"net/http"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"golang.org/x/time/rate"

	"github.com/odinnordico/privutil/internal/api"
)

const (
	// maxRequestBytes bounds a single MCP HTTP request body — a guardrail against
	// oversized payloads. Individual tools also enforce their own input caps.
	maxRequestBytes = 4 << 20 // 4 MiB

	// Rate-limit guardrail for the endpoint (shared across local agents).
	rateLimitPerSecond = 50
	rateLimitBurst     = 100
)

// NewServer builds an MCP server exposing a curated set of PrivUtil tools backed
// by the given API server. Tools are pure, offline computations with no side
// effects.
func NewServer(s *api.Server, version string) *mcp.Server {
	srv := mcp.NewServer(&mcp.Implementation{
		Name:    "privutil",
		Title:   "PrivUtil",
		Version: version,
	}, &mcp.ServerOptions{
		Instructions: "PrivUtil exposes offline developer utilities (hashing, encoding, UUIDs, JWT decoding, text diff, and more). Every tool runs locally with no network access and no side effects.",
	})
	registerTools(srv, s)
	return srv
}

// Handler returns the in-process Streamable HTTP handler for the MCP server.
// It is stateless (one shared server, no per-session state — no new instances)
// and bounds the request body. Mount it (e.g. at /mcp) behind the HTTP server's
// Host-allowlist and security middleware.
func Handler(s *api.Server, version string) http.Handler {
	srv := NewServer(s, version)
	h := mcp.NewStreamableHTTPHandler(
		func(*http.Request) *mcp.Server { return srv },
		&mcp.StreamableHTTPOptions{Stateless: true, JSONResponse: true},
	)
	return rateLimit(http.MaxBytesHandler(h, maxRequestBytes))
}

// rateLimit caps the request rate to the MCP endpoint.
func rateLimit(next http.Handler) http.Handler {
	limiter := rate.NewLimiter(rateLimitPerSecond, rateLimitBurst)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !limiter.Allow() {
			http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// errResult reports an in-band tool error (e.g. invalid input) to the client as
// an error result rather than a transport failure.
func errResult(msg string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		IsError: true,
		Content: []mcp.Content{&mcp.TextContent{Text: msg}},
	}
}

// addTool registers a tool whose handler is wrapped with panic recovery. The MCP
// path does not go through the Connect RecoveryInterceptor, and the SDK runs each
// tool in a goroutine with no recover of its own, so an unhandled panic in any
// handler would otherwise crash the whole process. A recovered panic becomes an
// error result instead.
func addTool[In, Out any](srv *mcp.Server, t *mcp.Tool, h func(context.Context, *mcp.CallToolRequest, In) (*mcp.CallToolResult, Out, error)) {
	mcp.AddTool(srv, t, func(ctx context.Context, req *mcp.CallToolRequest, in In) (res *mcp.CallToolResult, out Out, err error) {
		defer func() {
			if r := recover(); r != nil {
				name := ""
				if req != nil && req.Params != nil {
					name = req.Params.Name
				}
				log.Printf("mcp: recovered from panic in tool %q", name)
				var zero Out
				res, out, err = errResult("internal error"), zero, nil
			}
		}()
		return h(ctx, req, in)
	})
}
