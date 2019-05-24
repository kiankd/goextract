package main

import "testing"

// LoadSampleWords - get sample words
func LoadSampleWords() [][]string {
	l := ConstructLogger("silent")
	return ReadParseGz("../data/test_data/sample.txt.gz", false, l)
}

/* Check the newline parsing */
func TestNewLineParsing(t *testing.T) {
	documents := LoadSampleWords()
	if len(documents) == 0 {
		t.Error("Words not extracted!")
	}
	if len(documents) != 937 {
		t.Errorf("Incorrect newline count, we got %d lines!\n", len(documents))
	}
}

func TestParse(t *testing.T) {
	// How to test the parser?
}

/* Check the Cantor pairing function */
func TestCantorPairing(t *testing.T) {
	// Passes for all natural numbers!
	var (
		start1 = 97500
		end1   = 100000
		start2 = 1000
		end2   = 5000
	)
	uniques := make(map[int64]bool)
	for k1 := start1; k1 < end1; k1++ {
		for k2 := start2; k2 < end2; k2++ {
			cantor := CantorPairing(int64(k1), int64(k2))
			if !uniques[cantor] {
				uniques[cantor] = true
			} else {
				t.Errorf("cantor(%d,%d) is not unique!", k1, k2)
			}
			x, y := InverseCantor(cantor)
			if x != k1 && y != k2 {
				t.Errorf("Cantor code %d not invertible!", cantor)
				t.Errorf("Got: %d %d. Desire: %d %d\n", x, y, k1, k2)
			}
		}
	}
}
