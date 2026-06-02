// Package spellcheck provides a fully offline spell- and grammar-checking
// engine. Every language pack — its dictionary and grammar rules — is embedded
// into the binary, so no text ever leaves the process.
//
// Dictionaries under dict/ are word-frequency lists derived from the
// OpenSubtitles corpus (hermitdave/FrequencyWords, MIT-licensed). The
// frequency counts double as a ranking signal for spelling suggestions.
//
// Adding a language: drop a "<code>.txt" frequency list under dict/, add a
// newXxx() constructor returning a *Language, and register it in init().
package spellcheck

import "embed"

//go:embed dict/*.txt
var dictFS embed.FS
