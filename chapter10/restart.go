package main

import (
	"fmt"
	"math/rand"
	"time"
)

func longRunning(done, heartBeat chan struct{}) {
	for {
		select {
		case <-done:
			return
		case heartBeat <- struct{}{}:
		default:
		}
		time.Sleep(100 * time.Millisecond)
		if rand.Intn(100) > 95 {
			fmt.Println("function failed")
			return
		}
	}
}

func restart(done chan struct{}, f func(done, heartBeat chan struct{}), timeout time.Duration) {
	for {
		funcDone := make(chan struct{})
		heartBeat := make(chan struct{})
		go func() {
			f(funcDone, heartBeat)
		}()

		timer := time.NewTimer(timeout)
		retry := false

		for !retry {
			select {
			case <-done:
				close(funcDone)
				return
			case <-heartBeat:
				if !timer.Stop() {
					<-timer.C
				}
				timer.Reset(timeout)
			case <-timer.C:
				fmt.Println("function timed out, restarting")
				close(funcDone)
				retry = true
			}
		}
	}
}

func main() {
	done := make(chan struct{})
	restart(done, longRunning, 500*time.Millisecond)
}
