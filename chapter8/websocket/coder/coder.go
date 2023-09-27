package coder

import (
	"errors"
	"io"
)

func EncodeFromChan[T any](input <-chan T, encode func(T) ([]byte, error), write func([]byte) (int, error)) <-chan error {
	ret := make(chan error, 1)
	go func() {
		defer close(ret)
		for entry := range input {
			data, err := encode(entry)
			if err != nil {
				ret <- err
				return
			}
			if _, err := write(data); err != nil {
				if !errors.Is(err, io.EOF) {
					ret <- err
				}
				return
			}
		}
	}()
	return ret
}

func DecodeToChan[T any](decode func(*T) error) (<-chan T, <-chan error) {
	ret := make(chan T)
	errCh := make(chan error, 1)
	go func() {
		defer close(ret)
		defer close(errCh)
		var entry T
		for {
			if err := decode(&entry); err != nil {
				if !errors.Is(err, io.EOF) {
					errCh <- err
				}
				return
			}
			ret <- entry
		}
	}()
	return ret, errCh
}
