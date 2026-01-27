package api

import (
	"context"
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
		{"json", "{\n  \"a\": 1,\n  \"b\": 2\n}", "{\n  \"a\": 1,\n  \"b\": 3\n}", true},
		{"yaml", "a: 1\nb: 2", "a: 1\nb: 3", true},
		{"xml", "<a>\n  1\n</a>\n<b>\n  2\n</b>", "<a>\n  1\n</a>\n<b>\n  3\n</b>", true},
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
