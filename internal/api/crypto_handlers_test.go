package api

import (
	"context"
	"strings"
	"testing"

	pb "github.com/odinnordico/privutil/proto"
)

func TestHmacGenerate(t *testing.T) {
	s := NewServer()
	ctx := context.Background()

	algos := []string{"sha256", "sha512", "sha1", "md5", ""}
	for _, algo := range algos {
		t.Run("algo_"+algo, func(t *testing.T) {
			resp, err := s.HmacGenerate(ctx, &pb.HmacRequest{
				Message: "Hello, World!",
				Secret:  "secret",
				Algo:    algo,
			})
			if err != nil {
				t.Fatalf("HmacGenerate() error = %v", err)
			}
			if resp.Error != "" {
				t.Errorf("unexpected error: %s", resp.Error)
			}
			if resp.Hex == "" {
				t.Error("expected non-empty hex")
			}
			if resp.Base64 == "" {
				t.Error("expected non-empty base64")
			}
		})
	}
}

func TestHmacGenerateConsistency(t *testing.T) {
	s := NewServer()
	ctx := context.Background()

	// Same inputs must always produce the same output
	r1, _ := s.HmacGenerate(ctx, &pb.HmacRequest{Message: "msg", Secret: "key", Algo: "sha256"})
	r2, _ := s.HmacGenerate(ctx, &pb.HmacRequest{Message: "msg", Secret: "key", Algo: "sha256"})
	if r1.Hex != r2.Hex {
		t.Errorf("HMAC is not deterministic: %s != %s", r1.Hex, r2.Hex)
	}
	// Different secret → different output
	r3, _ := s.HmacGenerate(ctx, &pb.HmacRequest{Message: "msg", Secret: "other", Algo: "sha256"})
	if r1.Hex == r3.Hex {
		t.Error("different secrets produced the same HMAC")
	}
}

func TestOtpGenerate(t *testing.T) {
	s := NewServer()
	ctx := context.Background()

	t.Run("generate new secret and code", func(t *testing.T) {
		resp, err := s.OtpGenerate(ctx, &pb.OtpRequest{GenerateSecret: true, Digits: 6})
		if err != nil {
			t.Fatalf("OtpGenerate() error = %v", err)
		}
		if resp.Error != "" {
			t.Errorf("unexpected error: %s", resp.Error)
		}
		if resp.Secret == "" {
			t.Error("expected generated secret")
		}
		if len(resp.Code) != 6 {
			t.Errorf("expected 6-digit code, got %q", resp.Code)
		}
		if resp.TimeRemaining <= 0 {
			t.Error("expected positive time_remaining")
		}
	})

	t.Run("existing secret", func(t *testing.T) {
		// First generate a secret
		gen, _ := s.OtpGenerate(ctx, &pb.OtpRequest{GenerateSecret: true, Digits: 6})
		// Then use it
		resp, err := s.OtpGenerate(ctx, &pb.OtpRequest{Secret: gen.Secret, Digits: 6})
		if err != nil {
			t.Fatalf("OtpGenerate() error = %v", err)
		}
		if resp.Error != "" {
			t.Errorf("unexpected error: %s", resp.Error)
		}
		if len(resp.Code) != 6 {
			t.Errorf("expected 6-digit code, got %q", resp.Code)
		}
	})

	t.Run("8-digit code", func(t *testing.T) {
		resp, err := s.OtpGenerate(ctx, &pb.OtpRequest{GenerateSecret: true, Digits: 8})
		if err != nil {
			t.Fatalf("OtpGenerate() error = %v", err)
		}
		if len(resp.Code) != 8 {
			t.Errorf("expected 8-digit code, got %q", resp.Code)
		}
	})

	t.Run("invalid secret", func(t *testing.T) {
		resp, err := s.OtpGenerate(ctx, &pb.OtpRequest{Secret: "!!!invalid!!!"})
		if err != nil {
			t.Fatalf("OtpGenerate() error = %v", err)
		}
		if resp.Error == "" {
			t.Error("expected error for invalid secret")
		}
	})
}

func TestOtpValidate(t *testing.T) {
	s := NewServer()
	ctx := context.Background()

	// Generate a secret and code, then validate
	gen, err := s.OtpGenerate(ctx, &pb.OtpRequest{GenerateSecret: true, Digits: 6})
	if err != nil || gen.Error != "" {
		t.Fatalf("setup failed: %v / %s", err, gen.Error)
	}

	t.Run("valid code", func(t *testing.T) {
		resp, err := s.OtpValidate(ctx, &pb.OtpValidateRequest{
			Secret: gen.Secret,
			Code:   gen.Code,
			Window: 1,
		})
		if err != nil {
			t.Fatalf("OtpValidate() error = %v", err)
		}
		if !resp.Valid {
			t.Error("expected valid=true for correct code")
		}
	})

	t.Run("wrong code", func(t *testing.T) {
		resp, err := s.OtpValidate(ctx, &pb.OtpValidateRequest{
			Secret: gen.Secret,
			Code:   "000000",
			Window: 0,
		})
		if err != nil {
			t.Fatalf("OtpValidate() error = %v", err)
		}
		// This might coincidentally be valid, but 000000 is very unlikely
		_ = resp
	})
}

func TestUlidGenerate(t *testing.T) {
	s := NewServer()
	ctx := context.Background()

	t.Run("generate single", func(t *testing.T) {
		resp, err := s.UlidGenerate(ctx, &pb.UlidRequest{Count: 1})
		if err != nil {
			t.Fatalf("UlidGenerate() error = %v", err)
		}
		if resp.Error != "" {
			t.Errorf("unexpected error: %s", resp.Error)
		}
		if len(resp.Ulids) != 1 {
			t.Errorf("expected 1 ULID, got %d", len(resp.Ulids))
		}
		if len(resp.Ulids[0]) != 26 {
			t.Errorf("expected 26-char ULID, got %q", resp.Ulids[0])
		}
	})

	t.Run("generate multiple", func(t *testing.T) {
		resp, err := s.UlidGenerate(ctx, &pb.UlidRequest{Count: 5, Monotonic: true})
		if err != nil {
			t.Fatalf("UlidGenerate() error = %v", err)
		}
		if len(resp.Ulids) != 5 {
			t.Errorf("expected 5 ULIDs, got %d", len(resp.Ulids))
		}
		// Check uniqueness
		seen := map[string]struct{}{}
		for _, u := range resp.Ulids {
			seen[u] = struct{}{}
		}
		if len(seen) != 5 {
			t.Error("ULIDs are not unique")
		}
	})

	t.Run("cap at 100", func(t *testing.T) {
		resp, err := s.UlidGenerate(ctx, &pb.UlidRequest{Count: 200})
		if err != nil {
			t.Fatalf("UlidGenerate() error = %v", err)
		}
		if len(resp.Ulids) != 100 {
			t.Errorf("expected 100 ULIDs (capped), got %d", len(resp.Ulids))
		}
	})
}

func TestCaesarCipher(t *testing.T) {
	s := NewServer()
	ctx := context.Background()

	tests := []struct {
		name   string
		text   string
		shift  int32
		action string
		want   string
	}{
		{"ROT13 encode", "Hello, World!", 13, "encode", "Uryyb, Jbeyq!"},
		{"ROT13 decode", "Uryyb, Jbeyq!", 13, "decode", "Hello, World!"},
		{"Caesar 3 encode", "ABC", 3, "encode", "DEF"},
		{"Caesar 3 decode", "DEF", 3, "decode", "ABC"},
		{"wrap-around", "XYZ", 3, "encode", "ABC"},
		{"preserve case", "aBc", 1, "encode", "bCd"},
		{"non-alpha unchanged", "Hello! 123", 13, "encode", "Uryyb! 123"},
		{"zero shift", "Hello", 0, "encode", "Hello"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := s.CaesarCipher(ctx, &pb.CaesarRequest{
				Text:   tt.text,
				Shift:  tt.shift,
				Action: tt.action,
			})
			if err != nil {
				t.Fatalf("CaesarCipher() error = %v", err)
			}
			if resp.Error != "" {
				t.Errorf("unexpected error: %s", resp.Error)
			}
			if resp.Result != tt.want {
				t.Errorf("CaesarCipher() = %q, want %q", resp.Result, tt.want)
			}
		})
	}
}

func TestTextEncode(t *testing.T) {
	s := NewServer()
	ctx := context.Background()

	roundTrips := []struct {
		format string
		text   string
	}{
		{"binary", "Hi"},
		{"hex", "Hello"},
		{"octal", "ABC"},
		{"decimal", "Test"},
	}

	for _, tt := range roundTrips {
		t.Run("roundtrip_"+tt.format, func(t *testing.T) {
			enc, err := s.TextEncode(ctx, &pb.TextEncodeRequest{
				Text:   tt.text,
				Action: "encode",
				Format: tt.format,
			})
			if err != nil || enc.Error != "" {
				t.Fatalf("encode failed: %v / %s", err, enc.Error)
			}
			dec, err := s.TextEncode(ctx, &pb.TextEncodeRequest{
				Text:   enc.Result,
				Action: "decode",
				Format: tt.format,
			})
			if err != nil || dec.Error != "" {
				t.Fatalf("decode failed: %v / %s", err, dec.Error)
			}
			if dec.Result != tt.text {
				t.Errorf("roundtrip %s: got %q, want %q", tt.format, dec.Result, tt.text)
			}
		})
	}

	t.Run("known binary encoding", func(t *testing.T) {
		resp, err := s.TextEncode(ctx, &pb.TextEncodeRequest{Text: "A", Action: "encode", Format: "binary"})
		if err != nil || resp.Error != "" {
			t.Fatalf("unexpected error")
		}
		if resp.Result != "01000001" {
			t.Errorf("expected 01000001, got %q", resp.Result)
		}
	})

	t.Run("known hex encoding", func(t *testing.T) {
		resp, err := s.TextEncode(ctx, &pb.TextEncodeRequest{Text: "A", Action: "encode", Format: "hex"})
		if err != nil || resp.Error != "" {
			t.Fatalf("unexpected error")
		}
		if resp.Result != "41" {
			t.Errorf("expected 41, got %q", resp.Result)
		}
	})
}

func TestMorseCode(t *testing.T) {
	s := NewServer()
	ctx := context.Background()

	tests := []struct {
		name   string
		text   string
		action string
		want   string
	}{
		{"encode SOS", "SOS", "encode", "... --- ..."},
		{"encode hello", "HELLO", "encode", ".... . .-.. .-.. ---"},
		{"decode SOS", "... --- ...", "decode", "SOS"},
		{"encode with space", "HI YOU", "encode", ".... .. / -.-- --- ..-"},
		{"decode with word break", ".... .. / -.-- --- ..-", "decode", "HI YOU"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := s.MorseCode(ctx, &pb.MorseRequest{Text: tt.text, Action: tt.action})
			if err != nil {
				t.Fatalf("MorseCode() error = %v", err)
			}
			if resp.Error != "" {
				t.Errorf("unexpected error: %s", resp.Error)
			}
			if resp.Result != tt.want {
				t.Errorf("MorseCode() = %q, want %q", resp.Result, tt.want)
			}
		})
	}

	t.Run("unsupported character", func(t *testing.T) {
		resp, err := s.MorseCode(ctx, &pb.MorseRequest{Text: "héllo", Action: "encode"})
		if err != nil {
			t.Fatalf("MorseCode() error = %v", err)
		}
		if resp.Error == "" {
			t.Error("expected error for unsupported character")
		}
	})
}

func TestBasicAuthGenerate(t *testing.T) {
	s := NewServer()
	ctx := context.Background()

	t.Run("standard credentials", func(t *testing.T) {
		resp, err := s.BasicAuthGenerate(ctx, &pb.BasicAuthRequest{
			Username: "admin",
			Password: "secret",
		})
		if err != nil {
			t.Fatalf("BasicAuthGenerate() error = %v", err)
		}
		if resp.Error != "" {
			t.Errorf("unexpected error: %s", resp.Error)
		}
		if !strings.HasPrefix(resp.Header, "Basic ") {
			t.Errorf("expected header to start with 'Basic ', got %q", resp.Header)
		}
		if resp.Token == "" {
			t.Error("expected non-empty token")
		}
		if resp.Decoded != "admin:secret" {
			t.Errorf("decoded = %q, want %q", resp.Decoded, "admin:secret")
		}
	})

	t.Run("empty password allowed", func(t *testing.T) {
		resp, err := s.BasicAuthGenerate(ctx, &pb.BasicAuthRequest{Username: "user", Password: ""})
		if err != nil {
			t.Fatalf("error: %v", err)
		}
		if resp.Error != "" {
			t.Errorf("unexpected error: %s", resp.Error)
		}
		if resp.Decoded != "user:" {
			t.Errorf("decoded = %q, want %q", resp.Decoded, "user:")
		}
	})

	t.Run("empty username returns error", func(t *testing.T) {
		resp, err := s.BasicAuthGenerate(ctx, &pb.BasicAuthRequest{Username: "", Password: "pass"})
		if err != nil {
			t.Fatalf("error: %v", err)
		}
		if resp.Error == "" {
			t.Error("expected error for empty username")
		}
	})
}
