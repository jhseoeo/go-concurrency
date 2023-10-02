package main

import (
	"container/list"
	"math/rand"
	"sync"
)

func main() {
	wg := sync.WaitGroup{}
	wg.Add(2)
	ll := list.New()
	go func() {
		defer wg.Done()
		for i := 0; i < 100000; i++ {
			ll.PushBack(rand.Int())
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < 100000; i++ {
			ll.Remove(ll.Front())
		}
	}()
	wg.Wait()
}
