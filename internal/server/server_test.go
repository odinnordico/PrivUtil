package server

import (
	"io/fs"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNew(t *testing.T) {
	rpcHandler := http.NewServeMux()
	srv := New(":8090", "/privutil.PrivUtilService/", rpcHandler)

	if srv == nil {
		t.Fatal("New() returned nil")
	}
	if srv.addr != ":8090" {
		t.Errorf("New() addr = %v, want :8090", srv.addr)
	}
	if srv.rpcPath != "/privutil.PrivUtilService/" {
		t.Errorf("New() rpcPath = %v, want /privutil.PrivUtilService/", srv.rpcPath)
	}
	if srv.rpcHandler == nil {
		t.Error("New() rpcHandler not set correctly")
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
	return New(":0", "/privutil.PrivUtilService/", http.NewServeMux()).newHandler(distFS)
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
