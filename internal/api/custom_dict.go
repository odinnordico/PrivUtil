package api

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"unicode"
)

const (
	maxCustomWords   = 10_000 // cap the total number of custom words
	maxCustomWordLen = 64     // cap the length of a single custom word (runes)
)

// customDict is a thread-safe, file-backed set of extra accepted words for the
// spellchecker, shared across languages. Words are stored lowercased; matching
// is case-insensitive. An empty path keeps the dictionary in memory only (no
// persistence) — used by tests and when no dictionary file is configured.
type customDict struct {
	mu    sync.RWMutex
	path  string
	words map[string]struct{}
}

// newCustomDict builds a dictionary backed by path (loading it if present). A
// missing or unreadable file is not an error: the dictionary simply starts empty.
func newCustomDict(path string) *customDict {
	d := &customDict{path: path, words: make(map[string]struct{})}
	d.load()
	return d
}

func (d *customDict) load() {
	if d.path == "" {
		return
	}
	f, err := os.Open(d.path) // #nosec G304 -- path is operator-configured (flag/env), not user input
	if err != nil {
		return
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		if w := normalizeCustomWord(sc.Text()); w != "" && !hasControlChar(w) {
			d.words[w] = struct{}{}
		}
	}
}

// Has implements spellcheck.Lookup: reports whether word (case-insensitively) is
// in the custom dictionary.
func (d *customDict) Has(word string) bool {
	w := normalizeCustomWord(word)
	if w == "" {
		return false
	}
	d.mu.RLock()
	defer d.mu.RUnlock()
	_, ok := d.words[w]
	return ok
}

// List returns the custom words, sorted.
func (d *customDict) List() []string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.sortedLocked()
}

func (d *customDict) sortedLocked() []string {
	out := make([]string, 0, len(d.words))
	for w := range d.words {
		out = append(out, w)
	}
	sort.Strings(out)
	return out
}

// Add inserts words (each entry may contain several whitespace-separated words)
// and persists. It is transactional: every token is validated first, so an
// invalid word (too long, control characters) or hitting the cap rejects the
// whole batch without mutating memory or disk. On a persist failure the added
// words are rolled back so the in-memory set never gets ahead of the file.
// Returns the full list after the operation.
func (d *customDict) Add(words []string) ([]string, error) {
	// Validate + dedupe candidate tokens before taking the write lock.
	var candidates []string
	seen := make(map[string]struct{})
	for _, raw := range words {
		for _, tok := range strings.Fields(raw) {
			w := normalizeCustomWord(tok)
			if w == "" {
				continue
			}
			if len([]rune(w)) > maxCustomWordLen {
				return d.List(), fmt.Errorf("word %q is too long (limit %d characters)", tok, maxCustomWordLen)
			}
			if hasControlChar(w) {
				return d.List(), fmt.Errorf("word %q contains control characters", tok)
			}
			if _, dup := seen[w]; !dup {
				seen[w] = struct{}{}
				candidates = append(candidates, w)
			}
		}
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	var added []string
	for _, w := range candidates {
		if _, ok := d.words[w]; !ok {
			added = append(added, w)
		}
	}
	if len(added) == 0 {
		return d.sortedLocked(), nil // nothing new: skip the rewrite
	}
	if len(d.words)+len(added) > maxCustomWords {
		return d.sortedLocked(), fmt.Errorf("custom dictionary is full (limit %d words)", maxCustomWords)
	}
	for _, w := range added {
		d.words[w] = struct{}{}
	}
	if err := d.persistLocked(); err != nil {
		for _, w := range added {
			delete(d.words, w) // roll back so memory matches the unchanged file
		}
		return d.sortedLocked(), err
	}
	return d.sortedLocked(), nil
}

// Remove deletes a word and persists. Removing an absent word is a no-op (no
// file rewrite). On a persist failure the deletion is rolled back.
func (d *customDict) Remove(word string) ([]string, error) {
	w := normalizeCustomWord(word)
	d.mu.Lock()
	defer d.mu.Unlock()
	if _, ok := d.words[w]; !ok {
		return d.sortedLocked(), nil
	}
	delete(d.words, w)
	if err := d.persistLocked(); err != nil {
		d.words[w] = struct{}{} // roll back
		return d.sortedLocked(), err
	}
	return d.sortedLocked(), nil
}

// persistLocked writes the dictionary atomically (temp file + rename). It is a
// no-op when the dictionary is in-memory only (empty path). Caller holds d.mu.
func (d *customDict) persistLocked() error {
	if d.path == "" {
		return nil
	}
	dir := filepath.Dir(d.path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create dictionary directory: %w", err)
	}
	tmp, err := os.CreateTemp(dir, ".custom-dict-*.tmp")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName) // no-op after a successful rename

	w := bufio.NewWriter(tmp)
	for _, word := range d.sortedLocked() {
		if _, err := fmt.Fprintln(w, word); err != nil {
			_ = tmp.Close()
			return err
		}
	}
	if err := w.Flush(); err != nil {
		_ = tmp.Close()
		return err
	}
	// fsync before the rename so the data is durable, not just atomically
	// swapped — otherwise a crash after rename can leave a truncated file.
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, d.path)
}

// normalizeCustomWord lowercases and trims a stored/looked-up word.
func normalizeCustomWord(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

// hasControlChar reports whether s contains any control character, so such
// words are never stored (they can inject ANSI escapes when the dictionary file
// is later viewed in a terminal).
func hasControlChar(s string) bool {
	return strings.IndexFunc(s, unicode.IsControl) >= 0
}
