package main

import (
	"fmt"
	"math"
	"sort"
)

// Unigram - a dictionary struct that is sortable by counts.
type Unigram struct {
	encoder  map[string]int
	decoder  map[int]string
	counter  map[int]int
	idx      []int
	oovCount int
}

func (u *Unigram) addStr(str string, count int) {
	if code, ok := u.encoder[str]; ok {
		u.counter[code] += count
	} else {
		newCode := len(u.encoder)
		u.encoder[str] = newCode
		u.decoder[newCode] = str
		u.counter[newCode] = count
	}
}

// FillIdx - constructs the indices in a Unigram, if necessary.
func (u *Unigram) FillIdx() {
	if len(u.idx) == len(u.counter) {
		return
	}
	u.idx = make([]int, len(u.counter))
	for i := range u.idx {
		u.idx[i] = i
	}
}

// Decode - decodes a single code
func (u *Unigram) Decode(code int) string {
	if str, ok := u.decoder[code]; ok {
		return str
	}
	return OOV
}

// Encode - encodes a single string, returns "code" and a bool "is-OOV"
func (u *Unigram) Encode(str string) (int, bool) {
	if code, ok := u.encoder[str]; ok {
		return code, false
	}
	return -1, true
}

// Merge - Unigram u eats another Unigram u2.
func (u *Unigram) Merge(u2 *Unigram) {
	for str, code2 := range u2.encoder {
		u.addStr(str, u2.counter[code2])
	}
}

/*****  Sorting interface *****/
func (u Unigram) Swap(i, j int) {
	u.idx[i], u.idx[j] = u.idx[j], u.idx[i]
}
func (u Unigram) Len() int {
	return len(u.idx)
}
func (u Unigram) Less(i, j int) bool {
	return u.counter[u.idx[i]] > u.counter[u.idx[j]]
}

/*****  Helpful constructors *****/

// ConstructUnigram - constructs a unigram object without any allocated memory.
func ConstructUnigram() *Unigram {
	u := Unigram{
		encoder: make(map[string]int),
		decoder: make(map[int]string),
		counter: make(map[int]int)}
	return &u
}

// ConstructAllocatedUnigram - build a memory-allocated unigram exactly for size.
func ConstructAllocatedUnigram(size int) *Unigram {
	u := Unigram{
		encoder: make(map[string]int, size),
		decoder: make(map[int]string, size),
		counter: make(map[int]int, size),
		idx:     make([]int, size)}
	for i := range u.idx {
		u.idx[i] = i
	}
	return &u
}

/***** Helpers *****/

// DescribeUnigram - returns a string that describes unigram according to verbosity
func DescribeUnigram(u *Unigram, verbosity int) string {
	s := ""
	if verbosity == 0 {
		return s
	}
	if verbosity >= 1 {
		s += "Verbosity shallow:\n"
		s += fmt.Sprintf("\tencoder length: %d\n", len(u.encoder))
		s += fmt.Sprintf("\tdecoder length: %d\n", len(u.decoder))
		s += fmt.Sprintf("\tidx     length: %d\n", len(u.idx))
	}
	if verbosity >= 2 {
		s += "Verbosity deep:\n"
		for _, code := range u.idx[:(verbosity * 2)] {
			s += fmt.Sprintf("\tCode %4d: %16s, count=%d\n", code, u.decoder[code], u.counter[code])
		}
	}
	return s
}

/***** Primary utility functions *****/

// Private boy that actually does extraction. Could make it multi-processed, but its okay.
func extractWithUnigram(documents [][]string, u *Unigram) {
	for _, doc := range documents {
		for _, word := range doc {
			u.addStr(word, 1)
		}
	}
}

// ExtractUnigram - extract a unigram from a set of docs, used in testing.
func ExtractUnigram(documents [][]string) *Unigram {
	u := ConstructUnigram()
	extractWithUnigram(documents, u)
	u.FillIdx()
	return u
}

// FilterUnigram - filters a unigram object to correspond to a vocabulary size.
func FilterUnigram(u *Unigram, maxVocabSize int) (fu *Unigram) {
	vocabSize := int(math.Min(float64(maxVocabSize), float64(u.Len())))
	fu = ConstructAllocatedUnigram(vocabSize)
	sort.Sort(u)
	oovCount := 0
	for newCode, oldCode := range u.idx {
		if newCode >= vocabSize || u.decoder[oldCode] == OOV {
			oovCount += u.counter[oldCode]
			continue
		}
		word := u.decoder[oldCode]
		fu.encoder[word] = newCode
		fu.decoder[newCode] = word
		fu.counter[newCode] = u.counter[oldCode]
	}
	fu.oovCount = oovCount
	return
}

// UnigramEncode - encodes a string list into the unigram codes.
func UnigramEncode(u *Unigram, documents [][]string) [][]int {
	encodedDocs := make([][]int, len(documents))
	ch := make(chan []int, BUFFERSIZE)
	done := make(chan bool)

	// listener
	go func() {
		for i := 0; i < len(documents); i++ {
			codes := <-ch
			idx := codes[len(codes)-1]
			encodedDocs[idx] = codes[:len(codes)-1]
		}
		done <- true
	}()

	// Speaker, puts the idx in there to always retain order!
	for d, document := range documents {
		go func(idx int, doc []string) {
			codes := make([]int, 0, len(doc)+1)
			for _, word := range doc {
				// Purge OOV words!
				if code, oov := u.Encode(word); !oov {
					codes = append(codes, code)
				}
			}
			codes = append(codes, idx)
			ch <- codes
		}(d, document)
	}
	<-done
	return encodedDocs
}
