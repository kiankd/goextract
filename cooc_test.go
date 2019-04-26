package main

import (
	"fmt"
	"testing"
)

func TestExtractCooc(t *testing.T) {
	words := LoadSampleWords()
	u := ExtractUnigram(words)
	encodedWords, _ := UnigramEncode(u, words)

	// This extraction is currently being used improperly,
	// as it should be used on a doc-by-doc basis.

	c := ExtractCooc(encodedWords, 20)
	fmt.Printf("Length of Cooc: %d\n", len(c.counter))
	fmt.Printf("Length of doc: %d\n", len(words))

	// TODO: write actual tests for cooc extract
	i := 0
	for code, count := range c.counter {
		fmt.Printf("\t%d: %f\n", code, count)
		i++
		if i > 50 {
			break
		}
	}
}

func TestCoocMerge(t *testing.T) {
	words := LoadSampleWords()
	u := ExtractUnigram(words)

	// This extraction is currently being used improperly,
	// as it should be used on a doc-by-doc basis.
	encodedWords, _ := UnigramEncode(u, words)

	// The cooc that merges with the other "eats" it.
	eater1 := ExtractCooc(encodedWords, 2)
	c2 := ExtractCooc(encodedWords, 5)
	c1copy := eater1.deepCopy()
	eater2 := c2.deepCopy()

	eater1.merge(c2)
	eater2.merge(c1copy)

	if len(eater1.counter) != len(eater2.counter) {
		t.Error("Different lengths! Not a bijection!")
	}

	for cantor, c1count := range eater1.counter {
		if c1count != eater2.counter[cantor] {
			t.Error("Different counts for a cantor code!")
		}
	}
}
