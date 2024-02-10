Related to http servers and clients.

- [**ShutDown**](#shutdown)
- [**Graceful Shutdown**](#graceful-shutdown)
- [**Request Context**](#request-context)
- [**Client timeouts**](#client-timeouts)
- [**Server timeouts**](#server-timeouts)
- [**Custom Logger**](#custom-logger)
- [**Testing Handlers**](#testing-handlers)
- [**Transaction and Mutex**](#transaction-and-mutex)
- [**Headers against CSRF**](#headers-against-csrf)
- [**SQL Connection** *(100 Go Mistakes #78)*](#sql-connection-100-go-mistakes-78)
- [**Exhaust http.Response** *Efficient Go Chapter 11*](#exhaust-httpresponse-efficient-go-chapter-11)
- [**Encoding and Decoding**](#encoding-and-decoding)
- [**Validating data**](#validating-data)


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
	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout: time.Second,
			}).DialContext,
			TLSHandshakeTimeout:   time.Second,
			ResponseHeaderTimeout: time.Second,
		},
	}
	res, err := client.Get("INSERT-URL")
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

## **SQL Connection** *(100 Go Mistakes #78)*

From the `database/sql` package, for a given `sql.DB` struct:

| Setter             | Default   | When to set               |
| ------------------ | --------- | ------------------------- |
| SetMaxOpenConns    | unlimited | database limits           |
| SetMaxIdleConns    | 2         | avoid multiple reconnects |
| SetConnMaxIdleTime | unlimited | handle burs periods       |
| SetConnMaxLifetime | unlimited | load-balanced db server   |


## **Exhaust http.Response** *Efficient Go Chapter 11*

The [Documentation](https://pkg.go.dev/net/http#Client.Do) states that *If the Body is not both read to EOF and closed, the Client's underlying RoundTripper (typically Transport) may not be able to re-use a persistent TCP connection to the server for a subsequent "keep-alive" request.*


If not read then it is good practice to discard the response body like shown in https://github.com/efficientgo/core/blob/v1.0.0-rc.2/errcapture/do.go#L39 using `io.Copy(io.Discard, resp.Body)`.

Also close http client connections with:

```go
	c := http.Client{}
	defer c.CloseIdleConnections()
```


## [**Encoding and Decoding**](https://grafana.com/blog/2024/02/09/how-i-write-http-services-in-go-after-13-years/#handle-decodingencoding-in-one-place)

Using Generics:

```go
func encode[T any](w http.ResponseWriter, r *http.Request, status int, v T) error {
	w.WriteHeader(status)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		return fmt.Errorf("encode json: %w", err)
	}
	return nil
}

func decode[T any](r *http.Request) (T, error) {
	var v T
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		return v, fmt.Errorf("decode json: %w", err)
	}
	return v, nil
}
```

## [**Validating data**](https://grafana.com/blog/2024/02/09/how-i-write-http-services-in-go-after-13-years/#validating-data)

```go
// Validator is an object that can be validated.
type Validator interface {
	// Valid checks the object and returns any
	// problems. If len(problems) == 0 then
	// the object is valid.
	Valid(ctx context.Context) (problems map[string]string)
}
```

The implementation of the function `Valid(ctx context.Context) (problems map[string]string)` will check the validations (examples: max length, not null, range etc.).

```go
func decodeValid[T Validator](r *http.Request) (T, map[string]string, error) {
	var v T
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		return v, nil, fmt.Errorf("decode json: %w", err)
	}
	if problems := v.Valid(r.Context()); len(problems) > 0 {
		return v, problems, fmt.Errorf("invalid %T: %d problems", v, len(problems))
	}
	return v, nil, nil
}
```

The returned map will return the field name as the key and the validation error as the value. If `problems` is nil then `len(problems)` returns 0.