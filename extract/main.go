package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/pkg/profile"
)

func runExample() {
	l := ConstructLogger("print")
	u, c := FullExtraction("../data/sample.txt.gz", 1e4, 25, false, &l)
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

func loadExperimentPaths(extractPath string) []string {
	var paths []string
	if strings.HasSuffix(extractPath, ".paths") {
		if f, err := os.Open(extractPath); err == nil {
			defer f.Close()
			if bytes, err := ioutil.ReadAll(f); err == nil {
				paths = strings.Split(string(bytes), "\n")
			}
		} else {
			panic("The .paths file does not exist!")
		}
	} else {
		paths = make([]string, strings.Count(extractPath, " ")+1)
		for i, p := range strings.Split(extractPath, " ") {
			if _, err := os.Stat(p); os.IsNotExist(err) {
				panic(fmt.Sprintf("Error: path %s does not exist!", p))
			}
			paths[i] = p
		}
	}
	return paths
}

func main() {
	var extractPath string
	flag.StringVar(&extractPath, "extract", "../data/sample.txt.gz",
		"path to the target gz file we will be extracting")
	debug := flag.Bool("debug", false,
		"whether to run a debug profiler")
	cpuProfile := flag.Bool("pcpu", false,
		"whether to do CPU profiling (RAM profiling is the default)")
	logOption := flag.String("log", "print",
		"option for writing, printing, or silence [write, print, silent]")
	unigramPath := flag.String("u", "",
		"path to the unigram to pre-load, if desired")
	doExample := flag.Bool("example", false,
		"run a simple example run and exit")
	replaceDigits := flag.Bool("nodigits", false,
		"pass this to replace all digits with 0s")
	vocabSize := flag.Int("vocab", -1,
		"desired size of the vocabulary to perform extraction")
	flag.Parse()

	// Check if we just want to do an example run.
	if *doExample {
		runExample()
		return
	}

	// TODO: pass to the logger all args and log them.
	l := ConstructLogger(*logOption)

	// Very beginning - ensure the file exists, and parse if there are multiple.
	paths := loadExperimentPaths(extractPath)
	l.LogAll("Loading from paths:", paths)

	// Now check if we are doing debugging stuff.
	if *debug {
		if *cpuProfile {
			defer profile.Start().Stop()
		} else {
			defer profile.Start(profile.MemProfile).Stop()
		}
	}

	// Now check if we can load the unigram file.
	var unigram *Unigram
	uPth := *unigramPath
	if _, err := os.Stat(uPth); os.IsNotExist(err) {
		l.Log("Extracting unigram...")
		unigram = DynamicUnigramExtraction(paths, *replaceDigits, &l)
		l.Log("Serializing unigram...")
		SerializeUnigram(unigram, uPth)
	} else {
		l.Log(fmt.Sprintf("Loading unigram from %s...", uPth))
		unigram = LoadUnigram(uPth)
	}

	// Filtering the unigram.
	if *vocabSize > 0 {
		l.Log(fmt.Sprintf("Filtering unigram to %d most frequent tokens...", *vocabSize))
		unigram = FilterUnigram(unigram, *vocabSize)
	}
}
