package main

import (
	"fmt"
	"strings"
	"time"
)

// ReadParseGz - reads a gzip and then parses it into documents.
func ReadParseGz(filename string, replaceDigits bool, logger *Logger) [][]string {
	logger.Log(fmt.Sprintf("Reading GZ file %s...", filename))
	byteArr, _ := ReadGzFile(filename)

	logger.Log("\tconverting to strings...")
	fullStr := string(byteArr)
	docs := strings.Split(fullStr, "\n")

	// Using a channel in Parse to make this very fast.
	logger.Log(fmt.Sprintf("\tparsing %d initial documents...", len(docs)))
	return Parse(docs, replaceDigits)
}

/* Unigram Extraction */

// UnigramExtraction - to be used when using large amounts of data.
func UnigramExtraction(filename string, replaceDigits bool, logger *Logger) *Unigram {
	u := ConstructUnigram()
	documents := ReadParseGz(filename, replaceDigits, logger)
	logger.Log("\tdetermining the encoding and counting...")
	extractWithUnigram(documents, u)
	u.FillIdx()
	return u
}

/* Cooc Extraction */

// CoocMerger - manages merging for Coocs with concurrency in mind.
type CoocMerger struct {
	state *Cooc
	nDocs int
	input chan *Cooc
	done  chan bool
}

func (m *CoocMerger) listen() {
	for i := 0; i < m.nDocs; i++ {
		received := <-m.input
		m.state.Merge(received)
	}
	m.done <- true
}

// CoocExtraction - performs the full extraction pipeline.
func CoocExtraction(filename string, u *Unigram, window *Window, replaceDigits bool, logger *Logger) *Cooc {
	documents := ReadParseGz(filename, replaceDigits, logger)

	logger.Log("Encoding documents...")
	encodedDocs := UnigramEncode(u, documents)

	logger.Log(fmt.Sprintf("Extracting cooccurences from %d docs...", len(encodedDocs)))
	merger := CoocMerger{
		state: ConstructCooc(),
		nDocs: len(encodedDocs),
		input: make(chan *Cooc, BUFFERSIZE),
		done:  make(chan bool)}

	// listener
	go merger.listen()

	// speaker
	for i, doc := range encodedDocs {
		go func(document []int) {
			merger.input <- ExtractCooc(document, window)
		}(doc)
		if (i+1)%BUFFERSIZE == 0 { // Give some slack to let things catch up.
			for {
				if len(merger.input) == 0 {
					break
				} else {
					time.Sleep(250 * time.Millisecond)
				}
			}
			logger.Log(fmt.Sprintf("\t%d docs launched and merged", i+1))
		}
	}
	<-merger.done

	// Finished!
	logger.Log("\tfinished Cooc extraction!")
	return merger.state
}
