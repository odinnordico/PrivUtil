package api

import (
	"context"
	"strings"
	"testing"

	pb "github.com/odinnordico/privutil/proto"
)

// TestMathEvalDepthLimit ensures deeply nested expressions are rejected with an
// error instead of overflowing the goroutine stack (an unrecoverable throw).
func TestMathEvalDepthLimit(t *testing.T) {
	s := NewServer()
	deep := strings.Repeat("(", 5000) + "1" + strings.Repeat(")", 5000)

	resp, err := s.MathEval(context.Background(), &pb.MathEvalRequest{Expression: deep})
	if err != nil {
		t.Fatalf("MathEval() returned transport error = %v", err)
	}
	if resp.Error == "" {
		t.Fatal("MathEval() expected an error for over-nested expression, got none")
	}

	// A modestly nested expression well under the limit must still evaluate.
	ok := strings.Repeat("(", 10) + "2+3" + strings.Repeat(")", 10)
	resp, err = s.MathEval(context.Background(), &pb.MathEvalRequest{Expression: ok})
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
