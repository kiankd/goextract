package main

import "testing"

/* Check the newline parsing */
func TestNewLineParsing(t *testing.T) {
	words := LoadSampleWords()
	if len(words) == 0 {
		t.Error("Words not extracted!")
	}
	count := 0
	for _, w := range words {
		if w == NEWLINE {
			count++
		}
	}
	if count != 937 {
		t.Errorf("Incorrect newline count, we got %d lines!\n", count)
	}
}

func TestParse(t *testing.T) {
	// How to test the parser?
}

/* Check the Cantor pairing function */
func TestCantorPairing(t *testing.T) {
	// Passes for all natural numbers!
	var (
		start1 = 0
		end1   = 1000
		start2 = 0
		end2   = 2000
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
