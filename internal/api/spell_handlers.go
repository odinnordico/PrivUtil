package api

import (
	"context"
	"fmt"

	"github.com/odinnordico/privutil/internal/spellcheck"
	pb "github.com/odinnordico/privutil/proto"
)

// contextRadius is how many runes of surrounding text accompany each issue so
// the UI can show a snippet in the suggestions panel.
const contextRadius = 32

// maxTextRunes bounds the input size to keep edit-distance suggestion work
// from becoming pathological on very large pastes.
const maxTextRunes = 100_000

// SpellCheck runs the offline spell- and grammar-checking engine over the
// supplied text and returns located issues with suggested corrections.
func (s *Server) SpellCheck(ctx context.Context, req *pb.SpellCheckRequest) (*pb.SpellCheckResponse, error) {
	if req.Text == "" {
		return &pb.SpellCheckResponse{Language: spellcheck.DefaultLanguage}, nil
	}
	runes := []rune(req.Text)
	if len(runes) > maxTextRunes {
		return &pb.SpellCheckResponse{
			Language: spellcheck.DefaultLanguage,
			Error:    fmt.Sprintf("text too large: %d characters (limit %d)", len(runes), maxTextRunes),
		}, nil
	}

	issues, lang := spellcheck.Check(req.Text, req.Language)

	out := make([]*pb.SpellIssue, 0, len(issues))
	for _, is := range issues {
		ctxText, ctxOff := buildContext(runes, is.Offset, is.Length)
		out = append(out, &pb.SpellIssue{
			Id:            fmt.Sprintf("%s-%d-%d", is.Rule, is.Offset, is.Length),
			Offset:        int32(is.Offset), // #nosec G115
			Length:        int32(is.Length), // #nosec G115
			Text:          is.Text,
			Type:          is.Type,
			Rule:          is.Rule,
			Message:       is.Message,
			Replacements:  is.Replacements,
			Context:       ctxText,
			ContextOffset: int32(ctxOff), // #nosec G115
		})
	}

	return &pb.SpellCheckResponse{
		Issues:    out,
		Language:  lang,
		WordCount: int32(spellcheck.WordCount(req.Text)), // #nosec G115
	}, nil
}

// SpellLanguages reports the languages supported by the offline engine.
func (s *Server) SpellLanguages(ctx context.Context, _ *pb.SpellLanguagesRequest) (*pb.SpellLanguagesResponse, error) {
	infos := spellcheck.Languages()
	langs := make([]*pb.SpellLanguage, 0, len(infos))
	for _, in := range infos {
		langs = append(langs, &pb.SpellLanguage{Code: in.Code, Label: in.Label})
	}
	return &pb.SpellLanguagesResponse{Languages: langs}, nil
}

// buildContext returns a snippet of runes surrounding [offset,offset+length)
// and the rune offset of the flagged span within that snippet.
func buildContext(runes []rune, offset, length int) (string, int) {
	start := max(offset-contextRadius, 0)
	end := min(offset+length+contextRadius, len(runes))
	prefix, suffix := "", ""
	if start > 0 {
		prefix = "…"
	}
	if end < len(runes) {
		suffix = "…"
	}
	return prefix + string(runes[start:end]) + suffix, len([]rune(prefix)) + (offset - start)
}
