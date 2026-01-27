package api

import (
	"context"
	"testing"

	pb "github.com/odinnordico/privutil/proto"
)

func TestBase64Encode(t *testing.T) {
	s := NewServer()
	ctx := context.Background()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"simple", "hello", "aGVsbG8="},
		{"empty", "", ""},
		{"special chars", "hello world!", "aGVsbG8gd29ybGQh"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := s.Base64Encode(ctx, &pb.Base64Request{Text: tt.input})
			if err != nil {
				t.Fatalf("Base64Encode() error = %v", err)
			}
			if resp.Text != tt.want {
				t.Errorf("Base64Encode() = %v, want %v", resp.Text, tt.want)
			}
		})
	}
}

func TestBase64Decode(t *testing.T) {
	s := NewServer()
	ctx := context.Background()

	tests := []struct {
		name      string
		input     string
		want      string
		wantError bool
	}{
		{"simple", "aGVsbG8=", "hello", false},
		{"empty", "", "", false},
		{"invalid", "!!!invalid!!!", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := s.Base64Decode(ctx, &pb.Base64Request{Text: tt.input})
			if err != nil {
				t.Fatalf("Base64Decode() unexpected error = %v", err)
			}
			if tt.wantError && resp.Error == "" {
				t.Error("Base64Decode() expected error in response")
			}
			if !tt.wantError && resp.Text != tt.want {
				t.Errorf("Base64Decode() = %v, want %v", resp.Text, tt.want)
			}
		})
	}
}

func TestUrlEncodeDecode(t *testing.T) {
	s := NewServer()
	ctx := context.Background()

	original := "hello world & more"
	encResp, _ := s.UrlEncode(ctx, &pb.TextRequest{Text: original})
	decResp, _ := s.UrlDecode(ctx, &pb.TextRequest{Text: encResp.Text})

	if decResp.Text != original {
		t.Errorf("UrlEncode/Decode roundtrip failed: got %v, want %v", decResp.Text, original)
	}
}

func TestHtmlEncodeDecode(t *testing.T) {
	s := NewServer()
	ctx := context.Background()

	original := "<script>alert('xss')</script>"
	encResp, _ := s.HtmlEncode(ctx, &pb.TextRequest{Text: original})
	decResp, _ := s.HtmlDecode(ctx, &pb.TextRequest{Text: encResp.Text})

	if decResp.Text != original {
		t.Errorf("HtmlEncode/Decode roundtrip failed: got %v, want %v", decResp.Text, original)
	}
}

func TestStringEscape(t *testing.T) {
	s := NewServer()
	ctx := context.Background()

	tests := []struct {
		name   string
		mode   string
		action string
		text   string
	}{
		{"json escape", "json", "escape", `hello "world"`},
		{"json unescape", "json", "unescape", `"hello \"world\""`},
		{"url escape", "url", "escape", "hello world"},
		{"html escape", "html_entity", "escape", "<script>"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := s.StringEscape(ctx, &pb.EscapeRequest{
				Text:   tt.text,
				Mode:   tt.mode,
				Action: tt.action,
			})
			if err != nil {
				t.Fatalf("StringEscape() error = %v", err)
			}
			if resp.Error != "" {
				t.Errorf("StringEscape() error = %v", resp.Error)
			}
		})
	}
}
