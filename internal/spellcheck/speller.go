package spellcheck

import (
	"sort"
	"strings"
	"unicode"
)

// maxSuggestions caps how many corrections are offered per misspelled word.
const maxSuggestions = 5

// speller implements a Norvig-style edit-distance spelling corrector backed by
// a frequency dictionary. Candidates are generated within edit distance 1 (and
// 2 as a fallback), filtered to known words, and ranked by frequency.
type speller struct {
	dict     *Dictionary
	alphabet []rune
}

func newSpeller(dict *Dictionary, alphabet string) *speller {
	return &speller{dict: dict, alphabet: []rune(alphabet)}
}

// edits1 returns all strings within edit distance 1 of word (deletes,
// transposes, replaces, inserts), de-duplicated.
func (s *speller) edits1(word string) []string {
	rs := []rune(word)
	n := len(rs)
	seen := make(map[string]struct{}, (2*n+1)*len(s.alphabet))
	out := make([]string, 0, (2*n+1)*len(s.alphabet))
	add := func(w string) {
		if _, ok := seen[w]; !ok {
			seen[w] = struct{}{}
			out = append(out, w)
		}
	}

	// deletes
	for i := 0; i < n; i++ {
		add(string(rs[:i]) + string(rs[i+1:]))
	}
	// transposes
	for i := 0; i+1 < n; i++ {
		t := make([]rune, n)
		copy(t, rs)
		t[i], t[i+1] = t[i+1], t[i]
		add(string(t))
	}
	// replaces
	for i := 0; i < n; i++ {
		for _, c := range s.alphabet {
			if c == rs[i] {
				continue
			}
			add(string(rs[:i]) + string(c) + string(rs[i+1:]))
		}
	}
	// inserts
	for i := 0; i <= n; i++ {
		for _, c := range s.alphabet {
			add(string(rs[:i]) + string(c) + string(rs[i:]))
		}
	}
	return out
}

// rank sorts known words by descending frequency (ties broken alphabetically)
// and trims to maxSuggestions.
func (s *speller) rank(words []string) []string {
	sort.Slice(words, func(i, j int) bool {
		fi, fj := s.dict.freq(words[i]), s.dict.freq(words[j])
		if fi != fj {
			return fi > fj
		}
		return words[i] < words[j]
	})
	if len(words) > maxSuggestions {
		words = words[:maxSuggestions]
	}
	return words
}

// suggest returns ranked corrections for word, case-matched to the original.
// onlyStrong restricts results to edit distance 1 (high-confidence) matches,
// which is used to avoid second-guessing likely proper nouns.
func (s *speller) suggest(word string, onlyStrong bool) []string {
	lw := strings.ToLower(word)
	e1 := s.edits1(lw)

	if known := s.dict.knownSet(e1); len(known) > 0 {
		return matchCaseAll(s.rank(known), word)
	}
	if onlyStrong || len([]rune(lw)) > 15 {
		return nil
	}

	// edit distance 2: edits of the edits.
	seen := make(map[string]struct{}, 1<<14)
	e2 := make([]string, 0, 1<<14)
	for _, w := range e1 {
		for _, w2 := range s.edits1(w) {
			if _, ok := seen[w2]; !ok {
				seen[w2] = struct{}{}
				e2 = append(e2, w2)
			}
		}
	}
	if known := s.dict.knownSet(e2); len(known) > 0 {
		return matchCaseAll(s.rank(known), word)
	}
	return nil
}

// matchCaseAll applies the capitalization pattern of original to each word.
func matchCaseAll(words []string, original string) []string {
	out := make([]string, len(words))
	for i, w := range words {
		out[i] = matchCase(w, original)
	}
	return out
}

// matchCase mirrors original's casing onto candidate: ALL CAPS → upper,
// Titlecase → title, otherwise lowercase as-is.
func matchCase(candidate, original string) string {
	ro := []rune(original)
	if len(ro) == 0 {
		return candidate
	}
	if isAllUpper(original) && len(ro) > 1 {
		return strings.ToUpper(candidate)
	}
	if unicode.IsUpper(ro[0]) {
		rc := []rune(candidate)
		rc[0] = unicode.ToUpper(rc[0])
		return string(rc)
	}
	return candidate
}

func isAllUpper(word string) bool {
	hasLetter := false
	for _, r := range word {
		if unicode.IsLetter(r) {
			hasLetter = true
			if !unicode.IsUpper(r) {
				return false
			}
		}
	}
	return hasLetter
}
