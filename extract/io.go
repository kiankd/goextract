package main

import (
	"compress/gzip"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"log"
	"os"
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
func splitMap(m map[int64]float64) (m1, m2 map[int64]float64) {
	m1 = make(map[int64]float64, len(m)/2+1)
	m2 = make(map[int64]float64, len(m)/2+1)
	b := true
	for key, value := range m {
		if b {
			m1[key] = value
			b = false
		} else {
			m2[key] = value
			b = true
		}
	}
	return
}

// SerializeCooc - Helper to write a Cooc to disk in binary (gob).
func SerializeCooc(c *Cooc, fullPath string, l *Logger) {
	encodeFile, err := os.Create(fullPath + ".gob0")
	if err != nil {
		panic(err)
	}
	encoder := gob.NewEncoder(encodeFile)
	if err := encoder.Encode(c.Counter); err != nil {
		l.Log("Cooc too big, splitting into 2 files...")
		map1, map2 := splitMap(c.Counter)
		l.Log("\tencoding gob0...")
		e1 := encoder.Encode(map1)
		if e1 != nil {
			panic(e1)
		}
		l.Log("\tencoding gob1...")
		encodeFile2, err := os.Create(fullPath + ".gob1")
		if err != nil {
			panic(err)
		}
		encoder2 := gob.NewEncoder(encodeFile2)
		e2 := encoder2.Encode(map2)
		if e2 != nil {
			panic(e2)
		}
	}
	encodeFile.Close()
}

// LoadCooc - loads a cooc from the gob binary!
func LoadCooc(into *Cooc, fullPath string, l *Logger) {
	l.Log("\tloading gob0...")
	decodeFile, err := os.Open(fullPath + ".gob0")
	if err != nil {
		panic(err)
	}
	defer decodeFile.Close()
	decoder := gob.NewDecoder(decodeFile)
	decoder.Decode(&into.Counter)

	decodeFile2, err2 := os.Open(fullPath + ".gob1")
	if err2 != nil {
		return
	}
	l.Log("\tloading gob1...")
	defer decodeFile2.Close()
	decoder2 := gob.NewDecoder(decodeFile2)
	decoder2.Decode(&into.Counter)
}
