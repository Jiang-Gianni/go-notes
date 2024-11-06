package main

import (
	"fmt"

	"github.com/Jiang-Gianni/notes-golang/dt"
)

func main() {
	a := dt.Hello("world")
	b := dt.Hello("world")
	fmt.Println(a == b)
}
