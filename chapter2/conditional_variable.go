package main

import (
	"fmt"
	"math/rand"
	"sync"
)

func main() {

	lock := sync.Mutex{}
	fullCond := sync.NewCond(&lock)
	emptyCond := sync.NewCond(&lock)
	queue := NewQueue(10)

	producer := func(i int) {
		for {
			value := rand.Int()
			lock.Lock()
			for !queue.Enqueue(value) {
				fmt.Println("Queue is full, waiting from", i)
				fullCond.Wait()
			}
			lock.Unlock()
			emptyCond.Signal()
			// time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
		}
	}

	consumer := func(i int) {
		for {
			lock.Lock()
			var v int
			for {
				var ok bool
				if v, ok = queue.Dequeue(); !ok {
					fmt.Println("Queue is empty, waiting from", i)
					emptyCond.Wait()
					continue
				}
				break
			}

			lock.Unlock()
			fullCond.Signal()
			// time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
			fmt.Println("Consumed", v, "by", i)
		}
	}

	for i := 0; i < 10; i++ {
		go producer(i)
	}
	for i := 0; i < 10; i++ {
		go consumer(i)
	}
	select {}
}
