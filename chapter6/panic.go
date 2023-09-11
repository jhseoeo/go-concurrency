package main

import (
	"fmt"
	"net/http"
)

type Handler func(http.ResponseWriter, *http.Request)

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h(w, r)
}

func PanicHandler(next Handler) Handler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			err := recover()
			if err != nil {
				fmt.Println(err)
			}
		}()
		next(w, r)
	}
}

func main() {
	handler := func(w http.ResponseWriter, r *http.Request) {
		panic("Panic!")
	}

	http.Handle("/path", PanicHandler(handler))

	fmt.Println("Server started")
	http.ListenAndServe(":8080", nil)
}
