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
	errCh := make(chan error)
	defer close(errCh)
	canceled := make(chan struct{})

	errs := make([]error, 0)
	go func() {
		once := sync.Once{}
		for err := range errCh {
			errs = append(errs, err)
			once.Do(func() {
				close(canceled)
			})
		}
	}()

	resultCh1 := make(chan Result1)
	resultCh2 := make(chan Result2)

	go func() {
		defer close(resultCh1)
		result, err := compute(100)
		if err != nil {
			errCh <- err
			return
		}

		select {
		case <-canceled:
			return
		default:
		}
		resultCh1 <- Result1{Result: result}
	}()

	go func() {
		defer close(resultCh2)
		result, err := compute(100)
		if err != nil {
			errCh <- err
			return
		}

		select {
		case <-canceled:
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

	fmt.Println(result1, result2)
}
