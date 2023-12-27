Books:
* gRPC Go for Professionals
* gRPC Microservices in Go

- [**UnimplementedOrderServer**](#unimplementedorderserver)
- [**reflection.Register**](#reflectionregister)
- [**TLS**](#tls)
- [**Timestamp type**](#timestamp-type)
- [**Tags threshold**](#tags-threshold)
- [**Overhead**](#overhead)
- [**Mask**](#mask)
- [**Context and Metadata**](#context-and-metadata)
- [**Interceptors**](#interceptors)
- [**go-grpc-mideleware**](#go-grpc-mideleware)
  - [**Auth**](#auth)
  - [**Logging**](#logging)
  - [**Prometheus**](#prometheus)
  - [**Rate Limiting**](#rate-limiting)
- [**Retrying calls**](#retrying-calls)
- [**Message Compression**](#message-compression)
- [**Client Side Load Balancing**](#client-side-load-balancing)
- [**Request validation**](#request-validation)
- [**Testing**](#testing)
- [**ghz**](#ghz)
- [**grpcurl**](#grpcurl)
- [**gRPC Logs**](#grpc-logs)
- [**Channelz**](#channelz)


## **UnimplementedOrderServer**
UnimplementedOrderServer forward compatibility support

```go
type Adapter struct {
 api ports.APIPort
 port int
 order.UnimplementedOrderServer
}
```

## **reflection.Register**

`reflection.Register` allows grpcurl

```go
func (a Adapter) Run() {
	var err error
	listen, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		log.Fatalf("failed to listen on port %d, error: %v", a.port, err)
	}
	grpcServer := grpc.NewServer()
	order.RegisterOrderServer(grpcServer, a)
	if config.GetEnv() == "development" {
		reflection.Register(grpcServer)
	}
	if err := grpcServer.Serve(listen); err != nil {
		log.Fatalf("failed to serve grpc on port ")
	}
}
```

[gRPC Retry Middleware page 86/103](https://github.com/grpc-ecosystem/go-grpc-middleware)

[gRPC Circuit Breaker](https://github.com/sony/gobreaker)



## **TLS**

Example of certificates:

```bash
curl https://raw.githubusercontent.com/grpc/grpc-go/master/examples/data/x509/server_cert.pem --output server_cert.pem
curl https://raw.githubusercontent.com/grpc/grpc-go/master/examples/data/x509/server_key.pem --output server_key.pem
curl https://raw.githubusercontent.com/grpc/grpc-go/master/examples/data/x509/ca_cert.pem --output ca_cert.pem
```

```go
// Server
certFile := "server.crt"
keyFile := "server.pem"
creds, err := credentials.NewServerTLSFromFile(certFile, keyFile)
if err != nil {
 log.Fatalf("failed loading certificates: %v\n", err)
}
opts = append(opts, grpc.Creds(creds))
```

```go
// Client
certFile := "ca.crt"
creds, err := credentials.NewClientTLSFromFile(certFile, "")
if err != nil {
 log.Fatalf("error while loading CA trust certificate: %v\n", err)
}
opts = append(opts, grpc.WithTransportCredentials(creds))
```

[Dial Options](https://github.com/grpc/grpc-go/blob/master/dialoptions.go)

[Server](https://github.com/grpc/grpc-go/blob/master/server.go)



## **Timestamp type**

```proto
import "google/protobuf/timestamp.proto";

message AddTaskRequest {
  string description = 1;
  google.protobuf.Timestamp due_date = 2;
}
```


## **Tags threshold**

These are the tags threshold after which an extra byte is needed to serialize the tag

```proto
message Tags {
 int32 tag = 1;
 int32 tag2 = 16;
 int32 tag3 = 2048;
 int32 tag4 = 262_144;
 int32 tag5 = 33_554_432;
 int32 tag6 = 536_870_911;
}
```


## **Overhead**

This message

```proto
message UpdateTasksRequest {
 Task task = 1;
}
```

has a user-defined overhead (2 bytes: tag + type and length) over this:

```proto
message UpdateTasksRequest {
 uint64 id = 1;
 string description = 2;
 bool done = 3;
 google.protobuf.Timestamp due_date = 4;
}
```



## **Mask**

```proto
import "google/protobuf/field_mask.proto";
//...
message ListTasksRequest {
 google.protobuf.FieldMask mask = 1;
}
```

https://github.com/PacktPublishing/gRPC-Go-for-Professionals/blob/main/chapter6/server/impl.go#L17

Avoid fetching extra unused data

[gRPC Status Codes](https://pkg.go.dev/google.golang.org/grpc/codes#Code)



## **Context and Metadata**

The server can grab the context and metadata of the request's stream:

```go
func (s *server) ListTasks(req *pb.ListTasksRequest, stream pb.TodoService_ListTasksServer) error {
	ctx := stream.Context()
	md, _ := metadata.FromIncomingContext(ctx)
	if t, ok := md["auth_token"]; ok {
		switch {
		case len(t) != 1:
			return status.Errorf(
				codes.InvalidArgument,
				"auth_token should contain only 1 value",
			)
		case t[0] != "authd":
			return status.Errorf(
				codes.Unauthenticated,
				"incorrect auth_token",
			)
		}
	} else {
		return status.Errorf(
			codes.Unauthenticated,
			"failed to get auth_token",
		)
	}
  //
}
```

From the client:

```go
func updateTasks(c pb.TodoServiceClient, reqs ...*pb.
 UpdateTasksRequest) {
 ctx := context.Background()
 ctx = metadata.AppendToOutgoingContext(ctx, "auth_token", "authd")
 stream, err := c.UpdateTasks(ctx)
 //...
}
```


## **Interceptors**

Act like middlewares. Present for both client and server.

[UnaryServerInterceptor](https://pkg.go.dev/google.golang.org/grpc#UnaryServerInterceptor)

```go
type UnaryServerInterceptor func(ctx context.Context, req any, info *UnaryServerInfo, handler UnaryHandler) (resp any, err error)
```

```go
func unaryAuthInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	if err := validateAuthToken(ctx); err != nil {
		return nil, err
	}

	return handler(ctx, req)
}
```

[StreamServerInterceptor](https://pkg.go.dev/google.golang.org/grpc#StreamServerInterceptor)

```go
type StreamServerInterceptor func(srv any, ss ServerStream, info *StreamServerInfo, handler StreamHandler) error
```

```go
func streamAuthInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	if err := validateAuthToken(ss.Context()); err != nil {
		return err
	}

	return handler(srv, ss)
}
```


```go
// Server
	opts := []grpc.ServerOption{
		grpc.Creds(creds),
		grpc.ChainUnaryInterceptor(unaryAuthInterceptor, unaryLogInterceptor),
		grpc.ChainStreamInterceptor(streamAuthInterceptor, streamLogInterceptor),
	}
	s := grpc.NewServer(opts...)

	pb.RegisterTodoServiceServer(s, &server{
		d: New(),
	})
```


[UnaryClientInterceptor](https://pkg.go.dev/google.golang.org/grpc#UnaryClientInterceptor)

```go
type UnaryClientInterceptor func(ctx context.Context, method string, req, reply any, cc *ClientConn, invoker UnaryInvoker, opts ...CallOption) error
```

```go
func unaryAuthInterceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	ctx = metadata.AppendToOutgoingContext(ctx, authTokenKey, authTokenValue)
	err := invoker(ctx, method, req, reply, cc, opts...)

	return err
}
```

[StreamClientInterceptor](https://pkg.go.dev/google.golang.org/grpc#StreamClientInterceptor)

```go
type StreamClientInterceptor func(ctx context.Context, desc *StreamDesc, cc *ClientConn, method string, streamer Streamer, opts ...CallOption) (ClientStream, error)
```

```go
func streamAuthInterceptor(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	ctx = metadata.AppendToOutgoingContext(ctx, authTokenKey, authTokenValue)
	s, err := streamer(ctx, desc, cc, method, opts...)

	if err != nil {
		return nil, err
	}

	return s, nil
}
```

Use `grpc.WithChainUnaryInterceptor` in case of multiple interceptors.

```go
// Client
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(creds),
		grpc.WithUnaryInterceptor(unaryAuthInterceptor),
    // grpc.WithChainUnaryInterceptor(unaryLoggerInterceptor, unaryAuthInterceptor),
		grpc.WithStreamInterceptor(streamAuthInterceptor),
    grpc.WithDefaultCallOptions(grpc.UseCompressor(gzip.Name)),
		grpc.WithDefaultServiceConfig(`{"loadBalancingConfig": [{"round_robin":{}}]}`),
	}
	conn, err := grpc.Dial(addr, opts...)

	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}

	defer func(conn *grpc.ClientConn) {
		if err := conn.Close(); err != nil {
			log.Fatalf("unexpected error: %v", err)
		}
	}(conn)

	c := pb.NewTodoServiceClient(conn)
```

**Order dependent**: the first declared gets executed first



## **[go-grpc-mideleware](https://github.com/grpc-ecosystem/go-grpc-middleware)**

### **Auth**

```bash
go get github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/auth
```

This package let us write the grpc.ServerOption:

```go
opts := []grpc.ServerOption{
		grpc.Creds(creds),
		grpc.ChainUnaryInterceptor(
			//
			auth.UnaryServerInterceptor(validateAuthToken),
			//
		),
		grpc.ChainStreamInterceptor(
			//
			auth.StreamServerInterceptor(validateAuthToken),
			//
		),
	}
```

where `validateAuthToken` is of type:

```go
type AuthFunc func(ctx context.Context) (context.Context, error)
```


### **Logging**

```bash
go get github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging
```

```go
type loggerFunc func(ctx context.Context, lvl logging.Level, msg string, fields ...any)
```

```go
// More efficient if the service and method are always on the same positions
const grpcService = 5 // "grpc.service"
const grpcMethod = 7  //"grpc.method"

func logCalls(l *log.Logger) logging.Logger {
	return logging.LoggerFunc(func(_ context.Context, lvl logging.Level, msg string, fields ...any) {

		switch lvl {
		case logging.LevelDebug:
			msg = fmt.Sprintf("DEBUG :%v", msg)
		case logging.LevelInfo:
			msg = fmt.Sprintf("INFO :%v", msg)
		case logging.LevelWarn:
			msg = fmt.Sprintf("WARN :%v", msg)
		case logging.LevelError:
			msg = fmt.Sprintf("ERROR :%v", msg)
		default:
			panic(fmt.Sprintf("unknown level %v", lvl))
		}

		l.Println(msg, fields[grpcService], fields[grpcMethod])
	})
}

```

### **Prometheus**

```bash
go get github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus
```

```go
func newMetricsServer(httpAddr string, reg *prometheus.Registry) *http.Server {
	httpSrv := &http.Server{Addr: httpAddr}
	m := http.NewServeMux()
	m.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	httpSrv.Handler = m
	return httpSrv
}
```

See https://github.com/PacktPublishing/gRPC-Go-for-Professionals/blob/main/chapter8/server/main.go#L99 as example.

```go
srvMetrics := grpcprom.NewServerMetrics(
 grpcprom.WithServerHandlingTimeHistogram(
 grpcprom.WithHistogramBuckets([]float64{0.001, 0.01,
 0.1, 0.3, 0.6, 1, 3, 6, 9, 20, 30, 60, 90, 120}),
 ),
)
```

It shows how many requests were served under 0.001, 0.01, 0.1 ... seconds



### **Rate Limiting**

```bash
 go get golang.org/x/time/rate
 go get github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/ratelimit
```

The interceptors:

```go
ratelimit.UnaryServerInterceptor(limiter),
ratelimit.StreamServerInterceptor(limiter),
```

need a parameter that satisfies the interface:

```go
type Limiter interface {
	Limit(ctx context.Context) error
}
```

Example: https://github.com/PacktPublishing/gRPC-Go-for-Professionals/blob/main/chapter8/server/limit.go


## **Retrying calls**

```bash
go get github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/retry
```

```go
// Client
	retryOpts := []retry.CallOption{
		retry.WithMax(3),
		retry.WithBackoff(retry.BackoffExponential(100 * time.Millisecond)),
		retry.WithCodes(codes.Unavailable),
	}

  opts := []grpc.DialOption{
    //
		grpc.WithChainUnaryInterceptor(
			retry.UnaryClientInterceptor(retryOpts...),
		),
    //
	}
```

It is not available for client streaming


## **Message Compression**

Effective only on repetitive data.

**Before deciding to use a compression interceptor: compare the message sizes before and after the compression, it may not be worth it**

https://github.com/PacktPublishing/gRPC-Go-for-Professionals/blob/main/helpers/gzip.go

On the server side just a simple unnamed import is needed:

```go
// Server
import (
//
	_ "google.golang.org/grpc/encoding/gzip"
)
```

On the client side a Dial/Call Option has to be added either to the client grpc connection or to the client grpc service call:

```go
// Dial Option
grpc.WithDefaultCallOptions(grpc.UseCompressor(gzip.Name)),
//
// Per single Call
res, err := c.AddTask(context.Background(), req, grpc.UseCompressor(gzip.Name))
```



## **Client Side Load Balancing**

gRPC has a client load balancing implementation (all the servers addresses need to be known)

https://github.com/grpc/grpc/blob/master/doc/service_config.md

```go
grpc.WithDefaultServiceConfig(`{"loadBalancingConfig": [{"round_robin":{}}]}`),
```

Only works with the DNS scheme

```go
grpc.Dial("dns:///$HOSTNAME:50051", opts...)
```

```
kind create cluster --config k8s/kind.yaml
kubectl apply -f k8s/server.yaml
kubectl apply -f k8s/client.yaml
kubectl logs todo-server-{identifier}
```



## **Request validation**

https://github.com/bufbuild/protoc-gen-validate/blob/main/README.md

https://github.com/bufbuild/protovalidate

```go
go install github.com/envoyproxy/protoc-gen-validate@latest
```

Copy pasta the `validate.proto` from https://github.com/bufbuild/protoc-gen-validate/blob/main/validate/validate.proto

In your service.proto file:

```proto
import "path/to/file/validate.proto";

//

message AddTaskRequest {
  string description = 1 [
    (validate.rules).string.min_len = 1
  ];

  google.protobuf.Timestamp due_date = 2 [
    (validate.rules).timestamp.gt_now = true
  ];
}
```

A `{...}.pb.validate.go` file will be generated.

On the server side:

```go
func (s *server) AddTask(_ context.Context, in *pb.AddTaskRequest) (*pb.AddTaskResponse, error) {
  if err := in.Validate(); err != nil {
		return nil, err
	}
  //
}
```


## **Testing**

https://pkg.go.dev/google.golang.org/grpc/test/bufconn

Allows to create a buffered connection without needing any port

https://github.com/PacktPublishing/gRPC-Go-for-Professionals/blob/main/chapter9/server/server_test.go#L23

Use the same listener (that satisfies `net.Conn` interface) for both server and client

```go
// Server
lis = bufconn.Listen(bufSize)
s := grpc.NewServer()
//
err := s.Serve(lis)
//
```

```go
// Client
func newClient(t *testing.T) (*grpc.ClientConn, pb.TodoServiceClient) {
	ctx := context.TODO()
	creds := grpc.WithTransportCredentials(insecure.NewCredentials())
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), creds)

	if err != nil {
		t.Fatalf("failed to dial bufnet: %v", err)
	}

	return conn, pb.NewTodoServiceClient(conn)
}
```


To test bidirectional streaming one possible approach is to use a channel containing errors and the values to verify.


## **ghz**

* https://github.com/bojand/ghz

```bash
ghz --proto ./proto/todo/v2/todo.proto --import-paths=proto --call todo.v2.TodoService.AddTask --data '{"description":"task"}' --cacert ./certs/ca_cert.pem --cname "check.test.example.com" --metadata '{"auth_token":"authd"}' --total 500 0.0.0.0:5000
```

`--cacert` and `--cname` useful only for self signed certificates

the output is something like:

```bash

Summary:
  Count:        500
  Total:        47.31 ms
  Slowest:      13.37 ms
  Fastest:      0.26 ms
  Average:      3.21 ms
  Requests/sec: 10569.43

Response time histogram:
  0.263  [1]   |
  1.573  [150] |∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎
  2.884  [198] |∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎
  4.195  [54]  |∎∎∎∎∎∎∎∎∎∎∎
  5.505  [2]   |
  6.816  [29]  |∎∎∎∎∎∎
  8.127  [9]   |∎∎
  9.437  [23]  |∎∎∎∎∎
  10.748 [17]  |∎∎∎
  12.059 [14]  |∎∎∎
  13.370 [3]   |∎

Latency distribution:
  10 % in 0.87 ms
  25 % in 1.42 ms
  50 % in 2.28 ms
  75 % in 3.01 ms
  90 % in 9.12 ms
  95 % in 10.04 ms
  99 % in 11.57 ms

Status code distribution:
  [OK]   500 responses
```

Possible values for status codes and errors

```bash
Status code distribution:
 [Unavailable] 3 responses
 [PermissionDenied] 3 responses
 [OK] 186 responses
 [Internal] 8 responses

Error distribution:
[8] rpc error: code = Internal desc = Internal error.
[3] rpc error: code = PermissionDenied desc = Permission
 denied.
[3] rpc error: code = Unavailable desc = Service unavailable.
```


## **grpcurl**

* https://github.com/fullstorydev/grpcurl

```go
import (
 //...
 "google.golang.org/grpc/reflection"
)
//
	s := grpc.NewServer(opts...)

	pb.RegisterTodoServiceServer(s, &server{
		d: New(),
	})
	reflection.Register(s)
//
```

```bash
grpcurl -cacert ./certs/ca_cert.pem -authority "check.test.example.com" -reflect-header 'auth_token: authd' 0.0.0.0:5000 list
```

```bash
grpcurl -cacert ./certs/ca_cert.pem -authority "check.test.example.com" -reflect-header 'auth_token: authd' 0.0.0.0:5000 describe todo.v2.TodoService
```

```bash
grpcurl -cacert ./certs/ca_cert.pem -authority "check.test.example.com" -reflect-header 'auth_token: authd' 0.0.0.0:5000 describe todo.v2.TodoService.AddTask
```

```bash
grpcurl -cacert ./certs/ca_cert.pem -authority "check.test.example.com" -reflect-header 'auth_token: authd' 0.0.0.0:5000 describe todo.v2.AddTaskRequest
```

```bash
grpcurl -cacert ./certs/ca_cert.pem -authority "check.test.example.com" -rpc-header 'auth_token: authd' -reflect-header 'auth_token: authd' -d '{"description": "Hello World", "due_date":"2024-01-01T00:00:00Z"}' -use-reflection 0.0.0.0:5000 todo.v2.TodoService.AddTask
```

`–use-reflection` verifies that the data is valid

```bash
grpcurl -cacert ./certs/ca_cert.pem -authority "check.test.example.com" -rpc-header 'auth_token: authd' -reflect-header 'auth_token: authd' -use-reflection 0.0.0.0:5000 todo.v2.TodoService.ListTasks
```

For Client streaming

```bash
grpcurl -cacert ./certs/ca_cert.pem -authority "check.test.example.com" -rpc-header 'auth_token: authd' -reflect-header 'auth_token: authd' -use-reflection -d @ 0.0.0.0:5000 todo.v2.TodoService.UpdateTasks <<EOF
{ "id": 1, "description": "a better task!" }
{ "id": 2, "description": "another better task!" }
EOF
```

## **gRPC Logs**

Set the env variable `GRPC_GO_LOG_SEVERITY_LEVEL` to debug, info, or error

Set the env variable `GRPC_GO_LOG_VERBOSITY_LEVEL` between 2 (less verbose) and 99 (more verbose)

Set the env variable `GRPC_GO_LOG_FORMATTER` to json

```bash
GRPC_GO_LOG_SEVERITY_LEVEL=info GRPC_GO_LOG_VERBOSITY_LEVEL=99 GRPC_GO_LOG_FORMATTER=json go run ./server 0.0.0.0:5000 0.0.0.0:4000
```

## **Channelz**

* https://grpc.io/blog/a-short-introduction-to-channelz/
* https://pkg.go.dev/google.golang.org/grpc/channelz/service
* https://pkg.go.dev/google.golang.org/grpc/admin
* https://github.com/grpc/grpc-experiments.git
