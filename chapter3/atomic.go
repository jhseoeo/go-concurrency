package main

import (
	"fmt"
	"sync/atomic"
)

func main() {
	var i int
	var v atomic.Value
	done := make(chan struct{})
	go func() {
		i = 1
		v.Store(1)
	}()

	go func() {
		for {
			if val, _ := v.Load().(int); val == 1 {
				fmt.Println(i)
				close(done)
				return
			}
		}
	}()

	<-done
}
