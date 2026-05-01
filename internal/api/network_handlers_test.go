package api

import (
	"context"
	"slices"
	"strings"
	"testing"

	pb "github.com/odinnordico/privutil/proto"
)

func TestChmodCalc(t *testing.T) {
	s := NewServer()
	ctx := context.Background()

	tests := []struct {
		name         string
		input        string
		wantOctal    string
		wantSymbolic string
		wantError    bool
	}{
		{"755 octal", "755", "755", "rwxr-xr-x", false},
		{"644 octal", "644", "644", "rw-r--r--", false},
		{"0 octal", "0", "0", "---------", false},
		{"777 octal", "777", "777", "rwxrwxrwx", false},
		{"4755 setuid", "4755", "4755", "rwsr-xr-x", false},
		{"1777 sticky", "1777", "1777", "rwxrwxrwt", false},
		{"symbolic rwxr-xr-x", "rwxr-xr-x", "755", "rwxr-xr-x", false},
		{"symbolic rw-r--r--", "rw-r--r--", "644", "rw-r--r--", false},
		{"full symbolic -rwxr-xr-x", "-rwxr-xr-x", "755", "rwxr-xr-x", false},
		{"invalid input", "invalid", "", "", true},
		{"empty input", "", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := s.ChmodCalc(ctx, &pb.ChmodRequest{Input: tt.input})
			if err != nil {
				t.Fatalf("ChmodCalc() error = %v", err)
			}
			if tt.wantError {
				if resp.Error == "" {
					t.Error("ChmodCalc() expected error but got none")
				}
				return
			}
			if resp.Error != "" {
				t.Errorf("ChmodCalc() unexpected error = %v", resp.Error)
			}
			if resp.Octal != tt.wantOctal {
				t.Errorf("ChmodCalc() octal = %q, want %q", resp.Octal, tt.wantOctal)
			}
			if resp.Symbolic != tt.wantSymbolic {
				t.Errorf("ChmodCalc() symbolic = %q, want %q", resp.Symbolic, tt.wantSymbolic)
			}
		})
	}
}

func TestIpv4Convert(t *testing.T) {
	s := NewServer()
	ctx := context.Background()

	tests := []struct {
		name        string
		input       string
		wantDotted  string
		wantDecimal string
		wantHex     string
		wantError   bool
	}{
		{"dotted decimal", "192.168.1.1", "192.168.1.1", "3232235777", "0xC0A80101", false},
		{"decimal integer", "3232235777", "192.168.1.1", "3232235777", "0xC0A80101", false},
		{"hex with 0x", "0xC0A80101", "192.168.1.1", "3232235777", "0xC0A80101", false},
		{"hex without 0x", "C0A80101", "192.168.1.1", "3232235777", "0xC0A80101", false},
		{"binary dotted", "11000000.10101000.00000001.00000001", "192.168.1.1", "3232235777", "0xC0A80101", false},
		{"0.0.0.0", "0.0.0.0", "0.0.0.0", "0", "0x00000000", false},
		{"255.255.255.255", "255.255.255.255", "255.255.255.255", "4294967295", "0xFFFFFFFF", false},
		{"invalid", "not-an-ip", "", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := s.Ipv4Convert(ctx, &pb.Ipv4ConvertRequest{Input: tt.input})
			if err != nil {
				t.Fatalf("Ipv4Convert() error = %v", err)
			}
			if tt.wantError {
				if resp.Error == "" {
					t.Error("Ipv4Convert() expected error but got none")
				}
				return
			}
			if resp.Error != "" {
				t.Errorf("Ipv4Convert() unexpected error = %v", resp.Error)
			}
			if resp.Dotted != tt.wantDotted {
				t.Errorf("Ipv4Convert() dotted = %q, want %q", resp.Dotted, tt.wantDotted)
			}
			if resp.Decimal != tt.wantDecimal {
				t.Errorf("Ipv4Convert() decimal = %q, want %q", resp.Decimal, tt.wantDecimal)
			}
			if resp.Hex != tt.wantHex {
				t.Errorf("Ipv4Convert() hex = %q, want %q", resp.Hex, tt.wantHex)
			}
		})
	}
}

func TestIpv4RangeExpand(t *testing.T) {
	s := NewServer()
	ctx := context.Background()

	tests := []struct {
		name      string
		start     string
		end       string
		wantTotal int64
		wantCIDRs []string
		wantError bool
	}{
		{
			"single IP",
			"192.168.0.1", "192.168.0.1",
			1, []string{"192.168.0.1/32"}, false,
		},
		{
			"/24 network",
			"192.168.0.0", "192.168.0.255",
			256, []string{"192.168.0.0/24"}, false,
		},
		{
			"small range",
			"10.0.0.1", "10.0.0.3",
			3, []string{"10.0.0.1/32", "10.0.0.2/31"}, false,
		},
		{
			"start > end",
			"192.168.0.5", "192.168.0.1",
			0, nil, true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := s.Ipv4RangeExpand(ctx, &pb.Ipv4RangeRequest{Start: tt.start, End: tt.end})
			if err != nil {
				t.Fatalf("Ipv4RangeExpand() error = %v", err)
			}
			if tt.wantError {
				if resp.Error == "" {
					t.Error("Ipv4RangeExpand() expected error but got none")
				}
				return
			}
			if resp.Error != "" {
				t.Errorf("Ipv4RangeExpand() unexpected error = %v", resp.Error)
			}
			if resp.Total != tt.wantTotal {
				t.Errorf("Ipv4RangeExpand() total = %d, want %d", resp.Total, tt.wantTotal)
			}
			if len(tt.wantCIDRs) > 0 {
				for _, cidr := range tt.wantCIDRs {
					found := slices.Contains(resp.Cidrs, cidr)
					if !found {
						t.Errorf("Ipv4RangeExpand() CIDRs %v missing %q", resp.Cidrs, cidr)
					}
				}
			}
		})
	}
}

func TestGeneratePort(t *testing.T) {
	s := NewServer()
	ctx := context.Background()

	tests := []struct {
		name          string
		count         int32
		min           int32
		max           int32
		excludeSystem bool
		wantError     bool
	}{
		{"default count", 1, 0, 0, false, false},
		{"multiple ports", 10, 1024, 65535, false, false},
		{"exclude system", 5, 0, 65535, true, false},
		{"cap at 100", 200, 1024, 65535, false, false},
		{"impossible range", 5, 2000, 1000, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := s.GeneratePort(ctx, &pb.PortRequest{
				Count:         tt.count,
				Min:           tt.min,
				Max:           tt.max,
				ExcludeSystem: tt.excludeSystem,
			})
			if err != nil {
				t.Fatalf("GeneratePort() error = %v", err)
			}
			if tt.wantError {
				if resp.Error == "" {
					t.Error("GeneratePort() expected error but got none")
				}
				return
			}
			if resp.Error != "" {
				t.Errorf("GeneratePort() unexpected error = %v", resp.Error)
			}
			if len(resp.Ports) == 0 {
				t.Error("GeneratePort() expected non-empty ports")
			}
			if tt.excludeSystem {
				for _, p := range resp.Ports {
					if p <= 1023 {
						t.Errorf("GeneratePort() port %d is a system port", p)
					}
				}
			}
		})
	}
}

func TestGenerateMac(t *testing.T) {
	s := NewServer()
	ctx := context.Background()

	tests := []struct {
		name      string
		count     int32
		sep       string
		upper     bool
		oui       string
		wantError bool
	}{
		{"default colon", 3, ":", false, "", false},
		{"dash separator", 1, "-", false, "", false},
		{"cisco dot", 1, ".", false, "", false},
		{"uppercase", 1, ":", true, "", false},
		{"with OUI", 5, ":", false, "00:1A:2B", false},
		{"invalid OUI", 1, ":", false, "ZZ:ZZ", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := s.GenerateMac(ctx, &pb.MacRequest{
				Count:     tt.count,
				Separator: tt.sep,
				Uppercase: tt.upper,
				Oui:       tt.oui,
			})
			if err != nil {
				t.Fatalf("GenerateMac() error = %v", err)
			}
			if tt.wantError {
				if resp.Error == "" {
					t.Error("GenerateMac() expected error but got none")
				}
				return
			}
			if resp.Error != "" {
				t.Errorf("GenerateMac() unexpected error = %v", resp.Error)
			}
			if len(resp.Addresses) == 0 {
				t.Error("GenerateMac() expected non-empty addresses")
			}
			// Validate format of first address
			addr := resp.Addresses[0]
			if tt.upper && addr != strings.ToUpper(addr) {
				t.Errorf("GenerateMac() address %q not uppercase", addr)
			}
			if tt.oui != "" {
				normalAddr := strings.ToLower(strings.NewReplacer(":", "", "-", "", ".", "").Replace(addr))
				normalOUI := strings.ToLower(strings.NewReplacer(":", "", "-", "", ".", "").Replace(tt.oui))
				if !strings.HasPrefix(normalAddr, normalOUI) {
					t.Errorf("GenerateMac() address %q missing OUI prefix %q", addr, tt.oui)
				}
			}
		})
	}
}
