package api

import (
	"context"
	"testing"

	pb "github.com/odinnordico/privutil/proto"
)

func TestTokenCount(t *testing.T) {
	s := NewServer()
	ctx := context.Background()

	t.Run("empty input", func(t *testing.T) {
		resp, err := s.TokenCount(ctx, &pb.TokenCountRequest{Text: ""})
		if err != nil {
			t.Fatalf("TokenCount() error = %v", err)
		}
		if len(resp.Strategies) != 0 {
			t.Errorf("expected 0 strategies for empty input, got %d", len(resp.Strategies))
		}
	})

	t.Run("all strategies returned", func(t *testing.T) {
		resp, err := s.TokenCount(ctx, &pb.TokenCountRequest{Text: "Hello world"})
		if err != nil {
			t.Fatalf("TokenCount() error = %v", err)
		}
		if len(resp.Strategies) != 11 {
			t.Fatalf("expected 11 strategies, got %d", len(resp.Strategies))
		}
		if resp.CharCount != 11 {
			t.Errorf("CharCount = %d, want 11", resp.CharCount)
		}
		if resp.ByteCount != 11 {
			t.Errorf("ByteCount = %d, want 11", resp.ByteCount)
		}
	})

	t.Run("filter by strategy", func(t *testing.T) {
		resp, err := s.TokenCount(ctx, &pb.TokenCountRequest{Text: "Hello world", Strategy: "gpt-4o"})
		if err != nil {
			t.Fatalf("TokenCount() error = %v", err)
		}
		if len(resp.Strategies) != 1 {
			t.Fatalf("expected 1 strategy when filtering, got %d", len(resp.Strategies))
		}
		if resp.Strategies[0].Name != "gpt-4o" {
			t.Errorf("expected strategy gpt-4o, got %s", resp.Strategies[0].Name)
		}
		if !resp.Strategies[0].Exact {
			t.Error("gpt-4o should be exact (BPE)")
		}
		if resp.Strategies[0].Group != "openai" {
			t.Errorf("gpt-4o group = %s, want openai", resp.Strategies[0].Group)
		}
	})

	t.Run("unknown strategy", func(t *testing.T) {
		resp, err := s.TokenCount(ctx, &pb.TokenCountRequest{Text: "test", Strategy: "nonexistent"})
		if err != nil {
			t.Fatalf("TokenCount() error = %v", err)
		}
		if resp.Error == "" {
			t.Error("expected error for unknown strategy")
		}
	})

	t.Run("classic strategies counts", func(t *testing.T) {
		text := "Hello world. How are you? I'm fine!"
		resp, err := s.TokenCount(ctx, &pb.TokenCountRequest{Text: text})
		if err != nil {
			t.Fatalf("TokenCount() error = %v", err)
		}

		byName := make(map[string]*pb.TokenStrategy)
		for _, s := range resp.Strategies {
			byName[s.Name] = s
		}

		if ws := byName["whitespace"]; ws.Count != 7 {
			t.Errorf("whitespace count = %d, want 7", ws.Count)
		}
		if sn := byName["sentence"]; sn.Count != 3 {
			t.Errorf("sentence count = %d, want 3", sn.Count)
		}
	})

	t.Run("heuristic strategies are not exact", func(t *testing.T) {
		resp, err := s.TokenCount(ctx, &pb.TokenCountRequest{Text: "test", Strategy: "claude"})
		if err != nil {
			t.Fatalf("TokenCount() error = %v", err)
		}
		if resp.Strategies[0].Exact {
			t.Error("claude strategy should not be exact")
		}
		if resp.Strategies[0].Group != "anthropic" {
			t.Errorf("claude group = %s, want anthropic", resp.Strategies[0].Group)
		}
	})

	t.Run("BPE strategies return samples", func(t *testing.T) {
		text := "The quick brown fox jumps over the lazy dog."
		resp, err := s.TokenCount(ctx, &pb.TokenCountRequest{Text: text, Strategy: "gpt-4o"})
		if err != nil {
			t.Fatalf("TokenCount() error = %v", err)
		}
		s := resp.Strategies[0]
		if s.Count <= 0 {
			t.Errorf("gpt-4o count = %d, want > 0", s.Count)
		}
		if len(s.Sample) == 0 {
			t.Error("gpt-4o should return token samples")
		}
	})

	t.Run("unicode text", func(t *testing.T) {
		text := "こんにちは世界"
		resp, err := s.TokenCount(ctx, &pb.TokenCountRequest{Text: text})
		if err != nil {
			t.Fatalf("TokenCount() error = %v", err)
		}
		if resp.CharCount != 7 {
			t.Errorf("CharCount = %d, want 7", resp.CharCount)
		}
		if resp.ByteCount != 21 {
			t.Errorf("ByteCount = %d, want 21 (3 bytes per CJK char)", resp.ByteCount)
		}
	})

	t.Run("sample is capped", func(t *testing.T) {
		var sb []byte
		for range 1000 {
			sb = append(sb, "word "...)
		}
		resp, err := s.TokenCount(ctx, &pb.TokenCountRequest{Text: string(sb), Strategy: "whitespace"})
		if err != nil {
			t.Fatalf("TokenCount() error = %v", err)
		}
		if len(resp.Strategies[0].Sample) > 200 {
			t.Errorf("sample length = %d, want <= 200", len(resp.Strategies[0].Sample))
		}
	})
}
