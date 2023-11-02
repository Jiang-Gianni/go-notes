package main

import (
	"log"
	"net/http"
	"os"
	"time"
)

type Server struct {
	tls     bool
	timeout int
}

type option func(*Server)

func tls(val bool) option {
	return func(s *Server) {
		s.tls = val
	}
}

func timeout(ts int) option {
	return func(s *Server) {
		s.timeout = ts
	}
}

func NewServer(opts ...option) (*Server, error) {
	// Default Server
	s := &Server{
		tls:     false,
		timeout: 10,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s, nil
}

func main() {

	logger := log.New(os.Stdout, "yourPrefix ", log.LstdFlags|log.Lshortfile)

	logger.Println("Hello logger")

	http.HandleFunc("/", helloWorldHandler)
	server := &http.Server{
		Addr: ":3000",
	}

	http.TimeoutHandler(
		http.HandlerFunc(helloWorldHandler),
		time.Second,
		"timeout message",
	)

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}

}

func helloWorldHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello, World!"))
}
