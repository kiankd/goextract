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

// GOBLEN - max num of items for a .gob file. 50 million.
const GOBLEN int = 5 * 1e7

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
func divideMapData(m map[int64]float64) ([]int64, []float64) {
	keys := make([]int64, len(m))
	vals := make([]float64, len(m))
	i := 0
	for key, val := range m {
		keys[i] = key
		vals[i] = val
		i++
	}
	return keys, vals
}

// SerializeCooc - Helper to write a Cooc to disk in binary (gob).
func SerializeCooc(c *Cooc, fullPath string, l *Logger) {
	keys, vals := divideMapData(c.Counter)
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
		decodeFile, err := os.Open(f)
		if err != nil {
			panic(err)
		}
		coocData := CoocData{}
		decoder := gob.NewDecoder(decodeFile)
		decoder.Decode(&coocData)
		into.LoadCoocData(coocData)
		decodeFile.Close()
	}
}
