package server

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
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

type Server struct {
	addr       string
	rpcPath    string
	rpcHandler http.Handler
}

// New builds an HTTP server that routes connect RPC requests under rpcPath to
// rpcHandler and serves the embedded React SPA for everything else.
func New(addr, rpcPath string, rpcHandler http.Handler) *Server {
	return &Server{
		addr:       addr,
		rpcPath:    rpcPath,
		rpcHandler: rpcHandler,
	}
}

func (s *Server) newHandler(distFS fs.FS) http.Handler {
	fileServer := http.FileServer(http.FS(distFS))

	// Allow all origins, matching the previous grpc-web wrapper: PrivUtil is a
	// local utility and is also used cross-origin from the Vite dev server. The
	// connectcors helper supplies the headers the Connect/gRPC-Web protocols need.
	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: connectcors.AllowedMethods(),
		AllowedHeaders: connectcors.AllowedHeaders(),
		ExposedHeaders: connectcors.ExposedHeaders(),
	})

	mux := http.NewServeMux()
	mux.Handle(s.rpcPath, corsMiddleware.Handler(s.rpcHandler))
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
	return h2c.NewHandler(mux, &http2.Server{})
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
