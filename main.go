package main

import (
	"time"
)

func main() {
	tm := time.NewTimer(1)
	tm.Reset(100 * time.Millisecond)
	<-tm.C
	if !tm.Stop() {
		<-tm.C
	}
}
