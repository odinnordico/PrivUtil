package api

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"

	pb "github.com/odinnordico/privutil/proto"
)

// ─── Slugify ──────────────────────────────────────────────────────────────────

var multiSepRe = regexp.MustCompile(`-{2,}`)

func (s *Server) Slugify(_ context.Context, req *pb.SlugifyRequest) (*pb.SlugifyResponse, error) {
	text := strings.TrimSpace(req.Text)
	if text == "" {
		return &pb.SlugifyResponse{Error: "input is required"}, nil
	}

	sep := req.Separator
	if sep == "" {
		sep = "-"
	}
	if sep == "none" {
		sep = ""
	}

	// NFD-normalize then strip combining marks (diacritics)
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)))
	normalized, _, _ := transform.String(t, text)

	if !req.Uppercase {
		normalized = strings.ToLower(normalized)
	} else {
		normalized = strings.ToUpper(normalized)
	}

	// Replace non-alphanumeric runs with separator
	nonAlnum := regexp.MustCompile(`[^a-zA-Z0-9]+`)
	result := nonAlnum.ReplaceAllString(normalized, sep)

	if sep != "" {
		// Collapse multiple separators
		escapedSep := regexp.QuoteMeta(sep)
		multiRe := regexp.MustCompile(escapedSep + `{2,}`)
		result = multiRe.ReplaceAllString(result, sep)
		// Trim leading/trailing separators
		result = strings.Trim(result, sep)
	}

	if req.MaxLen > 0 && int32(len(result)) > req.MaxLen {
		result = result[:req.MaxLen]
		if sep != "" {
			result = strings.TrimRight(result, sep)
		}
	}

	return &pb.SlugifyResponse{Result: result}, nil
}

// ─── Hidden character detector ────────────────────────────────────────────────

var hiddenCharNames = map[rune]string{
	' ': "NO-BREAK SPACE",
	'­': "SOFT HYPHEN",
	'͏': "COMBINING GRAPHEME JOINER",
	'؜': "ARABIC LETTER MARK",
	'᠎': "MONGOLIAN VOWEL SEPARATOR",
	'​': "ZERO WIDTH SPACE",
	'‌': "ZERO WIDTH NON-JOINER",
	'‍': "ZERO WIDTH JOINER",
	'‎': "LEFT-TO-RIGHT MARK",
	'‏': "RIGHT-TO-LEFT MARK",
	'‪': "LEFT-TO-RIGHT EMBEDDING",
	'‫': "RIGHT-TO-LEFT EMBEDDING",
	'‬': "POP DIRECTIONAL FORMATTING",
	'‭': "LEFT-TO-RIGHT OVERRIDE",
	'‮': "RIGHT-TO-LEFT OVERRIDE",
	' ': "NARROW NO-BREAK SPACE",
	'⁠': "WORD JOINER",
	'⁡': "FUNCTION APPLICATION",
	'⁢': "INVISIBLE TIMES",
	'⁣': "INVISIBLE SEPARATOR",
	'⁤': "INVISIBLE PLUS",
	'⁪': "INHIBIT SYMMETRIC SWAPPING",
	'⁫': "ACTIVATE SYMMETRIC SWAPPING",
	'⁬': "INHIBIT ARABIC FORM SHAPING",
	'⁭': "ACTIVATE ARABIC FORM SHAPING",
	'⁮': "NATIONAL DIGIT SHAPES",
	'⁯': "NOMINAL DIGIT SHAPES",
	'　': "IDEOGRAPHIC SPACE",
	'\uFEFF': "ZERO WIDTH NO-BREAK SPACE (BOM)",
	'￹': "INTERLINEAR ANNOTATION ANCHOR",
	'￺': "INTERLINEAR ANNOTATION SEPARATOR",
	'￻': "INTERLINEAR ANNOTATION TERMINATOR",
}

func (s *Server) HiddenChars(_ context.Context, req *pb.HiddenCharsRequest) (*pb.HiddenCharsResponse, error) {
	text := req.Text
	if text == "" {
		return &pb.HiddenCharsResponse{}, nil
	}

	counts := make(map[rune]int)
	for _, r := range text {
		if _, known := hiddenCharNames[r]; known {
			counts[r]++
		}
	}

	if len(counts) == 0 {
		return &pb.HiddenCharsResponse{HasHidden: false, Cleaned: text}, nil
	}

	// Build sorted char info list
	infos := make([]*pb.HiddenCharInfo, 0, len(counts))
	for r, cnt := range counts {
		infos = append(infos, &pb.HiddenCharInfo{
			Name:      hiddenCharNames[r],
			Codepoint: fmt.Sprintf("U+%04X", r),
			Count:     int32(cnt), // #nosec G115
		})
	}
	sort.Slice(infos, func(i, j int) bool {
		return infos[i].Codepoint < infos[j].Codepoint
	})

	// Annotated: replace hidden chars with [U+XXXX] markers
	var annotated strings.Builder
	for _, r := range text {
		if _, known := hiddenCharNames[r]; known {
			annotated.WriteString(fmt.Sprintf("[U+%04X]", r))
		} else {
			annotated.WriteRune(r)
		}
	}

	// Cleaned: strip all hidden chars
	var cleaned strings.Builder
	for _, r := range text {
		if _, known := hiddenCharNames[r]; !known {
			cleaned.WriteRune(r)
		}
	}

	return &pb.HiddenCharsResponse{
		HasHidden: true,
		Chars:     infos,
		Annotated: annotated.String(),
		Cleaned:   cleaned.String(),
	}, nil
}

// ─── Text replacer ────────────────────────────────────────────────────────────

func (s *Server) TextReplace(_ context.Context, req *pb.TextReplaceRequest) (*pb.TextReplaceResponse, error) {
	if req.Find == "" {
		return &pb.TextReplaceResponse{Error: "find pattern is required"}, nil
	}

	var re *regexp.Regexp
	var err error

	if req.UseRegex {
		pattern := req.Find
		if req.CaseInsensitive {
			pattern = "(?i)" + pattern
		}
		re, err = regexp.Compile(pattern)
		if err != nil {
			return &pb.TextReplaceResponse{Error: "invalid regex: " + err.Error()}, nil
		}
	}

	count := 0
	var result string

	if req.UseRegex {
		// Count matches then replace
		matches := re.FindAllString(req.Text, -1)
		count = len(matches)
		result = re.ReplaceAllString(req.Text, req.ReplaceWith)
	} else {
		find := req.Find
		src := req.Text
		if req.CaseInsensitive {
			count = strings.Count(strings.ToLower(src), strings.ToLower(find))
			// Case-insensitive literal replacement
			pattern := "(?i)" + regexp.QuoteMeta(find)
			ciRe := regexp.MustCompile(pattern)
			result = ciRe.ReplaceAllString(src, req.ReplaceWith)
		} else {
			count = strings.Count(src, find)
			result = strings.ReplaceAll(src, find, req.ReplaceWith)
		}
	}

	return &pb.TextReplaceResponse{
		Result: result,
		Count:  int32(count), // #nosec G115
	}, nil
}

// ─── String obfuscator ────────────────────────────────────────────────────────

func (s *Server) StringObfuscate(_ context.Context, req *pb.StringObfuscateRequest) (*pb.StringObfuscateResponse, error) {
	text := req.Text
	if text == "" {
		return &pb.StringObfuscateResponse{Error: "input is required"}, nil
	}

	maskChar := req.MaskChar
	if maskChar == "" {
		maskChar = "*"
	}
	// Use only first rune of mask char
	maskRunes := []rune(maskChar)
	mask := string(maskRunes[0])

	runes_ := []rune(text)
	n := len(runes_)

	keepStart := int(req.KeepStart)
	keepEnd := int(req.KeepEnd)
	if keepStart < 0 {
		keepStart = 0
	}
	if keepEnd < 0 {
		keepEnd = 0
	}

	// If total visible >= length, just return the text
	if keepStart+keepEnd >= n {
		return &pb.StringObfuscateResponse{Result: text}, nil
	}

	masked := n - keepStart - keepEnd
	var sb strings.Builder
	sb.WriteString(string(runes_[:keepStart]))
	sb.WriteString(strings.Repeat(mask, masked))
	sb.WriteString(string(runes_[n-keepEnd:]))

	return &pb.StringObfuscateResponse{Result: sb.String()}, nil
}

// ─── Numeronym generator ──────────────────────────────────────────────────────

func (s *Server) NumeronymGenerate(_ context.Context, req *pb.NumeronymRequest) (*pb.NumeronymResponse, error) {
	text := strings.TrimSpace(req.Text)
	if text == "" {
		return &pb.NumeronymResponse{Error: "input is required"}, nil
	}

	rawWords := strings.Fields(text)
	results := make([]string, len(rawWords))
	for i, w := range rawWords {
		results[i] = numeronym(w)
	}

	return &pb.NumeronymResponse{
		Words:  results,
		Result: strings.Join(results, " "),
	}, nil
}

func numeronym(word string) string {
	runes_ := []rune(word)
	n := len(runes_)
	if n <= 3 {
		return word
	}
	middle := n - 2
	return string(runes_[0]) + strconv.Itoa(middle) + string(runes_[n-1])
}

// ─── NATO alphabet ────────────────────────────────────────────────────────────

var natoEncode = map[rune]string{
	'A': "Alfa", 'B': "Bravo", 'C': "Charlie", 'D': "Delta", 'E': "Echo",
	'F': "Foxtrot", 'G': "Golf", 'H': "Hotel", 'I': "India", 'J': "Juliett",
	'K': "Kilo", 'L': "Lima", 'M': "Mike", 'N': "November", 'O': "Oscar",
	'P': "Papa", 'Q': "Quebec", 'R': "Romeo", 'S': "Sierra", 'T': "Tango",
	'U': "Uniform", 'V': "Victor", 'W': "Whiskey", 'X': "X-ray",
	'Y': "Yankee", 'Z': "Zulu",
	'0': "Zero", '1': "One", '2': "Two", '3': "Three", '4': "Four",
	'5': "Five", '6': "Six", '7': "Seven", '8': "Eight", '9': "Nine",
	' ': "(Space)",
}

var natoDecode map[string]rune

func init() {
	natoDecode = make(map[string]rune, len(natoEncode))
	for k, v := range natoEncode {
		natoDecode[strings.ToLower(v)] = k
	}
}

func (s *Server) NatoAlphabet(_ context.Context, req *pb.NatoRequest) (*pb.NatoResponse, error) {
	if req.Text == "" {
		return &pb.NatoResponse{Error: "input is required"}, nil
	}

	if req.Action == "decode" {
		result, err := natoToText(req.Text)
		if err != nil {
			return &pb.NatoResponse{Error: err.Error()}, nil
		}
		return &pb.NatoResponse{Result: result}, nil
	}

	result, err := textToNato(req.Text)
	if err != nil {
		return &pb.NatoResponse{Error: err.Error()}, nil
	}
	return &pb.NatoResponse{Result: result}, nil
}

func textToNato(text string) (string, error) {
	var parts []string
	for _, r := range strings.ToUpper(text) {
		if word, ok := natoEncode[r]; ok {
			parts = append(parts, word)
		} else {
			return "", fmt.Errorf("no NATO word for character %q", r)
		}
	}
	return strings.Join(parts, " "), nil
}

func natoToText(nato string) (string, error) {
	var sb strings.Builder
	for _, word := range strings.Fields(nato) {
		r, ok := natoDecode[strings.ToLower(word)]
		if !ok {
			return "", fmt.Errorf("unknown NATO word %q", word)
		}
		if r == ' ' {
			sb.WriteRune(' ')
		} else {
			sb.WriteRune(r)
		}
	}
	return sb.String(), nil
}

// ─── List processor ───────────────────────────────────────────────────────────

func (s *Server) ListProcess(_ context.Context, req *pb.ListRequest) (*pb.ListResponse, error) {
	if strings.TrimSpace(req.Text) == "" {
		return &pb.ListResponse{Error: "input is required"}, nil
	}

	lines := strings.Split(req.Text, "\n")
	inputCount := int32(len(lines)) // #nosec G115

	keyOf := func(line string) string {
		if req.CaseInsensitive {
			return strings.ToLower(line)
		}
		return line
	}

	switch req.Action {
	case pb.ListAction_LIST_SORT_AZ:
		sorted := make([]string, len(lines))
		copy(sorted, lines)
		sort.SliceStable(sorted, func(i, j int) bool {
			return keyOf(sorted[i]) < keyOf(sorted[j])
		})
		return listResp(strings.Join(sorted, "\n"), inputCount, int32(len(sorted)), nil), nil // #nosec G115

	case pb.ListAction_LIST_SORT_ZA:
		sorted := make([]string, len(lines))
		copy(sorted, lines)
		sort.SliceStable(sorted, func(i, j int) bool {
			return keyOf(sorted[i]) > keyOf(sorted[j])
		})
		return listResp(strings.Join(sorted, "\n"), inputCount, int32(len(sorted)), nil), nil // #nosec G115

	case pb.ListAction_LIST_SORT_NUMERIC:
		sorted := make([]string, len(lines))
		copy(sorted, lines)
		sort.SliceStable(sorted, func(i, j int) bool {
			a, errA := strconv.ParseFloat(strings.TrimSpace(sorted[i]), 64)
			b, errB := strconv.ParseFloat(strings.TrimSpace(sorted[j]), 64)
			if errA != nil || errB != nil {
				return strings.ToLower(sorted[i]) < strings.ToLower(sorted[j])
			}
			return a < b
		})
		return listResp(strings.Join(sorted, "\n"), inputCount, int32(len(sorted)), nil), nil // #nosec G115

	case pb.ListAction_LIST_SHUFFLE:
		shuffled := make([]string, len(lines))
		copy(shuffled, lines)
		for i := len(shuffled) - 1; i > 0; i-- {
			j, err := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
			if err != nil {
				return &pb.ListResponse{Error: "shuffle failed"}, nil
			}
			shuffled[i], shuffled[int(j.Int64())] = shuffled[int(j.Int64())], shuffled[i]
		}
		return listResp(strings.Join(shuffled, "\n"), inputCount, int32(len(shuffled)), nil), nil // #nosec G115

	case pb.ListAction_LIST_DEDUPE:
		seen := make(map[string]bool)
		var out []string
		for _, l := range lines {
			if !seen[keyOf(l)] {
				seen[keyOf(l)] = true
				out = append(out, l)
			}
		}
		return listResp(strings.Join(out, "\n"), inputCount, int32(len(out)), nil), nil // #nosec G115

	case pb.ListAction_LIST_UNIQUE_ONLY:
		counts := countLines(lines, keyOf)
		var out []string
		for _, l := range lines {
			if counts[keyOf(l)] == 1 {
				out = append(out, l)
			}
		}
		return listResp(strings.Join(out, "\n"), inputCount, int32(len(out)), nil), nil // #nosec G115

	case pb.ListAction_LIST_DUPLICATES:
		counts := countLines(lines, keyOf)
		seen := make(map[string]bool)
		var out []string
		for _, l := range lines {
			k := keyOf(l)
			if counts[k] > 1 && !seen[k] {
				seen[k] = true
				out = append(out, l)
			}
		}
		return listResp(strings.Join(out, "\n"), inputCount, int32(len(out)), nil), nil // #nosec G115

	case pb.ListAction_LIST_FREQUENCY:
		counts := countLines(lines, keyOf)
		// Build sorted frequency list (most common first, alpha tiebreak)
		type kv struct {
			line  string
			count int
		}
		unique := make(map[string]string) // key → original line (first seen)
		for _, l := range lines {
			k := keyOf(l)
			if _, exists := unique[k]; !exists {
				unique[k] = l
			}
		}
		var kvs []kv
		for k, cnt := range counts {
			kvs = append(kvs, kv{unique[k], cnt})
		}
		sort.Slice(kvs, func(i, j int) bool {
			if kvs[i].count != kvs[j].count {
				return kvs[i].count > kvs[j].count
			}
			return kvs[i].line < kvs[j].line
		})
		freq := make([]*pb.ListFreqItem, len(kvs))
		var resultLines []string
		for i, item := range kvs {
			freq[i] = &pb.ListFreqItem{
				Line:  item.line,
				Count: int32(item.count), // #nosec G115
			}
			resultLines = append(resultLines, fmt.Sprintf("%d\t%s", item.count, item.line))
		}
		return listResp(strings.Join(resultLines, "\n"), inputCount, int32(len(freq)), freq), nil // #nosec G115

	case pb.ListAction_LIST_REVERSE:
		out := make([]string, len(lines))
		for i, l := range lines {
			out[len(lines)-1-i] = l
		}
		return listResp(strings.Join(out, "\n"), inputCount, int32(len(out)), nil), nil // #nosec G115

	case pb.ListAction_LIST_TRIM:
		out := make([]string, len(lines))
		for i, l := range lines {
			out[i] = strings.TrimSpace(l)
		}
		return listResp(strings.Join(out, "\n"), inputCount, int32(len(out)), nil), nil // #nosec G115

	case pb.ListAction_LIST_REMOVE_EMPTY:
		var out []string
		for _, l := range lines {
			if strings.TrimSpace(l) != "" {
				out = append(out, l)
			}
		}
		return listResp(strings.Join(out, "\n"), inputCount, int32(len(out)), nil), nil // #nosec G115

	default:
		return &pb.ListResponse{Error: "unknown action"}, nil
	}
}

func countLines(lines []string, keyOf func(string) string) map[string]int {
	counts := make(map[string]int, len(lines))
	for _, l := range lines {
		counts[keyOf(l)]++
	}
	return counts
}

func listResp(result string, in, out int32, freq []*pb.ListFreqItem) *pb.ListResponse {
	return &pb.ListResponse{
		Result:      result,
		InputCount:  in,
		OutputCount: out,
		Frequency:   freq,
	}
}
