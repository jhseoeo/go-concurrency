package main

import (
	"fmt"
	"time"
)

func monitor(heartbeat <-chan struct{}, done chan struct{}, tick <-chan time.Time) {
	var lastHeartbeat time.Time
	var numTicks int

	for {
		select {
		case <-tick:
			numTicks++
			if numTicks >= 2 {
				fmt.Printf("No progress since %s, exiting\n", lastHeartbeat)
				close(done)
				return
			}

		case <-heartbeat:
			lastHeartbeat = time.Now()
			numTicks = 0
		}
	}
}

func longRunningFunction(heartbeat chan<- struct{}, done chan struct{}) {
	for i := 0; i < 10; i++ {
		select {
		case <-done:
			return
		case heartbeat <- struct{}{}:
		}
		fmt.Printf("Job %d\n", i)
		time.Sleep(500 * time.Millisecond)
	}
	close(done)
}

func main() {
	heartbeat := make(chan struct{})
	defer close(heartbeat)
	done := make(chan struct{})
	tick := time.NewTicker(1 * time.Second)
	defer tick.Stop()

	go monitor(heartbeat, done, tick.C)
	go longRunningFunction(heartbeat, done)

	<-done
	fmt.Println("Long running function finished")
}
