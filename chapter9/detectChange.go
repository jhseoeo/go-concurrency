package main

import (
	"fmt"
	"sync/atomic"
)

type SomeStruct struct {
	Field1 int
}

func computeNewCopy(oldValue SomeStruct) SomeStruct {
	return SomeStruct{oldValue.Field1 + 1}
}

var sharedValue atomic.Pointer[SomeStruct]

func updateSharedValue() {
	myCopy := sharedValue.Load()
	newCopy := computeNewCopy(*myCopy)
	if sharedValue.CompareAndSwap(myCopy, &newCopy) {
		fmt.Println("Update successful.")
	} else {
		fmt.Println("Another goroutine modified sharedValue.")
	}
}

func main() {}
