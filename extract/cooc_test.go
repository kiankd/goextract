package main

import (
	"fmt"
	"testing"
)

func TestExtractCooc(t *testing.T) {
	documents := LoadSampleWords()
	u := ExtractUnigram(documents)
	encodedDocs := UnigramEncode(u, documents)

	c := ExtractCooc(encodedDocs[0], 20)
	fmt.Printf("Length of Cooc: %d\n", len(c.counter))
	fmt.Printf("Number of documents: %d\n", len(documents))

	// TODO: write actual tests for cooc extract
	if false {
		i := 0
		for code, count := range c.counter {
			fmt.Printf("\t%d: %f\n", code, count)
			i++
			if i > 50 {
				break
			}
		}
	}
}

func TestWeighting(t *testing.T) {
	window := 5
	values := []float64{0.2, 0.4, 0.6, 0.8, 1, 1, 0.8, 0.6, 0.4, 0.2}
	for i := window + 1; i < 1000-(window+1); i++ {
		start, end := getContexts(i, window, 1000)
		if end-start != len(values) {
			t.Errorf("Error with context extraction, got %d to %d.\n", start, end)
		}
		crt := 0
		for c := start; c < end; c++ {
			if i != c {
				w := Weighting(i, c, float64(window))
				if w != values[crt] {
					t.Errorf("Weight %f != %f!\n", w, values[crt])
				}
				crt++
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

	// The cooc that merges with the other "eats" it.
	eater1 := ExtractCooc(encodedDocs[0], 2)
	c2 := ExtractCooc(encodedDocs[1], 5)
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
