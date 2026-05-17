package api

import (
	"context"
	"fmt"
	"math"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/pkoukk/tiktoken-go"

	pb "github.com/odinnordico/privutil/proto"
)

const maxSampleTokens = 200

type strategyDef struct {
	name     string
	label    string
	group    string
	encoding string
	compute  func(string) (int, []string)
}

func buildStrategies() []strategyDef {
	return []strategyDef{
		// ── OpenAI ──
		{name: "gpt-4o", label: "GPT-4o / GPT-4o-mini", group: "openai", encoding: "o200k_base",
			compute: func(t string) (int, []string) { return bpeTokenize(t, "o200k_base") }},
		{name: "gpt-4", label: "GPT-4 / GPT-4-turbo", group: "openai", encoding: "cl100k_base",
			compute: func(t string) (int, []string) { return bpeTokenize(t, "cl100k_base") }},
		{name: "gpt-3.5", label: "GPT-3.5-turbo", group: "openai", encoding: "cl100k_base",
			compute: func(t string) (int, []string) { return bpeTokenize(t, "cl100k_base") }},
		// ── Anthropic ──
		{name: "claude", label: "Claude 3.5 / 4", group: "anthropic", encoding: "claude-bpe",
			compute: func(t string) (int, []string) { return heuristicTokenize(t, 3.5) }},
		// ── Meta ──
		{name: "llama-3", label: "Llama 3 / 3.1 / 3.2", group: "meta", encoding: "llama-spm",
			compute: func(t string) (int, []string) { return heuristicTokenize(t, 3.7) }},
		// ── Google ──
		{name: "gemini", label: "Gemini 1.5 / 2", group: "google", encoding: "gemini-spm",
			compute: func(t string) (int, []string) { return heuristicTokenize(t, 4.0) }},
		// ── Mistral ──
		{name: "mistral", label: "Mistral / Mixtral", group: "mistral", encoding: "mistral-spm",
			compute: func(t string) (int, []string) { return heuristicTokenize(t, 3.8) }},
		// ── Classic ──
		{name: "whitespace", label: "Whitespace", group: "classic", encoding: "whitespace",
			compute: func(t string) (int, []string) { return whitespaceTokenize(t) }},
		{name: "word", label: "Word", group: "classic", encoding: "word-boundary",
			compute: func(t string) (int, []string) { return wordTokenize(t) }},
		{name: "sentence", label: "Sentence", group: "classic", encoding: "sentence-boundary",
			compute: func(t string) (int, []string) { return sentenceTokenize(t) }},
		{name: "character", label: "Character", group: "classic", encoding: "unicode-rune",
			compute: func(t string) (int, []string) { return characterTokenize(t) }},
	}
}

func (s *Server) TokenCount(ctx context.Context, req *pb.TokenCountRequest) (*pb.TokenCountResponse, error) {
	text := req.Text
	if text == "" {
		return &pb.TokenCountResponse{}, nil
	}

	allDefs := buildStrategies()

	var defs []strategyDef
	if req.Strategy != "" {
		for _, d := range allDefs {
			if d.name == req.Strategy {
				defs = append(defs, d)
				break
			}
		}
		if len(defs) == 0 {
			return &pb.TokenCountResponse{Error: fmt.Sprintf("unknown strategy: %s", req.Strategy)}, nil
		}
	} else {
		defs = allDefs
	}

	strategies := make([]*pb.TokenStrategy, 0, len(defs))
	for _, d := range defs {
		count, sample := d.compute(text)
		exact := isBPEEncoding(d.encoding)
		strategies = append(strategies, &pb.TokenStrategy{
			Name:     d.name,
			Label:    d.label,
			Count:    int32(count), // #nosec G115
			Sample:   sample,
			Exact:    exact,
			Encoding: d.encoding,
			Group:    d.group,
		})
	}

	return &pb.TokenCountResponse{
		Strategies: strategies,
		CharCount:  int32(utf8.RuneCountInString(text)), // #nosec G115
		ByteCount:  int32(len(text)),                    // #nosec G115
	}, nil
}

func isBPEEncoding(enc string) bool {
	return enc == "cl100k_base" || enc == "o200k_base" || enc == "p50k_base" || enc == "r50k_base"
}

func bpeTokenize(text, encoding string) (int, []string) {
	enc, err := tiktoken.GetEncoding(encoding)
	if err != nil {
		count := int(math.Ceil(float64(len(text)) / 4.0))
		return count, nil
	}

	ids := enc.Encode(text, nil, nil)
	sample := make([]string, 0, min(len(ids), maxSampleTokens))
	for i, id := range ids {
		if i >= maxSampleTokens {
			break
		}
		sample = append(sample, enc.Decode([]int{id}))
	}
	return len(ids), sample
}

func heuristicTokenize(text string, charsPerToken float64) (int, []string) {
	count := int(math.Ceil(float64(len(text)) / charsPerToken))
	return count, nil
}

func whitespaceTokenize(text string) (int, []string) {
	tokens := strings.Fields(text)
	return len(tokens), capSample(tokens)
}

var wordRe = regexp.MustCompile(`[\p{L}\p{N}]+(?:[''\-][\p{L}\p{N}]+)*|[^\s\p{L}\p{N}]`)

func wordTokenize(text string) (int, []string) {
	tokens := wordRe.FindAllString(text, -1)
	return len(tokens), capSample(tokens)
}

var sentenceRe = regexp.MustCompile(`[^.!?]*[.!?]+[\s]*|[^.!?]+$`)

func sentenceTokenize(text string) (int, []string) {
	matches := sentenceRe.FindAllString(text, -1)
	var tokens []string
	for _, m := range matches {
		trimmed := strings.TrimSpace(m)
		if trimmed != "" {
			tokens = append(tokens, trimmed)
		}
	}
	if len(tokens) == 0 && strings.TrimSpace(text) != "" {
		tokens = []string{strings.TrimSpace(text)}
	}
	return len(tokens), capSample(tokens)
}

func characterTokenize(text string) (int, []string) {
	runes := []rune(text)
	sample := make([]string, 0, min(len(runes), maxSampleTokens))
	for i, r := range runes {
		if i >= maxSampleTokens {
			break
		}
		if unicode.IsSpace(r) {
			sample = append(sample, fmt.Sprintf("[U+%04X]", r))
		} else {
			sample = append(sample, string(r))
		}
	}
	return len(runes), sample
}

func capSample(tokens []string) []string {
	if len(tokens) <= maxSampleTokens {
		return tokens
	}
	return tokens[:maxSampleTokens]
}
