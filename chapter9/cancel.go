package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

func CancelSupport() (cancel func(), isCancelled func() bool) {
	v := atomic.Bool{}
	cancel = func() {
		v.Store(true)
	}
	isCancelled = func() bool {
		return v.Load()
	}
	return
}

func main() {
	cancel, isCancelled := CancelSupport()
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			time.Sleep(100 * time.Millisecond)
			if isCancelled() {
				fmt.Println("Cancelled")
				return
			}
		}
	}()
	time.AfterFunc(2*time.Second, cancel)
	wg.Wait()
}
