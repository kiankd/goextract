package main

import (
	"math"
	"testing"
)

func TestExtractCooc(t *testing.T) {
	documents := LoadSampleWords()
	u := ExtractUnigram(documents)
	encodedDocs := UnigramEncode(u, documents)

	win := MakeWindow(5, "")
	c := ExtractCooc(encodedDocs[0], *win)
	good := 0
	for cantor := range c.Counter {
		i, j := InverseCantor(cantor)
		lrC := float64(c.Counter[cantor])
		rlC := float64(c.Counter[CantorPairing(int64(j), int64(i))])
		if math.Abs(lrC-rlC) > 1e-3 { // float comparison
			t.Errorf("Not symmetric (%d, %d)! Got lr %f but rl %f\n", i, j, lrC, rlC)
		} else {
			good++
		}
	}
	if good != len(c.Counter) {
		t.Errorf("Only got %f percent symmetric extractions!\n", 100*float64(good)/float64(len(c.Counter)))
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
	win5 := MakeWindow(5, "")

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
