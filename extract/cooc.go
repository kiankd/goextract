package main

import (
	"fmt"
	"log"
	"math"
	"os"
)

// Cooc - Cooccurrence counter.
type Cooc struct {
	counter map[int64]float64
}

func (c1 *Cooc) deepCopy() *Cooc {
	c2 := ConstructCooc()
	for cantor, count := range c1.counter {
		c2.counter[cantor] = count
	}
	return c2
}

// Merge - Cooc c1 eats the input Cooc, c2
func (c1 *Cooc) Merge(c2 *Cooc) {
	for cantor, count := range c2.counter {
		c1.counter[cantor] += count
	}
}

// SerializeCooc - Helper to write a Cooc to disk.
func SerializeCooc(c *Cooc, fullPath string) error {
	if f, err := os.Create(fullPath); err == nil {
		defer f.Close()
		for cantor, count := range c.counter {
			f.WriteString(fmt.Sprintf("%d %f\n", cantor, count))
		}
	} else {
		log.Fatal("Cannot write unigram.")
		return err
	}
	return nil

}

// LoadCooc - loads a cooc!
func LoadCooc(fullPath string) *Cooc {
	c := ConstructCooc()
	return c
}

// ConstructCooc constructor
func ConstructCooc() *Cooc {
	cooc := Cooc{
		counter: make(map[int64]float64)}
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
	w := math.Floor(
		0.5 * (math.Sqrt(8*z+1) - 1))
	t := (w*w + w) / 2

	// k2 is defined first
	k2 = int(z - t)
	k1 = int(int(w) - k2)
	return
}

func getContexts(i, w, length int) (int, int) {
	return int(math.Max(float64(i-w), 0.0)),
		int(math.Min(float64(i+w), float64(length)))
}

// Weighting - do dynamic context window weighting, like SGNS.
// diff = abs(i-j) # 1, 2, 3, 4, 5...
// window - diff + 1 / window
// w = 5, diff=1: 5 - 1 + 1 / 5 = 1 good
// w = 5, diff=3: 5 - 3 + 1 / 5 = 3/5 good
// w = 5, diff=5: 5 - 5 + 1 / 5 = 1/5 good
func Weighting(termIdx int, contIdx int, window float64) float64 {
	return (window - math.Abs(float64(termIdx-contIdx)) + 1) / window
}

// ExtractCooc - extracts cooccurrence statistics from an encoded document.
func ExtractCooc(encodedDoc []int, window int) *Cooc {
	cooc := ConstructCooc()
	for i, term := range encodedDoc {
		start, end := getContexts(i, window, len(encodedDoc))
		for j := start; j < end; j++ {
			if i != j {
				cantor := CantorPairing(int64(term), int64(encodedDoc[j]))
				cooc.counter[cantor] += Weighting(i, j, float64(window))
			}
		}
	}
	return cooc
}
