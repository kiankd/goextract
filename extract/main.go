package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/pkg/profile"
)

/* Globals ("ewwwww!" - I know, and I'm sorry...) */
var (
	GOBLEN       = int(7 * 1e7) // max num of items for a .gob file. 70 million.
	STRBUF       = int(1e6)     // max num of strs for a .txt file write buffer, 1 million.
	MINCOUNT     = float32(100) // minimum value for the Nij statistics after merging coocs.
	VERYMINCOUNT = float32(5)   // min value for Nij statistics during preliminary extraction.
	OOV          = "<OOV>"      // default string for out-of-vocabulary.
	BUFFERSIZE   = 2500         // max number of threads used for the merging channels
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
func checkArgs(opt, exP, uP, cP *string, v, w *int, winF *string) {
	emptyExp := *exP == ""
	emptyUni := *uP == ""
	emptyCoo := *cP == ""
	emptyVoc := *v <= 0
	emptyWin := *w <= 0 && *winF == ""
	switch *opt {
	case "unigram-merge":
		if emptyUni || emptyVoc {
			panic("No path specified for unigram-merging & no vocab size passed!")
		}
	case "cooc-merge":
		if emptyCoo {
			panic("No path specified for cooc-merging!")
		} else if !strings.HasSuffix(*cP, "/") {
			panic("Trying to merge coocs, but need a directory!")
		}
	case "unigram":
		if emptyExp || emptyUni {
			panic("No paths specified for unigram extraction!")
		}
	case "cooc":
		if emptyExp || emptyCoo || emptyWin || emptyUni {
			panic("Cooc extraction needs exp, coocpath, unigram, and window! Missing!")
		}
	default:
		panic(fmt.Sprintf("Option %s is invalid!\n", *opt))
	}
}

// Merge these boys!
func mergeUnigrams(unigramPath string, vocabSize int, l *Logger) {
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
	fu := FilterUnigram(u, vocabSize)
	SerializeUnigram(fu, unigramPath+"merged.unigram")
}

// Merge those boys!
func mergeCoocs(coocsDir string, u *Unigram, l *Logger) {
	into := ConstructCooc()
	cFiles, _ := ioutil.ReadDir(coocsDir)
	for _, file := range cFiles {
		s := file.Name()
		if strings.Contains(s, ".cooc") && !strings.Contains(s, "merged") {
			l.Log(fmt.Sprintf("\tloading %s...", s))
			LoadSingleCooc(into, coocsDir+s)
		}
	}
	l.Log("\tsaving coocs...")
	SaveCooc(into, u, coocsDir+"merged.cooc")
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

	windowF := flag.String("window", "",
		"path to a file containing window weights, formatted as shown in example.w")

	// Optional arguments.
	debug := flag.Bool("debug", false,
		"whether to run a debug profiler")

	cpuProfile := flag.Bool("pcpu", false,
		"whether to do CPU profiling (RAM profiling is the default)")

	logOption := flag.String("log", "print",
		"option for writing, printing, or silence [write, print, silent]")

	replaceDigits := flag.Bool("nodigits", false,
		"replace all digits with 0s during extraction")

	mergeAsStr := flag.Bool("strkeep", false,
		"pass when using option \"cooc-merge\" to save as strings, not idxs")

	minNij := flag.Float64("minnij", 100,
		"value of the minimum Nij to be serialized")

	vminNij := flag.Float64("verynij", 5,
		"value of the minimum Nij to be serialized")

	flag.Parse()

	// Check args.
	checkArgs(extractOption, &extractPath, unigramPath, coocPath, vocabSize, window, windowF)

	// Put the Nij args into the globals (will be changed in future)
	MINCOUNT = float32(*minNij)
	VERYMINCOUNT = float32(*vminNij)

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
		mergeUnigrams(uPth, *vocabSize, l)

	case "cooc-merge":
		if *mergeAsStr {
			u := LoadUnigram(uPth)
			mergeCoocs(*coocPath, u, l)
		} else {
			mergeCoocs(*coocPath, nil, l)
		}

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
		window := MakeWindow(*window, *windowF)
		c := CoocExtraction(exPath, unigram, window, *replaceDigits, l)
		l.Log("Serializing coocs...")
		SerializeCooc(c, *coocPath, l)
	}
	l.Log("Finished.")
}
