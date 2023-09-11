package main

import (
	"errors"
	"fmt"
	"sync"
)

type Result1 struct {
	Result ResultType
	Error  error
}

type Result2 struct {
	Result ResultType
	Error  error
}

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
	resultCh1 := make(chan Result1)
	resultCh2 := make(chan Result2)
	canceled := make(chan struct{})
	cancelCh := make(chan struct{})
	defer close(cancelCh)

	go func() {
		once := sync.Once{}
		for range cancelCh {
			once.Do(func() {
				close(canceled)
			})
		}
	}()

	go func() {
		result, err := compute(100)
		if err != nil {
			cancelCh <- struct{}{}
			resultCh1 <- Result1{Error: err}
			return
		}

		select {
		case <-canceled:
			close(resultCh1)
			return
		default:
		}

		resultCh1 <- Result1{Result: result}
	}()

	go func() {
		result, err := compute(-1)
		if err != nil {
			cancelCh <- struct{}{}
			resultCh2 <- Result2{Error: err}
			return
		}

		select {
		case <-canceled:
			close(resultCh2)
			return
		default:
		}

		resultCh2 <- Result2{Result: result}
	}()

	result1, ok1 := <-resultCh1
	result2, ok2 := <-resultCh2

	if !ok1 || !ok2 {
		fmt.Println("canceled")
	}

	fmt.Println(result1.Result, result1.Error)
	fmt.Println(result2.Result, result2.Error)
}
