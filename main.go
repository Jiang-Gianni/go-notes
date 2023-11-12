// Main package documentation.
package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

func main() {

}

type Event struct {
	ID int
	time.Time
}

var ErrMy MyError = MyError{err: "my error"}

type MyError struct {
	err string
}

func (m MyError) Error() string {
	return m.err
}

func Work() (err error) {
	defer func() { err = WrapIfErr(err, "Work") }()
	err = DoWorkB()
	if err != nil {
		return err
	}
	return nil
}

func DoWorkB() (err error) {
	defer func() { err = WrapIfErr(err, "DoWorkB") }()
	return ErrMy
}

func WrapIfErr(err error, msg string) error {
	if err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	return nil
}

func PrintHttpResponse(url string) (err error) {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer func() {
		closeErr := resp.Body.Close()
		if err != nil {
			if closeErr != nil {
				log.Printf("failed to close response body: %v\n", closeErr)
			}
			return
		}
		err = closeErr
	}()
	bodyText, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	fmt.Printf("%s\n", bodyText)
	return nil
}
