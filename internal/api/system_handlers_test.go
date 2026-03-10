package api

import (
	"context"
	"testing"

	pb "github.com/odinnordico/privutil/proto"
)

func TestTimeConvert(t *testing.T) {
	s := NewServer()
	ctx := context.Background()

	tests := []struct {
		name  string
		input string
	}{
		{"now", "now"},
		{"empty", ""},
		{"unix", "1609459200"},
		{"unix ms", "1609459200000"},
		{"rfc3339", "2021-01-01T00:00:00Z"},
		{"date only", "2021-01-01"},
		{"datetime", "2021-01-01 00:00:00"},
		{"us date", "01/01/2021"},
		{"us date time", "01/01/2021 00:00:00"},
		{"ansic", "Fri Jan  1 00:00:00 2021"},
		{"rfc850", "Friday, 01-Jan-21 00:00:00 UTC"},
		{"rfc1123", "Fri, 01 Jan 2021 00:00:00 UTC"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := s.TimeConvert(ctx, &pb.TimeRequest{Input: tt.input})
			if err != nil {
				t.Fatalf("TimeConvert() error = %v", err)
			}
			if resp.Iso == "" && resp.Iso != "Invalid input format" {
				t.Error("TimeConvert() expected result")
			}
		})
	}

	t.Run("invalid input", func(t *testing.T) {
		resp, err := s.TimeConvert(ctx, &pb.TimeRequest{Input: "not-a-date"})
		if err != nil {
			t.Fatalf("TimeConvert() error = %v", err)
		}
		if resp.Iso != "Invalid input format" {
			t.Errorf("TimeConvert() expected 'Invalid input format', got %q", resp.Iso)
		}
	})
}

func TestCronExplain(t *testing.T) {
	s := NewServer()
	ctx := context.Background()

	tests := []struct {
		name       string
		expression string
		wantError  bool
	}{
		{"every minute", "* * * * *", false},
		{"every 5 min", "*/5 * * * *", false},
		{"invalid", "invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := s.CronExplain(ctx, &pb.CronRequest{Expression: tt.expression})
			if err != nil {
				t.Fatalf("CronExplain() error = %v", err)
			}
			if tt.wantError && resp.Error == "" {
				t.Error("CronExplain() expected error")
			}
			if !tt.wantError && resp.Description == "" {
				t.Error("CronExplain() expected description")
			}
		})
	}
}

func TestIpCalc(t *testing.T) {
	s := NewServer()
	ctx := context.Background()

	tests := []struct {
		name      string
		cidr      string
		wantError bool
	}{
		{"valid cidr", "192.168.1.0/24", false},
		{"single ip", "192.168.1.1", false},
		{"invalid", "invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := s.IpCalc(ctx, &pb.IpRequest{Cidr: tt.cidr})
			if err != nil {
				t.Fatalf("IpCalc() error = %v", err)
			}
			if tt.wantError && resp.Error == "" {
				t.Error("IpCalc() expected error")
			}
			if !tt.wantError && resp.Network == "" {
				t.Error("IpCalc() expected network")
			}
		})
	}
}
