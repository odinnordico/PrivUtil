package server

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"
	"time"

	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"google.golang.org/grpc"
)

//go:embed dist/*
var staticFiles embed.FS

type Server struct {
	addr       string
	grpcServer *grpc.Server
}

func New(addr string, grpcServer *grpc.Server) *Server {
	return &Server{
		addr:       addr,
		grpcServer: grpcServer,
	}
}

func (s *Server) Start() error {
	wrappedGrpc := grpcweb.WrapServer(s.grpcServer, grpcweb.WithOriginFunc(func(origin string) bool {
		return true // Allow all origins for dev
	}))

	// Setup static file server
	distFS, _ := fs.Sub(staticFiles, "dist")
	fileServer := http.FileServer(http.FS(distFS))

	httpServer := &http.Server{
		Addr:              s.addr,
		ReadHeaderTimeout: 3 * time.Second, // G112: Potential Slowloris Attack
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if wrappedGrpc.IsGrpcWebRequest(r) {
				wrappedGrpc.ServeHTTP(w, r)
				return
			}

			// Serve static files
			// If file exists in dist, serve it. Otherwise serve index.html for SPA routing
			path := r.URL.Path
			if path == "/" {
				path = "index.html"
			}

			// Check if file exists in the embedded FS
			f, err := distFS.Open(strings.TrimPrefix(path, "/"))
			if err != nil {
				// File not found, serve index.html for client-side routing
				r.URL.Path = "/"
				fileServer.ServeHTTP(w, r)
				return
			}
			if f != nil {
				_ = f.Close() // #nosec G104: Errors unhandled
			}

			fileServer.ServeHTTP(w, r)
		}),
	}

	return httpServer.ListenAndServe()
}
