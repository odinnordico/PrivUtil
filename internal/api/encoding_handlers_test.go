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

func TestBaseConvert(t *testing.T) {
	s := NewServer()
	ctx := context.Background()

	tests := []struct {
		name       string
		input      string
		sourceBase int32
		wantDec    string
		wantHex    string
		wantBin    string
		wantOct    string
		wantBase64 string
		wantErr    bool
	}{
		{"decimal to all", "255", 10, "255", "FF", "11111111", "377", "D/", false},
		{"hex to all", "FF", 16, "255", "FF", "11111111", "377", "D/", false},
		{"hex with prefix", "0xff", 16, "255", "FF", "11111111", "377", "D/", false},
		{"binary to all", "11111111", 2, "255", "FF", "11111111", "377", "D/", false},
		{"octal to all", "377", 8, "255", "FF", "11111111", "377", "D/", false},
		{"base64 to all", "D/", 64, "255", "FF", "11111111", "377", "D/", false},
		{"zero value", "0", 10, "0", "0", "0", "0", "A", false},
		{"invalid char", "zz", 16, "", "", "", "", "", true},
		{"invalid base64", "?", 64, "", "", "", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := s.BaseConvert(ctx, &pb.BaseConvertRequest{
				Input:      tt.input,
				SourceBase: tt.sourceBase,
			})
			if err != nil {
				t.Fatalf("BaseConvert() error = %v", err)
			}
			if tt.wantErr {
				if resp.Error == "" {
					t.Error("BaseConvert() expected error in response")
				}
				return
			}
			if resp.Error != "" {
				t.Errorf("BaseConvert() unexpected error: %s", resp.Error)
			}
			if resp.Decimal != tt.wantDec {
				t.Errorf("BaseConvert() dec = %v, want %v", resp.Decimal, tt.wantDec)
			}
			if resp.Hex != tt.wantHex {
				t.Errorf("BaseConvert() hex = %v, want %v", resp.Hex, tt.wantHex)
			}
			if resp.Binary != tt.wantBin {
				t.Errorf("BaseConvert() bin = %v, want %v", resp.Binary, tt.wantBin)
			}
			if resp.Octal != tt.wantOct {
				t.Errorf("BaseConvert() oct = %v, want %v", resp.Octal, tt.wantOct)
			}
			if resp.Base64 != tt.wantBase64 {
				t.Errorf("BaseConvert() b64 = %v, want %v", resp.Base64, tt.wantBase64)
			}
		})
	}
}

func TestMarkdownToHtml(t *testing.T) {
	s := NewServer()
	ctx := context.Background()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"heading", "# Hello", "<h1>Hello</h1>\n"},
		{"bold", "**bold**", "<p><strong>bold</strong></p>\n"},
		{"paragraph", "Hello world", "<p>Hello world</p>\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := s.MarkdownToHtml(ctx, &pb.TextRequest{Text: tt.input})
			if err != nil {
				t.Fatalf("MarkdownToHtml() error = %v", err)
			}
			if resp.Text != tt.want {
				t.Errorf("MarkdownToHtml() = %q, want %q", resp.Text, tt.want)
			}
		})
	}
}

func TestHtmlToMarkdown(t *testing.T) {
	s := NewServer()
	ctx := context.Background()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"heading", "<h1>Hello</h1>", "# Hello"},
		{"bold", "<p><strong>bold</strong></p>", "**bold**"},
		{"paragraph", "<p>Hello world</p>", "Hello world"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := s.HtmlToMarkdown(ctx, &pb.TextRequest{Text: tt.input})
			if err != nil {
				t.Fatalf("HtmlToMarkdown() error = %v", err)
			}
			if resp.Text != tt.want {
				t.Errorf("HtmlToMarkdown() = %q, want %q", resp.Text, tt.want)
			}
		})
	}
}
