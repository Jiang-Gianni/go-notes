https://github.com/nats-io/nats-server

https://github.com/wallyqs/practical-nats/

https://github.com/nats-io/nats-top

https://github.com/nats-io/prometheus-nats-exporter

There are three ports:
* port for connecting clients (default 4222)
* port for clustering
* port for monitoring (flag -m)

Custom port:

```bash
nats-server -p 4333 -a 127.0.0.1
```

Also open a monitoring endpoint at localhost:8222
```bash
nats-server -m 8222
```

Monitoring has `varz`, `connz` (clients), `routez` (cluster) and `subsz` as endpoints.

In the `connz` and `routez` endpoints, to show the subscriptions of each client:

```bash
curl http://localhost:8222/connz?subs=1
curl http://localhost:8222/routez?subs=1
```

To sort the client by number of subs:

```bash
curl http://localhost:8222/connz?subs=1&sort=subs
```

To get the client with the most messages:

```bash
curl http://127.0.0.1:8222/connz?subs=1&sort=msgs_to&limit=1
```

With credentials:

```bash
# User password
nats-server -m 8222 -user foo -pass secret
# Token
nats-server -m 8222 -auth bar
```

The cluster can be set with:

```bash
nats-server -m 8222 --cluster nats://127.0.0.1:6222
```

With pprof:

```bash
nats-server --profile 9090

# In another terminal
go tool pprof http://127.0.0.1:9090/debug/pprof/goroutine
```


With logging (nats-server is silent by default):

```bash
# -D Debug, -V Tracing, -l log file ./nats/log.txt
nats-server -DV -l ./nats/log.txt
# all of this will hit performance
```


With TLS:

| Flag               | Description                                        |
| ------------------ | -------------------------------------------------- |
| --tls              | Enable TLS, do not verify clients (default: false) |
| --tlscert <file>   | Server certificate file                            |
| --tlskey <file>    | Private key for server certificate                 |
| --tlsverify        | Enable TLS, verify client certificates             |
| --tlscacert <file> | Client certificate CA for verification             |




## Docker

Default ports for all three (client, monitoring, cluster)
```bash
 docker run -p 4222:4222 -p 8222:8222 -p 6222:6222 nats
```


## Cluster

It is a full mesh (10 nodes -> 45 extra tcp connections) so between 3 to 5 is the recommendation.

```bash
SERVERS=nats://127.0.0.1:6222,nats://127.0.0.1:6223,nats://127.0.0.1:6224
nats-server -m 8222 -T -V -p 4222 -cluster nats://127.0.0.1:6222 -routes $SERVERS &
nats-server -m 8223 -T -V -p 4223 -cluster nats://127.0.0.1:6223 -routes $SERVERS &
nats-server -m 8224 -T -V -p 4224 -cluster nats://127.0.0.1:6224 -routes $SERVERS &
```

```bash
pkill -f nats-server
```

In the Go code:

```go
// If the cluster addresses are unknown
nc.DiscoveredServers()
```

```go
// To connect directly to the cluster routes
servers := "nats://127.0.0.1:4222,nats://127.0.0.1:4223,nats://127.0.0.1:4224"
nc, err := nats.Connect(servers, nats.DontRandomize())
```


If there is a load balancer in between the client and the server then it is possible to make so that the cluster addresses are not revealed (useful if they are unreachable)

```bash
nats-server --no_advertise
```
