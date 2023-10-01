package main

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

type ProgressMeter struct {
	progress  int64
	timestamp int64
}

func (pm *ProgressMeter) Progress() {
	atomic.AddInt64(&pm.progress, 1)
	atomic.StoreInt64(&pm.timestamp, time.Now().UnixNano())
}

func (pm *ProgressMeter) Get() (int64, int64) {
	return atomic.LoadInt64(&pm.progress), atomic.LoadInt64(&pm.timestamp)
}

func longGoroutine(ctx context.Context, pm *ProgressMeter) {
	for {
		select {
		case <-ctx.Done():
			fmt.Println("Context cancelled")
			return
		default:
		}
		time.Sleep(time.Duration(rand.Intn(120)) * time.Millisecond)
		pm.Progress()
	}
}

func observer(ctx context.Context, cancel func(), progress *ProgressMeter) {
	tick := time.NewTicker(100 * time.Millisecond)
	defer tick.Stop()
	var lastProgress int64
	for {
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
			p, _ := progress.Get()
			if p == lastProgress {
				fmt.Println("No progress in the last 100ms")
				cancel()
				return
			} else {
				lastProgress = p
				fmt.Println("Progress:", p)
			}
		}
	}
}

func main() {
	var progress ProgressMeter
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		longGoroutine(ctx, &progress)
	}()
	go observer(ctx, cancel, &progress)
	wg.Wait()
}
