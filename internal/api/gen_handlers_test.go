package api

import (
	"context"
	"testing"

	pb "github.com/odinnordico/privutil/proto"
)

func TestGenerateUuid(t *testing.T) {
	s := NewServer()
	ctx := context.Background()

	tests := []struct {
		name      string
		count     int32
		version   string
		hyphen    bool
		uppercase bool
	}{
		{"single v4", 1, "v4", true, false},
		{"multiple", 5, "v4", true, false},
		{"no hyphen", 1, "v4", false, false},
		{"uppercase", 1, "v4", true, true},
		{"v1", 1, "v1", true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := s.GenerateUuid(ctx, &pb.UuidRequest{
				Count:     tt.count,
				Version:   tt.version,
				Hyphen:    tt.hyphen,
				Uppercase: tt.uppercase,
			})
			if err != nil {
				t.Fatalf("GenerateUuid() error = %v", err)
			}
			if len(resp.Uuids) != int(tt.count) {
				t.Errorf("GenerateUuid() got %d uuids, want %d", len(resp.Uuids), tt.count)
			}
		})
	}
}

func TestGenerateLorem(t *testing.T) {
	s := NewServer()
	ctx := context.Background()

	tests := []struct {
		name  string
		typ   string
		count int32
	}{
		{"word", "word", 5},
		{"sentence", "sentence", 3},
		{"paragraph", "paragraph", 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := s.GenerateLorem(ctx, &pb.LoremRequest{Type: tt.typ, Count: tt.count})
			if err != nil {
				t.Fatalf("GenerateLorem() error = %v", err)
			}
			if resp.Text == "" {
				t.Error("GenerateLorem() expected non-empty result")
			}
		})
	}
}

func TestGeneratePassword(t *testing.T) {
	s := NewServer()
	ctx := context.Background()

	tests := []struct {
		name   string
		length int32
		count  int32
	}{
		{"single", 16, 1},
		{"multiple", 12, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := s.GeneratePassword(ctx, &pb.PasswordRequest{
				Length:    tt.length,
				Count:     tt.count,
				Lowercase: true,
				Uppercase: true,
				Numbers:   true,
				Symbols:   true,
			})
			if err != nil {
				t.Fatalf("GeneratePassword() error = %v", err)
			}
			if len(resp.Passwords) != int(tt.count) {
				t.Errorf("GeneratePassword() got %d passwords, want %d", len(resp.Passwords), tt.count)
			}
		})
	}
}
