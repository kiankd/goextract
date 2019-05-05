package main

import "testing"

func TestUnigramIO(t *testing.T) {
	documents := LoadSampleWords()
	u := ExtractUnigram(documents)
	SerializeUnigram(u, "/tmp/ex.unigram")
	u2 := LoadUnigram("/tmp/ex.unigram")
	if u.Len() != u2.Len() {
		t.Error("Different length after serializing!")
	}
	for code, count := range u2.counter {
		if u.counter[code] != count {
			t.Error("Different counts after serializing!")
		}
		if u.decoder[code] != u2.decoder[code] {
			t.Error("Different codes after serializing!")
		}
	}
}

func TestCoocIO(t *testing.T) {
	l := ConstructLogger("print")
	documents := LoadSampleWords()
	u := ExtractUnigram(documents)
	c := CoocExtraction("../data/sample.txt.gz", u, 5, false, l)

	l.Log("Seriailizing...")
	SerializeCooc(c, "/tmp/ex.cooc")
	c2 := ConstructCooc()
	l.Log("Loading...")
	LoadCooc(c2, "/tmp/ex.cooc", l)

	if len(c.counter) != len(c2.counter) {
		t.Errorf("Different length after serializing! %d orig vs %d\n", len(c.counter), len(c2.counter))
	}
	for code, count := range c2.counter {
		if c.counter[code] != count {
			t.Error("Different counts after serializing!")
		}
	}
}
