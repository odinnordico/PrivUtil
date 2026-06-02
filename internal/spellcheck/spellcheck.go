package spellcheck

import (
	"sort"
	"unicode"
)

// DefaultLanguage is used when a request omits or names an unknown language.
const DefaultLanguage = "en"

// Language is a self-contained pack: a dictionary, a spelling corrector built
// on it, and an ordered set of grammar rules.
type Language struct {
	Code  string
	Label string

	dict    *Dictionary
	speller *speller
	rules   []ruleFunc
}

// LanguageInfo is the public description of an available language.
type LanguageInfo struct {
	Code  string
	Label string
}

var (
	registry = map[string]*Language{}
	order    []string
)

func register(l *Language) {
	registry[l.Code] = l
	order = append(order, l.Code)
}

func init() {
	register(newEnglish())
	register(newSpanish())
}

// Languages lists the available languages in registration order.
func Languages() []LanguageInfo {
	out := make([]LanguageInfo, 0, len(order))
	for _, code := range order {
		l := registry[code]
		out = append(out, LanguageInfo{Code: l.Code, Label: l.Label})
	}
	return out
}

// Check runs spelling and grammar checks over text in the given language.
// It returns the issues, the resolved language code, and (for symmetry with
// other engines) an error slot that is currently always nil.
func Check(text, lang string) ([]Issue, string) {
	if lang == "" {
		lang = DefaultLanguage
	}
	l, ok := registry[lang]
	if !ok {
		l, lang = registry[DefaultLanguage], DefaultLanguage
	}
	return l.check([]rune(text)), lang
}

// WordCount counts word tokens in text (language-independent).
func WordCount(text string) int {
	return len(tokenizeWords([]rune(text)))
}

func (l *Language) check(runes []rune) []Issue {
	issues := l.spellingIssues(runes)
	for _, rule := range l.rules {
		issues = append(issues, rule(runes)...)
	}
	return sortAndDedupe(issues)
}

// spellingIssues flags word tokens that are not in the dictionary, with ranked
// suggestions. Capitalized words that aren't at a sentence start are treated as
// likely proper nouns and only flagged when a strong (edit-distance-1)
// correction exists.
func (l *Language) spellingIssues(runes []rune) []Issue {
	var issues []Issue
	for _, t := range tokenizeWords(runes) {
		w := trimApostrophes(t.text)
		rw := []rune(w)
		if len(rw) < 2 || hasDigit(w) {
			continue
		}
		if isAllUpper(w) {
			continue // acronyms
		}
		if l.dict.contains(w) || acceptedContraction(l.dict, w) {
			continue
		}

		titleCased := unicode.IsUpper(rw[0]) && !isAllUpper(w)
		atStart := isSentenceStart(runes, t.offset)

		var suggestions []string
		if titleCased && !atStart {
			// Likely a proper noun; require a high-confidence correction.
			if suggestions = l.speller.suggest(w, true); len(suggestions) == 0 {
				continue
			}
		} else {
			suggestions = l.speller.suggest(w, false)
		}

		issues = append(issues, Issue{
			Offset:       t.offset,
			Length:       t.length,
			Text:         t.text,
			Type:         TypeSpelling,
			Rule:         "spelling",
			Message:      "Possible spelling mistake.",
			Replacements: suggestions,
		})
	}
	return issues
}

// sortAndDedupe orders issues by position and removes exact duplicates
// (same offset, length and rule).
func sortAndDedupe(issues []Issue) []Issue {
	sort.SliceStable(issues, func(i, j int) bool {
		if issues[i].Offset != issues[j].Offset {
			return issues[i].Offset < issues[j].Offset
		}
		return issues[i].Length < issues[j].Length
	})
	out := issues[:0]
	type key struct {
		off, length int
		rule        string
	}
	seen := map[key]struct{}{}
	for _, is := range issues {
		k := key{is.Offset, is.Length, is.Rule}
		if _, dup := seen[k]; dup {
			continue
		}
		seen[k] = struct{}{}
		out = append(out, is)
	}
	return out
}
