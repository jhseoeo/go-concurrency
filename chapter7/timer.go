package main

import (
	"fmt"
	"time"
)

// func main() {
// 	timer := time.NewTimer(10 * time.Millisecond)
// 	timeout := make(chan struct{})

// 	go func() {
// 		<-timer.C
// 		close(timeout)
// 		fmt.Println("Timer expired")
// 	}()

// 	x := 0
// 	done := false
// 	for !done {
// 		select {
// 		case <-timeout:
// 			done = true
// 		default:
// 		}

// 		time.Sleep(1 * time.Millisecond)
// 		x++
// 	}
// 	fmt.Println("x =", x)
// }

func main() {
	timeout := make(chan struct{})

	time.AfterFunc(3*time.Second, func() {
		close(timeout)
		fmt.Println("Timer expired")
	})

	<-timeout
}
