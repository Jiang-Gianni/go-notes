package main_test

import (
	"testing"
)

var ch1 = make(chan struct{}, 1)
var ch2 = make(chan struct{}, 1)

func Benchmark_Select_OneCase(b *testing.B) {
	for i := 0; i < b.N; i++ {
		select {
		case ch1 <- struct{}{}:
			<-ch1
		}
	}
}
func Benchmark_Select_TwoCases(b *testing.B) {
	for i := 0; i < b.N; i++ {
		select {
		case ch1 <- struct{}{}:
			<-ch1
		case ch2 <- struct{}{}:
			<-ch2
		}
	}
}

func Benchmark_Select_OneNil(b *testing.B) {
	ch2 = nil
	for i := 0; i < b.N; i++ {
		select {
		case ch1 <- struct{}{}:
			<-ch1
		case ch2 <- struct{}{}:
			<-ch2
		}
	}
}

var vx int
var vy string

func Benchmark_TwoChannels(b *testing.B) {
	var x = make(chan int)
	var y = make(chan string)
	go func() {
		for {
			x <- 1
		}
	}()
	go func() {
		for {
			y <- "hello"
		}
	}()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		select {
		case vx = <-x:
		case vy = <-y:
		}
	}
}
func Benchmark_OneChannel_Interface(b *testing.B) {
	var x = make(chan interface{})
	go func() {
		for {
			x <- 1
		}
	}()
	go func() {
		for {
			x <- "hello"
		}
	}()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		select {
		case v := <-x:
			switch v := v.(type) {
			case int:
				vx = v
			case string:
				vy = v
			}
		}
	}
}
func Benchmark_OneChannel_Struct(b *testing.B) {
	type T struct {
		x int
		y string
	}
	var x = make(chan T)
	go func() {
		for {
			x <- T{x: 1}
		}
	}()
	go func() {
		for {
			x <- T{y: "hello"}
		}
	}()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v := <-x
		if v.y == "" {
			vx = v.x
		} else {
			vy = v.y
		}
	}
}
