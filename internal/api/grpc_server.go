package api

import (
	"bytes"
	"context"
	"crypto/md5"  // #nosec G501
	"crypto/sha1" // #nosec G505
	"crypto/sha256"
	"crypto/sha512"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"html"
	"net"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/clbanning/mxj/v2"
	"github.com/google/uuid"
	"github.com/gorhill/cronexpr"
	"github.com/sergi/go-diff/diffmatchpatch"
	"gopkg.in/yaml.v3"

	pb "github.com/odinnordico/privutil/proto"
)

type Server struct {
	pb.UnimplementedPrivUtilServiceServer
}

func NewServer() *Server {
	return &Server{}
}

func (s *Server) Diff(ctx context.Context, req *pb.DiffRequest) (*pb.DiffResponse, error) {
	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(req.Text1, req.Text2, false)
	html := dmp.DiffPrettyHtml(diffs)

	return &pb.DiffResponse{
		DiffHtml: fmt.Sprintf("<div class='diff-output'>%s</div>", html),
	}, nil
}

func (s *Server) Base64Encode(ctx context.Context, req *pb.Base64Request) (*pb.Base64Response, error) {
	encoded := base64.StdEncoding.EncodeToString([]byte(req.Text))
	return &pb.Base64Response{
		Text: encoded,
	}, nil
}

func (s *Server) Base64Decode(ctx context.Context, req *pb.Base64Request) (*pb.Base64Response, error) {
	decoded, err := base64.StdEncoding.DecodeString(req.Text)
	if err != nil {
		return &pb.Base64Response{
			Error: fmt.Sprintf("Failed to decode: %v", err),
		}, nil
	}
	return &pb.Base64Response{
		Text: string(decoded),
	}, nil
}

func (s *Server) JsonFormat(ctx context.Context, req *pb.JsonFormatRequest) (*pb.JsonFormatResponse, error) {
	var data interface{}
	if err := json.Unmarshal([]byte(req.Text), &data); err != nil {
		return &pb.JsonFormatResponse{Error: fmt.Sprintf("Invalid JSON: %v", err)}, nil
	}

	var formatted []byte
	var err error

	if req.Indent == "min" {
		formatted, err = json.Marshal(data)
	} else {
		indent := "  " // default 2 spaces
		if req.Indent == "4" {
			indent = "    "
		} else if req.Indent == "tab" {
			indent = "\t"
		}

		formatted, err = json.MarshalIndent(data, "", indent)
	}

	if err != nil {
		return &pb.JsonFormatResponse{Error: fmt.Sprintf("Formatting failed: %v", err)}, nil
	}

	return &pb.JsonFormatResponse{Text: string(formatted)}, nil
}

func (s *Server) Convert(ctx context.Context, req *pb.ConvertRequest) (*pb.ConvertResponse, error) {
	var data interface{}
	var err error

	switch req.SourceFormat {
	case pb.DataFormat_JSON:
		err = json.Unmarshal([]byte(req.Data), &data)
	case pb.DataFormat_YAML:
		err = yaml.Unmarshal([]byte(req.Data), &data)
	case pb.DataFormat_XML:
		mv, merr := mxj.NewMapXml([]byte(req.Data))
		if merr != nil {
			err = merr
		} else {
			data = mv
		}
	}

	if err != nil {
		return &pb.ConvertResponse{Error: fmt.Sprintf("Parse failed: %v", err)}, nil
	}

	var output []byte
	switch req.TargetFormat {
	case pb.DataFormat_JSON:
		output, err = json.MarshalIndent(data, "", "  ")
	case pb.DataFormat_YAML:
		output, err = yaml.Marshal(data)
	case pb.DataFormat_XML:
		m, ok := data.(map[string]interface{})
		if !ok {
			b, _ := json.Marshal(data)
			mv, _ := mxj.NewMapJson(b)
			output, err = mv.XmlIndent("", "  ")
		} else {
			mv := mxj.Map(m)
			output, err = mv.XmlIndent("", "  ")
		}
	}

	if err != nil {
		return &pb.ConvertResponse{Error: fmt.Sprintf("Conversion failed: %v", err)}, nil
	}

	return &pb.ConvertResponse{Data: string(output)}, nil
}

func (s *Server) GenerateUuid(ctx context.Context, req *pb.UuidRequest) (*pb.UuidResponse, error) {
	var uuids []string
	count := req.Count
	if count <= 0 {
		count = 1
	}
	if count > 100 {
		count = 100
	}

	for i := 0; i < int(count); i++ {
		var u uuid.UUID
		var err error

		if req.Version == "v1" {
			u, err = uuid.NewUUID()
		} else {
			u, err = uuid.NewRandom()
		}

		if err != nil {
			return nil, err
		}

		str := u.String()
		if !req.Hyphen {
			str = strings.ReplaceAll(str, "-", "")
		}
		if req.Uppercase {
			str = strings.ToUpper(str)
		}
		uuids = append(uuids, str)
	}

	return &pb.UuidResponse{Uuids: uuids}, nil
}

func (s *Server) GenerateLorem(ctx context.Context, req *pb.LoremRequest) (*pb.LoremResponse, error) {
	var text string
	count := int(req.Count)
	if count <= 0 {
		count = 1
	}

	switch req.Type {
	case "word":
		var words []string
		for i := 0; i < count; i++ {
			words = append(words, gofakeit.Word())
		}
		text = strings.Join(words, " ")
	case "sentence":
		var sentences []string
		for i := 0; i < count; i++ {
			sentences = append(sentences, gofakeit.Sentence(10))
		}
		text = strings.Join(sentences, " ")
	default:
		var paragraphs []string
		for i := 0; i < count; i++ {
			paragraphs = append(paragraphs, gofakeit.Paragraph(3, 5, 10, "\n"))
		}
		text = strings.Join(paragraphs, "\n\n")
	}

	return &pb.LoremResponse{Text: text}, nil
}

func (s *Server) CalculateHash(ctx context.Context, req *pb.HashRequest) (*pb.HashResponse, error) {
	var hash string
	data := []byte(req.Text)

	switch req.Algo {
	case "md5":
		// #nosec G401 G501 - MD5 is intentionally provided as a utility feature
		sum := md5.Sum(data)
		hash = hex.EncodeToString(sum[:])
	case "sha1":
		// #nosec G401 G505 - SHA1 is intentionally provided as a utility feature
		sum := sha1.Sum(data)
		hash = hex.EncodeToString(sum[:])
	case "sha512":
		sum := sha512.Sum512(data)
		hash = hex.EncodeToString(sum[:])
	default:
		sum := sha256.Sum256(data)
		hash = hex.EncodeToString(sum[:])
	}

	return &pb.HashResponse{Hash: hash}, nil
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

func (s *Server) UrlEncode(ctx context.Context, req *pb.TextRequest) (*pb.TextResponse, error) {
	return &pb.TextResponse{Text: url.QueryEscape(req.Text)}, nil
}

func (s *Server) UrlDecode(ctx context.Context, req *pb.TextRequest) (*pb.TextResponse, error) {
	decoded, err := url.QueryUnescape(req.Text)
	if err != nil {
		return &pb.TextResponse{Text: "Error: " + err.Error()}, nil
	}
	return &pb.TextResponse{Text: decoded}, nil
}

func (s *Server) HtmlEncode(ctx context.Context, req *pb.TextRequest) (*pb.TextResponse, error) {
	return &pb.TextResponse{Text: html.EscapeString(req.Text)}, nil
}

func (s *Server) HtmlDecode(ctx context.Context, req *pb.TextRequest) (*pb.TextResponse, error) {
	return &pb.TextResponse{Text: html.UnescapeString(req.Text)}, nil
}

func (s *Server) TimeConvert(ctx context.Context, req *pb.TimeRequest) (*pb.TimeResponse, error) {
	input := strings.TrimSpace(req.Input)
	var t time.Time

	if input == "" || strings.EqualFold(input, "now") {
		t = time.Now()
	} else {
		if ts, err := strconv.ParseInt(input, 10, 64); err == nil {
			if ts > 10000000000 {
				t = time.UnixMilli(ts)
			} else {
				t = time.Unix(ts, 0)
			}
		} else {
			formats := []string{
				time.RFC3339,
				time.RFC3339Nano,
				time.Layout,
				"2006-01-02T15:04:05",
				"2006-01-02 15:04:05",
				time.RubyDate,
				time.UnixDate,
			}
			var parsed bool
			for _, layout := range formats {
				if pt, err := time.Parse(layout, input); err == nil {
					t = pt
					parsed = true
					break
				}
			}
			if !parsed {
				return &pb.TimeResponse{Iso: "Invalid input format"}, nil
			}
		}
	}

	return &pb.TimeResponse{
		Unix:  t.Unix(),
		Utc:   t.UTC().Format(time.RFC3339),
		Local: t.Local().Format("2006-01-02 15:04:05 -0700 MST"),
		Iso:   t.Format(time.RFC3339),
	}, nil
}

func (s *Server) JwtDecode(ctx context.Context, req *pb.JwtRequest) (*pb.JwtResponse, error) {
	parts := strings.Split(req.Token, ".")
	if len(parts) < 2 {
		return &pb.JwtResponse{Error: "Invalid JWT format"}, nil
	}

	decodeSegment := func(seg string) string {
		if l := len(seg) % 4; l > 0 {
			seg += strings.Repeat("=", 4-l)
		}
		b, err := base64.URLEncoding.DecodeString(seg)
		if err != nil {
			return fmt.Sprintf("Error decoding: %v", err)
		}
		var pretty bytes.Buffer
		if err := json.Indent(&pretty, b, "", "  "); err == nil {
			return pretty.String()
		}
		return string(b)
	}

	return &pb.JwtResponse{
		Header:  decodeSegment(parts[0]),
		Payload: decodeSegment(parts[1]),
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

func (s *Server) JsonToGo(ctx context.Context, req *pb.JsonToGoRequest) (*pb.JsonToGoResponse, error) {
	var data interface{}
	if err := json.Unmarshal([]byte(req.Json), &data); err != nil {
		return &pb.JsonToGoResponse{Error: fmt.Sprintf("Invalid JSON: %v", err)}, nil
	}

	name := req.StructName
	if name == "" {
		name = "AutoGenerated"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("type %s struct {\n", name))

	m, ok := data.(map[string]interface{})
	if ok {
		var keys []string
		for k := range m {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, k := range keys {
			val := m[k]
			fieldName := toPascalCase(k)
			typeName := "interface{}"

			switch v := val.(type) {
			case string:
				typeName = "string"
			case float64:
				typeName = "float64"
				if v == float64(int64(v)) {
					typeName = "int"
				}
			case bool:
				typeName = "bool"
			case []interface{}:
				typeName = "[]interface{}"
			case map[string]interface{}:
				typeName = "struct { ... }"
			}

			sb.WriteString(fmt.Sprintf("\t%s %s `json:\"%s\"`\n", fieldName, typeName, k))
		}
	} else {
		sb.WriteString("\t// Root must be an object\n")
	}

	sb.WriteString("}")
	return &pb.JsonToGoResponse{GoCode: sb.String()}, nil
}

func (s *Server) CronExplain(ctx context.Context, req *pb.CronRequest) (*pb.CronResponse, error) {
	expr, err := cronexpr.Parse(req.Expression)
	if err != nil {
		return &pb.CronResponse{Error: fmt.Sprintf("Invalid cron expression: %v", err)}, nil
	}

	nextTimes := expr.NextN(time.Now(), 5)
	var nextRuns []string
	for _, t := range nextTimes {
		nextRuns = append(nextRuns, t.Format(time.RFC3339))
	}

	desc := describeCron(req.Expression)

	return &pb.CronResponse{
		Description: desc,
		NextRuns:    strings.Join(nextRuns, "\n"),
	}, nil
}

func describeCron(expr string) string {
	parts := strings.Fields(expr)
	if len(parts) != 5 {
		return "Invalid cron expression format"
	}
	min, hour, dom, month, dow := parts[0], parts[1], parts[2], parts[3], parts[4]

	// Very basic heuristic description
	if expr == "* * * * *" {
		return "Every minute"
	}
	if strings.HasPrefix(min, "*/") && hour == "*" && dom == "*" && month == "*" && dow == "*" {
		return fmt.Sprintf("Every %s minutes", min[2:])
	}
	if min == "0" && hour == "*" && dom == "*" && month == "*" && dow == "*" {
		return "At the start of every hour"
	}
	if min == "0" && strings.HasPrefix(hour, "*/") && dom == "*" && month == "*" && dow == "*" {
		return fmt.Sprintf("At minute 0 past every %s hours", hour[2:])
	}
	if min == "0" && hour == "0" && dom == "*" && month == "*" && dow == "*" {
		return "At 00:00 every day"
	}

	// Complex cases fallback
	desc := "Run "
	if min != "*" {
		desc += fmt.Sprintf("at minute %s", min)
	} else {
		desc += "every minute"
	}

	if hour != "*" {
		desc += fmt.Sprintf(" of hour %s", hour)
	}

	if dom != "*" {
		desc += fmt.Sprintf(" on day-of-month %s", dom)
	}

	if dow != "*" {
		desc += fmt.Sprintf(" on day-of-week %s", dow)
	}

	return desc
}

func (s *Server) CertParse(ctx context.Context, req *pb.CertRequest) (*pb.CertResponse, error) {
	block, _ := pem.Decode([]byte(req.Data))
	if block == nil {
		return &pb.CertResponse{Error: "Failed to decode PEM block"}, nil
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return &pb.CertResponse{Error: fmt.Sprintf("Failed to parse certificate: %v", err)}, nil
	}

	return &pb.CertResponse{
		Subject:   cert.Subject.String(),
		Issuer:    cert.Issuer.String(),
		NotBefore: cert.NotBefore.Format(time.RFC3339),
		NotAfter:  cert.NotAfter.Format(time.RFC3339),
		Sans:      cert.DNSNames,
	}, nil
}

func (s *Server) ColorConvert(ctx context.Context, req *pb.ColorRequest) (*pb.ColorResponse, error) {
	input := strings.TrimSpace(req.Input)
	var r, g, b uint8
	var err error

	if strings.HasPrefix(input, "#") {
		input = strings.TrimPrefix(input, "#")
		if len(input) == 3 {
			input = string([]byte{input[0], input[0], input[1], input[1], input[2], input[2]})
		}
		if len(input) == 6 {
			if v, e := strconv.ParseUint(input, 16, 32); e == nil {
				r = uint8(v >> 16) // #nosec G115
				g = uint8(v >> 8)  // #nosec G115
				b = uint8(v)       // #nosec G115
			} else {
				err = e
			}
		} else {
			err = fmt.Errorf("invalid hex length")
		}
	} else if strings.HasPrefix(input, "rgb") {
		re := regexp.MustCompile(`rgb\(\s*(\d+)\s*,\s*(\d+)\s*,\s*(\d+)\s*\)`)
		matches := re.FindStringSubmatch(input)
		if len(matches) == 4 {
			ri, _ := strconv.Atoi(matches[1])
			gi, _ := strconv.Atoi(matches[2])
			bi, _ := strconv.Atoi(matches[3])
			r, g, b = uint8(ri), uint8(gi), uint8(bi) // #nosec G115
		} else {
			err = fmt.Errorf("invalid rgb format")
		}
	} else {
		err = fmt.Errorf("unsupported format (use #Hex or rgb(...) )")
	}

	if err != nil {
		return &pb.ColorResponse{Error: err.Error()}, nil
	}

	rf, gf, bf := float64(r)/255.0, float64(g)/255.0, float64(b)/255.0
	maxC := max(rf, max(gf, bf))
	minC := min(rf, min(gf, bf))
	delta := maxC - minC

	var hue, sat, lum float64
	lum = (maxC + minC) / 2

	if delta == 0 {
		hue = 0
		sat = 0
	} else {
		if lum < 0.5 {
			sat = delta / (maxC + minC)
		} else {
			sat = delta / (2 - maxC - minC)
		}

		switch maxC {
		case rf:
			hue = (gf - bf) / delta
			if gf < bf {
				hue += 6
			}
		case gf:
			hue = (bf-rf)/delta + 2
		case bf:
			hue = (rf-gf)/delta + 4
		}
		hue /= 6
	}

	return &pb.ColorResponse{
		Hex: fmt.Sprintf("#%02x%02x%02x", r, g, b),
		Rgb: fmt.Sprintf("rgb(%d, %d, %d)", r, g, b),
		Hsl: fmt.Sprintf("hsl(%.0f, %.0f%%, %.0f%%)", hue*360, sat*100, lum*100),
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

func (s *Server) StringEscape(ctx context.Context, req *pb.EscapeRequest) (*pb.EscapeResponse, error) {
	text := req.Text
	var res string
	var err error

	switch req.Mode {
	case "json":
		if req.Action == "escape" {
			b, _ := json.Marshal(text)
			res = string(b)
		} else {
			if strings.HasPrefix(text, "\"") && strings.HasSuffix(text, "\"") {
				if err := json.Unmarshal([]byte(text), &res); err != nil {
					return &pb.EscapeResponse{Error: "Invalid JSON string"}, nil
				}
			} else {
				if err := json.Unmarshal([]byte("\""+text+"\""), &res); err != nil {
					return &pb.EscapeResponse{Error: "Could not unescape"}, nil
				}
			}
		}
	case "html_entity":
		if req.Action == "escape" {
			res = html.EscapeString(text)
		} else {
			res = html.UnescapeString(text)
		}
	case "url":
		if req.Action == "escape" {
			res = url.QueryEscape(text)
		} else {
			res, err = url.QueryUnescape(text)
		}
	case "sql":
		if req.Action == "escape" {
			res = strings.ReplaceAll(text, "'", "''")
		} else {
			res = strings.ReplaceAll(text, "''", "'")
		}
	case "java":
		if req.Action == "escape" {
			res = strconv.Quote(text)
		} else {
			res, err = strconv.Unquote(text)
		}
	default:
		return &pb.EscapeResponse{Error: "Unknown mode"}, nil
	}

	if err != nil {
		return &pb.EscapeResponse{Error: err.Error()}, nil
	}

	return &pb.EscapeResponse{Result: res}, nil
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

func toPascalCase(s string) string {
	s = strings.ReplaceAll(s, "_", " ")
	s = strings.ReplaceAll(s, "-", " ")
	words := strings.Fields(s)
	for i, w := range words {
		if len(w) > 0 {
			words[i] = strings.ToUpper(w[:1]) + w[1:]
		}
	}
	return strings.Join(words, "")
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
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

func (s *Server) SqlFormat(ctx context.Context, req *pb.SqlRequest) (*pb.SqlResponse, error) {
	// Very basic custom storage formatter
	sql := req.Query
	// Keywords to uppercase
	keywords := []string{"select", "from", "where", "insert", "update", "delete", "create", "drop", "alter", "table", "into", "values", "join", "on", "order by", "group by", "limit", "offset", "and", "or", "not", "null", "as"}

	// Poor man's formatter: uppercase keywords and add format breaks
	// For production usage, a real SQL parsing library is needed.
	// We will settle for regex replacements for common keywords.
	formatted := sql
	for _, kw := range keywords {
		re := regexp.MustCompile(`(?i)\b` + regexp.QuoteMeta(kw) + `\b`)
		formatted = re.ReplaceAllStringFunc(formatted, strings.ToUpper)
	}

	// Add newlines before major clauses
	major := []string{"SELECT", "FROM", "WHERE", "INSERT", "UPDATE", "DELETE", "ORDER BY", "GROUP BY", "LIMIT"}
	for _, kw := range major {
		if strings.Contains(formatted, kw) {
			formatted = strings.ReplaceAll(formatted, kw, "\n"+kw)
		}
	}
	formatted = strings.TrimSpace(formatted)

	return &pb.SqlResponse{Formatted: formatted}, nil
}

func (s *Server) IpCalc(ctx context.Context, req *pb.IpRequest) (*pb.IpResponse, error) {
	// Using standard library net package
	// Parse CIDR
	ip, ipnet, err := net.ParseCIDR(req.Cidr)
	if err != nil {
		// Try parsing as single IP, assume /32 (v4) or /128 (v6) if valid IP
		if ip2 := net.ParseIP(req.Cidr); ip2 != nil {
			if ip2.To4() != nil {
				ip, ipnet, _ = net.ParseCIDR(req.Cidr + "/32")
			} else {
				ip, ipnet, _ = net.ParseCIDR(req.Cidr + "/128")
			}
		}

		if ip == nil {
			return &pb.IpResponse{Error: "Invalid IP or CIDR"}, nil
		}
	}

	ones, bits := ipnet.Mask.Size()

	// Calculate network, broadcast, first/last IP
	// This is complex for IPv6, implementing simplified version mostly for IPv4

	network := ipnet.IP
	var broadcast net.IP
	var netmask net.IP = net.IP(ipnet.Mask)

	if ip.To4() != nil {
		// IPv4 logic
		broadcast = make(net.IP, 4)
		for i := 0; i < 4; i++ {
			broadcast[i] = network[i] | ^ipnet.Mask[i]
		}
	}

	// Count
	var count int64
	if bits-ones < 63 { // Avoid overflow
		count = 1 << uint(bits-ones) // #nosec G115
	}

	return &pb.IpResponse{
		Network:   network.String(),
		Broadcast: broadcast.String(),
		Netmask:   netmask.String(),
		NumHosts:  count,
		FirstIp:   network.String(),   // Approximate
		LastIp:    broadcast.String(), // Approximate
	}, nil
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

func (s *Server) GeneratePassword(ctx context.Context, req *pb.PasswordRequest) (*pb.PasswordResponse, error) {
	length := int(req.Length)
	if length <= 0 {
		length = 16
	}
	if length > 128 {
		length = 128
	}

	count := int(req.Count)
	if count <= 0 {
		count = 1
	}
	if count > 100 {
		count = 100
	}

	// Build character set
	var charset string
	if req.CustomChars != "" {
		charset = req.CustomChars
	} else {
		if req.Lowercase {
			charset += "abcdefghijklmnopqrstuvwxyz"
		}
		if req.Uppercase {
			charset += "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
		}
		if req.Numbers {
			charset += "0123456789"
		}
		if req.Symbols {
			charset += "!@#$%^&*()-_=+[]{}|;:,.<>?"
		}
		// Default to all if nothing selected
		if charset == "" {
			charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*"
		}
	}

	passwords := make([]string, count)
	for i := 0; i < count; i++ {
		password := make([]byte, length)
		for j := 0; j < length; j++ {
			password[j] = charset[gofakeit.Number(0, len(charset)-1)]
		}
		passwords[i] = string(password)
	}

	return &pb.PasswordResponse{Passwords: passwords}, nil
}
