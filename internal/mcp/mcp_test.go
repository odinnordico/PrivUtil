package mcp

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/odinnordico/privutil/internal/api"
)

// rpc posts a single JSON-RPC message to the MCP endpoint and returns the raw
// response body. Stateless mode accepts calls without a prior initialize.
func rpc(t *testing.T, url, body string) (int, string) {
	t.Helper()
	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(body))
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, string(b)
}

func TestMCPHandler(t *testing.T) {
	ts := httptest.NewServer(Handler(api.NewServer(), "test"))
	defer ts.Close()

	t.Run("tools/list exposes tools", func(t *testing.T) {
		code, body := rpc(t, ts.URL, `{"jsonrpc":"2.0","id":1,"method":"tools/list"}`)
		if code != http.StatusOK {
			t.Fatalf("status = %d, body = %s", code, body)
		}
		// A sampling across every registration group.
		for _, name := range []string{
			"calculate_hash", "generate_uuid", "base64_encode", "jwt_decode", "text_diff",
			"convert", "case_convert", "regex_test", "token_count", "ip_calc", "url_parse",
			"math_eval", "unit_convert", "date_diff", "generate_password", "cron_explain",
		} {
			if !strings.Contains(body, name) {
				t.Errorf("tools/list missing %q; body = %s", name, body)
			}
		}
	})

	t.Run("tools/call with an enum argument (convert json->yaml)", func(t *testing.T) {
		code, body := rpc(t, ts.URL,
			`{"jsonrpc":"2.0","id":9,"method":"tools/call","params":{"name":"convert","arguments":{"data":"{\"a\":1}","source_format":"json","target_format":"yaml"}}}`)
		if code != http.StatusOK {
			t.Fatalf("status = %d, body = %s", code, body)
		}
		if strings.Contains(body, `"isError":true`) || !strings.Contains(body, "a:") {
			t.Errorf("convert json->yaml did not produce YAML; body = %s", body)
		}
	})

	t.Run("tools/call returns a result", func(t *testing.T) {
		// sha256("hello")
		const want = "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"
		code, body := rpc(t, ts.URL,
			`{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"calculate_hash","arguments":{"text":"hello","algo":"sha256"}}}`)
		if code != http.StatusOK {
			t.Fatalf("status = %d, body = %s", code, body)
		}
		if !strings.Contains(body, want) {
			t.Errorf("tools/call did not return the expected hash; body = %s", body)
		}
	})

	t.Run("in-band error is reported", func(t *testing.T) {
		code, body := rpc(t, ts.URL,
			`{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"json_format","arguments":{"text":"{not valid"}}}`)
		if code != http.StatusOK {
			t.Fatalf("status = %d, body = %s", code, body)
		}
		if !strings.Contains(body, "isError") && !strings.Contains(body, "error") {
			t.Errorf("expected an error result for invalid JSON; body = %s", body)
		}
	})
}
