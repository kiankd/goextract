package main

import (
	"fmt"

	"github.com/pkg/profile"
)

// LoadSampleWords - get sample words
func LoadSampleWords() [][]string {
	l := ConstructLogger("silent")
	return ReadParseGz("../data/sample.txt.gz", &l)
}

func main() {
	// defer profile.Start().Stop()
	defer profile.Start(profile.MemProfile).Stop()
	l := ConstructLogger("")

	// allWords := ReadParseGz("/Users/kiankd/sentpiece/spm-parsed/spm-parsed00.gz", &l)
	// u, _, _ := FullUnigramExtraction(&allWords, 50000, &l)
	// DescribeUnigram(u, 10)
	// SerializeUnigram(u, ".")
	// u2 := LoadUnigram("u.unigram")
	// fmt.Println(DescribeUnigram(u2, 10))

	// u, c := FullExtraction("/Users/kiankd/sentpiece/spm-parsed/spm-parsed00.gz", 50000, 5, &l)
	u, c := FullExtraction("../data/sample.txt.gz", 1e4, 25, &l)
	fmt.Printf("\nLength of c: %d\n\n", len(c.counter))

	if false {
		i := 0
		for cantorCode, count := range c.counter {
			x, y := InverseCantor(cantorCode)
			wX := u.decode(x)
			wY := u.decode(y)
			fmt.Printf("(%16s, %15s): %f\n", wX, wY, count)
			i++
			if i > 2 {
				break
			}
		}
		codeOf := u.encode("▁of")
		codeThe := u.encode("▁the")
		cantor := CantorPairing(int64(codeOf), int64(codeThe))
		fmt.Println()
		fmt.Printf("(▁of, ▁the): %f\n", c.counter[cantor])
	}
}
