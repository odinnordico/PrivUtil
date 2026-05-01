package api

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/clbanning/mxj/v2"
	toml "github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"

	pb "github.com/odinnordico/privutil/proto"
)

func (s *Server) JsonFormat(ctx context.Context, req *pb.JsonFormatRequest) (*pb.JsonFormatResponse, error) {
	var data any
	if err := json.Unmarshal([]byte(req.Text), &data); err != nil {
		return &pb.JsonFormatResponse{Error: fmt.Sprintf("Invalid JSON: %v", err)}, nil
	}

	var (
		formatted []byte
		err       error
	)

	if req.Indent == "min" {
		formatted, err = json.Marshal(data)
	} else {
		indent := "  "
		switch req.Indent {
		case "4":
			indent = "    "
		case "tab":
			indent = "\t"
		}
		formatted, err = json.MarshalIndent(data, "", indent)
	}

	if err != nil {
		return &pb.JsonFormatResponse{Error: fmt.Sprintf("Formatting failed: %v", err)}, nil
	}

	return &pb.JsonFormatResponse{Text: string(formatted)}, nil
}

// csvDelimiter returns the configured delimiter rune, defaulting to comma.
func csvDelimiter(raw string) rune {
	if raw == `\t` || raw == "tab" {
		return '\t'
	}
	if len(raw) > 0 {
		return rune(raw[0])
	}
	return ','
}

// parseSource parses req.Data from its source format into a generic value.
func parseSource(req *pb.ConvertRequest) (any, error) {
	var data any

	switch req.SourceFormat {
	case pb.DataFormat_JSON:
		if err := json.Unmarshal([]byte(req.Data), &data); err != nil {
			return nil, err
		}

	case pb.DataFormat_YAML:
		if err := yaml.Unmarshal([]byte(req.Data), &data); err != nil {
			return nil, err
		}

	case pb.DataFormat_XML:
		mv, err := mxj.NewMapXml([]byte(req.Data))
		if err != nil {
			return nil, err
		}
		data = map[string]any(mv)

	case pb.DataFormat_TOML:
		if err := toml.Unmarshal([]byte(req.Data), &data); err != nil {
			return nil, err
		}

	case pb.DataFormat_CSV:
		delim := csvDelimiter(req.CsvDelimiter)
		r := csv.NewReader(strings.NewReader(req.Data))
		r.Comma = delim
		r.TrimLeadingSpace = true
		records, err := r.ReadAll()
		if err != nil {
			return nil, err
		}
		if len(records) == 0 {
			return []any{}, nil
		}
		if req.CsvNoHeader {
			rows := make([]any, len(records))
			for i, row := range records {
				cells := make([]any, len(row))
				for j, cell := range row {
					cells[j] = cell
				}
				rows[i] = cells
			}
			data = rows
		} else {
			headers := records[0]
			rows := make([]any, len(records)-1)
			for i, row := range records[1:] {
				m := make(map[string]any, len(headers))
				for j, h := range headers {
					if j < len(row) {
						m[h] = row[j]
					} else {
						m[h] = ""
					}
				}
				rows[i] = m
			}
			data = rows
		}

	default:
		return nil, fmt.Errorf("unsupported source format")
	}

	return data, nil
}

// marshalTarget serialises data into the requested target format.
func marshalTarget(data any, req *pb.ConvertRequest) ([]byte, error) {
	switch req.TargetFormat {
	case pb.DataFormat_JSON:
		return json.MarshalIndent(data, "", "  ")

	case pb.DataFormat_YAML:
		return yaml.Marshal(data)

	case pb.DataFormat_XML:
		var mv mxj.Map
		switch m := data.(type) {
		case map[string]any:
			mv = mxj.Map(m)
		default:
			b, err := json.Marshal(data)
			if err != nil {
				return nil, err
			}
			var err2 error
			mv, err2 = mxj.NewMapJson(b)
			if err2 != nil {
				return nil, err2
			}
		}
		return mv.XmlIndent("", "  ")

	case pb.DataFormat_TOML:
		return toml.Marshal(data)

	case pb.DataFormat_CSV:
		delim := csvDelimiter(req.CsvDelimiter)
		var buf strings.Builder
		w := csv.NewWriter(&buf)
		w.Comma = delim

		// Unwrap single-key map (common from XML via mxj)
		if m, ok := data.(map[string]any); ok && len(m) == 1 {
			for _, v := range m {
				data = v
			}
		}

		rows, ok := data.([]any)
		if !ok {
			return nil, fmt.Errorf("CSV output requires an array at the root level")
		}
		if len(rows) == 0 {
			return []byte{}, nil
		}

		switch firstRow := rows[0].(type) {
		case map[string]any:
			// Collect all keys (union across all rows, sorted for determinism)
			keySet := make(map[string]struct{})
			for _, row := range rows {
				if m, ok := row.(map[string]any); ok {
					for k := range m {
						keySet[k] = struct{}{}
					}
				}
			}
			headers := make([]string, 0, len(keySet))
			for k := range keySet {
				headers = append(headers, k)
			}
			sort.Strings(headers)

			if !req.CsvNoHeader {
				if err := w.Write(headers); err != nil {
					return nil, err
				}
			}
			for _, row := range rows {
				m, ok := row.(map[string]any)
				if !ok {
					continue
				}
				record := make([]string, len(headers))
				for i, h := range headers {
					if v, ok := m[h]; ok {
						record[i] = fmt.Sprintf("%v", v)
					}
				}
				if err := w.Write(record); err != nil {
					return nil, err
				}
			}
			_ = firstRow

		case []any:
			for _, row := range rows {
				cells, ok := row.([]any)
				if !ok {
					continue
				}
				record := make([]string, len(cells))
				for i, c := range cells {
					record[i] = fmt.Sprintf("%v", c)
				}
				if err := w.Write(record); err != nil {
					return nil, err
				}
			}
			_ = firstRow

		default:
			return nil, fmt.Errorf("CSV output requires an array of objects or array of arrays")
		}

		w.Flush()
		if err := w.Error(); err != nil {
			return nil, err
		}
		return []byte(buf.String()), nil

	default:
		return nil, fmt.Errorf("unsupported target format")
	}
}

func (s *Server) Convert(ctx context.Context, req *pb.ConvertRequest) (*pb.ConvertResponse, error) {
	data, err := parseSource(req)
	if err != nil {
		return &pb.ConvertResponse{Error: fmt.Sprintf("Parse failed: %v", err)}, nil
	}

	output, err := marshalTarget(data, req)
	if err != nil {
		return &pb.ConvertResponse{Error: fmt.Sprintf("Conversion failed: %v", err)}, nil
	}

	return &pb.ConvertResponse{Data: string(output)}, nil
}

// offsetToLineCol converts a byte offset in s to 1-based line and column.
func offsetToLineCol(s string, offset int) (line, col int) {
	if offset > len(s) {
		offset = len(s)
	}
	line = 1
	col = 1
	for i, c := range s {
		if i >= offset {
			break
		}
		if c == '\n' {
			line++
			col = 1
		} else {
			col++
		}
	}
	return line, col
}

func (s *Server) ValidateData(ctx context.Context, req *pb.ValidateRequest) (*pb.ValidateResponse, error) {
	if strings.TrimSpace(req.Data) == "" {
		return &pb.ValidateResponse{Valid: false, Error: "input is empty"}, nil
	}

	switch req.Format {
	case pb.DataFormat_JSON:
		dec := json.NewDecoder(strings.NewReader(req.Data))
		for {
			_, err := dec.Token()
			if err == io.EOF {
				break
			}
			if err != nil {
				if se, ok := err.(*json.SyntaxError); ok { //nolint:errorlint
					l, c := offsetToLineCol(req.Data, int(se.Offset))
					return &pb.ValidateResponse{
						Valid:  false,
						Error:  err.Error(),
						Line:   int32(l), // #nosec G115
						Column: int32(c), // #nosec G115
					}, nil
				}
				return &pb.ValidateResponse{Valid: false, Error: err.Error()}, nil
			}
		}

	case pb.DataFormat_YAML:
		var node yaml.Node
		if err := yaml.Unmarshal([]byte(req.Data), &node); err != nil {
			return &pb.ValidateResponse{Valid: false, Error: err.Error()}, nil
		}

	case pb.DataFormat_XML:
		dec := xml.NewDecoder(strings.NewReader(req.Data))
		for {
			_, err := dec.Token()
			if err == io.EOF {
				break
			}
			if err != nil {
				if se, ok := err.(*xml.SyntaxError); ok { //nolint:errorlint
					return &pb.ValidateResponse{
						Valid: false,
						Error: se.Msg,
						Line:  int32(se.Line), // #nosec G115
					}, nil
				}
				return &pb.ValidateResponse{Valid: false, Error: err.Error()}, nil
			}
		}

	case pb.DataFormat_TOML:
		var v any
		if err := toml.Unmarshal([]byte(req.Data), &v); err != nil {
			return &pb.ValidateResponse{Valid: false, Error: err.Error()}, nil
		}

	default:
		return &pb.ValidateResponse{Valid: false, Error: "unsupported format for validation"}, nil
	}

	return &pb.ValidateResponse{Valid: true}, nil
}

func (s *Server) JsonToGo(ctx context.Context, req *pb.JsonToGoRequest) (*pb.JsonToGoResponse, error) {
	var data any
	if err := json.Unmarshal([]byte(req.Json), &data); err != nil {
		return &pb.JsonToGoResponse{Error: fmt.Sprintf("Invalid JSON: %v", err)}, nil
	}

	name := req.StructName
	if name == "" {
		name = "AutoGenerated"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("type %s struct {\n", name))

	m, ok := data.(map[string]any)
	if ok {
		var keys []string
		for k := range m {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, k := range keys {
			val := m[k]
			fieldName := toPascalCase(k)
			typeName := "any"

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
			case []any:
				typeName = "[]any"
			case map[string]any:
				typeName = "struct { ... }"
			}

			fmt.Fprintf(&sb, "\t%s %s `json:\"%s\"`\n", fieldName, typeName, k)
		}
	} else {
		sb.WriteString("\t// Root must be an object\n")
	}

	sb.WriteString("}")
	return &pb.JsonToGoResponse{GoCode: sb.String()}, nil
}

func (s *Server) SqlFormat(ctx context.Context, req *pb.SqlRequest) (*pb.SqlResponse, error) {
	sql := req.Query
	keywords := []string{
		"select", "from", "where", "insert", "update", "delete", "create",
		"drop", "alter", "table", "into", "values", "join", "on",
		"order by", "group by", "limit", "offset", "and", "or", "not", "null", "as",
	}

	formatted := sql
	for _, kw := range keywords {
		re := regexp.MustCompile(`(?i)\b` + regexp.QuoteMeta(kw) + `\b`)
		formatted = re.ReplaceAllStringFunc(formatted, strings.ToUpper)
	}

	for _, kw := range []string{"SELECT", "FROM", "WHERE", "INSERT", "UPDATE", "DELETE", "ORDER BY", "GROUP BY", "LIMIT"} {
		if strings.Contains(formatted, kw) {
			formatted = strings.ReplaceAll(formatted, kw, "\n"+kw)
		}
	}
	formatted = strings.TrimSpace(formatted)

	return &pb.SqlResponse{Formatted: formatted}, nil
}

func (s *Server) ColorConvert(ctx context.Context, req *pb.ColorRequest) (*pb.ColorResponse, error) {
	input := strings.TrimSpace(req.Input)
	var r, g, b uint8
	var err error

	switch {
	case strings.HasPrefix(input, "#"):
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
	case strings.HasPrefix(input, "rgb"):
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
	default:
		err = fmt.Errorf("unsupported format (use #Hex or rgb(...))")
	}

	if err != nil {
		return &pb.ColorResponse{Error: err.Error()}, nil
	}

	rf, gf, bf := float64(r)/255.0, float64(g)/255.0, float64(b)/255.0
	maxC := max(rf, gf, bf)
	minC := min(rf, gf, bf)
	delta := maxC - minC

	var hue, sat, lum float64
	lum = (maxC + minC) / 2

	if delta != 0 {
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
