package server

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"mime"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	connectcors "connectrpc.com/cors"
	"github.com/rs/cors"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

//go:embed dist/*
var staticFiles embed.FS

func init() {
	// Go's MIME table has no .webmanifest entry, so http.FileServer would serve
	// site.webmanifest as text/plain and browsers would reject the manifest.
	_ = mime.AddExtensionType(".webmanifest", "application/manifest+json")
}

type Server struct {
	addr         string
	rpcPath      string
	rpcHandler   http.Handler
	allowedHosts map[string]struct{}
	mcpHandler   http.Handler // optional; mounted at /mcp when set
}

// SetMCPHandler mounts an MCP handler at /mcp (in-process). It is served behind
// the same Host-header allowlist and security middleware as everything else,
// plus an Origin guard, and without any permissive CORS.
func (s *Server) SetMCPHandler(h http.Handler) { s.mcpHandler = h }

// mcpGuard rejects cross-origin browser requests to the MCP endpoint. Local/CLI
// agents send no Origin header and pass through; a browser page from another
// origin (whose Origin is not allowlisted) is refused. Combined with the absence
// of CORS headers on /mcp, this blocks cross-origin browser access entirely.
func (s *Server) mcpGuard(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if o := r.Header.Get("Origin"); o != "" && !s.allowedOrigin(o) {
			http.Error(w, "forbidden: origin not allowed", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// New builds an HTTP server that routes connect RPC requests under rpcPath to
// rpcHandler and serves the embedded React SPA for everything else. allowedHosts
// is the set of Host-header hostnames (no port) accepted for both request
// admission and CORS — the DNS-rebinding / cross-origin defense.
func New(addr, rpcPath string, rpcHandler http.Handler, allowedHosts []string) *Server {
	set := make(map[string]struct{}, len(allowedHosts))
	for _, h := range allowedHosts {
		set[strings.ToLower(h)] = struct{}{}
	}
	return &Server{
		addr:         addr,
		rpcPath:      rpcPath,
		rpcHandler:   rpcHandler,
		allowedHosts: set,
	}
}

// hostAllowed reports whether the given Host header (which may include a port)
// has a hostname in the allowlist.
func (s *Server) hostAllowed(hostport string) bool {
	host := hostport
	if h, _, err := net.SplitHostPort(hostport); err == nil {
		host = h
	}
	_, ok := s.allowedHosts[strings.ToLower(host)]
	return ok
}

// allowedOrigin authorizes a CORS Origin whose hostname is in the allowlist
// (any port/scheme) — scopes the RPC surface to localhost-family origins.
func (s *Server) allowedOrigin(origin string) bool {
	u, err := url.Parse(origin)
	if err != nil {
		return false
	}
	_, ok := s.allowedHosts[strings.ToLower(u.Hostname())]
	return ok
}

// checkHost rejects requests whose Host header is not in the allowlist, the
// canonical defense against DNS rebinding (which survives a loopback bind).
func (s *Server) checkHost(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !s.hostAllowed(r.Host) {
			http.Error(w, "forbidden: host not allowed", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) newHandler(distFS fs.FS) http.Handler {
	fileServer := http.FileServer(http.FS(distFS))

	// Scope CORS to the allowlisted hosts (localhost family + any configured
	// host) instead of "*", so a website the user visits cannot read RPC
	// responses cross-origin. The connectcors helper supplies the headers the
	// Connect/gRPC-Web protocols need.
	corsMiddleware := cors.New(cors.Options{
		AllowOriginFunc: s.allowedOrigin,
		AllowedMethods:  connectcors.AllowedMethods(),
		AllowedHeaders:  connectcors.AllowedHeaders(),
		ExposedHeaders:  connectcors.ExposedHeaders(),
	})

	mux := http.NewServeMux()
	mux.Handle(s.rpcPath, corsMiddleware.Handler(s.rpcHandler))

	// In-process MCP endpoint (no CORS: cross-origin browsers are blocked, local
	// agents reach it directly). The outer securityHeaders/checkHost middleware
	// still applies, and mcpGuard adds an Origin check.
	if s.mcpHandler != nil {
		guarded := s.mcpGuard(s.mcpHandler)
		mux.Handle("/mcp", guarded)
		mux.Handle("/mcp/", guarded)
	}

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}

		if _, err := fs.Stat(distFS, path); err != nil {
			r.URL.Path = "/"
			fileServer.ServeHTTP(w, r)
			return
		}

		fileServer.ServeHTTP(w, r)
	})

	// Serve cleartext HTTP/2 (h2c) so native gRPC clients work without TLS; the
	// browser uses gRPC-Web over HTTP/1.1, which the same handler also serves.
	// checkHost runs first (rebinding defense), then security headers, then mux.
	return h2c.NewHandler(securityHeaders(s.checkHost(mux)), &http2.Server{})
}

// contentSecurityPolicy backstops the SPA's parent origin. Scripts must be
// same-origin ('self', no inline/eval); the theme bootstrap is an external
// /theme.js for that reason. Preview iframes (srcDoc/blob) need frame-src, and
// decoded media needs data:/blob: in img/frame-src. Untrusted markup is still
// isolated in per-iframe sandboxes with their own stricter CSP.
const contentSecurityPolicy = "default-src 'self'; " +
	"script-src 'self'; " +
	"style-src 'self' 'unsafe-inline'; " +
	"img-src 'self' data: blob:; " +
	"media-src 'self' data: blob:; " + // decoded audio/video preview (blob URLs)
	"font-src 'self' data:; " +
	"connect-src 'self'; " +
	"frame-src 'self' data: blob:; " +
	"object-src 'none'; " +
	"base-uri 'none'; " +
	"frame-ancestors 'none'"

// securityHeaders adds hardening headers to every response.
func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "no-referrer")
		h.Set("Content-Security-Policy", contentSecurityPolicy)
		next.ServeHTTP(w, r)
	})
}

func (s *Server) Start() error {
	distFS, err := fs.Sub(staticFiles, "dist")
	if err != nil {
		return fmt.Errorf("failed to access embedded assets: %w", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	httpServer := &http.Server{
		Addr:              s.addr,
		ReadHeaderTimeout: 3 * time.Second,
		Handler:           s.newHandler(distFS),
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = httpServer.Shutdown(shutdownCtx)
	}()

	if err := httpServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}
