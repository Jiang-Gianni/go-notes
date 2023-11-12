// Package arst My Arst Package is here
package arst

import "fmt"

// Arst This is a comment.
const Arst = "Arst"

// Qwfp hello.
func Qwfp() {
	fmt.Println("Hello World")
}

var CustomError = myCustomError("my custom error")

// var CustomError = errors.New("astneio")

type myCustomError string

func (m myCustomError) Error() string {
	return string(m)
}
