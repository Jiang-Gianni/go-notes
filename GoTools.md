Go tools

- [**ldflags**](#ldflags)
- [**pprof on web server**](#pprof-on-web-server)
- [**pprof on performance**](#pprof-on-performance)
- [**gcflags**](#gcflags)
- [**Go Documentation**](#go-documentation)



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
go build -ldflags="-X 'main.gitCommit=${GIT_COMMIT}'" main.go
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


**/debug/pprof/profile** will activate CPU profiling (30 seconds duration with 10 ms rate sampling). After the duration a CPU profiler file is generated can can be downloaded and served locally with:

```bash
go tool pprof -http=:8000 profile
```

Use the following endpoint to change the duration: `http://localhost:3000/debug/pprof/profile?seconds=5`


**debug/pprof/heap/?debug=0** will generate a heap profiling file `heap` that can be served locally with:

```bash
go tool pprof -http=:8000 heap
```

Refer to *100 Go Mistakes #98*

```bash
go tool pprof -noinlines http://localhost:3000/debug/pprof/allocs
(pprof) top 10 -cum
(pprof) list myFunction
(pprof) web myFunction
```


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