package spellcheck

import (
	"strings"
	"unicode"
)

// token is a word occurrence located by rune offset within the source text.
type token struct {
	text   string // original substring (including internal apostrophes)
	offset int    // rune offset of the first rune
	length int    // rune length
}

func (t token) end() int { return t.offset + t.length }

// isWordRune reports whether r can be part of a word (letters and the
// apostrophes used in contractions).
func isWordRune(r rune) bool {
	return unicode.IsLetter(r) || r == '\'' || r == '’'
}

// tokenizeWords splits runes into word tokens, trimming leading/trailing
// apostrophes that aren't part of the word itself (e.g. quotes).
func tokenizeWords(runes []rune) []token {
	var tokens []token
	i, n := 0, len(runes)
	for i < n {
		if !isWordRune(runes[i]) {
			i++
			continue
		}
		start := i
		for i < n && isWordRune(runes[i]) {
			i++
		}
		// Trim surrounding apostrophes.
		s, e := start, i
		for s < e && isApostrophe(runes[s]) {
			s++
		}
		for e > s && isApostrophe(runes[e-1]) {
			e--
		}
		if e <= s {
			continue
		}
		tokens = append(tokens, token{
			text:   string(runes[s:e]),
			offset: s,
			length: e - s,
		})
	}
	return tokens
}

func isApostrophe(r rune) bool { return r == '\'' || r == '’' }

func trimApostrophes(w string) string {
	return strings.TrimFunc(w, isApostrophe)
}

func containsApostrophe(w string) bool {
	return strings.ContainsFunc(w, isApostrophe)
}

// stripApostrophes removes all apostrophes from w. The frequency lists store
// contractions in this de-apostrophe'd form (e.g. "dont", "youre").
func stripApostrophes(w string) string {
	return strings.Map(func(r rune) rune {
		if isApostrophe(r) {
			return -1
		}
		return r
	}, w)
}

// contractionSuffixes are the English suffixes that follow an apostrophe in
// contractions and possessives.
var contractionSuffixes = map[string]bool{
	"s": true, "t": true, "ll": true, "re": true, "ve": true, "d": true, "m": true,
}

// acceptedContraction reports whether an apostrophe-bearing token should be
// treated as valid: either its de-apostrophe'd form is known, or it splits into
// a known base word plus a recognized contraction suffix (e.g. "dog's").
func acceptedContraction(dict *Dictionary, word string) bool {
	if !containsApostrophe(word) {
		return false
	}
	if dict.contains(stripApostrophes(word)) {
		return true
	}
	lw := strings.ToLower(word)
	for _, ap := range []string{"'", "’"} {
		if i := strings.Index(lw, ap); i > 0 && i < len(lw)-1 {
			base, suffix := lw[:i], lw[i+len(ap):]
			if dict.contains(base) && contractionSuffixes[suffix] {
				return true
			}
		}
	}
	return false
}

func hasDigit(w string) bool {
	for _, r := range w {
		if unicode.IsDigit(r) {
			return true
		}
	}
	return false
}

// sentenceTerminators end a sentence for the purposes of capitalization and
// proper-noun heuristics.
var sentenceTerminators = map[rune]bool{
	'.': true, '!': true, '?': true, '\n': true,
}

// sentenceLeadingSkips are runes that may sit between a terminator and the
// first word of the next sentence (whitespace, opening quotes/brackets, and
// Spanish opening marks). They are transparent to sentence-start detection.
func sentenceLeadingSkip(r rune) bool {
	switch r {
	case ' ', '\t', '"', '“', '‘', '(', '«', '¿', '¡':
		return true
	}
	return isApostrophe(r)
}

// isSentenceStart reports whether the token at the given rune offset begins a
// sentence (start of text, or first word after a terminator).
func isSentenceStart(runes []rune, offset int) bool {
	for i := offset - 1; i >= 0; i-- {
		if sentenceLeadingSkip(runes[i]) {
			continue
		}
		return sentenceTerminators[runes[i]]
	}
	return true
}

// wordsOnlyBetween reports whether the runes in [from,to) are all spaces or
// tabs — i.e. two tokens are separated only by intra-line whitespace.
func wordsOnlyBetween(runes []rune, from, to int) bool {
	if from >= to {
		return false
	}
	for i := from; i < to; i++ {
		if runes[i] != ' ' && runes[i] != '\t' {
			return false
		}
	}
	return true
}
