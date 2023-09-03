package main

import (
	"fmt"
	"math/rand"
	"time"
)

var cnt_ [5]int

func philosopher_channel(index int, leftFork, rightFork chan bool) {
	for {
		//fmt.Printf("Philosopher %d is thinking.\n", index)
		time.Sleep(time.Millisecond * time.Duration(rand.Intn(10)))
		select {
		case <-leftFork:
			select {
			case <-rightFork:
				//fmt.Printf("Philosopher %d is eating.\n", index)
				cnt_[index] = cnt_[index] + 1
				time.Sleep(time.Millisecond * time.Duration(rand.Intn(10)))
				rightFork <- true
			default:
			}
			leftFork <- true
		}
	}
}

func main() {
	var forks [5]chan bool
	for i := range forks {
		forks[i] = make(chan bool, 1)
		forks[i] <- true
	}

	go philosopher_channel(0, forks[0], forks[1])
	go philosopher_channel(1, forks[1], forks[2])
	go philosopher_channel(2, forks[2], forks[3])
	go philosopher_channel(3, forks[3], forks[4])
	go philosopher_channel(4, forks[4], forks[0])

	select {
	case <-time.After(time.Second * 10):
	}

	for i := 0; i < 5; i++ {
		fmt.Printf("philosopher %d ate %d times\n", i, cnt_[i])
	}
}
