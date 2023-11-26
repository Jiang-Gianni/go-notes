package main

import "testing"

func TestDoWorkB(t *testing.T) {
	err := DoWorkB()
	if err == nil {
		t.Fatalf("err: %s", err)
	}
}
