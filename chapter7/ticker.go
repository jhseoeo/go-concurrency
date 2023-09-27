package main

import (
	"fmt"
	"time"
)

func main() {
	start := time.Now()
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	done := time.After(1 * time.Second)

	for {
		select {
		case <-ticker.C:
			fmt.Println("tick:", time.Since(start).Milliseconds())
		case <-done:
			return
		}
	}
}
