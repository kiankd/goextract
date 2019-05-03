package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strings"

	"github.com/pkg/profile"
)

func runExample() {
}

// 	l := ConstructLogger("print")
// 	u, c := FullExtraction("../data/sample.txt.gz", 1e4, 25, false, l)
// 	fmt.Printf("\nLength of c: %d\n\n", len(c.counter))
// 	i := 0
// 	for cantorCode, count := range c.counter {
// 		x, y := InverseCantor(cantorCode)
// 		wX := u.Decode(x)
// 		wY := u.Decode(y)
// 		fmt.Printf("(%16s, %15s): %f\n", wX, wY, count)
// 		i++
// 		if i > 2 {
// 			break
// 		}
// 	}
// 	codeOf := u.Encode("▁of")
// 	codeThe := u.Encode("▁the")
// 	cantor := CantorPairing(int64(codeOf), int64(codeThe))
// 	fmt.Println()
// 	fmt.Printf("(▁of, ▁the): %f\n", c.counter[cantor])
// }

func loadExperimentPath(extractPath string) string {
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
	if len(paths) > 1 {
		panic("It is inefficient to parse multiple at once, do one at a time with a bash script.")
	}
	return paths[0]
}

// Merge these boys!
func mergeUnigrams(unigramPath string, l *Logger) {
	u := ConstructUnigram()
	uFiles, _ := ioutil.ReadDir(unigramPath)
	for _, file := range uFiles {
		n := file.Name()
		if strings.HasSuffix(n, ".unigram") && !strings.HasPrefix(n, "merged") {
			l.Log(fmt.Sprintf("\tmerging %s...\n", n))
			u2 := LoadUnigram(unigramPath + n)
			u.Merge(u2)
		}
	}
	u.FillIdx()
	sort.Sort(u)
	SerializeUnigram(u, unigramPath+"merged.unigram")
}

// Merge those boys!
func mergeCoocs(coocPath string, l *Logger) {
	c := ConstructCooc()
	cFiles, _ := ioutil.ReadDir(coocPath)
	for _, file := range cFiles {
		n := file.Name()
		if strings.HasSuffix(n, ".cooc") && !strings.HasPrefix(n, "merged") {
			l.Log(fmt.Sprintf("\tmerging %s...\n", n))
			c2 := LoadCooc(coocPath + n)
			c.Merge(c2)
		}
	}
	SerializeCooc(c, coocPath+"merged.cooc")
}

func main() {
	var extractPath string
	flag.StringVar(&extractPath, "extract", "../data/sample.txt.gz",
		"path to the target gz file we will be extracting")

	extractOption := flag.String("option", "unigram",
		"option for extraction, \"unigram\" or \"cooc\"; add \"-merge\" to merge?")

	debug := flag.Bool("debug", false,
		"whether to run a debug profiler")

	cpuProfile := flag.Bool("pcpu", false,
		"whether to do CPU profiling (RAM profiling is the default)")

	logOption := flag.String("log", "print",
		"option for writing, printing, or silence [write, print, silent]")

	unigramPath := flag.String("u", "",
		"path to the unigram to pre-load, if desired")

	coocPath := flag.String("c", "",
		"path for where to save Coocs, if desired")

	doExample := flag.Bool("example", false,
		"run a simple example run and exit")

	replaceDigits := flag.Bool("nodigits", false,
		"replace all digits with 0s during extraction")

	vocabSize := flag.Int("vocab", -1,
		"desired size of the vocabulary to perform extraction")

	window := flag.Int("window", 5,
		"window size, an integer indicating it (only dynamic weighting for now)")

	flag.Parse()

	// Check if we just want to do an example run.
	if *doExample {
		runExample()
		return
	}

	// TODO: pass to the logger all args and log them.
	l := ConstructLogger(*logOption)

	// Very beginning - ensure the file exists, and parse if there are multiple.
	exPath := loadExperimentPath(extractPath)

	// Now check if we are doing debugging stuff.
	if *debug {
		if *cpuProfile {
			defer profile.Start().Stop()
		} else {
			defer profile.Start(profile.MemProfile).Stop()
		}
	}

	// Now check if we can load the unigram file or if something else is happening.
	var unigram *Unigram
	uPth := *unigramPath

	if *extractOption == "unigram-merge" {
		mergeUnigrams(uPth, l)

	} else if *extractOption == "unigram" {
		l.Log(fmt.Sprintf("Will extract from path %s...", exPath))
		if _, err := os.Stat(uPth); os.IsNotExist(err) {
			l.Log("\textracting its unigram...")
			unigram = UnigramExtraction(extractPath, *replaceDigits, l)
			l.Log("\tserializing its unigram...")
			SerializeUnigram(unigram, uPth)
		}

	} else if *extractOption == "cooc" {
		l.Log(fmt.Sprintf("Loading unigram from %s...", uPth))
		unigram = LoadUnigram(uPth)
		if *vocabSize > 0 {
			l.Log(fmt.Sprintf("Filtering unigram to %d most frequent tokens...", *vocabSize))
			unigram = FilterUnigram(unigram, *vocabSize)
		}
		c := CoocExtraction(exPath, unigram, *window, *replaceDigits, l)
		cPth := *coocPath
		if cPth == "" {
			cPth = fmt.Sprintf("coocs/%s.cooc", exPath)
		}
		l.Log("Serializing coocs...")
		SerializeCooc(c, cPth)

	} else if *extractOption == "cooc-merge" {
		mergeCoocs(*coocPath, l)
	}
	l.Log("Finished.")
}
