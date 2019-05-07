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

// Does checks for the CLI.
func checkArgs(opt, exP, uP, cP *string, v, w *int) {
	emptyExp := *exP == ""
	emptyUni := *uP == ""
	emptyCoo := *cP == ""
	emptyVoc := *v <= 0
	emptyWin := *w <= 0
	switch *opt {
	case "unigram-merge":
		if emptyUni {
			panic("No path specified for unigram-merging!")
		}
	case "cooc-merge":
		if emptyCoo {
			panic("No path specified for cooc-merging!")
		} else if !strings.HasPrefix(*cP, "/") {
			panic("Trying to merge coocs, but need a directory!")
		}

	case "unigram":
		if emptyExp || emptyUni {
			panic("No paths specified for unigram extraction!")
		}
	case "cooc":
		if emptyExp || emptyCoo || emptyVoc || emptyWin {
			panic("Cooc extraction needs exp, coocpath, vocabsize, and window! Missing!")
		}
	default:
		panic(fmt.Sprintf("Option %s is invalid!\n", *opt))
	}
}

// Merge these boys!
func mergeUnigrams(unigramPath string, l *Logger) {
	var u *Unigram
	uFiles, _ := ioutil.ReadDir(unigramPath)
	for _, file := range uFiles {
		n := file.Name()
		if strings.HasSuffix(n, ".unigram") && !strings.HasPrefix(n, "merged") {
			l.Log(fmt.Sprintf("\tmerging %s...\n", n))
			u2 := LoadUnigram(unigramPath + n)
			if u == nil {
				u = u2
			} else {
				u.Merge(u2)
			}
		}
	}
	u.FillIdx()
	sort.Sort(u)
	SerializeUnigram(u, unigramPath+"merged.unigram")
}

// Merge those boys!
func mergeCoocs(coocsDir string, l *Logger) {
	into := ConstructCooc()
	cFiles, _ := ioutil.ReadDir(coocsDir)
	for _, file := range cFiles {
		s := file.Name()
		if strings.Contains(s, ".cooc") && !strings.Contains(s, "merged") {
			l.Log(fmt.Sprintf("\tloading %s...", s))
			LoadSingleCooc(into, coocsDir+s)
		}
	}
	SaveCooc(into, coocsDir+"merged.cooc", l)
}

func main() {
	var extractPath string

	// Required argument
	extractOption := flag.String("option", "",
		"option for extraction, \"unigram\" or \"cooc\"; add \"-merge\" to merge?")

	// possibly required arguments
	flag.StringVar(&extractPath, "e", "",
		"path to the target gz file we will be extracting")

	unigramPath := flag.String("U", "",
		"path to the unigram to pre-load, if desired")

	coocPath := flag.String("C", "",
		"path for where to save Coocs, if desired")

	vocabSize := flag.Int("v", -1,
		"desired size of the vocabulary to perform extraction")

	window := flag.Int("w", -1,
		"window size, an integer indicating it (only dynamic weighting for now)")

	// Optional arguments.
	debug := flag.Bool("debug", false,
		"whether to run a debug profiler")

	cpuProfile := flag.Bool("pcpu", false,
		"whether to do CPU profiling (RAM profiling is the default)")

	logOption := flag.String("log", "print",
		"option for writing, printing, or silence [write, print, silent]")

	replaceDigits := flag.Bool("nodigits", false,
		"replace all digits with 0s during extraction")

	flag.Parse()

	// Check args.
	checkArgs(extractOption, &extractPath, unigramPath, coocPath, vocabSize, window)

	// TODO: pass to the logger all args and log them.
	l := ConstructLogger(*logOption)

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

	switch *extractOption {
	case "unigram-merge":
		mergeUnigrams(uPth, l)

	case "cooc-merge":
		mergeCoocs(*coocPath, l)

	case "unigram":
		exPath := loadExperimentPath(extractPath)
		l.Log(fmt.Sprintf("Will extract from path %s...", exPath))
		if _, err := os.Stat(uPth); os.IsNotExist(err) {
			l.Log("\textracting its unigram...")
			unigram = UnigramExtraction(extractPath, *replaceDigits, l)
			l.Log("\tserializing its unigram...")
			SerializeUnigram(unigram, uPth)
		}
	case "cooc":
		exPath := loadExperimentPath(extractPath)
		l.Log(fmt.Sprintf("Loading unigram from %s...", uPth))
		unigram = LoadUnigram(uPth)
		l.Log(fmt.Sprintf("Filtering unigram to %d most frequent tokens...", *vocabSize))
		unigram = FilterUnigram(unigram, *vocabSize)
		c := CoocExtraction(exPath, unigram, *window, *replaceDigits, l)
		l.Log("Serializing coocs...")
		SerializeCooc(c, *coocPath, l)
	}
	l.Log("Finished.")
}
