# **Go Notes**
Random Go notes

- [**Go Notes**](#go-notes)
  - [**Cancellation**](#cancellation)
    - [defer](#defer)
    - [Request Context](#request-context)
  - [**Client**](#client)
    - [Client timeouts](#client-timeouts)
  - [**Constructor**](#constructor)
    - [Functional Options](#functional-options)
  - [**Goroutine**](#goroutine)
  - [**Server**](#server)
    - [Server timeouts](#server-timeouts)
    - [Custom Logger](#custom-logger)
    - [Testing Handlers](#testing-handlers)
  - [**Other resources**](#other-resources)




## **Cancellation**

### [defer](https://www.youtube.com/watch?v=YF1qSfkDGAQ&list=PL4WJSMupJdF8WPlGJQy4nlvWVWIPv7c3B&t=25m45s)

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



### [Request Context](https://www.youtube.com/watch?v=YF1qSfkDGAQ&list=PL4WJSMupJdF8WPlGJQy4nlvWVWIPv7c3B&t=31m50s)

If the request gets interrupted then the request's Context gets closed

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



## **Client**

###  [Client timeouts](https://www.youtube.com/watch?v=YF1qSfkDGAQ&list=PL4WJSMupJdF8WPlGJQy4nlvWVWIPv7c3B&t=9m)

DefaultClient in http/net doesn't timeout

```go
	client := http.Client{
		Timeout: time.Second,
	}
	res, err := client.Get("INSERT-URL")
```



## **Constructor**

### [Functional Options](https://www.youtube.com/watch?v=jZ1ZsULRyE0&t=32m30s)

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


## **Goroutine**

**Never use a goroutine (or any resource) without knowing how to release it**







## **Server**

### [Server timeouts](https://www.youtube.com/watch?v=YF1qSfkDGAQ&list=PL4WJSMupJdF8WPlGJQy4nlvWVWIPv7c3B&t=13m40s)

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




### [Custom Logger](https://www.youtube.com/watch?v=wxkEQxvxs3w&t=18m)

Use a custom logger: with `log.Lshortfile` the file name and line will be displayed

```go
logger := log.New(os.Stdout, "yourPrefix ", log.LstdFlags|log.Lshortfile)
```


### [Testing Handlers](https://www.youtube.com/watch?v=wxkEQxvxs3w&t=29m)

Use `*httptest.ResponseRecorder` from `net/http/httptest` and call the handler passing `httptest.NewRecorder` and `httptest.NewRequest`




## **Other resources**

* ### [50 Shades of Go](https://golang50shad.es/)
