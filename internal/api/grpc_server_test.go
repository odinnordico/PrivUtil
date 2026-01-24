package api

import (
	"context"
	"strings"
	"testing"

	pb "github.com/odinnordico/privutil/proto"
)

func TestDiff(t *testing.T) {
	s := NewServer()
	ctx := context.Background()

	tests := []struct {
		name     string
		text1    string
		text2    string
		wantHTML bool
	}{
		{"identical", "hello", "hello", true},
		{"different", "hello", "world", true},
		{"empty", "", "", true},
		{"one empty", "hello", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := s.Diff(ctx, &pb.DiffRequest{Text1: tt.text1, Text2: tt.text2})
			if err != nil {
				t.Fatalf("Diff() error = %v", err)
			}
			if tt.wantHTML && resp.DiffHtml == "" {
				t.Error("Diff() expected non-empty HTML")
			}
		})
	}
}

func TestBase64Encode(t *testing.T) {
	s := NewServer()
	ctx := context.Background()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"simple", "hello", "aGVsbG8="},
		{"empty", "", ""},
		{"special chars", "hello world!", "aGVsbG8gd29ybGQh"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := s.Base64Encode(ctx, &pb.Base64Request{Text: tt.input})
			if err != nil {
				t.Fatalf("Base64Encode() error = %v", err)
			}
			if resp.Text != tt.want {
				t.Errorf("Base64Encode() = %v, want %v", resp.Text, tt.want)
			}
		})
	}
}

func TestBase64Decode(t *testing.T) {
	s := NewServer()
	ctx := context.Background()

	tests := []struct {
		name      string
		input     string
		want      string
		wantError bool
	}{
		{"simple", "aGVsbG8=", "hello", false},
		{"empty", "", "", false},
		{"invalid", "!!!invalid!!!", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := s.Base64Decode(ctx, &pb.Base64Request{Text: tt.input})
			if err != nil {
				t.Fatalf("Base64Decode() unexpected error = %v", err)
			}
			if tt.wantError && resp.Error == "" {
				t.Error("Base64Decode() expected error in response")
			}
			if !tt.wantError && resp.Text != tt.want {
				t.Errorf("Base64Decode() = %v, want %v", resp.Text, tt.want)
			}
		})
	}
}

func TestJsonFormat(t *testing.T) {
	s := NewServer()
	ctx := context.Background()

	tests := []struct {
		name      string
		input     string
		indent    string
		wantError bool
	}{
		{"valid json", `{"key":"value"}`, "2", false},
		{"minify", `{"key": "value"}`, "min", false},
		{"tab indent", `{"key":"value"}`, "tab", false},
		{"invalid json", `{invalid}`, "2", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := s.JsonFormat(ctx, &pb.JsonFormatRequest{Text: tt.input, Indent: tt.indent})
			if err != nil {
				t.Fatalf("JsonFormat() error = %v", err)
			}
			if tt.wantError && resp.Error == "" {
				t.Error("JsonFormat() expected error")
			}
			if !tt.wantError && resp.Text == "" {
				t.Error("JsonFormat() expected non-empty result")
			}
		})
	}
}

func TestConvert(t *testing.T) {
	s := NewServer()
	ctx := context.Background()

	tests := []struct {
		name   string
		data   string
		source pb.DataFormat
		target pb.DataFormat
	}{
		{"json to yaml", `{"key":"value"}`, pb.DataFormat_JSON, pb.DataFormat_YAML},
		{"yaml to json", "key: value", pb.DataFormat_YAML, pb.DataFormat_JSON},
		{"json to xml", `{"root":"value"}`, pb.DataFormat_JSON, pb.DataFormat_XML},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := s.Convert(ctx, &pb.ConvertRequest{
				Data:         tt.data,
				SourceFormat: tt.source,
				TargetFormat: tt.target,
			})
			if err != nil {
				t.Fatalf("Convert() error = %v", err)
			}
			if resp.Error != "" {
				t.Errorf("Convert() error = %v", resp.Error)
			}
			if resp.Data == "" {
				t.Error("Convert() expected non-empty result")
			}
		})
	}
}

func TestGenerateUuid(t *testing.T) {
	s := NewServer()
	ctx := context.Background()

	tests := []struct {
		name      string
		count     int32
		version   string
		hyphen    bool
		uppercase bool
	}{
		{"single v4", 1, "v4", true, false},
		{"multiple", 5, "v4", true, false},
		{"no hyphen", 1, "v4", false, false},
		{"uppercase", 1, "v4", true, true},
		{"v1", 1, "v1", true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := s.GenerateUuid(ctx, &pb.UuidRequest{
				Count:     tt.count,
				Version:   tt.version,
				Hyphen:    tt.hyphen,
				Uppercase: tt.uppercase,
			})
			if err != nil {
				t.Fatalf("GenerateUuid() error = %v", err)
			}
			if len(resp.Uuids) != int(tt.count) {
				t.Errorf("GenerateUuid() got %d uuids, want %d", len(resp.Uuids), tt.count)
			}
		})
	}
}

func TestGenerateLorem(t *testing.T) {
	s := NewServer()
	ctx := context.Background()

	tests := []struct {
		name  string
		typ   string
		count int32
	}{
		{"word", "word", 5},
		{"sentence", "sentence", 3},
		{"paragraph", "paragraph", 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := s.GenerateLorem(ctx, &pb.LoremRequest{Type: tt.typ, Count: tt.count})
			if err != nil {
				t.Fatalf("GenerateLorem() error = %v", err)
			}
			if resp.Text == "" {
				t.Error("GenerateLorem() expected non-empty result")
			}
		})
	}
}

func TestCalculateHash(t *testing.T) {
	s := NewServer()
	ctx := context.Background()

	tests := []struct {
		name string
		algo string
		text string
	}{
		{"md5", "md5", "hello"},
		{"sha1", "sha1", "hello"},
		{"sha256", "sha256", "hello"},
		{"sha512", "sha512", "hello"},
		{"default", "", "hello"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := s.CalculateHash(ctx, &pb.HashRequest{Text: tt.text, Algo: tt.algo})
			if err != nil {
				t.Fatalf("CalculateHash() error = %v", err)
			}
			if resp.Hash == "" {
				t.Error("CalculateHash() expected non-empty hash")
			}
		})
	}
}

func TestTextInspect(t *testing.T) {
	s := NewServer()
	ctx := context.Background()

	resp, err := s.TextInspect(ctx, &pb.TextInspectRequest{Text: "hello world\nnew line"})
	if err != nil {
		t.Fatalf("TextInspect() error = %v", err)
	}
	if resp.CharCount != 20 {
		t.Errorf("TextInspect() CharCount = %d, want 20", resp.CharCount)
	}
	if resp.WordCount != 4 {
		t.Errorf("TextInspect() WordCount = %d, want 4", resp.WordCount)
	}
	if resp.LineCount != 2 {
		t.Errorf("TextInspect() LineCount = %d, want 2", resp.LineCount)
	}
}

func TestTextManipulate(t *testing.T) {
	s := NewServer()
	ctx := context.Background()

	tests := []struct {
		name   string
		action pb.TextAction
		input  string
	}{
		{"sort az", pb.TextAction_SORT_AZ, "c\nb\na"},
		{"sort za", pb.TextAction_SORT_ZA, "a\nb\nc"},
		{"reverse", pb.TextAction_REVERSE, "a\nb\nc"},
		{"dedupe", pb.TextAction_DEDUPE, "a\na\nb"},
		{"remove empty", pb.TextAction_REMOVE_EMPTY, "a\n\nb"},
		{"trim", pb.TextAction_TRIM, " a \n b "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := s.TextManipulate(ctx, &pb.TextManipulateRequest{Text: tt.input, Action: tt.action})
			if err != nil {
				t.Fatalf("TextManipulate() error = %v", err)
			}
			if resp.Text == "" {
				t.Error("TextManipulate() expected non-empty result")
			}
		})
	}
}

func TestUrlEncodeDecode(t *testing.T) {
	s := NewServer()
	ctx := context.Background()

	original := "hello world & more"
	encResp, _ := s.UrlEncode(ctx, &pb.TextRequest{Text: original})
	decResp, _ := s.UrlDecode(ctx, &pb.TextRequest{Text: encResp.Text})

	if decResp.Text != original {
		t.Errorf("UrlEncode/Decode roundtrip failed: got %v, want %v", decResp.Text, original)
	}
}

func TestHtmlEncodeDecode(t *testing.T) {
	s := NewServer()
	ctx := context.Background()

	original := "<script>alert('xss')</script>"
	encResp, _ := s.HtmlEncode(ctx, &pb.TextRequest{Text: original})
	decResp, _ := s.HtmlDecode(ctx, &pb.TextRequest{Text: encResp.Text})

	if decResp.Text != original {
		t.Errorf("HtmlEncode/Decode roundtrip failed: got %v, want %v", decResp.Text, original)
	}
}

func TestTimeConvert(t *testing.T) {
	s := NewServer()
	ctx := context.Background()

	tests := []struct {
		name  string
		input string
	}{
		{"now", "now"},
		{"empty", ""},
		{"unix", "1609459200"},
		{"unix ms", "1609459200000"},
		{"rfc3339", "2021-01-01T00:00:00Z"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := s.TimeConvert(ctx, &pb.TimeRequest{Input: tt.input})
			if err != nil {
				t.Fatalf("TimeConvert() error = %v", err)
			}
			if resp.Iso == "" && resp.Iso != "Invalid input format" {
				t.Error("TimeConvert() expected result")
			}
		})
	}
}

func TestJwtDecode(t *testing.T) {
	s := NewServer()
	ctx := context.Background()

	// Valid JWT structure (not actually signed)
	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"

	resp, err := s.JwtDecode(ctx, &pb.JwtRequest{Token: token})
	if err != nil {
		t.Fatalf("JwtDecode() error = %v", err)
	}
	if resp.Header == "" || resp.Payload == "" {
		t.Error("JwtDecode() expected header and payload")
	}
}

func TestRegexTest(t *testing.T) {
	s := NewServer()
	ctx := context.Background()

	tests := []struct {
		name      string
		pattern   string
		text      string
		wantMatch bool
		wantError bool
	}{
		{"simple match", "hello", "hello world", true, false},
		{"no match", "xyz", "hello world", false, false},
		{"invalid regex", "[invalid", "hello", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := s.RegexTest(ctx, &pb.RegexRequest{Pattern: tt.pattern, Text: tt.text})
			if err != nil {
				t.Fatalf("RegexTest() error = %v", err)
			}
			if tt.wantError && resp.Error == "" {
				t.Error("RegexTest() expected error")
			}
			if !tt.wantError && resp.Match != tt.wantMatch {
				t.Errorf("RegexTest() match = %v, want %v", resp.Match, tt.wantMatch)
			}
		})
	}
}

func TestJsonToGo(t *testing.T) {
	s := NewServer()
	ctx := context.Background()

	resp, err := s.JsonToGo(ctx, &pb.JsonToGoRequest{
		Json:       `{"name": "test", "count": 42, "active": true}`,
		StructName: "MyStruct",
	})
	if err != nil {
		t.Fatalf("JsonToGo() error = %v", err)
	}
	if !strings.Contains(resp.GoCode, "type MyStruct struct") {
		t.Error("JsonToGo() expected struct definition")
	}
}

func TestCronExplain(t *testing.T) {
	s := NewServer()
	ctx := context.Background()

	tests := []struct {
		name       string
		expression string
		wantError  bool
	}{
		{"every minute", "* * * * *", false},
		{"every 5 min", "*/5 * * * *", false},
		{"invalid", "invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := s.CronExplain(ctx, &pb.CronRequest{Expression: tt.expression})
			if err != nil {
				t.Fatalf("CronExplain() error = %v", err)
			}
			if tt.wantError && resp.Error == "" {
				t.Error("CronExplain() expected error")
			}
			if !tt.wantError && resp.Description == "" {
				t.Error("CronExplain() expected description")
			}
		})
	}
}

func TestColorConvert(t *testing.T) {
	s := NewServer()
	ctx := context.Background()

	tests := []struct {
		name      string
		input     string
		wantError bool
	}{
		{"hex", "#ff0000", false},
		{"short hex", "#f00", false},
		{"rgb", "rgb(255, 0, 0)", false},
		{"invalid", "invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := s.ColorConvert(ctx, &pb.ColorRequest{Input: tt.input})
			if err != nil {
				t.Fatalf("ColorConvert() error = %v", err)
			}
			if tt.wantError && resp.Error == "" {
				t.Error("ColorConvert() expected error")
			}
			if !tt.wantError && resp.Hex == "" {
				t.Error("ColorConvert() expected hex output")
			}
		})
	}
}

func TestCaseConvert(t *testing.T) {
	s := NewServer()
	ctx := context.Background()

	resp, err := s.CaseConvert(ctx, &pb.CaseRequest{Text: "hello world"})
	if err != nil {
		t.Fatalf("CaseConvert() error = %v", err)
	}
	if resp.Camel != "helloWorld" {
		t.Errorf("CaseConvert() Camel = %v, want helloWorld", resp.Camel)
	}
	if resp.Pascal != "HelloWorld" {
		t.Errorf("CaseConvert() Pascal = %v, want HelloWorld", resp.Pascal)
	}
	if resp.Snake != "hello_world" {
		t.Errorf("CaseConvert() Snake = %v, want hello_world", resp.Snake)
	}
	if resp.Kebab != "hello-world" {
		t.Errorf("CaseConvert() Kebab = %v, want hello-world", resp.Kebab)
	}
}

func TestStringEscape(t *testing.T) {
	s := NewServer()
	ctx := context.Background()

	tests := []struct {
		name   string
		mode   string
		action string
		text   string
	}{
		{"json escape", "json", "escape", `hello "world"`},
		{"json unescape", "json", "unescape", `"hello \"world\""`},
		{"url escape", "url", "escape", "hello world"},
		{"html escape", "html_entity", "escape", "<script>"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := s.StringEscape(ctx, &pb.EscapeRequest{
				Text:   tt.text,
				Mode:   tt.mode,
				Action: tt.action,
			})
			if err != nil {
				t.Fatalf("StringEscape() error = %v", err)
			}
			if resp.Error != "" {
				t.Errorf("StringEscape() error = %v", resp.Error)
			}
		})
	}
}

func TestTextSimilarity(t *testing.T) {
	s := NewServer()
	ctx := context.Background()

	tests := []struct {
		name     string
		text1    string
		text2    string
		wantDist int32
	}{
		{"identical", "hello", "hello", 0},
		{"one char", "hello", "hallo", 1},
		{"kitten/sitting", "kitten", "sitting", 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := s.TextSimilarity(ctx, &pb.SimilarityRequest{Text1: tt.text1, Text2: tt.text2})
			if err != nil {
				t.Fatalf("TextSimilarity() error = %v", err)
			}
			if resp.Distance != tt.wantDist {
				t.Errorf("TextSimilarity() distance = %d, want %d", resp.Distance, tt.wantDist)
			}
		})
	}
}

func TestSqlFormat(t *testing.T) {
	s := NewServer()
	ctx := context.Background()

	resp, err := s.SqlFormat(ctx, &pb.SqlRequest{Query: "select * from users where id = 1"})
	if err != nil {
		t.Fatalf("SqlFormat() error = %v", err)
	}
	if !strings.Contains(resp.Formatted, "SELECT") {
		t.Error("SqlFormat() expected uppercase SELECT")
	}
}

func TestIpCalc(t *testing.T) {
	s := NewServer()
	ctx := context.Background()

	tests := []struct {
		name      string
		cidr      string
		wantError bool
	}{
		{"valid cidr", "192.168.1.0/24", false},
		{"single ip", "192.168.1.1", false},
		{"invalid", "invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := s.IpCalc(ctx, &pb.IpRequest{Cidr: tt.cidr})
			if err != nil {
				t.Fatalf("IpCalc() error = %v", err)
			}
			if tt.wantError && resp.Error == "" {
				t.Error("IpCalc() expected error")
			}
			if !tt.wantError && resp.Network == "" {
				t.Error("IpCalc() expected network")
			}
		})
	}
}

func TestCertParse(t *testing.T) {
	s := NewServer()
	ctx := context.Background()

	// Invalid PEM
	resp, err := s.CertParse(ctx, &pb.CertRequest{Data: "invalid"})
	if err != nil {
		t.Fatalf("CertParse() error = %v", err)
	}
	if resp.Error == "" {
		t.Error("CertParse() expected error for invalid PEM")
	}
}
