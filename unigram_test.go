package main

import (
	"math"
	"sort"
	"testing"
)

/* Validator functions */
func validateUnigramLengths(u Unigram) bool {
	return len(u.encoder) == len(u.decoder) &&
		len(u.encoder) == len(u.counter) &&
		len(u.encoder) == len(u.idx) &&
		len(u.encoder) == u.Len()
}
func validateEncoderDecoder(u Unigram) bool {
	for word, code := range u.encoder {
		if u.decoder[code] != word {
			return false
		}
	}
	for code, word := range u.decoder {
		if u.encoder[word] != code {
			return false
		}
	}
	return true
}
func validateIndexing(u Unigram) bool {
	allCodes := make(map[int]bool)
	for _, code := range u.idx {
		if code < 0 || code >= len(u.idx) {
			return false
		}

		// ensure the code in the idx is in the decoder & encoder
		if word, ok := u.decoder[code]; ok {
			if _, ok2 := u.encoder[word]; !ok2 {
				return false
			}
		} else {
			return false
		}

		// check if a code is repeated
		_, ok := allCodes[code]
		if !ok {
			allCodes[code] = true
		} else {
			return false
		}
	}
	return len(allCodes) == len(u.idx)
}
func validateAll(u Unigram, t *testing.T) {
	if !validateUnigramLengths(u) {
		t.Error("Integrity lost: lengths in unigram do not equal!")
	}
	if !validateEncoderDecoder(u) {
		t.Error("Integrity lost: encoder and decoder do not align!")
	}
	if !validateIndexing(u) {
		t.Error("Integrity lost: indexing in unigram is not appropriate!")
	}
	if u.encoder[OOV] != 0 {
		t.Errorf("OOV code is nonzero! The code is %d instead!", u.encoder[OOV])
	}
}
func validateSorting(u Unigram, t *testing.T) {
	sort.Sort(u)
	prevCode := u.idx[0]
	for _, nextCode := range u.idx {
		if !(u.counter[prevCode] >= u.counter[nextCode]) {
			t.Errorf(`Unigram should be sorted in decreasing order but it is not!
			%f is < %f (but should be >=)!`, u.counter[prevCode], u.counter[nextCode])
			return
		}
		prevCode = nextCode
	}
}
func validateFiltering(u Unigram, fu Unigram, t *testing.T) {
	sort.Sort(u)
	sort.Sort(fu)
	minCount := 1e99
	for _, count := range fu.counter {
		minCount = math.Min(count, minCount)
	}
	for _, code := range u.idx {
		word := u.decoder[code]
		if word == OOV {
			continue
		}
		// If we have the word, then make sure its the same count.
		if fcode, ok := fu.encoder[word]; ok {
			if u.counter[code] != fu.counter[fcode] {
				t.Errorf("Original and filtered have different counts for: %s", word)
			}
		} else { // otherwise, make sure the count is smaller than the min.
			if !(u.counter[code] <= minCount) {
				t.Errorf("We did not properly filter, too big count for: %s", word)
			}
		}
	}
}

/* Tests */

func TestExtractUnigram(t *testing.T) {
	words := LoadSampleWords()
	u := ExtractUnigram(words)
	validateAll(u, t)
	validateSorting(u, t)
}

func TestFilterUnigram(t *testing.T) {
	words := LoadSampleWords()
	u := ExtractUnigram(words)

	vocabSizes := []int{10, 500, 1000, 50000}
	for _, vocabSize := range vocabSizes {
		fu := FilterUnigram(u, vocabSize)
		validateAll(fu, t)
		validateSorting(fu, t)
		validateFiltering(u, fu, t)
	}
}

func TestUnigramEncode(t *testing.T) {
	words := LoadSampleWords()
	u := ExtractUnigram(words)
	codes, docIdxs := UnigramEncode(u, words)
	newDocCodeCount := 0
	for i, code := range codes {
		if code == NEWDOC {
			newDocCodeCount++
			continue
		}
		word := words[i]
		wordHat := u.decode(code)
		if wordHat != word {
			t.Errorf("Decoding error: should be %s but got %s\n", word, wordHat)
		}
		codeHat := u.encode(word)
		if code != codeHat {
			t.Errorf("Encoding error: should be %d but got %d\n", code, codeHat)
		}
	}
	if len(docIdxs) != 937 && newDocCodeCount != 937 {
		t.Errorf("Different number of line breaks (%d) vs docidxs (%d) vs line break codes (%d)!\n",
			937, len(docIdxs), newDocCodeCount)
	}

}