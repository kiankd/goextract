package main

import (
	"flag"
	"fmt"
	"strings"

	"github.com/pkg/profile"
)

// LoadSampleWords - get sample words
func LoadSampleWords() [][]string {
	l := ConstructLogger("silent")
	return ReadParseGz("../data/sample.txt.gz", &l)
}

func main() {
	var path string
	flag.StringVar(&path, "extract", "../data/sample.txt.gz",
		"path to the target gz file we will be extracting")
	debug := flag.Bool("debug", false,
		"whether to run a debug profiler")
	cpuProfile := flag.Bool("pcpu", false,
		"whether to do CPU profiling (RAM is default)")
	flag.Parse()

	// TODO: pass to the logger all args and log them.
	l := ConstructLogger("")

	// Very beginning - ensure the file exists, and parse if there are multiple.
	nPaths := strings.Count(path, " ") + 1
	paths := make([]string, nPaths)
	for i, p := range strings.Split(path, " ") {
		paths[i] = p
		fmt.Println(p)
	}
	return

	// First check if we are doing debugging stuff.
	if *debug {
		if *cpuProfile {
			defer profile.Start().Stop()
		} else {
			defer profile.Start(profile.MemProfile).Stop()
		}
	}

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
