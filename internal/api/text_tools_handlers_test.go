package api

import (
	"context"
	"strings"
	"testing"

	pb "github.com/odinnordico/privutil/proto"
)

var textSrv = &Server{}

// ─── Slugify ──────────────────────────────────────────────────────────────────

func TestSlugify_Basic(t *testing.T) {
	resp, err := textSrv.Slugify(context.Background(), &pb.SlugifyRequest{Text: "Hello World"})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Result != "hello-world" {
		t.Errorf("got %q, want %q", resp.Result, "hello-world")
	}
}

func TestSlugify_Diacritics(t *testing.T) {
	resp, _ := textSrv.Slugify(context.Background(), &pb.SlugifyRequest{Text: "Ünïcödé"})
	if resp.Result != "unicode" {
		t.Errorf("got %q, want %q", resp.Result, "unicode")
	}
}

func TestSlugify_CustomSeparator(t *testing.T) {
	resp, _ := textSrv.Slugify(context.Background(), &pb.SlugifyRequest{Text: "foo bar", Separator: "_"})
	if resp.Result != "foo_bar" {
		t.Errorf("got %q, want %q", resp.Result, "foo_bar")
	}
}

func TestSlugify_NoSeparator(t *testing.T) {
	resp, _ := textSrv.Slugify(context.Background(), &pb.SlugifyRequest{Text: "foo bar", Separator: "none"})
	if resp.Result != "foobar" {
		t.Errorf("got %q, want %q", resp.Result, "foobar")
	}
}

func TestSlugify_Uppercase(t *testing.T) {
	resp, _ := textSrv.Slugify(context.Background(), &pb.SlugifyRequest{Text: "Hello World", Uppercase: true})
	if resp.Result != "HELLO-WORLD" {
		t.Errorf("got %q, want %q", resp.Result, "HELLO-WORLD")
	}
}

func TestSlugify_MaxLen(t *testing.T) {
	resp, _ := textSrv.Slugify(context.Background(), &pb.SlugifyRequest{Text: "hello world foo bar", MaxLen: 11})
	if len(resp.Result) > 11 {
		t.Errorf("result %q exceeds max length 11", resp.Result)
	}
}

func TestSlugify_EmptyInput(t *testing.T) {
	resp, _ := textSrv.Slugify(context.Background(), &pb.SlugifyRequest{Text: ""})
	if resp.Error == "" {
		t.Error("expected error for empty input")
	}
}

// ─── Hidden character detector ────────────────────────────────────────────────

func TestHiddenChars_NoHidden(t *testing.T) {
	resp, err := textSrv.HiddenChars(context.Background(), &pb.HiddenCharsRequest{Text: "normal text"})
	if err != nil {
		t.Fatal(err)
	}
	if resp.HasHidden {
		t.Error("expected no hidden chars")
	}
	if resp.Cleaned != "normal text" {
		t.Errorf("cleaned mismatch: %q", resp.Cleaned)
	}
}

func TestHiddenChars_ZeroWidthSpace(t *testing.T) {
	input := "hello​world"
	resp, _ := textSrv.HiddenChars(context.Background(), &pb.HiddenCharsRequest{Text: input})
	if !resp.HasHidden {
		t.Error("expected hidden chars found")
	}
	if resp.Cleaned != "helloworld" {
		t.Errorf("cleaned got %q", resp.Cleaned)
	}
	if !strings.Contains(resp.Annotated, "[U+200B]") {
		t.Errorf("annotated should contain [U+200B], got %q", resp.Annotated)
	}
}

func TestHiddenChars_BOM(t *testing.T) {
	input := "\uFEFF" + "text"
	resp, _ := textSrv.HiddenChars(context.Background(), &pb.HiddenCharsRequest{Text: input})
	if !resp.HasHidden {
		t.Error("expected BOM detected")
	}
	if resp.Cleaned != "text" {
		t.Errorf("cleaned got %q", resp.Cleaned)
	}
}

func TestHiddenChars_Empty(t *testing.T) {
	resp, _ := textSrv.HiddenChars(context.Background(), &pb.HiddenCharsRequest{Text: ""})
	if resp.HasHidden {
		t.Error("empty string should not have hidden chars")
	}
}

// ─── Text replacer ────────────────────────────────────────────────────────────

func TestTextReplace_Literal(t *testing.T) {
	resp, err := textSrv.TextReplace(context.Background(), &pb.TextReplaceRequest{
		Text: "foo bar foo", Find: "foo", ReplaceWith: "baz",
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Result != "baz bar baz" {
		t.Errorf("got %q", resp.Result)
	}
	if resp.Count != 2 {
		t.Errorf("count got %d, want 2", resp.Count)
	}
}

func TestTextReplace_CaseInsensitive(t *testing.T) {
	resp, _ := textSrv.TextReplace(context.Background(), &pb.TextReplaceRequest{
		Text: "Foo foo FOO", Find: "foo", ReplaceWith: "bar", CaseInsensitive: true,
	})
	if resp.Count != 3 {
		t.Errorf("count got %d, want 3", resp.Count)
	}
}

func TestTextReplace_Regex(t *testing.T) {
	resp, _ := textSrv.TextReplace(context.Background(), &pb.TextReplaceRequest{
		Text: "abc 123 def", Find: `\d+`, ReplaceWith: "NUM", UseRegex: true,
	})
	if resp.Result != "abc NUM def" {
		t.Errorf("got %q", resp.Result)
	}
	if resp.Count != 1 {
		t.Errorf("count got %d, want 1", resp.Count)
	}
}

func TestTextReplace_InvalidRegex(t *testing.T) {
	resp, _ := textSrv.TextReplace(context.Background(), &pb.TextReplaceRequest{
		Text: "text", Find: "[invalid", UseRegex: true,
	})
	if resp.Error == "" {
		t.Error("expected error for invalid regex")
	}
}

func TestTextReplace_EmptyFind(t *testing.T) {
	resp, _ := textSrv.TextReplace(context.Background(), &pb.TextReplaceRequest{
		Text: "text", Find: "",
	})
	if resp.Error == "" {
		t.Error("expected error for empty find pattern")
	}
}

// ─── String obfuscator ────────────────────────────────────────────────────────

func TestStringObfuscate_Default(t *testing.T) {
	resp, err := textSrv.StringObfuscate(context.Background(), &pb.StringObfuscateRequest{
		Text: "hello world", KeepStart: 2, KeepEnd: 2,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(resp.Result, "he") {
		t.Errorf("should start with 'he', got %q", resp.Result)
	}
	if !strings.HasSuffix(resp.Result, "ld") {
		t.Errorf("should end with 'ld', got %q", resp.Result)
	}
	middle := resp.Result[2 : len(resp.Result)-2]
	for _, c := range middle {
		if c != '*' {
			t.Errorf("middle should be '*', got %q", string(c))
		}
	}
}

func TestStringObfuscate_CustomMask(t *testing.T) {
	resp, _ := textSrv.StringObfuscate(context.Background(), &pb.StringObfuscateRequest{
		Text: "secret", KeepStart: 1, KeepEnd: 1, MaskChar: "#",
	})
	if resp.Result != "s####t" {
		t.Errorf("got %q, want %q", resp.Result, "s####t")
	}
}

func TestStringObfuscate_FullVisible(t *testing.T) {
	resp, _ := textSrv.StringObfuscate(context.Background(), &pb.StringObfuscateRequest{
		Text: "hi", KeepStart: 2, KeepEnd: 2,
	})
	if resp.Result != "hi" {
		t.Errorf("got %q, want %q", resp.Result, "hi")
	}
}

func TestStringObfuscate_Empty(t *testing.T) {
	resp, _ := textSrv.StringObfuscate(context.Background(), &pb.StringObfuscateRequest{Text: ""})
	if resp.Error == "" {
		t.Error("expected error for empty input")
	}
}

// ─── Numeronym generator ──────────────────────────────────────────────────────

func TestNumeronymGenerate_Basic(t *testing.T) {
	resp, err := textSrv.NumeronymGenerate(context.Background(), &pb.NumeronymRequest{Text: "internationalization"})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Result != "i18n" {
		t.Errorf("got %q, want %q", resp.Result, "i18n")
	}
}

func TestNumeronymGenerate_ShortWord(t *testing.T) {
	resp, _ := textSrv.NumeronymGenerate(context.Background(), &pb.NumeronymRequest{Text: "hi"})
	if resp.Result != "hi" {
		t.Errorf("short word should be unchanged: got %q", resp.Result)
	}
}

func TestNumeronymGenerate_MultipleWords(t *testing.T) {
	resp, _ := textSrv.NumeronymGenerate(context.Background(), &pb.NumeronymRequest{Text: "kubernetes accessibility"})
	if resp.Result != "k8s a11y" {
		t.Errorf("got %q, want %q", resp.Result, "k8s a11y")
	}
}

func TestNumeronymGenerate_Empty(t *testing.T) {
	resp, _ := textSrv.NumeronymGenerate(context.Background(), &pb.NumeronymRequest{Text: ""})
	if resp.Error == "" {
		t.Error("expected error for empty input")
	}
}

// ─── NATO alphabet ────────────────────────────────────────────────────────────

func TestNatoAlphabet_Encode(t *testing.T) {
	resp, err := textSrv.NatoAlphabet(context.Background(), &pb.NatoRequest{Text: "SOS", Action: "encode"})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Result != "Sierra Oscar Sierra" {
		t.Errorf("got %q", resp.Result)
	}
}

func TestNatoAlphabet_EncodeWithNumbers(t *testing.T) {
	resp, _ := textSrv.NatoAlphabet(context.Background(), &pb.NatoRequest{Text: "A1", Action: "encode"})
	if resp.Result != "Alfa One" {
		t.Errorf("got %q", resp.Result)
	}
}

func TestNatoAlphabet_Decode(t *testing.T) {
	resp, _ := textSrv.NatoAlphabet(context.Background(), &pb.NatoRequest{Text: "Sierra Oscar Sierra", Action: "decode"})
	if resp.Result != "SOS" {
		t.Errorf("got %q, want SOS", resp.Result)
	}
}

func TestNatoAlphabet_UnknownWord(t *testing.T) {
	resp, _ := textSrv.NatoAlphabet(context.Background(), &pb.NatoRequest{Text: "Unknown", Action: "decode"})
	if resp.Error == "" {
		t.Error("expected error for unknown NATO word")
	}
}

func TestNatoAlphabet_Empty(t *testing.T) {
	resp, _ := textSrv.NatoAlphabet(context.Background(), &pb.NatoRequest{Text: ""})
	if resp.Error == "" {
		t.Error("expected error for empty input")
	}
}

// ─── List processor ───────────────────────────────────────────────────────────

func TestListProcess_SortAZ(t *testing.T) {
	resp, err := textSrv.ListProcess(context.Background(), &pb.ListRequest{
		Text: "banana\napple\ncherry", Action: pb.ListAction_LIST_SORT_AZ,
	})
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(resp.Result, "\n")
	if lines[0] != "apple" || lines[1] != "banana" || lines[2] != "cherry" {
		t.Errorf("sort failed: %v", lines)
	}
}

func TestListProcess_SortZA(t *testing.T) {
	resp, _ := textSrv.ListProcess(context.Background(), &pb.ListRequest{
		Text: "banana\napple\ncherry", Action: pb.ListAction_LIST_SORT_ZA,
	})
	lines := strings.Split(resp.Result, "\n")
	if lines[0] != "cherry" || lines[2] != "apple" {
		t.Errorf("sort ZA failed: %v", lines)
	}
}

func TestListProcess_Dedupe(t *testing.T) {
	resp, _ := textSrv.ListProcess(context.Background(), &pb.ListRequest{
		Text: "a\nb\na\nc\nb", Action: pb.ListAction_LIST_DEDUPE,
	})
	lines := strings.Split(resp.Result, "\n")
	if len(lines) != 3 {
		t.Errorf("expected 3 unique lines, got %d: %v", len(lines), lines)
	}
	if resp.OutputCount != 3 {
		t.Errorf("output count got %d, want 3", resp.OutputCount)
	}
}

func TestListProcess_UniqueOnly(t *testing.T) {
	resp, _ := textSrv.ListProcess(context.Background(), &pb.ListRequest{
		Text: "a\nb\na\nc", Action: pb.ListAction_LIST_UNIQUE_ONLY,
	})
	lines := strings.Split(resp.Result, "\n")
	if len(lines) != 2 {
		t.Errorf("expected 2 unique-only lines, got %d", len(lines))
	}
}

func TestListProcess_Duplicates(t *testing.T) {
	resp, _ := textSrv.ListProcess(context.Background(), &pb.ListRequest{
		Text: "a\nb\na\nc\nb", Action: pb.ListAction_LIST_DUPLICATES,
	})
	lines := strings.Split(resp.Result, "\n")
	if len(lines) != 2 {
		t.Errorf("expected 2 duplicate lines, got %d: %v", len(lines), lines)
	}
}

func TestListProcess_Frequency(t *testing.T) {
	resp, _ := textSrv.ListProcess(context.Background(), &pb.ListRequest{
		Text: "a\nb\na\na\nb", Action: pb.ListAction_LIST_FREQUENCY,
	})
	if len(resp.Frequency) != 2 {
		t.Errorf("expected 2 frequency items, got %d", len(resp.Frequency))
	}
	if resp.Frequency[0].Line != "a" || resp.Frequency[0].Count != 3 {
		t.Errorf("top item should be a(3), got %v", resp.Frequency[0])
	}
}

func TestListProcess_Reverse(t *testing.T) {
	resp, _ := textSrv.ListProcess(context.Background(), &pb.ListRequest{
		Text: "a\nb\nc", Action: pb.ListAction_LIST_REVERSE,
	})
	if resp.Result != "c\nb\na" {
		t.Errorf("reverse got %q", resp.Result)
	}
}

func TestListProcess_Trim(t *testing.T) {
	resp, _ := textSrv.ListProcess(context.Background(), &pb.ListRequest{
		Text: "  hello  \n  world  ", Action: pb.ListAction_LIST_TRIM,
	})
	lines := strings.Split(resp.Result, "\n")
	if lines[0] != "hello" || lines[1] != "world" {
		t.Errorf("trim failed: %v", lines)
	}
}

func TestListProcess_RemoveEmpty(t *testing.T) {
	resp, _ := textSrv.ListProcess(context.Background(), &pb.ListRequest{
		Text: "a\n\nb\n  \nc", Action: pb.ListAction_LIST_REMOVE_EMPTY,
	})
	if resp.OutputCount != 3 {
		t.Errorf("expected 3 non-empty lines, got %d", resp.OutputCount)
	}
}

func TestListProcess_Shuffle(t *testing.T) {
	input := "a\nb\nc\nd\ne\nf\ng\nh"
	resp, err := textSrv.ListProcess(context.Background(), &pb.ListRequest{
		Text: input, Action: pb.ListAction_LIST_SHUFFLE,
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.OutputCount != 8 {
		t.Errorf("shuffle changed count, got %d", resp.OutputCount)
	}
}

func TestListProcess_SortNumeric(t *testing.T) {
	resp, _ := textSrv.ListProcess(context.Background(), &pb.ListRequest{
		Text: "10\n2\n30\n4", Action: pb.ListAction_LIST_SORT_NUMERIC,
	})
	lines := strings.Split(resp.Result, "\n")
	if lines[0] != "2" || lines[3] != "30" {
		t.Errorf("numeric sort failed: %v", lines)
	}
}

func TestListProcess_CaseInsensitiveDedupe(t *testing.T) {
	resp, _ := textSrv.ListProcess(context.Background(), &pb.ListRequest{
		Text: "Apple\napple\nAPPLE", Action: pb.ListAction_LIST_DEDUPE, CaseInsensitive: true,
	})
	if resp.OutputCount != 1 {
		t.Errorf("case-insensitive dedupe got %d, want 1", resp.OutputCount)
	}
}

func TestListProcess_Empty(t *testing.T) {
	resp, _ := textSrv.ListProcess(context.Background(), &pb.ListRequest{Text: ""})
	if resp.Error == "" {
		t.Error("expected error for empty input")
	}
}
