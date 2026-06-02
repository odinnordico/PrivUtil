package api

import (
	"context"
	"slices"
	"strings"
	"testing"

	pb "github.com/odinnordico/privutil/proto"
)

func TestSpellCheck(t *testing.T) {
	s := NewServer()
	ctx := context.Background()

	t.Run("empty input", func(t *testing.T) {
		resp, err := s.SpellCheck(ctx, &pb.SpellCheckRequest{Text: ""})
		if err != nil {
			t.Fatalf("SpellCheck() error = %v", err)
		}
		if len(resp.Issues) != 0 {
			t.Errorf("expected 0 issues for empty input, got %d", len(resp.Issues))
		}
		if resp.Language != "en" {
			t.Errorf("language = %q, want en", resp.Language)
		}
	})

	t.Run("spelling issue with context and stable id", func(t *testing.T) {
		resp, err := s.SpellCheck(ctx, &pb.SpellCheckRequest{Text: "i have a tset here", Language: "en"})
		if err != nil {
			t.Fatalf("SpellCheck() error = %v", err)
		}
		if resp.WordCount != 5 {
			t.Errorf("word count = %d, want 5", resp.WordCount)
		}
		var found *pb.SpellIssue
		for _, is := range resp.Issues {
			if is.Text == "tset" {
				found = is
			}
		}
		if found == nil {
			t.Fatalf("expected issue for 'tset', got %+v", resp.Issues)
		}
		if found.Id == "" {
			t.Error("issue id should not be empty")
		}
		if !slices.Contains(found.Replacements, "test") {
			t.Errorf("expected 'test' in replacements, got %v", found.Replacements)
		}
		// Context must contain the flagged word at the reported offset.
		ctxRunes := []rune(found.Context)
		span := string(ctxRunes[found.ContextOffset : found.ContextOffset+found.Length])
		if span != "tset" {
			t.Errorf("context offset maps to %q, want 'tset' (context=%q)", span, found.Context)
		}
	})

	t.Run("spanish opening mark", func(t *testing.T) {
		resp, err := s.SpellCheck(ctx, &pb.SpellCheckRequest{Text: "como estas?", Language: "es"})
		if err != nil {
			t.Fatalf("SpellCheck() error = %v", err)
		}
		if resp.Language != "es" {
			t.Errorf("language = %q, want es", resp.Language)
		}
		var found bool
		for _, is := range resp.Issues {
			if is.Rule == "missing-opening-mark" && strings.HasPrefix(is.Replacements[0], "¿") {
				found = true
			}
		}
		if !found {
			t.Errorf("expected missing-opening-mark issue, got %+v", resp.Issues)
		}
	})
}

func TestSpellLanguages(t *testing.T) {
	s := NewServer()
	resp, err := s.SpellLanguages(context.Background(), &pb.SpellLanguagesRequest{})
	if err != nil {
		t.Fatalf("SpellLanguages() error = %v", err)
	}
	if len(resp.Languages) < 2 {
		t.Fatalf("expected at least 2 languages, got %d", len(resp.Languages))
	}
	codes := make([]string, len(resp.Languages))
	for i, l := range resp.Languages {
		codes[i] = l.Code
	}
	for _, want := range []string{"en", "es"} {
		if !slices.Contains(codes, want) {
			t.Errorf("missing language %q in %v", want, codes)
		}
	}
}
