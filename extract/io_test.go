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
	l := ConstructLogger("silent")
	documents := LoadSampleWords()
	u := ExtractUnigram(documents)
	win2 := MakeWindow(2, "")
	c := CoocExtraction("../data/test_data/sample.txt.gz", u, win2, false, l)

	l.Log("Seriailizing...")
	SerializeCooc(c, float32(5.0), "/tmp/ex.cooc", l)
	c2 := ConstructCooc()
	l.Log("Loading...")
	LoadCooc(c2, "/tmp/ex.cooc", l)

	for code, count := range c2.Counter {
		if c.Counter[code] != count {
			t.Error("Different counts after serializing!")
		}
	}
}

func TestMerge(t *testing.T) {
	l := ConstructLogger("silent")
	documents := LoadSampleWords()
	u := ExtractUnigram(documents)
	win2 := MakeWindow(2, "")
	c := CoocExtraction("../data/test_data/sample.txt.gz", u, win2, false, l)
	l.Log("Serializing...")
	SerializeCooc(c, float32(5.0), "/tmp/ex.cooc", l)
	l.Log("Merging...")
	mergeCoocs(nil, float32(5.0), "/tmp/", l)
}
