package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
)

// OOV - default string for out-of-vocabulary.
const OOV = "<OOV>"

// BUFFERSIZE - for the merging channels
const BUFFERSIZE = 2500

// Unigram - a dictionary struct that is sortable by counts.
type Unigram struct {
	encoder map[string]int
	decoder map[int]string
	counter map[int]float64
	idx     []int
}

func (u *Unigram) addStr(str string, count float64) {
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
	return u.decoder[u.encoder[OOV]]
}

// Encode - encodes a single string
func (u *Unigram) Encode(str string) int {
	if code, ok := u.encoder[str]; ok {
		return code
	}
	return u.encoder[OOV]
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
		counter: make(map[int]float64)}
	u.encoder[OOV] = 0
	u.decoder[0] = OOV
	u.counter[0] = 0.0
	return &u
}

// ConstructAllocatedUnigram - build a memory-allocated unigram exactly for size.
func ConstructAllocatedUnigram(size int) *Unigram {
	u := Unigram{
		encoder: make(map[string]int, size),
		decoder: make(map[int]string, size),
		counter: make(map[int]float64, size),
		idx:     make([]int, size)}
	u.encoder[OOV] = 0
	u.decoder[0] = OOV
	u.counter[0] = 0.0
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
			s += fmt.Sprintf("\tCode %4d: %16s, count=%f\n", code, u.decoder[code], u.counter[code])
		}
	}
	return s
}

// SerializeUnigram - writes the unigram to disk in a nice way
func SerializeUnigram(u *Unigram, fullPath string) error {
	if f, err := os.Create(fullPath); err == nil {
		defer f.Close()
		for _, code := range u.idx {
			word := u.Decode(code)
			count := u.counter[code]
			s := fmt.Sprintf("%d %s %f\n", code, word, count)
			f.WriteString(s)
		}
	} else {
		log.Fatal("Cannot write unigram.")
		return err
	}
	return nil
}

// Helper function for LoadUnigram.
func parseUnigramLine(trip string) (string, int, float64) {
	split := strings.Split(trip, " ")
	if len(split) != 3 {
		panic(fmt.Sprintf("Corrupted unigram encoding - %d spaces!\n", len(split)))
	}
	word := split[1]
	code, err1 := strconv.Atoi(split[0])
	count, err2 := strconv.ParseFloat(split[2], 64)
	if err1 != nil && err2 != nil {
		panic(fmt.Sprintf("Corrupted unigram encoding! Str is: %s", trip))
	}
	return word, code, count
}

// LoadUnigram - reads a unigram from disk
func LoadUnigram(fullPath string) *Unigram {
	if f, err := os.Open(fullPath); err == nil {
		defer f.Close()

		if bytes, err := ioutil.ReadAll(f); err == nil {
			fullStr := string(bytes)
			triples := strings.Split(fullStr, "\n")

			// minus 1 for trailing newline at end of a unigram doc
			u := ConstructAllocatedUnigram(len(triples) - 1)
			for _, trip := range triples {
				if len(trip) == 0 {
					continue
				}
				word, code, count := parseUnigramLine(trip)
				u.counter[code] = count
				u.decoder[code] = word
				u.encoder[word] = code
			}
			return u
		}
	}
	panic("File does not exist or is corrupted.")
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
	oovCount := 0.0
	for i, oldCode := range u.idx {
		newCode := i + 1 // don't overwrite OOV, which should always have code 0
		if newCode >= vocabSize || u.decoder[oldCode] == OOV {
			oovCount += u.counter[oldCode]
			continue
		}
		word := u.decoder[oldCode]
		fu.encoder[word] = newCode
		fu.decoder[newCode] = word
		fu.counter[newCode] = u.counter[oldCode]
	}
	fu.counter[fu.encoder[OOV]] = oovCount
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

	// speaker, puts the idx in there to always retain order!
	for d, document := range documents {
		go func(idx int, doc []string) {
			codes := make([]int, len(doc)+1)
			for i, word := range doc {
				codes[i] = u.Encode(word)
			}
			codes[len(doc)] = idx
			ch <- codes
		}(d, document)
	}
	<-done
	return encodedDocs
}
