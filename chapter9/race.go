package main

import "fmt"

func main() {
	var str string
	var done bool
	go func() {
		str = "Done!"
		done = true
	}()
	for !done {
	}
	fmt.Println(str)
}
