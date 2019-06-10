package main

import (
	"testing"
)

func WindowsEqualTest(w1, w2 *Window, t *testing.T) {
	if len(w1.lWeights) != len(w2.lWeights) &&
		len(w1.rWeights) != len(w2.rWeights) {
		t.Error("Different number of weights!!!")
		return
	}
	for i := 0; i < len(w1.lWeights); i++ {
		w1i, w2i := w1.lWeights[i], w2.lWeights[i]
		if w1i != w2i {
			t.Errorf("Different weight values at %d: %f vs %f\n", i, w1i, w2i)
		}
	}
	for i := 0; i < len(w1.rWeights); i++ {
		w1i, w2i := w1.rWeights[i], w2.rWeights[i]
		if w1i != w2i {
			t.Errorf("Different weight values at %d: %f vs %f\n", i, w1i, w2i)
		}
	}
}

func WindowValidate(targs []float32, win *Window, t *testing.T) {
	tmap := make(map[float32]int)
	for _, w := range targs {
		if w > 0 {
			tmap[w]++
		}
	}
	wmap := make(map[float32]int)
	for _, w := range win.lWeights {
		wmap[w]++
	}
	for _, w := range win.rWeights {
		wmap[w]++
	}
	for key, value := range tmap {
		if wmap[key] != value {
			t.Error("Error, different values of weights than expected!")
		}
	}
}

func TestBasicWeighting(t *testing.T) {
	win := MakeWindow(5, "")
	values := []float32{0.2, 0.4, 0.6, 0.8, 1, 1, 0.8, 0.6, 0.4, 0.2}
	WindowValidate(values, win, t)
}

func TestCustomWeighting(t *testing.T) {
	// Tiny doc, sorted in increasing order for easy testing.
	doc := []int{10, 11, 12, 13, 14}

	// Ensure that a 10 token window equals what happens when you load the same weights.
	w10 := MakeWindow(10, "")
	w10c := MakeWindow(-1, "../data/test_data/sample_w10.w")
	WindowsEqualTest(w10, w10c, t)

	// Right custpm assymetric window testing.
	win := MakeWindow(-1, "../data/test_data/sample_asymmetricR.w")
	wtargs := []float32{1, 0.8, 0.6, 0.4, 0.2}
	WindowValidate(wtargs, win, t)
	cooc := ExtractCooc(doc, *win)
	for code := range cooc.Counter {
		i, j := InverseCantor(code)
		if i > j {
			t.Error("Bad right assymmetric extraction!")
		}
	}

	// Right custom assymetric window testing.
	win = MakeWindow(-1, "../data/test_data/sample_asymmetricL.w")
	wtargs = []float32{0.2, 0.4, 0.6, 0.8, 1}
	WindowValidate(wtargs, win, t)
	cooc = ExtractCooc(doc, *win)
	for code := range cooc.Counter {
		i, j := InverseCantor(code)
		if i < j {
			t.Error("Bad left assymmetric extraction!")
		}
	}

	// Big context window testing.
	win = MakeWindow(-1, "../data/test_data/sample_receptive.w")
	wtargs = []float32{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0.5, 1, 0.5,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0.5, 1, 0.5}
	if win.lstart != 10 {
		t.Error("Left start is not 10 when it should be!")
	}
	if win.rstart != 10 {
		t.Error("Right start is not 10 when it should be!")
	}
	WindowValidate(wtargs, win, t)
	cooc = ExtractCooc(doc, *win)
	if len(cooc.Counter) > 0 {
		t.Error("Bad big window extraction, got counts when it shouldnt!")
	}
	// field := []float32{0.5, 1, 0.5, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	// 	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0.5, 1, 0.5}
}
