package spellcheck

import (
	"fmt"
	"slices"
)

const esAlphabet = "abcdefghijklmnopqrstuvwxyzñáéíóúü"

// esAbbreviations are abbreviations whose trailing period does not end a
// sentence.
var esAbbreviations = map[string]bool{
	"sr": true, "sra": true, "srta": true, "dr": true, "dra": true,
	"ud": true, "uds": true, "vd": true, "vds": true, "etc": true,
	"ej": true, "p": true, "pág": true, "núm": true, "art": true,
	"av": true, "avda": true, "núms": true, "tel": true,
}

// esRepeatWhitelist holds words that can legitimately appear doubled in Spanish.
var esRepeatWhitelist = map[string]bool{}

func newSpanish() *Language {
	dict := newDictionary("dict/es.txt")
	return &Language{
		Code:    "es",
		Label:   "Español (Latinoamérica)",
		dict:    dict,
		speller: newSpeller(dict, esAlphabet),
		rules: []ruleFunc{
			repeatedWordRule(esRepeatWhitelist),
			doubleSpaceRule,
			spaceBeforePunctuationRule,
			missingSpaceAfterCommaRule,
			sentenceCapitalizationRule(esAbbreviations),
			esOpeningMarkRule,
		},
	}
}

// esOpeningMarkRule flags Spanish questions and exclamations that are missing
// their opening "¿" or "¡". The mark is anchored at the start of the question
// clause: after the last comma/semicolon/colon if present, otherwise at the
// start of the sentence.
func esOpeningMarkRule(runes []rune) []Issue {
	var issues []Issue
	n := len(runes)
	segStart := 0

	flush := func(end int, term rune) {
		var open rune
		switch term {
		case '?':
			open = '¿'
		case '!':
			open = '¡'
		default:
			return
		}
		seg := runes[segStart:end]
		if slices.Contains(seg, open) {
			return // already opened
		}
		// Anchor after the last comma/semicolon/colon, else at the start.
		searchFrom := 0
		for idx, r := range seg {
			if r == ',' || r == ';' || r == ':' {
				searchFrom = idx + 1
			}
		}
		wstart := -1
		for idx := searchFrom; idx < len(seg); idx++ {
			if isWordRune(seg[idx]) {
				wstart = idx
				break
			}
		}
		if wstart < 0 {
			return
		}
		wend := wstart
		for wend < len(seg) && isWordRune(seg[wend]) {
			wend++
		}
		word := string(seg[wstart:wend])
		issues = append(issues, Issue{
			Offset:       segStart + wstart,
			Length:       wend - wstart,
			Text:         word,
			Type:         TypePunctuation,
			Rule:         "missing-opening-mark",
			Message:      fmt.Sprintf("Spanish questions and exclamations open with “%c”.", open),
			Replacements: []string{string(open) + word},
		})
	}

	for i := 0; i < n; i++ {
		switch runes[i] {
		case '?', '!':
			flush(i, runes[i])
			j := i
			for j < n && (runes[j] == '?' || runes[j] == '!' || runes[j] == '.') {
				j++
			}
			segStart = j
			i = j - 1
		case '.', '\n':
			segStart = i + 1
		}
	}
	return issues
}
