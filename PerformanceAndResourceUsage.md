Performance and resourse usage, some are not strictly Go related

- [**Default case with sleep in a select statement**](#default-case-with-sleep-in-a-select-statement)
- [**Substring and memory leaks** *(100 Go Mistakes #41)*](#substring-and-memory-leaks-100-go-mistakes-41)
- [**Nil Channel** *(100 Go Mistakes #66)*](#nil-channel-100-go-mistakes-66)
- [**time.After memory leaks** *(100 Go Mistakes #76)*](#timeafter-memory-leaks-100-go-mistakes-76)
- [**Reduce allocations** *(100 Go Mistakes #96 and go-perfbook)*](#reduce-allocations-100-go-mistakes-96-and-go-perfbook)
- [**No Allocation** *(Go Optimizations 101)*](#no-allocation-go-optimizations-101)
- [**Index tables vs maps** *(Go Optimizations 101)*](#index-tables-vs-maps-go-optimizations-101)
- [**Select Read with channels** *(Go Optimizations 101)*](#select-read-with-channels-go-optimizations-101)
- [**Calling interface methods** *(Go Optimizations 101)*](#calling-interface-methods-go-optimizations-101)
- [**Pre-Allocate** *Efficient Go Chapter 11*](#pre-allocate-efficient-go-chapter-11)
- [**The Power of Two Random Choices**](#the-power-of-two-random-choices)
- [**Batching to reduce overhead**](#batching-to-reduce-overhead)
- [**Avoid multiple same function calls and prefer multiplication over division**](#avoid-multiple-same-function-calls-and-prefer-multiplication-over-division)
- [**Slicing vs Offset**](#slicing-vs-offset)
- [Reading materials](#reading-materials)
- [Tools](#tools)
- [Examples](#examples)


## [**Default case with sleep in a select statement**](https://www.youtube.com/watch?v=19bxBMPOlyA&t=13m10s)

Consider using a `time.Sleep` in the default case of a select statement (with multiple channel reads/writes cases) to avoid unnecessary computation checks

## **Substring and memory leaks** *(100 Go Mistakes #41)*

If a substring needs to be extracted and saved, make a copy (`strings.Clone`) so that the backing array size of the output substring is not pointing to the original full string (do not use `subString := originalHugeString[:10]`)


## **Nil Channel** *(100 Go Mistakes #66)*

Once a channel is closed and don't need to be read anymore, assign `nil` to it so that it can't be read in a `select` statement: if the channel is just closed then `zero-value, false` are read.

```go
	for ch1 != nil || ch2 != nil {
		select {
		case v, open := <-ch1:
			if !open {
				ch1 = nil
				break
			}
			ch <- v
		case v, open := <-ch2:
			if !open {
				ch2 = nil
				break
			}
			ch <- v
		}
	}
```


## **time.After memory leaks** *(100 Go Mistakes #76)*


```go
func consumer(ch <-chan Event) {
	for {
		select {
		case event := <-ch:
			handle(event)
		case <-time.After(time.Hour):
			log.Println("warning: no messages received")
		}
	}
}
```

For each loop a new time.Time channel is returned by `time.After`: this channel is closed only when the timeout expires which can cause memory leaks.

A solution is to use instantiate a single `*time.Timer` with `time.NewTimer` and to use the `Reset` function:

```go
func consumer(ch <-chan Event) {
	timer := time.NewTimer(time.Hour)
	defer timer.Stop()
	for {
		timer.Reset(time.Hour)
		select {
		case event := <-ch:
			handle(event)
		case <-timer.C:
			log.Println("warning: no messages received")
		}
	}
}
```

## **Reduce allocations** *(100 Go Mistakes #96 and go-perfbook)*


* Prefer share down approach to prevent auto escape to the heap

```go
// Share down
type Reader interface {
	Read(p []byte) (n int, err error)
}

// Share up
type Reader interface {
	Read(n int) (p []byte, err error)
}
```

* Allow passing in buffers so caller can reuse and slice can be modified in place
* use error variables instead of errors.New() / fmt.Errorf() at call site (performance or style? interface requires pointer, so it escapes to heap anyway)
* Use strconv instead of fmt if possible
* Use `strings.EqualFold(str1, str2)` instead of `strings.ToLower(str1) == strings.ToLower(str2)` or `strings.ToUpper(str1) == strings.ToUpper(str2)` to efficiently compare strings if possible.
* Use `string(yourByteSlice)` to access a `map[string]any`
* Use `sync.Pool` to reuse already allocated memory


```go
	var pool = sync.Pool{
		New: func() any {
			return make([]byte, 1024)
		},
	}

	write := func(w io.Writer) {
		buffer := pool.Get().([]byte)
		buffer = buffer[:0]
		defer pool.Put(buffer)
		//
	}
```



## **No Allocation** *(Go Optimizations 101)*

* If the input slice is allowed to be mutated, then avoid allocations.

```go
func check(v int) bool {
	return v%2 == 0
}

func FilterOneAllocation(data []int) []int {
	var r = make([]int, 0, len(data))
	for _, v := range data {
		if check(v) {
			r = append(r, v)
		}
	}
	return r
}

func FilterNoAllocations(data []int) []int {
	var k = 0
	for i, v := range data {
		if check(v) {
			data[i] = data[k]
			data[k] = v
			k++
		}
	}
	return data[:k]
}
```

* `strings.EqualFold(a, b)` is more performant than `strings.ToLower(a) == strings.ToLower(b)`


## **Index tables vs maps** *(Go Optimizations 101)*

`MapSwitch` is around 10 times slower than `IfElse` and `IndexTable`

```go
func IfElse(x bool) func() {
	if x {
		return f
	} else {
		return g
	}
}

var m = map[bool]func(){true: f, false: g}

func MapSwitch(x bool) func() {
	return m[x]
}

func b2i(b bool) int {
	if b {
		return 1
	}
	return 0
}

var a = [2]func(){g, f}

func IndexTable(x bool) func() {
	return a[b2i(x)]
}
```


## **Select Read with channels** *(Go Optimizations 101)*

If possible make it so that there are fewer channel read cases in a `select` statement by merging the channels and differentiating by a field condition. See [select_test.go](./select_test.go).

```bash
# go version go1.21.2 linux/amd64
go test -bench=. select_test.go

Benchmark_Select_OneCase-12             57681339                20.64 ns/op
Benchmark_Select_TwoCases-12            23804878                49.43 ns/op
Benchmark_Select_OneNil-12              35653648                33.27 ns/op
Benchmark_TwoChannels-12                 7241138               165.4 ns/op
Benchmark_OneChannel_Interface-12       10793074               110.0 ns/op
Benchmark_OneChannel_Struct-12          10981497               112.5 ns/op

```


## **Calling interface methods** *(Go Optimizations 101)*

Calling an interface method requires some extra cost. See [interface_test.go](./interface_test.go)

```bash
# go version go1.21.2 linux/amd64
go test -bench=. interface_test.go

Benchmark_Add_Inline-12         1000000000               0.3437 ns/op
Benchmark_Add_NotInlined-12     775364438                1.545 ns/op
Benchmark_Add_Interface-12      771482826                1.557 ns/op
```


## **Pre-Allocate** *Efficient Go Chapter 11*

If the size of a slice is known it is better to allocate the needed memory at the beginning. `append` will re-allocate the memory in case the size is going to exceed the slice capacity.

```go
func ReadAll1(r io.Reader, size int) ([]byte, error) {
	buf := bytes.Buffer{}
	buf.Grow(size)
	n, err := io.Copy(&buf, r)
	return buf.Bytes()[:n], err
}
```

```go
func ReadAll2(r io.Reader, size int) ([]byte, error) {
	buf := make([]byte, size)
	n, err := io.ReadFull(r, buf)
	if err == io.EOF {
		err = nil
	}
	return buf[:n], err
}
```

Both `ReadAll1` and `ReadAll2` are faster than the [`io.ReadAll`](https://pkg.go.dev/io#ReadAll) from the standard library but the byte slice size must be known.

An example is extracting the [`Content-Length`](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Length) header from an http response to preallocate the slice needed to read the body




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


## **[Avoid multiple same function calls and prefer multiplication over division](https://github.com/golang/go/commit/ed6c6c9c11496ed8e458f6e0731103126ce60223)**






## **[Slicing vs Offset](https://github.com/golang/go/commit/b85433975aedc2be2971093b6bbb0a7dc264c8fd)**

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