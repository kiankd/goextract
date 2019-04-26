package main

import (
	"fmt"
)

// LoadSampleWords - get sample words
func LoadSampleWords() []string {
	l := ConstructLogger("silent")
	return ReadParseGz("sample.txt.gz", &l)
}

func main() {
	l := ConstructLogger("")
	// allWords := ReadParseGz("/Users/kiankd/sentpiece/spm-parsed/spm-parsed00.gz", &l)
	// u, _, _ := FullUnigramExtraction(&allWords, 50000, &l)
	// DescribeUnigram(u, 10)
	// SerializeUnigram(u, ".")
	u2 := LoadUnigram("u.unigram")
	fmt.Println(DescribeUnigram(u2, 10))

	if true {
		// u, c := FullExtraction("/Users/kiankd/sentpiece/spm-parsed/spm-parsed00.gz", 50000, 5, &l)
		u, c := FullExtraction("sample.txt.gz", 1e4, 5, &l)
		fmt.Printf("\nLength of c: %d\n\n", len(c.counter))
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
