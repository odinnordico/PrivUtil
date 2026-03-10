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
		cost *int32
	}{
		{"md5", "md5", "hello", nil},
		{"sha1", "sha1", "hello", nil},
		{"sha256", "sha256", "hello", nil},
		{"sha512", "sha512", "hello", nil},
		{"bcrypt default", "bcrypt", "hello", nil},
		{"bcrypt custom cost", "bcrypt", "hello", func(i int32) *int32 { return &i }(4)},
		{"default", "", "hello", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &pb.HashRequest{Text: tt.text, Algo: tt.algo}
			if tt.cost != nil {
				req.Cost = tt.cost
			}
			resp, err := s.CalculateHash(ctx, req)
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

func TestGenerateRsaKeyPair(t *testing.T) {
	s := NewServer()
	ctx := context.Background()

	tests := []struct {
		name      string
		bits      int32
		wantError bool
	}{
		{"default bits", 0, false},
		{"2048 bits", 2048, false},
		{"invalid small bits", 512, true},
		{"invalid large bits", 16384, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := s.GenerateRsaKeyPair(ctx, &pb.RsaKeyRequest{Bits: tt.bits})
			if err != nil {
				t.Fatalf("GenerateRsaKeyPair() error = %v", err)
			}
			if tt.wantError {
				if resp.Error == "" {
					t.Error("GenerateRsaKeyPair() expected error but got none")
				}
				return
			}
			if resp.Error != "" {
				t.Fatalf("GenerateRsaKeyPair() unexpected error: %s", resp.Error)
			}
			if resp.PrivateKey == "" || resp.PublicKey == "" {
				t.Error("GenerateRsaKeyPair() expected non-empty keys")
			}
		})
	}
}
