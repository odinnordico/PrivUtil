package api

import (
	"context"
	"strings"
	"testing"

	pb "github.com/odinnordico/privutil/proto"
)

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
		name         string
		data         string
		source       pb.DataFormat
		target       pb.DataFormat
		csvDelimiter string
		csvNoHeader  bool
		wantError    bool
		wantContains string
	}{
		{"json to yaml", `{"key":"value"}`, pb.DataFormat_JSON, pb.DataFormat_YAML, "", false, false, "key:"},
		{"yaml to json", "key: value", pb.DataFormat_YAML, pb.DataFormat_JSON, "", false, false, `"key"`},
		{"json to xml", `{"root":"value"}`, pb.DataFormat_JSON, pb.DataFormat_XML, "", false, false, "<root>"},
		{"json to toml", `{"title":"hello","count":3}`, pb.DataFormat_JSON, pb.DataFormat_TOML, "", false, false, "title"},
		{"toml to json", "title = \"hello\"\ncount = 3", pb.DataFormat_TOML, pb.DataFormat_JSON, "", false, false, `"title"`},
		{"toml to yaml", "title = \"hello\"", pb.DataFormat_TOML, pb.DataFormat_YAML, "", false, false, "title:"},
		{"csv to json (with header)", "name,age\nAlice,30\nBob,25", pb.DataFormat_CSV, pb.DataFormat_JSON, ",", false, false, `"name"`},
		{"csv to yaml (with header)", "name,age\nAlice,30", pb.DataFormat_CSV, pb.DataFormat_YAML, ",", false, false, "name:"},
		{"csv to json (no header)", "Alice,30\nBob,25", pb.DataFormat_CSV, pb.DataFormat_JSON, ",", true, false, "Alice"},
		{"csv semicolon delimiter", "name;age\nAlice;30", pb.DataFormat_CSV, pb.DataFormat_JSON, ";", false, false, `"name"`},
		{"json array to csv", `[{"name":"Alice","age":30},{"name":"Bob","age":25}]`, pb.DataFormat_JSON, pb.DataFormat_CSV, ",", false, false, "Alice"},
		{"invalid json source", `{invalid}`, pb.DataFormat_JSON, pb.DataFormat_YAML, "", false, true, ""},
		{"invalid toml source", `[invalid toml`, pb.DataFormat_TOML, pb.DataFormat_JSON, "", false, true, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := s.Convert(ctx, &pb.ConvertRequest{
				Data:         tt.data,
				SourceFormat: tt.source,
				TargetFormat: tt.target,
				CsvDelimiter: tt.csvDelimiter,
				CsvNoHeader:  tt.csvNoHeader,
			})
			if err != nil {
				t.Fatalf("Convert() error = %v", err)
			}
			if tt.wantError {
				if resp.Error == "" {
					t.Error("Convert() expected error but got none")
				}
				return
			}
			if resp.Error != "" {
				t.Errorf("Convert() unexpected error = %v", resp.Error)
			}
			if resp.Data == "" {
				t.Error("Convert() expected non-empty result")
			}
			if tt.wantContains != "" && !strings.Contains(resp.Data, tt.wantContains) {
				t.Errorf("Convert() output %q does not contain %q", resp.Data, tt.wantContains)
			}
		})
	}
}

func TestValidateData(t *testing.T) {
	s := NewServer()
	ctx := context.Background()

	tests := []struct {
		name      string
		data      string
		format    pb.DataFormat
		wantValid bool
		wantLine  int32
	}{
		{"valid json", `{"key":"value"}`, pb.DataFormat_JSON, true, 0},
		{"invalid json", `{"key":}`, pb.DataFormat_JSON, false, 0},
		{"valid yaml", "key: value", pb.DataFormat_YAML, true, 0},
		{"invalid yaml", "key: :\n  bad", pb.DataFormat_YAML, false, 0},
		{"valid xml", `<root><child>val</child></root>`, pb.DataFormat_XML, true, 0},
		{"invalid xml", `<root><child>val</root>`, pb.DataFormat_XML, false, 0},
		{"valid toml", "title = \"hello\"\ncount = 3", pb.DataFormat_TOML, true, 0},
		{"invalid toml", "title = ", pb.DataFormat_TOML, false, 0},
		{"empty input", "", pb.DataFormat_JSON, false, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := s.ValidateData(ctx, &pb.ValidateRequest{
				Data:   tt.data,
				Format: tt.format,
			})
			if err != nil {
				t.Fatalf("ValidateData() error = %v", err)
			}
			if resp.Valid != tt.wantValid {
				t.Errorf("ValidateData() valid = %v, want %v (error: %s)", resp.Valid, tt.wantValid, resp.Error)
			}
			if !tt.wantValid && resp.Error == "" {
				t.Error("ValidateData() expected error message for invalid input")
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
