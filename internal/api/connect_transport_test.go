package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	connect "connectrpc.com/connect"
	pb "github.com/odinnordico/privutil/proto"
	protoconnect "github.com/odinnordico/privutil/proto/protoconnect"
)

// newGRPCWebClient spins up the connect handler exactly as main.go wires it and
// returns a client that speaks the gRPC-Web protocol — the same protocol the
// browser frontend (nice-grpc-web) uses. This guards the transport migration:
// it must accept gRPC-Web requests and map handler status codes correctly.
func newGRPCWebClient(t *testing.T) protoconnect.PrivUtilServiceClient {
	t.Helper()
	path, handler := protoconnect.NewPrivUtilServiceHandler(
		NewConnectServer(NewServer()),
		connect.WithInterceptors(RecoveryInterceptor()),
	)
	mux := http.NewServeMux()
	mux.Handle(path, handler)
	ts := httptest.NewServer(mux)
	t.Cleanup(ts.Close)

	return protoconnect.NewPrivUtilServiceClient(ts.Client(), ts.URL, connect.WithGRPCWeb())
}

func TestConnectGRPCWebSuccess(t *testing.T) {
	client := newGRPCWebClient(t)
	resp, err := client.SpellLanguages(context.Background(), connect.NewRequest(&pb.SpellLanguagesRequest{}))
	if err != nil {
		t.Fatalf("SpellLanguages over gRPC-Web: %v", err)
	}
	if len(resp.Msg.Languages) == 0 {
		t.Error("expected at least one supported language")
	}
}

func TestConnectGRPCWebPreservesInvalidArgument(t *testing.T) {
	client := newGRPCWebClient(t)
	// SvgOptimize returns codes.InvalidArgument when svg is empty. This is the
	// error-status path that previously crashed under improbable-eng/grpc-web,
	// and the code must survive the gRPC-status -> connect-code translation.
	_, err := client.SvgOptimize(context.Background(), connect.NewRequest(&pb.SvgOptimizeRequest{Svg: ""}))
	if err == nil {
		t.Fatal("expected error for empty svg, got nil")
	}
	if got := connect.CodeOf(err); got != connect.CodeInvalidArgument {
		t.Errorf("error code = %v, want %v", got, connect.CodeInvalidArgument)
	}
}
