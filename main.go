package main

import "fmt"

func main() {
	var limit = 2
	hugeSlice := []string{
		"task 1",
		"task 2",
		"task 3",
		"task 4",
	}
	sem := make(chan struct{}, limit)
	for _, task := range hugeSlice {
		sem <- struct{}{}
		go func(task string) {
			// perform task
			fmt.Println(task)
			<-sem
		}(task)
	}
	for n := limit; n > 0; n-- {
		sem <- struct{}{}
	}
}
