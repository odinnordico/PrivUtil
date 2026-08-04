package server

import (
	"io/fs"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNew(t *testing.T) {
	rpcHandler := http.NewServeMux()
	srv := New(":8090", "/privutil.PrivUtilService/", rpcHandler, []string{"localhost", "127.0.0.1"})

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
	return New(":0", "/privutil.PrivUtilService/", http.NewServeMux(), []string{"localhost", "127.0.0.1", "::1"}).newHandler(distFS)
}

func TestHostAllowlistAndCSP(t *testing.T) {
	ts := httptest.NewServer(testHandler(t))
	defer ts.Close()
	client := ts.Client()

	cases := []struct {
		name string
		host string
		want int
	}{
		{"loopback ip allowed", "127.0.0.1", http.StatusOK},
		{"localhost allowed", "localhost", http.StatusOK},
		{"foreign host blocked", "evil.example.com", http.StatusForbidden},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, ts.URL+"/", nil)
			if err != nil {
				t.Fatalf("NewRequest: %v", err)
			}
			req.Host = tc.host
			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("Do: %v", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != tc.want {
				t.Errorf("Host %q status = %d, want %d", tc.host, resp.StatusCode, tc.want)
			}
			if tc.want == http.StatusOK && resp.Header.Get("Content-Security-Policy") == "" {
				t.Error("expected Content-Security-Policy header on allowed request")
			}
		})
	}
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
