package spellcheck

import (
	"fmt"
	"strings"
	"unicode"
)

// Issue types reported to callers.
const (
	TypeSpelling    = "spelling"
	TypeGrammar     = "grammar"
	TypePunctuation = "punctuation"
	TypeStyle       = "style"
)

// Issue is a single detected problem in the input text. Offsets and lengths are
// expressed in runes so the frontend can map them onto the original string.
type Issue struct {
	Offset       int
	Length       int
	Text         string
	Type         string
	Rule         string
	Message      string
	Replacements []string
}

// ruleFunc inspects the full rune slice and returns any issues it finds.
type ruleFunc func(runes []rune) []Issue

// ── Shared rules ──────────────────────────────────────────────────────────────

// repeatedWordRule flags a word immediately repeated (separated only by
// whitespace), e.g. "the the". whitelist holds lowercase words that may
// legitimately double (e.g. "had had").
func repeatedWordRule(whitelist map[string]bool) ruleFunc {
	return func(runes []rune) []Issue {
		tokens := tokenizeWords(runes)
		var issues []Issue
		for i := 0; i+1 < len(tokens); i++ {
			a, b := tokens[i], tokens[i+1]
			la := strings.ToLower(a.text)
			if la != strings.ToLower(b.text) {
				continue
			}
			if whitelist[la] || !isAlpha(a.text) {
				continue
			}
			if !wordsOnlyBetween(runes, a.end(), b.offset) {
				continue
			}
			issues = append(issues, Issue{
				Offset:       a.offset,
				Length:       b.end() - a.offset,
				Text:         string(runes[a.offset:b.end()]),
				Type:         TypeGrammar,
				Rule:         "repeated-word",
				Message:      fmt.Sprintf("Repeated word %q.", a.text),
				Replacements: []string{a.text},
			})
		}
		return issues
	}
}

// doubleSpaceRule flags runs of two or more spaces/tabs (not spanning a line
// break) and offers to collapse them to a single space.
func doubleSpaceRule(runes []rune) []Issue {
	var issues []Issue
	n := len(runes)
	for i := 0; i < n; {
		if runes[i] != ' ' && runes[i] != '\t' {
			i++
			continue
		}
		start := i
		for i < n && (runes[i] == ' ' || runes[i] == '\t') {
			i++
		}
		if i-start >= 2 {
			issues = append(issues, Issue{
				Offset:       start,
				Length:       i - start,
				Text:         string(runes[start:i]),
				Type:         TypePunctuation,
				Rule:         "double-space",
				Message:      "Multiple consecutive spaces.",
				Replacements: []string{" "},
			})
		}
	}
	return issues
}

// closingPunctuation are marks that should hug the preceding word.
var closingPunctuation = map[rune]bool{
	',': true, '.': true, ';': true, ':': true, '!': true, '?': true, ')': true,
}

// spaceBeforePunctuationRule flags whitespace inserted before closing
// punctuation, e.g. "word ." → "word.".
func spaceBeforePunctuationRule(runes []rune) []Issue {
	var issues []Issue
	n := len(runes)
	for i := 0; i < n; {
		if runes[i] != ' ' && runes[i] != '\t' {
			i++
			continue
		}
		start := i
		for i < n && (runes[i] == ' ' || runes[i] == '\t') {
			i++
		}
		// Need a word char before the run and closing punctuation after it.
		if start == 0 || i >= n || !closingPunctuation[runes[i]] {
			continue
		}
		if !unicode.IsLetter(runes[start-1]) && !unicode.IsDigit(runes[start-1]) {
			continue
		}
		issues = append(issues, Issue{
			Offset:       start,
			Length:       (i - start) + 1,
			Text:         string(runes[start : i+1]),
			Type:         TypePunctuation,
			Rule:         "space-before-punctuation",
			Message:      fmt.Sprintf("Remove the space before %q.", string(runes[i])),
			Replacements: []string{string(runes[i])},
		})
	}
	return issues
}

// missingSpaceAfterCommaRule flags a comma or semicolon wedged between two
// letters, e.g. "one,two" → "one, two".
func missingSpaceAfterCommaRule(runes []rune) []Issue {
	var issues []Issue
	n := len(runes)
	for i := 1; i+1 < n; i++ {
		if runes[i] != ',' && runes[i] != ';' {
			continue
		}
		if !unicode.IsLetter(runes[i-1]) || !unicode.IsLetter(runes[i+1]) {
			continue
		}
		issues = append(issues, Issue{
			Offset:       i - 1,
			Length:       3,
			Text:         string(runes[i-1 : i+2]),
			Type:         TypePunctuation,
			Rule:         "missing-space-after-punctuation",
			Message:      fmt.Sprintf("Add a space after %q.", string(runes[i])),
			Replacements: []string{string(runes[i-1]) + string(runes[i]) + " " + string(runes[i+1])},
		})
	}
	return issues
}

// sentenceCapitalizationRule flags sentences that begin with a lowercase
// letter. abbrevs holds lowercase abbreviations whose trailing period does not
// end a sentence (e.g. "etc", "mr").
func sentenceCapitalizationRule(abbrevs map[string]bool) ruleFunc {
	return func(runes []rune) []Issue {
		tokens := tokenizeWords(runes)
		var issues []Issue
		for _, t := range tokens {
			first := []rune(t.text)[0]
			if !unicode.IsLetter(first) || !unicode.IsLower(first) {
				continue
			}
			if !isSentenceStart(runes, t.offset) {
				continue
			}
			if precededByAbbreviation(runes, t.offset, abbrevs) {
				continue
			}
			issues = append(issues, Issue{
				Offset:       t.offset,
				Length:       1,
				Text:         string(first),
				Type:         TypeGrammar,
				Rule:         "sentence-capitalization",
				Message:      "Sentence should start with a capital letter.",
				Replacements: []string{string(unicode.ToUpper(first))},
			})
		}
		return issues
	}
}

// precededByAbbreviation guards sentence capitalization against false sentence
// breaks: a period after a known abbreviation, after a single initial (e.g.
// "U.S."), or after a digit (decimals, numbered lists).
func precededByAbbreviation(runes []rune, offset int, abbrevs map[string]bool) bool {
	// Find the terminator immediately preceding this token.
	i := offset - 1
	for i >= 0 && (runes[i] == ' ' || runes[i] == '\t') {
		i--
	}
	if i < 0 || runes[i] != '.' {
		return false
	}
	// Collect the word ending right before the period.
	end := i
	start := end
	for start > 0 && isWordRune(runes[start-1]) {
		start--
	}
	if start == end {
		// Period preceded by a non-word rune; check for digit (e.g. "3.").
		if start-1 >= 0 && unicode.IsDigit(runes[start-1]) {
			return true
		}
		return false
	}
	word := strings.ToLower(strings.TrimFunc(string(runes[start:end]), isApostrophe))
	if abbrevs[word] {
		return true
	}
	// Single-letter initials like "U." in "U.S.".
	if len([]rune(word)) == 1 {
		return true
	}
	return false
}

func isAlpha(s string) bool {
	hasLetter := false
	for _, r := range s {
		if isApostrophe(r) {
			continue
		}
		if !unicode.IsLetter(r) {
			return false
		}
		hasLetter = true
	}
	return hasLetter
}
