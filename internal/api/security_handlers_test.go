package api

import (
	"context"
	"testing"

	pb "github.com/odinnordico/privutil/proto"
)

func TestCalculateHash(t *testing.T) {
	s := NewServer()
	ctx := context.Background()

	tests := []struct {
		name string
		algo string
		text string
	}{
		{"md5", "md5", "hello"},
		{"sha1", "sha1", "hello"},
		{"sha256", "sha256", "hello"},
		{"sha512", "sha512", "hello"},
		{"default", "", "hello"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := s.CalculateHash(ctx, &pb.HashRequest{Text: tt.text, Algo: tt.algo})
			if err != nil {
				t.Fatalf("CalculateHash() error = %v", err)
			}
			if resp.Hash == "" {
				t.Error("CalculateHash() expected non-empty hash")
			}
		})
	}
}

func TestJwtDecode(t *testing.T) {
	s := NewServer()
	ctx := context.Background()

	// Valid JWT structure (not actually signed)
	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"

	resp, err := s.JwtDecode(ctx, &pb.JwtRequest{Token: token})
	if err != nil {
		t.Fatalf("JwtDecode() error = %v", err)
	}
	if resp.Header == "" || resp.Payload == "" {
		t.Error("JwtDecode() expected header and payload")
	}
}

func TestCertParse(t *testing.T) {
	s := NewServer()
	ctx := context.Background()

	// Invalid PEM
	resp, err := s.CertParse(ctx, &pb.CertRequest{Data: "invalid"})
	if err != nil {
		t.Fatalf("CertParse() error = %v", err)
	}
	if resp.Error == "" {
		t.Error("CertParse() expected error for invalid PEM")
	}
}
