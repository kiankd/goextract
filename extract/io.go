package main

import (
	"compress/gzip"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

/* Basic IO helpers. */

// ReadGzFile - reads a gzip file.
func ReadGzFile(filename string) ([]byte, error) {
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

/* IO for Unigrams. */

// SerializeUnigram - writes the unigram to disk in a nice way
func SerializeUnigram(u *Unigram, fullPath string) error {
	if f, err := os.Create(fullPath); err == nil {
		defer f.Close()
		// oov is the header.
		f.WriteString(fmt.Sprintf("%s %d\n", OOV, u.oovCount))
		for _, code := range u.idx {
			word := u.Decode(code)
			count := u.counter[code]
			s := fmt.Sprintf("%d %s %d\n", code, word, count)
			f.WriteString(s)
		}
	} else {
		log.Fatal("Cannot write unigram.")
		return err
	}
	return nil
}

// Helper functions for LoadUnigram.
func parseUnigramLine(trip string) (string, int, int) {
	split := strings.Split(trip, " ")
	if len(split) != 3 {
		panic(fmt.Sprintf("Corrupted unigram encoding - %d spaces!\n", len(split)))
	}
	word := split[1]
	code, err1 := strconv.Atoi(split[0])
	count, err2 := strconv.Atoi(split[2])
	if err1 != nil && err2 != nil {
		panic(fmt.Sprintf("Corrupted unigram encoding! Str is: %s", trip))
	}
	return word, code, count
}

func parseOovCount(header string) int {
	split := strings.Split(header, " ")
	if len(split) != 2 {
		panic(fmt.Sprintf("Corrupted unigram header - %d spaces!\n", len(split)))
	}
	word := split[0]
	if word != OOV {
		panic(fmt.Sprintf("Corrupted unigram header - should be OOV but got %s!\n", word))
	}
	count, err := strconv.Atoi(split[1])
	if err != nil {
		panic(fmt.Sprintf("Corrupted unigram header! Header is: %s", header))
	}
	return count
}

// LoadUnigram - reads a unigram from disk
func LoadUnigram(fullPath string) *Unigram {
	if f, err := os.Open(fullPath); err == nil {
		defer f.Close()

		if bytes, err := ioutil.ReadAll(f); err == nil {
			fullStr := string(bytes)
			triples := strings.Split(fullStr, "\n")

			// minus 1 for trailing newline at end of a unigram doc
			// another minus for the OOV header.
			u := ConstructAllocatedUnigram(len(triples) - 2)
			u.oovCount = parseOovCount(triples[0])
			for _, trip := range triples[1:] {
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

/* IO for Coocs. */

// Filters out the counts that are too small before serializing.
func divideAndFilterMapData(m map[int64]float32, mincount float32) ([]int64, []float32) {
	keys := make([]int64, 0, len(m))
	vals := make([]float32, 0, len(m))
	for key, count := range m {
		if count > mincount {
			keys = append(keys, key)
			vals = append(vals, count)
		}
	}
	return keys, vals
}

// SerializeCooc - Helper to write a Cooc to disk in binary (gob).
func SerializeCooc(c *Cooc, mincount float32, fullPath string, l *Logger) {
	keys, vals := divideAndFilterMapData(c.Counter, mincount)
	start := 0
	end := GOBLEN
	for fnum := 0; start < len(keys); fnum++ {
		if end > len(keys) {
			end = len(keys)
		}
		encodeFile, err := os.Create(fmt.Sprintf("%s.gob%d", fullPath, fnum))
		if err != nil {
			panic(err)
		}
		l.Log("\tserializing " + encodeFile.Name())
		encoder := gob.NewEncoder(encodeFile)
		err = encoder.Encode(CoocData{keys[start:end], vals[start:end]})
		if err != nil {
			panic(err)
		}
		encodeFile.Close()
		start += GOBLEN
		end += GOBLEN
	}
}

// LoadCooc - loads a cooc from the gob binary!
func LoadCooc(into *Cooc, fullPath string, l *Logger) {
	files, err := filepath.Glob(fullPath + ".gob*")
	if err != nil {
		panic(err)
	}
	for _, f := range files {
		l.Log("\tloading " + f + "...")
		LoadSingleCooc(into, f)
	}
}

// LoadSingleCooc - loads a single cooc file into a Cooc
func LoadSingleCooc(into *Cooc, fullPath string) {
	decodeFile, err := os.Open(fullPath)
	if err != nil {
		panic(err)
	}
	coocData := CoocData{}
	decoder := gob.NewDecoder(decodeFile)
	decoder.Decode(&coocData)
	into.LoadCoocData(coocData)
	decodeFile.Close()
}

// SaveCooc - saves it into easy-readable text format.
func SaveCooc(c *Cooc, u *Unigram, mincount float32, fullPath string) {
	fi, err := os.Create(fullPath)
	if err != nil {
		panic(err)
	}
	defer fi.Close()

	i, b := 0, 0
	var str strings.Builder
	for cantor, count := range c.Counter {
		if b == STRBUF || i == len(c.Counter)-1 {
			fi.WriteString(str.String())
			str.Reset()
			b = 0
		}
		if count >= mincount {
			k1, k2 := InverseCantor(cantor)
			if u == nil {
				str.WriteString(fmt.Sprintf("%d %d %f\n", k1, k2, count))
			} else {
				s1 := u.Decode(k1)
				s2 := u.Decode(k2)
				str.WriteString(fmt.Sprintf("%s %s %f\n", s1, s2, count))
			}
			b++
		}
		i++
	}
}

func parseWeightsStr(wstr []string) []float32 {
	weights := make([]float32, len(wstr))
	for i := 0; i < len(wstr); i++ {
		w, err := strconv.ParseFloat(wstr[i], 64)
		if err != nil {
			panic(err)
		}
		weights[i] = float32(w)
	}
	return weights
}

// LoadCustomWeights - helps for loading custom weight files.
func LoadCustomWeights(fullPath string) ([]float32, []float32) {
	wFile, err := os.Open(fullPath)
	if err != nil {
		panic(err)
	}
	if bytes, err := ioutil.ReadAll(wFile); err == nil {
		var lWeightsStr []string
		var rWeightsStr []string

		fullStr := string(bytes)
		lines := strings.Split(fullStr, "\n")
		for _, line := range lines {
			// Allows user to make comments in their weight files.
			if strings.HasPrefix(line, "#") || line == "\n" {
				continue
			}
			if len(lWeightsStr) == 0 {
				lWeightsStr = strings.Split(line, " ")
			} else if len(rWeightsStr) == 0 {
				rWeightsStr = strings.Split(line, " ")
				break
			}
		}
		if len(lWeightsStr) == 0 || len(rWeightsStr) == 0 {
			panic("Empty weight string!")
		}
		lW := parseWeightsStr(lWeightsStr)
		rW := parseWeightsStr(rWeightsStr)
		if (len(lW) > 1 && lW[len(lW)-1] == 0) || (len(rW) > 1 && rW[len(rW)-1] == 0) {
			panic(fmt.Sprintf("Improperly formatted weight strings! %s\n%s", lWeightsStr, rWeightsStr))
		}
		return lW, rW
	}
	panic(err)
}
