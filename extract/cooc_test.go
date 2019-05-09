package main

import (
	"fmt"
	"testing"
)

func TestExtractCooc(t *testing.T) {
	documents := LoadSampleWords()
	u := ExtractUnigram(documents)
	encodedDocs := UnigramEncode(u, documents)

	win := MakeWindow(20, "")
	c := ExtractCooc(encodedDocs[0], *win)
	// fmt.Printf("Length of Cooc: %d\n", len(c.Counter))
	// fmt.Printf("Number of documents: %d\n", len(documents))

	// TODO: write actual tests for cooc extract
	if false {
		i := 0
		for code, count := range c.Counter {
			fmt.Printf("\t%d: %f\n", code, count)
			i++
			if i > 50 {
				break
			}
		}
	}
}

func TestCoocMerge(t *testing.T) {
	words := LoadSampleWords()
	u := ExtractUnigram(words)

	// This extraction is currently being used improperly,
	// as it should be used on a doc-by-doc basis.
	encodedDocs := UnigramEncode(u, words)

	// Make window daddy
	win2 := MakeWindow(2, "")
	win5 := MakeWindow(2, "")

	// The cooc that merges with the other "eats" it.
	eater1 := ExtractCooc(encodedDocs[0], *win2)
	c2 := ExtractCooc(encodedDocs[1], *win5)
	c1copy := eater1.deepCopy()
	eater2 := c2.deepCopy()

	eater1.Merge(c2)
	eater2.Merge(c1copy)

	if len(eater1.Counter) != len(eater2.Counter) {
		t.Error("Different lengths! Not a bijection!")
	}

	for cantor, c1count := range eater1.Counter {
		if c1count != eater2.Counter[cantor] {
			t.Error("Different counts for a cantor code!")
		}
	}
}
