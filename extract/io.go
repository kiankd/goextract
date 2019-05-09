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

// GOBLEN - max num of items for a .gob file. 70 million.
const GOBLEN int = 7 * 1e7

// STRBUF - max num of strs for a .txt file write buffer, 1 million.
const STRBUF int = 1e6

// MINCOUNT - minimum value for the Nij statistics.
const MINCOUNT float64 = 100

// VERYMINCOUNT - the very min value for Nij statistics during preliminary extraction.
const VERYMINCOUNT float64 = 5

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

/* IO for Coocs. */

// Filters out the counts that are too small before serializing.
func divideAndFilterMapData(m map[int64]float64) ([]int64, []float64) {
	keys := make([]int64, 0, len(m))
	vals := make([]float64, 0, len(m))
	for key, count := range m {
		if count > VERYMINCOUNT {
			keys = append(keys, key)
			vals = append(vals, count)
		}
	}
	return keys, vals
}

// SerializeCooc - Helper to write a Cooc to disk in binary (gob).
func SerializeCooc(c *Cooc, fullPath string, l *Logger) {
	keys, vals := divideAndFilterMapData(c.Counter)
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
func SaveCooc(c *Cooc, fullPath string) {
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
		if count >= MINCOUNT {
			k1, k2 := InverseCantor(cantor)
			str.WriteString(fmt.Sprintf("%d %d %f\n", k1, k2, count))
			b++
		}
		i++
	}
}

func parseWeightsStr(wstr []string) []float64 {
	weights := make([]float64, len(wstr))
	for i := 0; i < len(wstr); i++ {
		w, err := strconv.ParseFloat(wstr[i], 64)
		if err != nil {
			panic(err)
		}
		weights[i] = w
	}
	return weights
}

// LoadCustomWeights - helps for loading custom weight files.
func LoadCustomWeights(fullPath string) ([]float64, []float64) {
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
