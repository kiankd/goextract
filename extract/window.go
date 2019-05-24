package main

import (
	"math"
)

/* Window struct for help in having generalized windows. */

// Window - allows for arbitrarily defined context windows.
type Window struct {
	weights   []float32
	lnearest  int
	lfurthest int
	rnearest  int
	rfurthest int
	crtIdx    int
	nextIdx   int
	nextWIdx  int
	max       int
	ok        bool
}

// Start - starts the window for the next iteration.
func (w *Window) Start(idx, docLen int) {
	w.crtIdx = idx
	w.nextIdx = w.crtIdx - w.lfurthest - 1
	w.nextWIdx = -1
	w.max = docLen
	w.ok = true
}

// Next - gives the next context idx and weight for iteration.
func (w *Window) Next() (int, float32, bool) {
	if !w.ok {
		return -1, -1, false
	}
	w.nextIdx++
	w.nextWIdx++
	// need to do checks to see if we need to augment the counter.
	if w.nextIdx < 0 { // check for beginning of doc
		w.nextIdx = 0
	}
	diff := int(math.Abs(float64(w.crtIdx - w.nextIdx)))
	if diff < w.lnearest {
		w.nextIdx = w.crtIdx + w.rnearest
	}
	// check if at end of weights or if we are at end of doc
	if w.nextWIdx >= len(w.weights) || w.nextIdx >= w.max {
		w.ok = false
		return -1, -1, false
	}
	return w.nextIdx, w.weights[w.nextWIdx], true
}

func fillNearestFurthestWeights(into *Window, weights []float32, left bool) {
	// This checks for if it is totally asymmetric.
	if weights[0] == 0 && len(weights) == 1 {
		return
	}

	// Otherwise, we have some left and right context.
	nearest := 1
	for i := 0; i < len(weights); i++ {
		if weights[i] == 0 {
			nearest++
		} else {
			break
		}
	}
	if left {
		into.lnearest = nearest
		into.lfurthest = len(weights)
		for l := into.lfurthest; l >= into.lnearest; l-- {
			into.weights = append(into.weights, weights[l-1])
		}
	} else {
		into.rnearest = nearest
		into.rfurthest = len(weights)
		for r := into.rnearest; r <= into.rfurthest; r++ {
			into.weights = append(into.weights, weights[r-1])
		}
	}
}

// MakeWindow - creates a Window struct given a window size or a path.
func MakeWindow(w int, wPath string) *Window {
	if w != -1 && wPath != "" {
		panic("Ahh! Multiple window options provided!")
	}
	if w != -1 {
		weights := make([]float32, 2*w)
		fw := float32(w)
		// left side, tricky tricky!
		for i := 0; i < w; i++ {
			weights[i] = (fw - (fw - float32(i) - 1)) / fw
		}
		// right side
		for i := w; i < 2*w; i++ {
			weights[i] = (fw + (fw - float32(i))) / fw
		}
		win := Window{weights: weights,
			lnearest: 1, lfurthest: w,
			rnearest: 1, rfurthest: w}
		return &win
	}
	// Otherwise, we are doing a special custom window.
	lWeights, rWeights := LoadCustomWeights(wPath)
	weights := make([]float32, 0, len(lWeights)+len(rWeights))
	win := Window{weights: weights,
		lnearest: 1, lfurthest: 0,
		rnearest: 1, rfurthest: 0}
	fillNearestFurthestWeights(&win, lWeights, true)
	fillNearestFurthestWeights(&win, rWeights, false)
	return &win
}
