package spellcheck

import (
	"fmt"
	"strings"
)

const enAlphabet = "abcdefghijklmnopqrstuvwxyz'"

// enAbbreviations are abbreviations whose trailing period does not end a
// sentence, used to suppress false capitalization warnings.
var enAbbreviations = map[string]bool{
	"mr": true, "mrs": true, "ms": true, "dr": true, "prof": true,
	"sr": true, "jr": true, "st": true, "vs": true, "etc": true,
	"inc": true, "ltd": true, "co": true, "no": true, "fig": true,
	"approx": true, "dept": true, "est": true, "min": true, "max": true,
}

// enRepeatWhitelist holds words that can legitimately appear doubled.
var enRepeatWhitelist = map[string]bool{
	"had": true, "that": true, "is": true, "the": false,
}

func newEnglish() *Language {
	dict := newDictionary("dict/en.txt")
	return &Language{
		Code:    "en",
		Label:   "English",
		dict:    dict,
		speller: newSpeller(dict, enAlphabet),
		rules: []ruleFunc{
			repeatedWordRule(enRepeatWhitelist),
			doubleSpaceRule,
			spaceBeforePunctuationRule,
			missingSpaceAfterCommaRule,
			sentenceCapitalizationRule(enAbbreviations),
			enArticleRule,
			enCapitalizeIRule,
			enWouldOfRule,
		},
	}
}

// enArticleRule flags "a"/"an" used against the following word's sound, e.g.
// "a apple" → "an apple", "an car" → "a car".
func enArticleRule(runes []rune) []Issue {
	tokens := tokenizeWords(runes)
	var issues []Issue
	for i := 0; i+1 < len(tokens); i++ {
		art := strings.ToLower(tokens[i].text)
		if art != "a" && art != "an" {
			continue
		}
		next := tokens[i+1]
		if !wordsOnlyBetween(runes, tokens[i].end(), next.offset) {
			continue
		}
		if !isAlpha(next.text) {
			continue
		}
		wantAn := startsWithVowelSound(next.text)
		if wantAn && art == "a" {
			issues = append(issues, articleIssue(tokens[i], "an"))
		} else if !wantAn && art == "an" {
			issues = append(issues, articleIssue(tokens[i], "a"))
		}
	}
	return issues
}

func articleIssue(t token, correct string) Issue {
	replacement := matchCase(correct, t.text)
	return Issue{
		Offset:       t.offset,
		Length:       t.length,
		Text:         t.text,
		Type:         TypeGrammar,
		Rule:         "article-a-an",
		Message:      fmt.Sprintf("Use %q before this word.", replacement),
		Replacements: []string{replacement},
	}
}

// vowelSoundExceptions start with a vowel letter but a consonant sound
// (→ use "a"). consonantSoundVowelExceptions start with a consonant letter but
// a vowel sound (→ use "an"); these are the silent-h words.
var consonantStartButVowelSound = []string{"hour", "honest", "honor", "honour", "heir"}
var vowelStartButConsonantSound = []string{"uni", "use", "usu", "eu", "ewe", "one", "once", "ubiqu"}

func startsWithVowelSound(word string) bool {
	lw := strings.ToLower(word)
	for _, p := range consonantStartButVowelSound {
		if strings.HasPrefix(lw, p) {
			return true
		}
	}
	for _, p := range vowelStartButConsonantSound {
		if strings.HasPrefix(lw, p) {
			return false
		}
	}
	r := []rune(lw)[0]
	switch r {
	case 'a', 'e', 'i', 'o', 'u':
		return true
	}
	return false
}

// enCapitalizeIRule flags the lowercase pronoun "i" (including contractions
// like "i'm", "i've").
func enCapitalizeIRule(runes []rune) []Issue {
	tokens := tokenizeWords(runes)
	var issues []Issue
	for _, t := range tokens {
		rt := []rune(t.text)
		if rt[0] != 'i' {
			continue
		}
		lw := strings.ToLower(t.text)
		if lw == "i" || isContractionOfI(lw) {
			issues = append(issues, Issue{
				Offset:       t.offset,
				Length:       1,
				Text:         "i",
				Type:         TypeGrammar,
				Rule:         "capitalize-i",
				Message:      "Capitalize the pronoun “I”.",
				Replacements: []string{"I"},
			})
		}
	}
	return issues
}

func isContractionOfI(lw string) bool {
	switch lw {
	case "i'm", "i’m", "i've", "i’ve", "i'll", "i’ll", "i'd", "i’d":
		return true
	}
	return false
}

// enWouldOfRule flags "would of" / "could of" / "should of" (and friends),
// which should use "have".
var modalsBeforeHave = map[string]bool{
	"would": true, "could": true, "should": true,
	"must": true, "might": true, "may": true,
}

func enWouldOfRule(runes []rune) []Issue {
	tokens := tokenizeWords(runes)
	var issues []Issue
	for i := 0; i+1 < len(tokens); i++ {
		if !modalsBeforeHave[strings.ToLower(tokens[i].text)] {
			continue
		}
		if strings.ToLower(tokens[i+1].text) != "of" {
			continue
		}
		if !wordsOnlyBetween(runes, tokens[i].end(), tokens[i+1].offset) {
			continue
		}
		of := tokens[i+1]
		replacement := matchCase("have", of.text)
		issues = append(issues, Issue{
			Offset:       of.offset,
			Length:       of.length,
			Text:         of.text,
			Type:         TypeGrammar,
			Rule:         "would-of",
			Message:      fmt.Sprintf("Did you mean “%s have”?", tokens[i].text),
			Replacements: []string{replacement},
		})
	}
	return issues
}
