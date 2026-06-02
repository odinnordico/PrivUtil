package spellcheck

import (
	"slices"
	"strings"
	"testing"
)

// findRule returns the first issue matching rule, or nil.
func findRule(issues []Issue, rule string) *Issue {
	for i := range issues {
		if issues[i].Rule == rule {
			return &issues[i]
		}
	}
	return nil
}

// findText returns the first issue whose flagged text equals text, or nil.
func findText(issues []Issue, text string) *Issue {
	for i := range issues {
		if issues[i].Text == text {
			return &issues[i]
		}
	}
	return nil
}

func TestSpelling(t *testing.T) {
	issues, lang := Check("i have a tset here", "en")
	if lang != "en" {
		t.Fatalf("lang = %q, want en", lang)
	}
	is := findText(issues, "tset")
	if is == nil {
		t.Fatalf("expected a spelling issue for 'tset', got %+v", issues)
	}
	if is.Type != TypeSpelling {
		t.Errorf("type = %q, want spelling", is.Type)
	}
	if !slices.Contains(is.Replacements, "test") {
		t.Errorf("expected 'test' among suggestions, got %v", is.Replacements)
	}
}

func TestSpellingPreservesCase(t *testing.T) {
	// Capitalized misspelling at sentence start keeps a capitalized suggestion.
	issues, _ := Check("Tset the code.", "en")
	is := findText(issues, "Tset")
	if is == nil {
		t.Fatalf("expected issue for 'Tset', got %+v", issues)
	}
	if !slices.Contains(is.Replacements, "Test") {
		t.Errorf("expected 'Test' (capitalized) among suggestions, got %v", is.Replacements)
	}
}

func TestProperNounNotFlaggedWithoutStrongMatch(t *testing.T) {
	// A capitalized unknown word mid-sentence with no close correction is
	// treated as a proper noun and left alone.
	issues, _ := Check("I met Zyqxwv yesterday.", "en")
	if is := findText(issues, "Zyqxwv"); is != nil {
		t.Errorf("proper-noun-like token should not be flagged, got %+v", is)
	}
}

func TestRepeatedWord(t *testing.T) {
	issues, _ := Check("the the cat", "en")
	is := findRule(issues, "repeated-word")
	if is == nil {
		t.Fatalf("expected repeated-word issue, got %+v", issues)
	}
	if !slices.Contains(is.Replacements, "the") {
		t.Errorf("expected replacement 'the', got %v", is.Replacements)
	}
}

func TestArticleAAn(t *testing.T) {
	cases := []struct {
		text string
		want string
	}{
		{"I ate a apple", "an"},
		{"It was an car", "a"},
		{"an hour ago", ""},  // silent h → correct
		{"a university", ""}, // long-u → correct
	}
	for _, c := range cases {
		issues, _ := Check(c.text, "en")
		is := findRule(issues, "article-a-an")
		if c.want == "" {
			if is != nil {
				t.Errorf("%q: expected no article issue, got %+v", c.text, is)
			}
			continue
		}
		if is == nil {
			t.Fatalf("%q: expected article issue", c.text)
		}
		if !slices.Contains(is.Replacements, c.want) {
			t.Errorf("%q: replacement %v, want %q", c.text, is.Replacements, c.want)
		}
	}
}

func TestCapitalizeI(t *testing.T) {
	issues, _ := Check("yesterday i went home", "en")
	is := findRule(issues, "capitalize-i")
	if is == nil {
		t.Fatalf("expected capitalize-i issue, got %+v", issues)
	}
	if is.Length != 1 || !slices.Contains(is.Replacements, "I") {
		t.Errorf("unexpected issue %+v", is)
	}
}

func TestWouldOf(t *testing.T) {
	issues, _ := Check("I would of known", "en")
	is := findRule(issues, "would-of")
	if is == nil {
		t.Fatalf("expected would-of issue, got %+v", issues)
	}
	if !slices.Contains(is.Replacements, "have") {
		t.Errorf("expected 'have', got %v", is.Replacements)
	}
}

func TestDoubleSpaceAndSpaceBeforePunctuation(t *testing.T) {
	issues, _ := Check("hello  world .", "en")
	if findRule(issues, "double-space") == nil {
		t.Errorf("expected double-space issue, got %+v", issues)
	}
	if findRule(issues, "space-before-punctuation") == nil {
		t.Errorf("expected space-before-punctuation issue, got %+v", issues)
	}
}

func TestMissingSpaceAfterComma(t *testing.T) {
	issues, _ := Check("apples,oranges", "en")
	is := findRule(issues, "missing-space-after-punctuation")
	if is == nil {
		t.Fatalf("expected missing-space issue, got %+v", issues)
	}
	if !slices.Contains(is.Replacements, "s, o") {
		t.Errorf("replacement %v, want 's, o'", is.Replacements)
	}
}

func TestSentenceCapitalization(t *testing.T) {
	issues, _ := Check("hello there. how are you", "en")
	caps := 0
	for _, is := range issues {
		if is.Rule == "sentence-capitalization" {
			caps++
		}
	}
	// Both "hello" and "how" start sentences lowercase.
	if caps < 2 {
		t.Errorf("expected 2 sentence-capitalization issues, got %d (%+v)", caps, issues)
	}
}

func TestSentenceCapitalizationSkipsAbbreviation(t *testing.T) {
	issues, _ := Check("Email me at the address etc. then we talk.", "en")
	for _, is := range issues {
		if is.Rule == "sentence-capitalization" && is.Text == "t" {
			t.Errorf("should not flag 'then' after abbreviation 'etc.': %+v", is)
		}
	}
}

func TestSpanishOpeningMark(t *testing.T) {
	issues, lang := Check("como estas?", "es")
	if lang != "es" {
		t.Fatalf("lang = %q, want es", lang)
	}
	is := findRule(issues, "missing-opening-mark")
	if is == nil {
		t.Fatalf("expected missing-opening-mark issue, got %+v", issues)
	}
	if !slices.Contains(is.Replacements, "¿como") {
		t.Errorf("replacement %v, want '¿como'", is.Replacements)
	}
}

func TestSpanishOpeningMarkAfterComma(t *testing.T) {
	// The mark anchors at the question clause, after the comma.
	issues, _ := Check("Hola, como estas?", "es")
	is := findRule(issues, "missing-opening-mark")
	if is == nil {
		t.Fatalf("expected missing-opening-mark issue, got %+v", issues)
	}
	if is.Text != "como" {
		t.Errorf("anchor text = %q, want 'como'", is.Text)
	}
}

func TestSpanishOpeningMarkNotDuplicated(t *testing.T) {
	issues, _ := Check("¿Cómo estás?", "es")
	if is := findRule(issues, "missing-opening-mark"); is != nil {
		t.Errorf("well-formed question should not be flagged, got %+v", is)
	}
}

func TestSpanishSpelling(t *testing.T) {
	issues, _ := Check("Tengo una compuradora nueva.", "es")
	is := findText(issues, "compuradora")
	if is == nil {
		t.Fatalf("expected spelling issue for 'compuradora', got %+v", issues)
	}
	if !slices.Contains(is.Replacements, "computadora") {
		t.Errorf("expected 'computadora' among suggestions, got %v", is.Replacements)
	}
}

func TestContractionsNotFlagged(t *testing.T) {
	// Common contractions and possessives must not be reported as misspellings.
	text := "I don't think it's the dog's fault; they're sure we've won't and can't."
	issues, _ := Check(text, "en")
	for _, w := range []string{"don't", "it's", "dog's", "they're", "we've", "won't", "can't"} {
		if is := findText(issues, w); is != nil && is.Type == TypeSpelling {
			t.Errorf("contraction %q should not be flagged as a spelling mistake", w)
		}
	}
}

func TestUnknownLanguageFallsBackToEnglish(t *testing.T) {
	_, lang := Check("hello", "xx")
	if lang != "en" {
		t.Errorf("lang = %q, want en (fallback)", lang)
	}
}

func TestLanguages(t *testing.T) {
	langs := Languages()
	if len(langs) < 2 {
		t.Fatalf("expected at least 2 languages, got %d", len(langs))
	}
	codes := make([]string, len(langs))
	for i, l := range langs {
		codes[i] = l.Code
	}
	for _, want := range []string{"en", "es"} {
		if !slices.Contains(codes, want) {
			t.Errorf("missing language %q in %v", want, codes)
		}
	}
}

func TestOffsetsAreRuneBased(t *testing.T) {
	// Accented prefix shifts byte offsets but not rune offsets.
	text := "café  end" // double space after accented word
	issues, _ := Check(text, "en")
	is := findRule(issues, "double-space")
	if is == nil {
		t.Fatalf("expected double-space issue, got %+v", issues)
	}
	got := []rune(text)[is.Offset : is.Offset+is.Length]
	if strings.TrimSpace(string(got)) != "" || len(got) != 2 {
		t.Errorf("offset/length %d/%d do not map to the double space (got %q)", is.Offset, is.Length, string(got))
	}
}
