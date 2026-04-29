package server

import (
	"io/fs"
	"net/http"
	"net/http/httptest"
	"testing"

	"google.golang.org/grpc"
)

func TestNew(t *testing.T) {
	grpcServer := grpc.NewServer()
	srv := New(":8090", grpcServer)

	if srv == nil {
		t.Fatal("New() returned nil")
	}
	if srv.addr != ":8090" {
		t.Errorf("New() addr = %v, want :8090", srv.addr)
	}
	if srv.grpcServer != grpcServer {
		t.Error("New() grpcServer not set correctly")
	}
}

func TestStaticFilesEmbedded(t *testing.T) {
	_, err := staticFiles.ReadFile("dist/index.html")
	if err != nil {
		t.Errorf("Expected dist/index.html to be embedded: %v", err)
	}
}

func testHandler(t *testing.T) http.Handler {
	t.Helper()
	distFS, err := fs.Sub(staticFiles, "dist")
	if err != nil {
		t.Fatalf("fs.Sub: %v", err)
	}
	return New(":0", grpc.NewServer()).newHandler(distFS)
}

func TestServerHandlerStaticFiles(t *testing.T) {
	ts := httptest.NewServer(testHandler(t))
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/")
	if err != nil {
		t.Fatalf("GET /: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("GET / status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestServerHandlerSPAFallback(t *testing.T) {
	ts := httptest.NewServer(testHandler(t))
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/some/client/route")
	if err != nil {
		t.Fatalf("GET /some/client/route: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("GET /some/client/route status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}
