package api

import (
	"context"
	"strings"
	"testing"

	pb "github.com/odinnordico/privutil/proto"
)

// TestMathEvalDepthLimit ensures deeply nested expressions are rejected with an
// error instead of overflowing the goroutine stack (an unrecoverable throw).
// It covers all three recursion cycles — parentheses, prefix unary chains, and
// right-associative power chains — each of which must hit the depth guard.
func TestMathEvalDepthLimit(t *testing.T) {
	s := NewServer()

	// Each case stays under the 10k expression-length cap but far exceeds the
	// 256 recursion-depth cap. Before the fix, only the parenthesis case was
	// guarded; the unary and power chains overflowed the stack.
	cases := map[string]string{
		"parens": strings.Repeat("(", 1000) + "1" + strings.Repeat(")", 1000),
		"unary":  strings.Repeat("-", 1000) + "1",
		"power":  "1" + strings.Repeat("^1", 1000),
	}
	for name, expr := range cases {
		resp, err := s.MathEval(context.Background(), &pb.MathEvalRequest{Expression: expr})
		if err != nil {
			t.Fatalf("%s: MathEval() returned transport error = %v", name, err)
		}
		if resp.Error == "" {
			t.Fatalf("%s: MathEval() expected an error for over-deep expression, got none", name)
		}
	}

	// A modestly nested expression well under the limit must still evaluate.
	ok := strings.Repeat("(", 10) + "2+3" + strings.Repeat(")", 10)
	resp, err := s.MathEval(context.Background(), &pb.MathEvalRequest{Expression: ok})
	if err != nil {
		t.Fatalf("MathEval() error = %v", err)
	}
	if resp.Error != "" {
		t.Fatalf("MathEval() unexpected error for valid nesting: %s", resp.Error)
	}
	if resp.RawValue != 5 {
		t.Errorf("MathEval() = %v, want 5", resp.RawValue)
	}
}

// TestGenerateLoremCountCap ensures the Lorem generator caps count so it cannot
// be driven into unbounded allocation.
func TestGenerateLoremCountCap(t *testing.T) {
	s := NewServer()
	resp, err := s.GenerateLorem(context.Background(), &pb.LoremRequest{Type: "word", Count: 1_000_000})
	if err != nil {
		t.Fatalf("GenerateLorem() error = %v", err)
	}
	words := strings.Fields(resp.Text)
	if len(words) > 100 {
		t.Errorf("GenerateLorem() produced %d words, want <= 100", len(words))
	}
}

// TestBcryptCostCapped ensures a hostile high cost is clamped rather than
// accepted (bcrypt work grows as 2^cost).
func TestBcryptCostCapped(t *testing.T) {
	s := NewServer()
	cost := int32(31)
	resp, err := s.CalculateHash(context.Background(), &pb.HashRequest{
		Text: "hello", Algo: "bcrypt", Cost: &cost,
	})
	if err != nil {
		t.Fatalf("CalculateHash() error = %v", err)
	}
	// bcrypt hash string encodes the cost as the second $-field: $2a$<cost>$...
	if !strings.HasPrefix(resp.Hash, "$2a$10$") {
		t.Errorf("bcrypt cost not clamped to default: hash prefix = %q", resp.Hash[:7])
	}
}

// TestTextSimilarityInputLimit ensures oversized inputs are rejected before the
// quadratic Levenshtein computation runs.
func TestTextSimilarityInputLimit(t *testing.T) {
	s := NewServer()
	big := strings.Repeat("a", maxSimilarityRunes+1)
	resp, err := s.TextSimilarity(context.Background(), &pb.SimilarityRequest{Text1: big, Text2: "b"})
	if err != nil {
		t.Fatalf("TextSimilarity() error = %v", err)
	}
	if resp.Error == "" {
		t.Error("TextSimilarity() expected error for oversized input, got none")
	}
}

// TestOtpValidate8Digit ensures an 8-digit TOTP code validates — previously
// OtpValidate hard-coded 6 digits, so 8-digit codes always failed.
func TestOtpValidate8Digit(t *testing.T) {
	s := NewServer()
	ctx := context.Background()

	gen, err := s.OtpGenerate(ctx, &pb.OtpRequest{
		Type: "totp", Algo: "sha1", Digits: 8, Period: 30, GenerateSecret: true,
	})
	if err != nil {
		t.Fatalf("OtpGenerate() error = %v", err)
	}
	if gen.Error != "" {
		t.Fatalf("OtpGenerate() returned error: %s", gen.Error)
	}
	if len(gen.Code) != 8 {
		t.Fatalf("expected an 8-digit code, got %q", gen.Code)
	}

	val, err := s.OtpValidate(ctx, &pb.OtpValidateRequest{
		Secret: gen.Secret, Code: gen.Code, Algo: "sha1", Period: 30, Window: 1, Digits: 8,
	})
	if err != nil {
		t.Fatalf("OtpValidate() error = %v", err)
	}
	if !val.Valid {
		t.Error("OtpValidate() rejected a freshly generated 8-digit code")
	}
}

// TestBaseConvertInputLimit ensures oversized input is rejected before the
// superlinear big.Int / O(n^2) base64 conversion runs.
func TestBaseConvertInputLimit(t *testing.T) {
	s := NewServer()
	resp, err := s.BaseConvert(context.Background(), &pb.BaseConvertRequest{
		Input: strings.Repeat("9", 10_001), SourceBase: 10,
	})
	if err != nil {
		t.Fatalf("BaseConvert() error = %v", err)
	}
	if resp.Error == "" {
		t.Error("BaseConvert() expected error for oversized input, got none")
	}
}

// TestLeapYearListCap ensures the comma-separated form is bounded like the range
// form (both capped at 400 entries).
func TestLeapYearListCap(t *testing.T) {
	s := NewServer()
	years := make([]string, 500)
	for i := range years {
		years[i] = "2000"
	}
	resp, err := s.LeapYear(context.Background(), &pb.LeapYearRequest{Input: strings.Join(years, ",")})
	if err != nil {
		t.Fatalf("LeapYear() error = %v", err)
	}
	if resp.Error == "" {
		t.Error("LeapYear() expected error for over-long list, got none")
	}
}

// TestCaseConvertUnicode ensures the first character is treated as a rune, not a
// byte, so a leading multi-byte character is not corrupted.
func TestCaseConvertUnicode(t *testing.T) {
	s := NewServer()
	resp, err := s.CaseConvert(context.Background(), &pb.CaseRequest{Text: "über cafe"})
	if err != nil {
		t.Fatalf("CaseConvert() error = %v", err)
	}
	if resp.Pascal != "ÜberCafe" {
		t.Errorf("Pascal = %q, want %q", resp.Pascal, "ÜberCafe")
	}
	if resp.Camel != "überCafe" {
		t.Errorf("Camel = %q, want %q", resp.Camel, "überCafe")
	}
}

// TestColorConvertRejectsOutOfRange ensures RGB components above 255 or
// unparseable values are rejected rather than silently truncated.
func TestColorConvertRejectsOutOfRange(t *testing.T) {
	s := NewServer()
	for _, in := range []string{"rgb(999,0,0)", "rgb(256,0,0)"} {
		resp, err := s.ColorConvert(context.Background(), &pb.ColorRequest{Input: in})
		if err != nil {
			t.Fatalf("ColorConvert(%q) error = %v", in, err)
		}
		if resp.Error == "" {
			t.Errorf("ColorConvert(%q) expected error, got none", in)
		}
	}

	// A valid color still converts cleanly.
	resp, err := s.ColorConvert(context.Background(), &pb.ColorRequest{Input: "rgb(255,0,0)"})
	if err != nil {
		t.Fatalf("ColorConvert() error = %v", err)
	}
	if resp.Error != "" {
		t.Errorf("ColorConvert() unexpected error for valid input: %s", resp.Error)
	}
}
