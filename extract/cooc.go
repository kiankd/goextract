package main

import (
	"math"
)

/* CoocData struct for assist in storage. */

// CoocData - for storage later
type CoocData struct {
	Keys []int64
	Vals []float32
}

// LoadCoocData - load serialized data into it
func (c *Cooc) LoadCoocData(d CoocData) {
	for i := 0; i < len(d.Keys); i++ {
		c.Counter[d.Keys[i]] += d.Vals[i]
	}
}

/* Cooc struct for the primary extraction. */

// Cooc - Cooccurrence counter.
type Cooc struct {
	Counter map[int64]float32
}

func (c *Cooc) deepCopy() *Cooc {
	c2 := ConstructCooc()
	for cantor, count := range c.Counter {
		c2.Counter[cantor] = count
	}
	return c2
}

// Merge - Cooc c1 eats the input Cooc, c2
func (c *Cooc) Merge(c2 *Cooc) {
	for cantor, count := range c2.Counter {
		c.Counter[cantor] += count
	}
}

// AddAll - Cooc adds list of all terms and contexts for a single weight value.
func (c *Cooc) AddAll(tids []int, cids []int, weight float32) {
	// No Min function between ints in Go :(
	size := len(tids)
	if len(cids) < size {
		size = len(cids)
	}
	// Iterate over all of them and add up!
	for i := 0; i < size; i++ {
		cantor := CantorPairing(int64(tids[i]), int64(cids[i]))
		c.Counter[cantor] += weight
	}
}

// ConstructCooc constructor
func ConstructCooc() *Cooc {
	cooc := Cooc{
		Counter: make(map[int64]float32)}
	return &cooc
}

/* See https://en.wikipedia.org/wiki/Pairing_function#Cantor_pairing_function */

// CantorPairing - unique, invertible code for all pairs of words = amazing
func CantorPairing(k1, k2 int64) int64 {
	return (k1+k2)*(k1+k2+1)/2 + k2
}

// InverseCantor - gets back the original pair
func InverseCantor(cantor int64) (k1, k2 int) {
	z := float64(cantor)
	w := math.Floor(0.5 * (math.Sqrt(8*z+1) - 1))
	t := (w*w + w) / 2

	// k2 is defined first
	k2 = int(z - t)
	k1 = int(int(w) - k2)
	return
}

// ExtractCooc - extracts cooccurrence statistics from an encoded document.
func ExtractCooc(encodedDoc []int, win Window) *Cooc {
	cooc := ConstructCooc()
	lstart, lend := win.GetLeftStartEnd()
	for i := lstart; i < lend; i++ {
		weight := win.lWeights[i]
		if weight > 0 && i+1 < len(encodedDoc) {
			offset := i + 1
			terms := encodedDoc[offset:]
			conts := encodedDoc[:len(encodedDoc)-offset]
			cooc.AddAll(terms, conts, weight)
		}
	}
	rstart, rend := win.GetRightStartEnd()
	for i := rstart; i < rend; i++ {
		weight := win.rWeights[i]
		if weight > 0 && i+1 < len(encodedDoc) {
			offset := i + 1
			terms := encodedDoc[:len(encodedDoc)-offset]
			conts := encodedDoc[offset:]
			cooc.AddAll(terms, conts, weight)
		}
	}
	return cooc
}
