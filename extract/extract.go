package main

import (
	"compress/gzip"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

// readGzFile - reads a gzip file.
func readGzFile(filename string) ([]byte, error) {
	fi, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer fi.Close()

	fz, err := gzip.NewReader(fi)
	if err != nil {
		return nil, err
	}
	defer fz.Close()

	s, err := ioutil.ReadAll(fz)
	if err != nil {
		return nil, err
	}
	return s, nil
}

// ReadParseGz - reads a gzip and then parses it into documents.
func ReadParseGz(filename string, replaceDigits bool, logger *Logger) [][]string {
	logger.Log(fmt.Sprintf("Reading GZ file %s...", filename))
	byteArr, _ := readGzFile(filename)

	logger.Log("\tconverting to strings...")
	fullStr := string(byteArr)
	docs := strings.Split(fullStr, "\n")

	// Using a channel in Parse to make this very fast.
	logger.Log(fmt.Sprintf("\tparsing %d initial documents...", len(docs)))
	return Parse(docs, replaceDigits)
}

// CoocMerger - manages merging for Coocs with concurrency in mind.
type CoocMerger struct {
	state *Cooc
	nDocs int
	input chan *Cooc
	done  chan bool
}

func (merger *CoocMerger) listen() {
	for i := 0; i < merger.nDocs; i++ {
		received := <-merger.input
		merger.state.merge(received)
	}
	merger.done <- true
}

// FullExtraction - performs the full extraction pipeline.
func FullExtraction(
	filename string,
	maxVocabSize int,
	window int,
	replaceDigits bool,
	logger *Logger) (*Unigram, *Cooc) {

	documents := ReadParseGz(filename, replaceDigits, logger)
	u, encodedDocs := FullUnigramExtraction(documents, maxVocabSize, logger)

	logger.Log(fmt.Sprintf("Extracting cooccurences from %d docs...", len(encodedDocs)))
	merger := CoocMerger{
		state: ConstructCooc(),
		nDocs: len(encodedDocs),
		input: make(chan *Cooc, BUFFERSIZE),
		done:  make(chan bool)}
	go merger.listen()

	for _, doc := range encodedDocs {
		go func(document []int) {
			merger.input <- ExtractCooc(document, window)
		}(doc)
	}
	<-merger.done
	logger.Log("Finished.")
	return u, merger.state
}
