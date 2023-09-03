package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

var cnt [5]int

//normal mutex version
//func philosopher(index int, firstFork, secondFork *sync.Mutex) {
//	for {
//		fmt.Printf("philosopher %d is thinking\n", index)
//		time.Sleep(time.Millisecond * time.Duration(rand.Intn(10)))
//
//		firstFork.Lock()
//		secondFork.Lock()
//
//		fmt.Printf("philosopher %d is eating\n", index)
//		cnt[index] = cnt[index] + 1
//		time.Sleep(time.Millisecond * time.Duration(rand.Intn(10)))
//
//		secondFork.Unlock()
//		firstFork.Unlock()
//	}
//}

// trylock version
func philosopher(index int, leftFork, rightFork *sync.Mutex) {
	for {
		//fmt.Printf("philosopher %d is thinking\n", index)
		time.Sleep(time.Millisecond * time.Duration(rand.Intn(10)))

		leftFork.Lock()
		if rightFork.TryLock() {
			//fmt.Printf("philosopher %d is eating\n", index)
			cnt[index] = cnt[index] + 1
			time.Sleep(time.Millisecond * time.Duration(rand.Intn(10)))
			rightFork.Unlock()
		}
		leftFork.Unlock()
	}
}

func main() {
	forks := [5]sync.Mutex{}

	//go philosopher(0, &forks[0], &forks[1])
	go philosopher(0, &forks[1], &forks[0])
	go philosopher(1, &forks[1], &forks[2])
	go philosopher(2, &forks[2], &forks[3])
	go philosopher(3, &forks[3], &forks[4])
	go philosopher(4, &forks[4], &forks[0])

	select {
	case <-time.After(time.Second * 10):
	}

	for i := 0; i < 5; i++ {
		fmt.Printf("philosopher %d ate %d times\n", i, cnt[i])
	}
}
