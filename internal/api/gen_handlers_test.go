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
		namespace string
		hyphen    bool
		uppercase bool
	}{
		{"single v4", 1, "v4", "", true, false},
		{"multiple", 5, "v4", "", true, false},
		{"no hyphen", 1, "v4", "", false, false},
		{"uppercase", 1, "v4", "", true, true},
		{"v1", 1, "v1", "", true, false},
		{"v2", 1, "v2", "", true, false},
		{"v3", 1, "v3", "", true, false},
		{"v3 with dns", 1, "v3", "dns", true, false},
		{"v3 with url", 1, "v3", "url", true, false},
		{"v3 with oid", 1, "v3", "oid", true, false},
		{"v3 with x500", 1, "v3", "x500", true, false},
		{"v5", 1, "v5", "", true, false},
		{"v5 with url", 1, "v5", "url", true, false},
		{"v6", 1, "v6", "", true, false},
		{"v7", 1, "v7", "", true, false},
		{"v8", 1, "v8", "", true, false},
		{"v8 with url", 1, "v8", "url", true, false},
		{"v2 multiple", 3, "v2", "", true, false},
		{"v3 no hyphen", 1, "v3", "", false, false},
		{"v5 uppercase", 1, "v5", "", true, true},
		{"v6 multiple", 5, "v6", "", true, false},
		{"v7 no hyphen uppercase", 1, "v7", "", false, true},
		{"v8 multiple", 3, "v8", "", true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := s.GenerateUuid(ctx, &pb.UuidRequest{
				Count:     tt.count,
				Version:   tt.version,
				Namespace: tt.namespace,
				Hyphen:    tt.hyphen,
				Uppercase: tt.uppercase,
			})
			if err != nil {
				t.Fatalf("GenerateUuid() error = %v", err)
			}
			if len(resp.Uuids) != int(tt.count) {
				t.Errorf("GenerateUuid() got %d uuids, want %d", len(resp.Uuids), tt.count)
			}

			// Validate each UUID
			for _, uuidStr := range resp.Uuids {
				// Check hyphen format
				if tt.hyphen && !containsHyphens(uuidStr) {
					t.Errorf("GenerateUuid() expected hyphens in UUID: %s", uuidStr)
				}
				if !tt.hyphen && containsHyphens(uuidStr) {
					t.Errorf("GenerateUuid() expected no hyphens in UUID: %s", uuidStr)
				}

				// Check case
				if tt.uppercase && !isUpperCase(uuidStr) {
					t.Errorf("GenerateUuid() expected uppercase UUID: %s", uuidStr)
				}
				if !tt.uppercase && isUpperCase(uuidStr) {
					t.Errorf("GenerateUuid() expected lowercase UUID: %s", uuidStr)
				}

				// Validate UUID format (with or without hyphens)
				if tt.hyphen {
					// Standard format: 8-4-4-4-12
					if len(uuidStr) != 36 {
						t.Errorf("GenerateUuid() UUID with hyphens should be 36 chars, got %d: %s", len(uuidStr), uuidStr)
					}
				} else {
					// No hyphens: 32 hex chars
					if len(uuidStr) != 32 {
						t.Errorf("GenerateUuid() UUID without hyphens should be 32 chars, got %d: %s", len(uuidStr), uuidStr)
					}
				}
			}
		})
	}
}

// Helper function to check if string contains hyphens
func containsHyphens(s string) bool {
	for _, c := range s {
		if c == '-' {
			return true
		}
	}
	return false
}

// Helper function to check if string is uppercase
func isUpperCase(s string) bool {
	hasAlpha := false
	for _, c := range s {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') {
			hasAlpha = true
			if c >= 'a' && c <= 'z' {
				return false
			}
		}
	}
	return hasAlpha
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
