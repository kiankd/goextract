package main

import (
	"fmt"
	"testing"
)

func WeightingIntegrationTest(win *Window, maxIter int, t *testing.T) {
	count := 0
	for {
		_, w, ok := win.Next()
		if !ok {
			break
		}
		if w == 0 {
			t.Error("We have a 0 weight for some reason!")
		}
		count++
	}
	if count > maxIter {
		t.Error("More iterations than total number of weights!")
	}
}

func TestBasicWeighting(t *testing.T) {
	win := MakeWindow(5, "")
	window := 5
	values := []float32{0.2, 0.4, 0.6, 0.8, 1, 1, 0.8, 0.6, 0.4, 0.2}
	WindowValidate(values, win, t)
	for i := 0; i < 1000; i++ {
		win.Start(i, 1000)
		WeightingIntegrationTest(win, len(values), t)
		win.Start(i, 1000)

		// Functionality test
		if i > window && i < 1000-window {
			start, end := i-window, i+window
			crt := 0
			for c := start; c <= end; c++ {
				if i != c {
					wcont, wweight, ok := win.Next()
					if !ok {
						t.Error("Iterator says not okay, but it should be okay!")
					}
					if wcont != c {
						t.Errorf("Got context %d, wanted %d\n", wcont, c)
					}
					if wweight != values[crt] {
						t.Errorf("(%d, %d) Got weight %f, wanted %f!\n", i, c, wweight, values[crt])
					}
					crt++
				}
			}
			x, y, ok := win.Next()
			if ok {
				t.Errorf("Window should have ended iteration, but still says its ok! For %d it gives: %d %f\n", i, x, y)
			}
		}
	}
}

func WindowsEqualTest(w1, w2 *Window, t *testing.T) {
	if w1.lnearest != w2.lnearest {
		t.Error("Left nearests not equal!")
	}
	if w1.lfurthest != w2.lfurthest {
		t.Error("Left furthests not equal!")
	}
	if w1.rnearest != w2.rnearest {
		t.Error("Right nearests not equal")
	}
	if w1.rfurthest != w2.rfurthest {
		t.Error("Right furthests not equal!")
	}
	if len(w1.weights) != len(w2.weights) {
		t.Error("Different number of weights!!!")
		fmt.Println(w1.weights)
		fmt.Println(w2.weights)
		return
	}
	for i := 0; i < len(w1.weights); i++ {
		w1i, w2i := w1.weights[i], w2.weights[i]
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
	for _, w := range win.weights {
		if w == 0 {
			t.Error("Error, window has an inappropriate 0 value")
			break
		}
		wmap[w]++
	}
	for key, value := range tmap {
		if wmap[key] != value {
			t.Error("Error, different values of weights than expected!")
		}
	}
}

func TestCustomWeighting(t *testing.T) {
	w10 := MakeWindow(10, "")
	w10c := MakeWindow(-1, "../data/test_data/sample_w10.w")
	WindowsEqualTest(w10, w10c, t)

	// wfiles := []string{,
	// ../data/sample_asymmetricL.w
	// 	"../data/sample_crazy.w",
	// 	"../data/sample_receptive.w"}

	win := MakeWindow(-1, "../data/test_data/sample_asymmetricR.w")
	wtargs := []float32{1, 0.8, 0.6, 0.4, 0.2}
	WindowValidate(wtargs, win, t)
	for i := 0; i < 1000; i++ {
		win.Start(i, 1000)
		WeightingIntegrationTest(win, len(wtargs), t)
		win.Start(i, 1000)
		if i > 100 && i < 900 {
			widx := 0
			for c := i + 1; c <= i+len(wtargs); c++ {
				targW := wtargs[widx]
				wcont, wweight, ok := win.Next()
				if !ok {
					t.Error("Iterator says not okay, but it should be okay!")
				}
				if wcont != c {
					t.Errorf("Wrong context idx, got %d, wanted %d\n", wcont, c)
				}
				if wweight != targW {
					t.Errorf("Wrong weight, got %f, wanted %f\n", wweight, targW)
				}
				widx++
			}
			x, y, ok := win.Next()
			if ok {
				t.Errorf("Gave okay but shouldn't have! Given term %d, got context %d & weight %f\n", i, x, y)
			}
		}
	}

	win = MakeWindow(-1, "../data/test_data/sample_asymmetricL.w")
	wtargs = []float32{0.2, 0.4, 0.6, 0.8, 1}
	WindowValidate(wtargs, win, t)
	for i := 0; i < 1000; i++ {
		win.Start(i, 1000)
		WeightingIntegrationTest(win, len(wtargs), t)
		win.Start(i, 1000)
		if i > 100 && i < 900 {
			widx := 0
			for c := i - len(wtargs); c < i; c++ {
				targW := wtargs[widx]
				wcont, wweight, ok := win.Next()
				if !ok {
					t.Error("Iterator says not okay, but it should be okay!")
				}
				if wcont != c {
					t.Errorf("Wrong context idx, got %d, wanted %d\n", wcont, c)
				}
				if wweight != targW {
					t.Errorf("Wrong weight, got %f, wanted %f\n", wweight, targW)
				}
				widx++
			}
			x, y, ok := win.Next()
			if ok {
				t.Errorf("Gave okay but shouldn't have! Given term %d, got context %d & weight %f\n", i, x, y)
			}
		}
	}

	win = MakeWindow(-1, "../data/test_data/sample_receptive.w")
	wtargs = []float32{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0.5, 1, 0.5,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0.5, 1, 0.5}
	WindowValidate(wtargs, win, t)
	field := []float32{0.5, 1, 0.5, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0.5, 1, 0.5}
	for i := 0; i < 1000; i++ {
		win.Start(i, 1000)
		WeightingIntegrationTest(win, 6, t)
		win.Start(i, 1000)
		if i >= 100 && i <= 900 {
			widx := 0
			for c := i - (len(field) / 2); c <= i+(len(field)/2); c++ {
				targW := field[widx]
				if c != i {
					widx++
				} else {
					continue
				}
				if targW == 0 {
					continue
				}
				wcont, wweight, ok := win.Next()
				if !ok {
					t.Error("Iterator says not okay, but it should be okay!")
				}
				if wcont != c {
					t.Errorf("Wrong context idx, got %d, wanted %d\n", wcont, c)
				}
				if wweight != targW {
					t.Errorf("Wrong weight, got %f, wanted %f\n", wweight, targW)
				}
			}
			x, y, ok := win.Next()
			if ok {
				t.Errorf("Gave okay but shouldn't have! Given term %d, got context %d & weight %f\n", i, x, y)
			}
		}
	}
}
