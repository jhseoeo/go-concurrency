package main

import (
	"errors"
	"fmt"
	"sync"
)

type ResultType int

func compute(step int) (ResultType, error) {
	result := ResultType(0)

	if step < 0 {
		return result, errors.New("invalid step")
	}
	for i := 1; i <= step; i++ {
		result += ResultType(i)
	}

	return result, nil
}

func main() {
	wg := sync.WaitGroup{}
	wg.Add(2)

	var err1, err2 error
	resultCh1 := make(chan ResultType, 1)
	resultCh2 := make(chan ResultType, 1)

	go func() {
		defer wg.Done()
		defer close(resultCh1)
		result, err := compute(100)
		if err != nil {
			err1 = err
			return
		}
		resultCh1 <- result
	}()
	go func() {
		defer wg.Done()
		defer close(resultCh2)
		result, err := compute(100)
		if err != nil {
			err2 = err
			return
		}
		resultCh2 <- result
	}()

	wg.Wait()
	if err1 != nil {
		fmt.Println("1:", err1)
	}
	if err2 != nil {
		fmt.Println("2:", err2)
	}

	result1, ok1 := <-resultCh1
	result2, ok2 := <-resultCh2

	if !ok1 || !ok2 {
		fmt.Println("channel closed")
	}

	fmt.Println(result1, result2)
}
