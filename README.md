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
	- [**Init order execution** *(100 Go Mistakes #3)*](#init-order-execution-100-go-mistakes-3)
	- [**Generics Constraints** *(100 Go Mistakes #9)*](#generics-constraints-100-go-mistakes-9)
	- [**Go Documentation**](#go-documentation)
	- [**Linters** *(100 Go Mistakes #16)*](#linters-100-go-mistakes-16)
	- [**Integer literals** *(100 Go Mistakes #17)*](#integer-literals-100-go-mistakes-17)
	- [**Slice Copy** *(100 Go Mistakes #25)*](#slice-copy-100-go-mistakes-25)
	- [**Map initalization** *(100 Go Mistakes #27)*](#map-initalization-100-go-mistakes-27)
	- [**Break** *(100 Go Mistakes #34)*](#break-100-go-mistakes-34)
	- [**Rune** *(100 Go Mistakes #36 and #37)*](#rune-100-go-mistakes-36-and-37)
	- [**Substring and memory leaks** *(100 Go Mistakes #41)*](#substring-and-memory-leaks-100-go-mistakes-41)
	- [**Error type and value check** *(100 Go Mistakes #50 and #51)*](#error-type-and-value-check-100-go-mistakes-50-and-51)
	- [**Exporting constant error**](#exporting-constant-error)
	- [**Error from deferred function** *(100 Go Mistakes #54)*](#error-from-deferred-function-100-go-mistakes-54)




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


## **Generics Constraints** *(100 Go Mistakes #9)*

```go
type customConstraint interface {
	~int | ~string
}

func getKeys[K customConstraint, V any](m map[K]V) []V {
	keys := make([]V, len(m))
	i := 0
	for k := range m {
		keys[i] = m[k]
		i++
	}
	return keys
}
```

* | is the union operator
* ~int and ~string include all the types whose underlying type is an int or a string



## **Go Documentation**

Use `// Deprecated: yourComment` comment above the function / variable to mark them as deprecated

```bash
go install golang.org/x/pkgsite/cmd/pkgsite@latest
```

```bash
pkgsite -http=:7000
```

This spins up a localhost server on port 7000 serving the standard library documentation and the current project

```bash
wget -r -np -N -E -p -k http://localhost:7000/YOUR_MODULE_NAME
# Example
# wget -r -np -N -E -p -k http://localhost:7000/github.com/Jiang-Gianni/notes-golang
```

This will download all the assets (html, css, js, images) served from the localhost pkgsite server


## **Linters** *(100 Go Mistakes #16)*

```bash
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

```bash
golangci-lint --enable-all -v run
```

[Go vet](https://pkg.go.dev/cmd/go/internal/vet)

```bash
go install golang.org/x/tools/go/analysis/passes/shadow/cmd/shadow@latest
```

```bash
go vet -vettool=$(which shadow)
```

## **Integer literals** *(100 Go Mistakes #17)*

| Literal     | Prefixes        |
| ----------- | --------------- |
| Binary      | 0b, 0B          |
| Octal       | 0,       0o, 0O |
| Hexadecimal | 0x, 0X          |

```go
// Output: 102
fmt.Println(100 + 0b10)

// Output: 108
fmt.Println(100 + 010)

// Output: 108
fmt.Println(100 + 0o10)

// Output: 116
fmt.Println(100 + 0x10)
```


## **Slice Copy** *(100 Go Mistakes #25)*

`copy` copies the minimum number of elements between the length of the source and the length of the destination

An alternative is `append([]int{}, src...)`

```go
	src := []int{0, 1, 2}
	var dst1 []int
	copy(dst1, src)
	// []
	fmt.Println(dst1)
	var dst2 = make([]int, len(src))
	copy(dst2, src)
	// [0, 1, 2]
	fmt.Println(dst2)
	var dst3 = append([]int{}, src...)
	// [0, 1, 2]
	fmt.Println(dst3)
```


## **Map initalization** *(100 Go Mistakes #27)*

Like for slices, an insertion can be an O(n) operation.

If the maximum possible length is known, then use `make` to initialize the map.

```go
m := make(map[int]int, len(valuesToStore))
```


## **Break** *(100 Go Mistakes #34)*

`break` terminates the execution of the innermost `for`, `switch` or `select` statement.

Use a label to specify which statement to stop.

`continue` works similarly.

```go
fast:
	for i := 0; i < 5; i++ {
		fmt.Println(i)
		switch i {
		default:
		case 2:
			break fast
		}
	}
```


## **Rune** *(100 Go Mistakes #36 and #37)*

* charset = set of characters
* encoding = how to translate a charset to binary (in UTF-8: 汉 => []byte{0xE6, 0xB1, 0x89}, between 1 and 4 bytes)
  ```go
  	s := string([]byte{0xE6, 0xB1, 0x89})
	fmt.Printf("%s\n", s) // Prints 汉
  ```
* code point = single value (汉 => U+6C49)
* len(myString) returns the number of bytes
  ```go
  	s := "汉"
	fmt.Println(len(s)) // Prints 3
  ```
* rune = int32 = code point

Example:

```go
	s := "hêllo"
	for i, r := range s {
		fmt.Printf("position %d: %c, %c\n", i, s[i], r)
	}
	fmt.Printf("len=%d\n", len(s))
	// Output
	// position 0: h, h
	// position 1: Ã, ê
	// position 3: l, l
	// position 4: l, l
	// position 5: o, o
	// len=6
```

* ê requires 2 bytes (len = 6)
* range iterates over the start of each rune (i = 2 is skipped)
* s[i] prints the UTF-8 representation of the byte



## **Substring and memory leaks** *(100 Go Mistakes #41)*

If a substring needs to be extracted and saved, make a copy (`strings.Clone`) so that the backing array size of the substring is not equal to the size of the original full string (do not use `subString := originalHugeString[:10]`)



## **Error type and value check** *(100 Go Mistakes #50 and #51)*

```go
var ErrMy MyError = MyError{err: "my error"}

type MyError struct {
	err string
}

func (m MyError) Error() string {
	return m.err
}
```

If an error of type `MyError` has been wrapped using `fmt.Errorf("wrapper message: %w", err)` then the wrapper type becomes `*fmt.wrapError`.

Use `errors.As(err, &MyError{})` to check if one of the wrapping chained errors is of type `MyError`

Use `errors.Is(err, ErrMy)` to check if one of the wrapping chained errors has the same value as `ErrMy`



## [**Exporting constant error**](https://www.youtube.com/watch?v=d7A81rIMwxI&t=1m15s)

`errors.New` can't be used to define constant error, which means that the variable can be reassigned once exported

```go
var CustomError = errors.New("my custom error")
```

By using a **unexportable** error struct, the error variable can't be modified outside the package (`myCustomError` is not exported)

```go
var CustomError = myCustomError("my custom error")

type myCustomError string

func (m myCustomError) Error() string {
	return string(m)
}
```


## **Error from deferred function** *(100 Go Mistakes #54)*

```go
func PrintHttpResponse(url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	bodyText, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	fmt.Printf("%s\n", bodyText)
	return nil
}
```

`resp.Body.Close()` returns an error but it isn't handled.

One possible solution using named output parameters that gives less priority to the error from `resp.Body.Close()` is:

```go
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
```