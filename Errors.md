Errors stuff

- [**Error type and value check** *(100 Go Mistakes #50 and #51)*](#error-type-and-value-check-100-go-mistakes-50-and-51)
- [**Exporting constant error**](#exporting-constant-error)
- [**Error from deferred function** *(100 Go Mistakes #54)*](#error-from-deferred-function-100-go-mistakes-54)
- [**Defer Error Wrapping**](#defer-error-wrapping)
- [**errgroup**](#errgroup)


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


## **Defer Error Wrapping**

https://github.com/golang/pkgsite/blob/master/internal/derrors/derrors.go#L240

If named return `error` is used then the defer wrap can be set at the beginning of the function.

```go
// Wrap adds context to the error and allows
// unwrapping the result to recover the original error.
//
// Example:
//
//	defer derrors.Wrap(&err, "copy(%s, %s)", src, dst)
//
// See Add for an equivalent function that does not allow
// the result to be unwrapped.
func Wrap(errp *error, format string, args ...any) {
	if *errp != nil {
		*errp = fmt.Errorf("%s: %w", fmt.Sprintf(format, args...), *errp)
	}
}
```


## [**errgroup**](https://www.storj.io/blog/production-concurrency)

```go
// on failure, waits other goroutines
// to stop on their own
var g errgroup.Group
g.Go(func() error {
	return publicServer.Run(ctx)
})
g.Go(func() error {
	return grpcServer.Run(ctx)
})
err := g.Wait()
```

```go
// on failure, cancels other goroutines
g, ctx := errgroup.WithContext(ctx)
g.Go(func() error {
	return publicServer.Run(ctx)
})
g.Go(func() error {
	return grpcServer.Run(ctx)
})
err := g.Wait()
```
