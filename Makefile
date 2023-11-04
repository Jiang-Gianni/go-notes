build:
	go build -ldflags="-X 'main.environment=TEST'" main.go

pprof:
	go tool pprof -http=:8000 mem.pprof

trace:
	go tool trace trace.out

gc:
	go build -gcflags='-m -m' main.go

ssa:
	go build -gcflags=-d=ssa/check_bce/debug=1 main.go