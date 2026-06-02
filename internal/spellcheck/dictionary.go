package spellcheck

import (
	"bufio"
	"strconv"
	"strings"
	"sync"
)

// Dictionary is a lazily-loaded word→frequency map. It backs two operations:
// membership testing (is a token a real word?) and ranking candidate
// corrections by how common the word is. Words are stored lowercased.
type Dictionary struct {
	file  string
	once  sync.Once
	words map[string]int
	err   error
}

func newDictionary(file string) *Dictionary { return &Dictionary{file: file} }

func (d *Dictionary) load() {
	d.once.Do(func() {
		f, err := dictFS.Open(d.file)
		if err != nil {
			d.err = err
			return
		}
		defer f.Close()

		words := make(map[string]int, 1<<16)
		sc := bufio.NewScanner(f)
		sc.Buffer(make([]byte, 0, 64*1024), 1<<20)
		for sc.Scan() {
			line := strings.TrimSpace(sc.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			word, count := line, 1
			if i := strings.IndexByte(line, ' '); i > 0 {
				word = line[:i]
				if n, err := strconv.Atoi(strings.TrimSpace(line[i+1:])); err == nil {
					count = n
				}
			}
			word = strings.ToLower(word)
			if word == "" {
				continue
			}
			if count > words[word] {
				words[word] = count
			}
		}
		d.err = sc.Err()
		d.words = words
	})
}

// contains reports whether word (case-insensitively) is a known word.
func (d *Dictionary) contains(word string) bool {
	d.load()
	_, ok := d.words[strings.ToLower(word)]
	return ok
}

// freq returns the frequency count for word, or 0 if unknown.
func (d *Dictionary) freq(word string) int {
	d.load()
	return d.words[strings.ToLower(word)]
}

// knownSet returns the unique members of candidates that exist in the
// dictionary (already lowercased input expected).
func (d *Dictionary) knownSet(candidates []string) []string {
	d.load()
	seen := make(map[string]struct{}, len(candidates))
	out := make([]string, 0, 8)
	for _, c := range candidates {
		if _, dup := seen[c]; dup {
			continue
		}
		if _, ok := d.words[c]; ok {
			seen[c] = struct{}{}
			out = append(out, c)
		}
	}
	return out
}
