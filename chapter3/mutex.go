package main

import (
	"fmt"
	"sync"
)

func main() {
	var m sync.Mutex
	var a int
	done := make(chan struct{})
	m.Lock()

	go func() {
		m.Lock()
		fmt.Println(a)
		m.Unlock()
		close(done)
	}()

	go func() {
		a = 1
		m.Unlock()
	}()

	<-done
}
