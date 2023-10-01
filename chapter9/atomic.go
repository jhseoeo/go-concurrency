package main

import (
	"fmt"
	"sync/atomic"
)

func main() {
	var done atomic.Bool
	var a int
	go func() {
		a = 1
		done.Store(true)
	}()
	if done.Load() {
		fmt.Println(a)
	}
}
