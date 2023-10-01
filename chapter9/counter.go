package main

import (
	"fmt"
	"sync/atomic"
)

func main() {
	var count int64
	for i := 0; i < 10000; i++ {
		go func() {
			atomic.AddInt64(&count, 1)
		}()
	}
	for {
		v := atomic.LoadInt64(&count)
		fmt.Println(v)
		if v == 10000 {
			break
		}
	}
}
