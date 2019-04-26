package main

import (
	"strings"
)

type docMerger struct {
	nDocs    int
	realDocs int
	state    [][]string
	input    chan []string
	done     chan bool
}

func (m *docMerger) listen() {
	for i := 0; i < m.nDocs; i++ {
		words := <-m.input
		if len(words) > 0 {
			m.state = append(m.state, words)
			m.realDocs++
		}
	}
	m.done <- true
}

// Parse - parses into words
func Parse(documents []string) [][]string {
	merger := docMerger{
		nDocs: len(documents),
		state: make([][]string, 0, len(documents)),
		input: make(chan []string, 100),
		done:  make(chan bool)}
	go merger.listen()

	// Now send all the jobs.
	for _, docStr := range documents {
		go func(s string) {
			words := strings.Fields(s)
			merger.input <- words
		}(docStr)
	}
	<-merger.done
	return merger.state
}
