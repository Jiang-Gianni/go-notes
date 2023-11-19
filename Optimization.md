# Optimization

Not necessarily Go related

- [Optimization](#optimization)
  - [**The Power of Two Random Choices**](#the-power-of-two-random-choices)
  - [**Batching to reduce overhead**](#batching-to-reduce-overhead)
  - [Avoid multiple same function calls and prefer multiplication over division](#avoid-multiple-same-function-calls-and-prefer-multiplication-over-division)
  - [Slicing vs Offset](#slicing-vs-offset)
  - [Reading materials](#reading-materials)
  - [Tools](#tools)
  - [Examples](#examples)


## [**The Power of Two Random Choices**](https://github.com/dgryski/go-perfbook/blob/master/performance.md?plain=1#L672)

For selection problems make two random picks and choose the best from those



## [**Batching to reduce overhead**](https://lemire.me/blog/2018/04/17/iterating-in-batches-over-data-structures-can-be-much-faster/)

Instead of looping over a huge list, divide it into smaller batches / buffers and double loop over them: each batch / buffer (within a certain range) is moved to a cache / faster memory

```go
buffer := make([]uint, 256)
j := uint(0)
j, buffer = bitmap.NextSetMany(j, buffer)
for ; len(buffer) > 0; j, buffer = bitmap.NextSetMany(j, buffer) {
     for k := range buffer {
        // do something with buffer[k]
     }
     j += 1
}
```


## [Avoid multiple same function calls and prefer multiplication over division](https://github.com/golang/go/commit/ed6c6c9c11496ed8e458f6e0731103126ce60223)






## [Slicing vs Offset](https://github.com/golang/go/commit/b85433975aedc2be2971093b6bbb0a7dc264c8fd)

Scan a slice by keeping track of an offset



## Reading materials
* https://dave.cheney.net/high-performance-go-workshop/dotgo-paris.html
* https://users.ece.cmu.edu/~franzf/papers/gttse07.pdf
* https://smallmemory.charlesweir.com/book.html


## Tools

* http://golang.org/pkg/testing/
* https://pkg.go.dev/golang.org/x/perf/benchstat
* https://github.com/tsenart/vegeta
* https://github.com/aclements/perflock
* https://tip.golang.org/doc/diagnostics
* https://github.com/google/pprof


## Examples

* https://benhoyt.com/writings/count-words/#go
* https://github.com/dgryski/go-perfbook/blob/master/performance.md?plain=1#L1055