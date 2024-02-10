Testing

- [**Testing: parallel shuffle flags** *(100 Go Mistakes #84)*](#testing-parallel-shuffle-flags-100-go-mistakes-84)
- [**Check Function type for Testing**](#check-function-type-for-testing)
- [**Cleanup function for Testing**](#cleanup-function-for-testing)
- [**t.Cleanup**](#tcleanup)
- [**Test Coverage**](#test-coverage)
- [**Test Tags**](#test-tags)


## **Testing: parallel shuffle flags** *(100 Go Mistakes #84)*

Test execution order: sequential and then parallel

`t.Parallel()` marks the test as parallel

```go
func TestFoo(t *testing.T) {
	t.Parallel()
	// ...
}
```

To increase the number of maximum executing parallel tests for a given time (default `GOMAXPROCS`) use:

```bash
go test -parallel 16 .
```

The `shuffle` flag make sure that the tests are run in a random order.

`-v` will print out the seed value number.

```bash
go test -shuffle=on -v .
```

Run the following to repeat the same order

```bash
go test -shuffle=YOUR_SEED_VALUE -v .
```



## [**Check Function type for Testing**](https://www.youtube.com/watch?v=TGg6cc0QCzw&t=30m)

Write a function type with the inputs that match the output signature of the function to test + `*testing.T`


```go
func DoStuff(i int) (int, error)
type checkDoStuff func(int, error, *testing.T)
func hasError(want error) checkDoStuff {
	return func(_ int, got error, t *testing.T) {
		t.Helper()
		if want != got {
			t.Errorf("Expected error %v, got: %v", want, got)
		}
	}
}
checks := func(cs ...checkDoStuff) []checkDoStuff { return cs }
```

This makes table test more readable.

```go
	testCases := []struct {
		name   string
		input  int
		checks []checkDoStuff
	}{
		{
			name: "invalid input",
			input: -1,
			checks: checks(hasError(ErrInvalidInput)),
		},
	}
```


## [**Cleanup function for Testing**](https://www.youtube.com/watch?v=8hQG7QlcLBk&t=18m15s)

Make the function that prepares the test also return a cleanup function that can be defer called.

```go
func testTempFile(t *testing.T) (string, func()) {
	t.Helper()
	tf, err := os.CreateTemp("", "test")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	tf.Close()
	return tf.Name(), func() { os.Remove(tf.Name()) }
}
func TestThing(t *testing.T) {
	tf, tfclose := testTempFile(t)
	defer tfclose()
	//...
}
```

## [**t.Cleanup**](https://stackoverflow.com/questions/61609085/what-is-useful-for-t-cleanup)

Use `t.Cleanup` to register a cleanup function that will be executed when the test ends.

The following test doesn't work because sub test 1 and 2 are run in parallel, so the main test returns and calls `cleanup()`

```go
func TestSomething(t *testing.T){
   setup()
   defer cleanup()
   t.Run("parallel subtest 1", func(t *testing.T){
      t.Parallel()
      (...)
   })
   t.Run("parallel subtest 2", func(t *testing.T){
      t.Parallel()
      (...)
   })
}
```

By using `t.Cleanup` the registered function is called only when the test ends.

```go
func TestSomething(t *testing.T){
   setup()
   t.Cleanup(cleanup)
   t.Run("parallel subtest 1", func(t *testing.T){
      t.Parallel()
      (...)
   })
   t.Run("parallel subtest 2", func(t *testing.T){
      t.Parallel()
      (...)
   })
}
```

For example, to test a server:

```go
func Test(t *testing.T) {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	t.Cleanup(cancel)
	go run(ctx)
	// test code goes here
}
```


## [**Test Coverage**](https://www.youtube.com/watch?v=ndmB0bj7eyw&t=4m50s)

```bash
go test -coverprofile=cover.out ./...
go tool cover -func=cover.out
go tool cover -html=cover.out -o=cover.html
```


## [**Test Tags**](https://medium.com/@tharun208/build-tags-in-go-f21ccf44a1b8)

At the beginning of the file add a build tag to that the test contained are executed only when calling `go test -tags yourTag`

```go
// +build integration

package api_test
```

Add an exclamation point `!` to negate the execution in case that specific tag is included in the call

```go
// +build !integration

package api_test
```


Use a comma `,` between multiple tag names to execute the test(s) only when all the tags are included

```go
// +build integration,api

package api_test
```


Run the test with a `count` flag to ensure the test is being run even after it is cached

```go
go test -tags yourTag -count=1
```

Tags can also be used to select target operating systems (example `// +build linux`). Another approach is to add the OS / architecture as suffix in the file name (example `api_linux.go`, `api_windows.go`, `api_darwin.go`)
