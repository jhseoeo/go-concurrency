package main

import "sync"

func philosopher(left, right *sync.Mutex) {
	for {
		left.Lock()
		right.Lock()
		// eat
		right.Unlock()
		left.Unlock()
	}
}

func main() {
	forks := [2]sync.Mutex{}
	go philosopher(&forks[0], &forks[1])
	go philosopher(&forks[1], &forks[0])
	select {}
}
