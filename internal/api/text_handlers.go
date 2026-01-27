package api

import (
	"context"
	"fmt"
	"html"
	"regexp"
	"sort"
	"strings"
	"unicode"

	"github.com/sergi/go-diff/diffmatchpatch"

	pb "github.com/odinnordico/privutil/proto"
)

func (s *Server) Diff(ctx context.Context, req *pb.DiffRequest) (*pb.DiffResponse, error) {
	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(req.Text1, req.Text2, false)

	var buffer strings.Builder
	for _, diff := range diffs {
		escapedText := html.EscapeString(diff.Text)

		switch diff.Type {
		case diffmatchpatch.DiffInsert:
			buffer.WriteString("<ins style='background:#ccffd8; color:#004d0d; text-decoration:none; padding:1px 2px; border-radius:2px;'>")
			buffer.WriteString(escapedText)
			buffer.WriteString("</ins>")
		case diffmatchpatch.DiffDelete:
			buffer.WriteString("<del style='background:#ffd7d5; color:#991b1b; text-decoration:line-through; padding:1px 2px; border-radius:2px;'>")
			buffer.WriteString(escapedText)
			buffer.WriteString("</del>")
		case diffmatchpatch.DiffEqual:
			buffer.WriteString("<span>")
			buffer.WriteString(escapedText)
			buffer.WriteString("</span>")
		}
	}

	return &pb.DiffResponse{
		DiffHtml: fmt.Sprintf("<div class='diff-output' style='white-space: pre-wrap; font-family: monospace;'>%s</div>", buffer.String()),
	}, nil
}

func (s *Server) TextInspect(ctx context.Context, req *pb.TextInspectRequest) (*pb.TextInspectResponse, error) {
	text := req.Text
	return &pb.TextInspectResponse{
		CharCount: int32(len([]rune(text))),              // #nosec G115
		WordCount: int32(len(strings.Fields(text))),      // #nosec G115
		LineCount: int32(len(strings.Split(text, "\n"))), // #nosec G115
		ByteCount: int32(len(text)),                      // #nosec G115
	}, nil
}

func (s *Server) TextManipulate(ctx context.Context, req *pb.TextManipulateRequest) (*pb.TextManipulateResponse, error) {
	text := req.Text
	lines := strings.Split(text, "\n")
	var result string

	switch req.Action {
	case pb.TextAction_SORT_AZ:
		sort.Strings(lines)
		result = strings.Join(lines, "\n")
	case pb.TextAction_SORT_ZA:
		sort.Sort(sort.Reverse(sort.StringSlice(lines)))
		result = strings.Join(lines, "\n")
	case pb.TextAction_REVERSE:
		for i, j := 0, len(lines)-1; i < j; i, j = i+1, j-1 {
			lines[i], lines[j] = lines[j], lines[i]
		}
		result = strings.Join(lines, "\n")
	case pb.TextAction_DEDUPE:
		seen := make(map[string]bool)
		var unique []string
		for _, line := range lines {
			if !seen[line] {
				seen[line] = true
				unique = append(unique, line)
			}
		}
		result = strings.Join(unique, "\n")
	case pb.TextAction_REMOVE_EMPTY:
		var nonEmpty []string
		for _, line := range lines {
			if strings.TrimSpace(line) != "" {
				nonEmpty = append(nonEmpty, line)
			}
		}
		result = strings.Join(nonEmpty, "\n")
	case pb.TextAction_TRIM:
		var trimmed []string
		for _, line := range lines {
			trimmed = append(trimmed, strings.TrimSpace(line))
		}
		result = strings.Join(trimmed, "\n")
	default:
		result = text
	}

	return &pb.TextManipulateResponse{Text: result}, nil
}

func (s *Server) TextSimilarity(ctx context.Context, req *pb.SimilarityRequest) (*pb.SimilarityResponse, error) {
	// Simple Levenshtein implementation
	s1, s2 := req.Text1, req.Text2
	r1, r2 := []rune(s1), []rune(s2)
	n, m := len(r1), len(r2)
	if n > m {
		r1, r2 = r2, r1
		n, m = m, n
	}

	currentRow := make([]int, n+1)
	for i := 0; i <= n; i++ {
		currentRow[i] = i
	}

	for i := 1; i <= m; i++ {
		previousRow := currentRow
		currentRow = make([]int, n+1)
		currentRow[0] = i
		for j := 1; j <= n; j++ {
			add, del, change := previousRow[j]+1, currentRow[j-1]+1, previousRow[j-1]
			if r1[j-1] != r2[i-1] {
				change++
			}
			currentRow[j] = minInt(add, minInt(del, change))
		}
	}
	dist := currentRow[n]

	// Calculate similarity 0.0 - 1.0
	maxLen := maxInt(n, m)
	var sim float32
	if maxLen == 0 {
		sim = 1.0
	} else {
		sim = 1.0 - float32(dist)/float32(maxLen)
	}

	return &pb.SimilarityResponse{
		Distance:   int32(dist), // #nosec G115
		Similarity: sim,
	}, nil
}

func (s *Server) RegexTest(ctx context.Context, req *pb.RegexRequest) (*pb.RegexResponse, error) {
	re, err := regexp.Compile(req.Pattern)
	if err != nil {
		return &pb.RegexResponse{Error: fmt.Sprintf("Invalid Pattern: %v", err)}, nil
	}

	matches := re.FindAllString(req.Text, -1)
	return &pb.RegexResponse{
		Match:   len(matches) > 0,
		Matches: matches,
	}, nil
}

func (s *Server) CaseConvert(ctx context.Context, req *pb.CaseRequest) (*pb.CaseResponse, error) {
	text := req.Text
	words := splitIntoWords(text)

	toCamel := func(words []string) string {
		if len(words) == 0 {
			return ""
		}
		res := strings.ToLower(words[0])
		for _, w := range words[1:] {
			if len(w) > 0 {
				res += strings.ToUpper(w[:1]) + strings.ToLower(w[1:])
			}
		}
		return res
	}

	toPascal := func(words []string) string {
		var res string
		for _, w := range words {
			if len(w) > 0 {
				res += strings.ToUpper(w[:1]) + strings.ToLower(w[1:])
			}
		}
		return res
	}

	toSnake := func(words []string) string {
		var lower []string
		for _, w := range words {
			lower = append(lower, strings.ToLower(w))
		}
		return strings.Join(lower, "_")
	}

	toKebab := func(words []string) string {
		var lower []string
		for _, w := range words {
			lower = append(lower, strings.ToLower(w))
		}
		return strings.Join(lower, "-")
	}

	toConstant := func(words []string) string {
		var upper []string
		for _, w := range words {
			upper = append(upper, strings.ToUpper(w))
		}
		return strings.Join(upper, "_")
	}

	return &pb.CaseResponse{
		Camel:    toCamel(words),
		Pascal:   toPascal(words),
		Snake:    toSnake(words),
		Kebab:    toKebab(words),
		Constant: toConstant(words),
		Title:    strings.Join(words, " "),
	}, nil
}

func splitIntoWords(s string) []string {
	s = strings.ReplaceAll(s, "-", " ")
	s = strings.ReplaceAll(s, "_", " ")
	s = strings.ReplaceAll(s, ".", " ")

	var sb strings.Builder
	runes := []rune(s)
	for i, r := range runes {
		if i > 0 && i < len(runes)-1 {
			prev := runes[i-1]
			next := runes[i+1]
			if unicode.IsLower(prev) && unicode.IsUpper(r) {
				sb.WriteRune(' ')
			}
			if unicode.IsUpper(prev) && unicode.IsUpper(r) && unicode.IsLower(next) {
				sb.WriteRune(' ')
			}
		}
		sb.WriteRune(r)
	}
	s = sb.String()

	return strings.Fields(s)
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
