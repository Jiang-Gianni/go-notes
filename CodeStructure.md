Code structure and patterns

- [**Main abstraction**](#main-abstraction)
- [**Configuration**](#configuration)
- [**Functional Options**](#functional-options)
- [**Functional Programming**](#functional-programming)
- [**Line of Sight**](#line-of-sight)
- [**Init order execution** *(100 Go Mistakes #3)*](#init-order-execution-100-go-mistakes-3)


## [**Main abstraction**](https://www.youtube.com/watch?v=IV0wrVb31Pg&t=10m40s)

https://grafana.com/blog/2024/02/09/how-i-write-http-services-in-go-after-13-years/#func-main-only-calls-run

Create a `run` function that handles all the initial settings and returns an error in case of failure. This avoids multiple repeated error handling in case they need to be handled the same way.

```go
func main() {
	ctx := context.Background()
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()
	if err := run(ctx, os.Stdout, os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, w io.Writer, args []string) error {
	// ...
}
```

Passing `io.Writer`, `io.Reader` as input to the `run` enables more parallel testing.

This can also be applied to flags and environment variables:

```go
args := []string{
	"myapp",
	"--out", outFile,
	"--fmt", "markdown",
}
go run(ctx, args, etc.)
```

```go
getenv := func(key string) string {
	switch key {
	case "MYAPP_FORMAT":
		return "markdown"
	case "MYAPP_TIMEOUT":
		return "5s"
	default:
		return ""
}
go run(ctx, args, getenv)
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


## **Init order execution** *(100 Go Mistakes #3)*

```go
package main

import "fmt"

// Executed first
var a = func() int {
	fmt.Println("var")
	return 0
}()

// Executed second
func init() {
	fmt.Println("init")
}

// Executed last
func main() {
	fmt.Println("main")
}
```

* The lowest packages init functions / variables of the import dependency tree are executed / evaluated first
* Multiple init functions can be defined on the same package: the execution order is based on the alphabetical order of the filenames and declaration order (in case of multiple inits in the same file)

A bit tricky, init?
