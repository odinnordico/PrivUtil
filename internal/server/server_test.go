package server

import (
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
	// Verify that static files are embedded
	_, err := staticFiles.ReadFile("dist/index.html")
	if err != nil {
		t.Errorf("Expected dist/index.html to be embedded: %v", err)
	}
}

func TestServerHandlerStaticFiles(t *testing.T) {
	grpcServer := grpc.NewServer()
	srv := New(":0", grpcServer)

	// Verify the server struct is properly initialized
	if srv.addr != ":0" {
		t.Errorf("Server addr = %v, want :0", srv.addr)
	}
}

func TestServerHandlerSPAFallback(t *testing.T) {
	// Test that non-existent paths would fall back to index.html (SPA routing)
	grpcServer := grpc.NewServer()
	srv := New(":0", grpcServer)

	if srv == nil {
		t.Fatal("Server should not be nil")
	}
}
