package api

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"html"
	"regexp"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/sergi/go-diff/diffmatchpatch"

	pb "github.com/odinnordico/privutil/proto"
)

// maxDiffFileBytes caps each uploaded file for the file-diff comparison. Text
// files are diffed (diffmatchpatch has its own 1s timeout); binary files are
// only checksummed, so this bounds the work either way.
const maxDiffFileBytes = 10 << 20 // 10 MiB

func (s *Server) Diff(ctx context.Context, req *pb.DiffRequest) (*pb.DiffResponse, error) {
	return &pb.DiffResponse{DiffHtml: buildDiffHTML(req.Text1, req.Text2)}, nil
}

// buildDiffHTML renders an inline (unified) diff of two texts as HTML. Every
// segment is HTML-escaped, so user content can never inject markup.
func buildDiffHTML(text1, text2 string) string {
	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(text1, text2, false)

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

	return fmt.Sprintf("<div class='diff-output' style='white-space: pre-wrap; font-family: monospace;'>%s</div>", buffer.String())
}

// isReadableText reports whether data is safe to treat as text: valid UTF-8 with
// no NUL byte (the standard binary-vs-text heuristic).
func isReadableText(data []byte) bool {
	return utf8.Valid(data) && !bytes.Contains(data, []byte{0})
}

// DiffFiles compares two uploaded files. When both are readable text it returns
// an inline diff; otherwise it reports that a file is not readable and compares
// the files by SHA-256 checksum.
func (s *Server) DiffFiles(_ context.Context, req *pb.DiffFilesRequest) (*pb.DiffFilesResponse, error) {
	f1, f2 := req.File1, req.File2
	if len(f1) > maxDiffFileBytes || len(f2) > maxDiffFileBytes {
		return &pb.DiffFilesResponse{
			Error: fmt.Sprintf("file too large: limit %d bytes (%d MiB) per file", maxDiffFileBytes, maxDiffFileBytes>>20),
		}, nil
	}

	sum1 := sha256.Sum256(f1)
	sum2 := sha256.Sum256(f2)
	cs1 := hex.EncodeToString(sum1[:])
	cs2 := hex.EncodeToString(sum2[:])

	resp := &pb.DiffFilesResponse{
		Checksum1:      cs1,
		Checksum2:      cs2,
		ChecksumsMatch: cs1 == cs2,
		ChecksumAlgo:   "SHA-256",
	}

	if isReadableText(f1) && isReadableText(f2) {
		resp.IsText = true
		resp.DiffHtml = buildDiffHTML(string(f1), string(f2))
	} else {
		resp.Message = "One or both files are not readable as text (binary). Compared by SHA-256 checksum instead."
	}
	return resp, nil
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

// maxSimilarityRunes bounds each input, and maxSimilarityCells bounds their
// product, to the quadratic Levenshtein matrix so two large pastes cannot pin a
// CPU core (O(n·m) with no timeout otherwise). The cell cap is the tighter of
// the two: two 100k inputs would be 10^10 cells (~10s) under the per-input
// limit alone, so the product is capped to ~5·10^7 cells (sub-100ms).
const (
	maxSimilarityRunes = 100_000
	maxSimilarityCells = 50_000_000
)

func (s *Server) TextSimilarity(ctx context.Context, req *pb.SimilarityRequest) (*pb.SimilarityResponse, error) {
	// Simple Levenshtein implementation
	s1, s2 := req.Text1, req.Text2
	r1, r2 := []rune(s1), []rune(s2)
	if len(r1) > maxSimilarityRunes || len(r2) > maxSimilarityRunes {
		return &pb.SimilarityResponse{
			Error: fmt.Sprintf("input too large: limit %d characters per text", maxSimilarityRunes),
		}, nil
	}
	n, m := len(r1), len(r2)
	if n > m {
		r1, r2 = r2, r1
		n, m = m, n
	}
	if n*m > maxSimilarityCells {
		return &pb.SimilarityResponse{
			Error: fmt.Sprintf("inputs too large to compare: %d cells exceeds limit %d", n*m, maxSimilarityCells),
		}, nil
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
			currentRow[j] = min(add, del, change)
		}
	}
	dist := currentRow[n]

	// Calculate similarity 0.0 - 1.0
	maxLen := max(n, m)
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
		var res strings.Builder
		res.WriteString(strings.ToLower(words[0]))
		for _, w := range words[1:] {
			if len(w) > 0 {
				rw := []rune(w)
				res.WriteString(strings.ToUpper(string(rw[:1])) + strings.ToLower(string(rw[1:])))
			}
		}
		return res.String()
	}

	toPascal := func(words []string) string {
		var res strings.Builder
		for _, w := range words {
			if len(w) > 0 {
				rw := []rune(w)
				res.WriteString(strings.ToUpper(string(rw[:1])) + strings.ToLower(string(rw[1:])))
			}
		}
		return res.String()
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
