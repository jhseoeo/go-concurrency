package filters

import "github.com/jhseoeo/go-concurrency/chapter8/streaming/store"

func MovingAvg(threshold float64, windowSize int, in <-chan store.Entry) <-chan store.AboveThresholdEntry {

	window := make(chan float64, windowSize)
	out := make(chan store.AboveThresholdEntry)
	go func() {
		defer close(out)
		var runningTotal float64
		for input := range in {
			if len(window) == windowSize {
				avg := runningTotal / float64(windowSize)
				if avg >= threshold {
					out <- store.AboveThresholdEntry{
						Entry: input,
						Avg:   avg,
					}
				}
				runningTotal -= <-window
			}
			window <- input.Value
			runningTotal += input.Value
		}
	}()
	return out
}
