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
