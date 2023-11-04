# **Go Notes**
Random Go notes

- [**Go Notes**](#go-notes)
	- [**ShutDown**](#shutdown)
	- [**Graceful Shutdown**](#graceful-shutdown)
	- [**Request Context**](#request-context)
	- [**Client timeouts**](#client-timeouts)
	- [**Main abstraction**](#main-abstraction)
	- [**Configuration**](#configuration)
	- [**Functional Options**](#functional-options)
	- [**Functional Programming**](#functional-programming)
	- [**Goroutine**](#goroutine)
	- [**ldflags**](#ldflags)
	- [**pprof on web server**](#pprof-on-web-server)
	- [**pprof on performance**](#pprof-on-performance)
	- [**gcflags**](#gcflags)
	- [**Server timeouts**](#server-timeouts)
	- [**Custom Logger**](#custom-logger)
	- [**Testing Handlers**](#testing-handlers)
	- [**Transaction and Mutex**](#transaction-and-mutex)
	- [**Line of Sight**](#line-of-sight)
	- [**Headers against CSRF**](#headers-against-csrf)
	- [**Reflect to assign value**](#reflect-to-assign-value)
	- [**Worker Pool Semaphore pattern**](#worker-pool-semaphore-pattern)
	- [**Other resources**](#other-resources)




## [**ShutDown**](https://www.youtube.com/watch?v=YF1qSfkDGAQ&list=PL4WJSMupJdF8WPlGJQy4nlvWVWIPv7c3B&t=25m45s)

`defer` calls are not executed on interrupt / cancel, so use a signal channel to listen to the signal

```go
	sigChn := make(chan os.Signal)
	signal.Notify(sigChn, os.Interrupt, syscall.SIGTERM)

	defer func() {
		fmt.Println("done")
	}()

	timeout := time.NewTimer(3 * time.Second)
	select {
	case <-timeout.C:
	case <-sigChn:
	}
```


## [**Graceful Shutdown**](https://www.youtube.com/watch?v=9Q1RMueVHAg&t=5m50s)

```go
func GracefulShutdown(ctx context.Context, server *http.Server) error {
	sigChn := make(chan os.Signal, 1)
	signal.Notify(sigChn, os.Interrupt, syscall.SIGTERM)
	<-sigChn
	timeout := time.Duration(5*time.Millisecond) * time.Second
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	return server.Shutdown(ctx)
}
```


## [**Request Context**](https://www.youtube.com/watch?v=YF1qSfkDGAQ&list=PL4WJSMupJdF8WPlGJQy4nlvWVWIPv7c3B&t=31m50s)

If the request gets interrupted then the request's Context gets closed. Useful to interrupt long processing without wasting resources.

```go
func helloWorldHandler(w http.ResponseWriter, r *http.Request) {
	timer := time.NewTimer(time.Second * 3)
	select {
	case <-timer.C:
		log.Println("Hello World")
	case <-r.Context().Done():
		log.Println("Cancelled")
		return
	}
}
```


## [**Client timeouts**](https://www.youtube.com/watch?v=YF1qSfkDGAQ&list=PL4WJSMupJdF8WPlGJQy4nlvWVWIPv7c3B&t=9m)

DefaultClient in http/net doesn't timeout

```go
	client := http.Client{
		Timeout: time.Second,
	}
	res, err := client.Get("INSERT-URL")
```


## [**Main abstraction**](https://www.youtube.com/watch?v=IV0wrVb31Pg&t=10m40s)

Create a `run` function that handles all the initial settings and returns an error in case of failure (avoid multiple repeated error handling).

```go
func main() {
	if err := run(); err != nil {
		log.Println("error :", err)
		os.Exit(1)
	}
}
```

## [**Configuration**](https://www.youtube.com/watch?v=IV0wrVb31Pg&t=15m)

Put all the configurations in one place and define some default values.

To make sure that the right ones are being used when the server starts print out the configuration (no sensible data)

```go
type Configuration struct {
	Web struct {
		APIHost         string
		ReadTimeout     time.Duration
		WriteTimeout    time.Duration
		ShutdownTimeout time.Duration
	}
}
```


## [**Functional Options**](https://www.youtube.com/watch?v=jZ1ZsULRyE0&t=32m30s)

```go
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
```


## [**Functional Programming**](https://www.youtube.com/watch?v=nxydu5aPtjQ&t=9m20s)

By returning a function it is possible to declare variables that only are bound to that function's instance, sort of like a private state.

```go
func Accumulator() func(int) int {
	var acc int
	return func(i int) int {
		acc += i
		return acc
	}
}
```

```go
func Factorial() (f func(int) int) {
	cache := map[int]int{}
	return func(i int) int {
		if r, ok := cache[i]; ok {
			return r
		}
		switch {
		case i < 0:
			panic(i)
		case i == 0:
			return 1
		default:
			cache[i] = i * f(i-1)
			return cache[i]
		}
	}
}
```

```go
func myHandler() http.HandlerFunc {
	var myHandlerCounter = 0
	type request struct {
		Name string
	}
	type response struct {
		Greeting string `json:"greeting"`
	}
	var init sync.Once
	return func(w http.ResponseWriter, r *http.Request) {
		init.Do(func() {
			// init your stuff only once
		})
		myHandlerCounter++
		fmt.Fprintf(w, "myHandlerCounter:%d", myHandlerCounter)
	}
}
```


## **Goroutine**

**Never use a goroutine (or any resource) without knowing how to release it**





## [**ldflags**](https://www.youtube.com/watch?v=IV0wrVb31Pg&t=23m10s)

If there is a global variable in a package (example `environment`)

```go
package main

import (
	"fmt"
)

var environment = "DEV"

func main() {
	fmt.Println(environment)
}
```

it is possible to override the value during the build using `ldflags`.

```bash
go build -ldflags="-X 'main.environment=TEST'" main.go
```

Useful to print out the build version at the start of the server

```bash
go build -ldflags="-X 'main.version=${GIT_COMMIT}'" main.go
```




## [**pprof on web server**](https://www.youtube.com/watch?v=IV0wrVb31Pg&t=28m)

pprof to profile the memory usage. Using `chi` [middleware](https://github.com/go-chi/chi/blob/master/middleware/profiler.go):

```go
	r := chi.NewRouter()
	r.Mount("/debug", middleware.Profiler())
	http.ListenAndServe(":3000", r)
```

`http://localhost:3000/debug/vars`

`http://localhost:3000/debug/pprof`


## [**pprof on performance**](https://www.youtube.com/watch?v=nok0aYiGiYA&t=5m25s)

```bash
go get github.com/pkg/profile
```

```go
package main

import (
	"fmt"

	"github.com/pkg/profile"
)

func main() {
	defer profile.Start(profile.MemProfile, profile.ProfilePath(".")).Stop()
	var slice []int
	for i := 0; i < 1000; i++ {
		slice = append(slice, i)
	}
	fmt.Println(slice)
}
```

It creates a `mem.pprof` file which can be analyzed on a local server by running:

```bash
go tool pprof -http=:8000 mem.pprof
```

If `profile.TraceProfile` then a `trace.out` file is generated which can be analyzed on a local server by running:
```bash
go tool trace trace.out
```




## [**gcflags**](https://www.youtube.com/watch?v=oE_vm7KeV_E&t11m25s)

```bash
# If a variable `escapes to heap` then it cause allocations
go build -gcflags='-m -m' main.go
```

```bash
# can tell if we are indexing an out of bound value of a slice
# ./main.go:13:19: Found IsInBounds
go build -gcflags=-d=ssa/check_bce/debug=1 main.go
```




## [**Server timeouts**](https://www.youtube.com/watch?v=YF1qSfkDGAQ&list=PL4WJSMupJdF8WPlGJQy4nlvWVWIPv7c3B&t=13m40s)

ListenAndServe in net/http doesn't timeout

```go
	http.HandleFunc("/", helloWorldHandler)
	server := &http.Server{
		Addr:         ":3000",
		ReadTimeout:  time.Second,
		WriteTimeout: time.Second,
		IdleTimeout:  time.Second,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
```

`ReadHeaderTimeout: time.Second,` can be added too




## [**Custom Logger**](https://www.youtube.com/watch?v=wxkEQxvxs3w&t=18m)

Use a custom logger: with `log.Lshortfile` the file name and line will be displayed

```go
logger := log.New(os.Stdout, "yourPrefix ", log.LstdFlags|log.Lshortfile)
```


## [**Testing Handlers**](https://www.youtube.com/watch?v=wxkEQxvxs3w&t=29m)

Use `*httptest.ResponseRecorder` from `net/http/httptest` and call the handler passing `httptest.NewRecorder` and `httptest.NewRequest`



## [**Transaction and Mutex**](https://www.youtube.com/watch?v=GtsSzbs-xb8&t=34m30s)

```go
func (s *data) advance(v value){
	s.mu.Lock()
	defer s.mu.Unlock()
	s.a = s.nextA(v)
	s.b = s.nextB(v)
}
```

If `nextB` panics then the state of `s` stays corrupted because of `nextA` so one solution is something like:

```go
func (s *data) advance(v value){
	s.startTransaction()
	defer s.endTransaction()
	s.a = s.nextA(v)
	s.b = s.nextB(v)
	s.commitTransaction()
}
```

where:
* `startTransaction` locks the mutex and saves the state of the data
* `commitTransaction` sets a flag to true
* `endTransaction` unlocks the mutex and checks the flag, if false then it reverse the state of the data


## [**Line of Sight**](https://www.youtube.com/watch?v=zdKHq9Xo4OY&t=9m)

* Avoid `else` and nesting
* Return early
* Wrap in functions to make the code more readable

Instead of

```go
	if err != nil {
		if strings.Contains(err.Error(), "special case") {
			return fmt.Errorf("Special error")
		} else {
			return fmt.Errorf("Generic error")
		}
	}
	return nil
```

use

```go
	if err != nil && strings.Contains(err.Error(), "special case") {
		return fmt.Errorf("Special error")
	}
	if err != nil {
		return fmt.Errorf("Generic error")
	}
	return nil
```

## [**Headers against CSRF**](https://www.youtube.com/watch?v=wvdE0M8UEEQ&t=12m30s)

```go
func Allowed(r http.Request) bool {
	site := r.Header.Get("sec-fetch-site")
	mode := r.Header.Get("sec-fetch-mode")
	// Same site or direct url or not supported by browser
	if site == "" || site == "none" || site == "same-site" || site == "same-origin" {
		return true
	}
	// Cross site
	if mode == "navigate" && r.Method == http.MethodGet {
		return true
	}
	return false
}
```


## [**Reflect to assign value**](https://www.youtube.com/watch?v=hz6d7rzqJ6Q&t=6m15s)

It is possible to scan a struct and replace the value of the fields based on struct tags ([11m 20s](https://www.youtube.com/watch?v=hz6d7rzqJ6Q&t=6m15s))



## [**Worker Pool Semaphore pattern**](https://www.youtube.com/watch?v=5zXAHh5tJqQ&t=31m30s)

Only one goroutine is blocked at time, which is the one waiting for the signal on the semaphore channel:

```go
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
```




## **Other resources**

* ### [50 Shades of Go](https://golang50shad.es/)
