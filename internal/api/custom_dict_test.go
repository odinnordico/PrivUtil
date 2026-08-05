package api

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	pb "github.com/odinnordico/privutil/proto"
)

func TestCustomDictAddListRemove(t *testing.T) {
	d := newCustomDict(filepath.Join(t.TempDir(), "dict.txt"))

	if _, err := d.Add([]string{"Foo", "bar baz", "foo"}); err != nil {
		t.Fatalf("Add: %v", err)
	}
	// Stored lowercased, deduped, whitespace-split, sorted.
	if got := d.List(); strings.Join(got, ",") != "bar,baz,foo" {
		t.Errorf("List = %v, want [bar baz foo]", got)
	}
	if !d.Has("FOO") || !d.Has("foo") {
		t.Error("Has should be case-insensitive")
	}
	if _, err := d.Remove("BAR"); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	if d.Has("bar") {
		t.Error("bar should be removed")
	}
}

func TestCustomDictPersistence(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sub", "dict.txt") // dir must be created on write

	d1 := newCustomDict(path)
	if _, err := d1.Add([]string{"widget", "gizmo"}); err != nil {
		t.Fatalf("Add: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("dictionary file not written: %v", err)
	}

	// A fresh instance on the same path loads the persisted words.
	d2 := newCustomDict(path)
	if !d2.Has("widget") || !d2.Has("gizmo") {
		t.Errorf("reloaded dict missing words: %v", d2.List())
	}
}

func TestCustomDictValidation(t *testing.T) {
	d := newCustomDict("") // in-memory
	long := strings.Repeat("a", maxCustomWordLen+1)
	if _, err := d.Add([]string{long}); err == nil {
		t.Error("expected error for over-long word")
	}
	if d.Has(long) {
		t.Error("over-long word should not have been stored")
	}
}

func TestCustomDictInMemoryNoFile(t *testing.T) {
	d := newCustomDict("")
	if _, err := d.Add([]string{"ephemeral"}); err != nil {
		t.Fatalf("Add: %v", err)
	}
	if !d.Has("ephemeral") {
		t.Error("in-memory add failed")
	}
}

// TestSpellCheckHonorsCustomDict verifies an added word stops being flagged.
func TestSpellCheckHonorsCustomDict(t *testing.T) {
	s := NewServer(WithCustomDictPath(filepath.Join(t.TempDir(), "dict.txt")))
	ctx := context.Background()
	const word = "zzblarg"

	before, err := s.SpellCheck(ctx, &pb.SpellCheckRequest{Text: "hello " + word + " world"})
	if err != nil {
		t.Fatalf("SpellCheck: %v", err)
	}
	if !flagsWord(before.Issues, word) {
		t.Fatalf("expected %q to be flagged before adding it", word)
	}

	if _, err := s.AddCustomWords(ctx, &pb.AddCustomWordsRequest{Words: []string{word}}); err != nil {
		t.Fatalf("AddCustomWords: %v", err)
	}

	after, err := s.SpellCheck(ctx, &pb.SpellCheckRequest{Text: "hello " + word + " world"})
	if err != nil {
		t.Fatalf("SpellCheck: %v", err)
	}
	if flagsWord(after.Issues, word) {
		t.Errorf("%q should not be flagged after adding to the custom dictionary", word)
	}

	got, err := s.GetCustomWords(ctx, &pb.GetCustomWordsRequest{})
	if err != nil {
		t.Fatalf("GetCustomWords: %v", err)
	}
	if len(got.Words) != 1 || got.Words[0] != word {
		t.Errorf("GetCustomWords = %v, want [%s]", got.Words, word)
	}
}

func flagsWord(issues []*pb.SpellIssue, word string) bool {
	for _, is := range issues {
		if strings.EqualFold(is.Text, word) {
			return true
		}
	}
	return false
}
