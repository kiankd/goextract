package main

/* Window struct for help in having generalized windows. */

// Window - allows for arbitrarily defined context windows.
type Window struct {
	lWeights []float32
	rWeights []float32
	rstart   int
	lstart   int
}

// GetLeftStartEnd - gets the left start and end idxs of the Window
func (w *Window) GetLeftStartEnd() (int, int) {
	return w.lstart, len(w.lWeights)
}

// GetRightStartEnd - gets the right start and end idxs of the Window
func (w *Window) GetRightStartEnd() (int, int) {
	return w.rstart, len(w.rWeights)
}

// MakeWindow - creates a Window struct given a window size or a path.
func MakeWindow(w int, wPath string) *Window {
	if w != -1 && wPath != "" {
		panic("Ahh! Multiple window options provided!")
	}
	var (
		lWeights []float32
		rWeights []float32
	)

	// Integer-based weighting (dynamic only.)
	if w != -1 {
		weights := make([]float32, w)
		for i := 0; i < w; i++ {
			weights[i] = float32(w-i) / float32(w)
		}
		lWeights = weights
		rWeights = weights
	} else { // Otherwise, just load in the weights!
		lWeights, rWeights = LoadCustomWeights(wPath)
	}

	// Finally...
	l := 0
	for l < len(lWeights) && lWeights[l] == 0 {
		l++
	}
	r := 0
	for r < len(rWeights) && rWeights[r] == 0 {
		r++
	}
	win := Window{
		lWeights: lWeights,
		rWeights: rWeights,
		lstart:   l,
		rstart:   r}
	return &win
}
