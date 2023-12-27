package main

import (
	"encoding/json"
	"log"
	"runtime"
	"time"

	"github.com/nats-io/nats.go"
)

func main() {
	nc, err := nats.Connect("nats://127.0.0.1:4222",
		nats.Name("practical-nats-client"),
		nats.UserInfo("foo", "secret"),

		// -1 => client never stops trying to reconnect
		// unless it receives an error (example full buffer)
		nats.MaxReconnects(-1),

		// No reconnect
		nats.NoReconnect(),
		nats.ReconnectBufSize(1024),

		// On event handlers
		nats.DisconnectHandler(func(nc *nats.Conn) {
			log.Printf("Disconnected!\n")
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			log.Printf("Reconnected to %v!\n",
				nc.ConnectedUrl())
		}),
		nats.ClosedHandler(func(nc *nats.Conn) {
			log.Printf("Connection closed. Reason: %q\n", nc.LastError())
		}),
		nats.DiscoveredServersHandler(func(nc *nats.Conn) {
			log.Printf("Server discovered\n")
		}),
	)
	if err != nil {
		log.Fatalf("Error: %s", err)
	}

	g, err := nc.Subscribe("greeting", func(m *nats.Msg) {
		log.Printf("[Received] %s", string(m.Data))
	})
	if err != nil {
		log.Fatal(err)
	}
	g.AutoUnsubscribe(1000)

	count := 0
	var w *nats.Subscription
	w, err = nc.Subscribe(">", func(m *nats.Msg) {
		log.Printf("[Wildcard] %s", string(m.Data))
		count++
		if count == 5 {
			// Unsubscribe within the callback
			w.Unsubscribe()
		}
		time.Sleep(1 * time.Second)
	})
	if err != nil {
		log.Fatal(err)
	}

	for i := range [10]int{} {
		payload := struct {
			RequestID int
			Data      []byte
		}{
			RequestID: i,
			Data:      []byte("encoded data"),
		}
		msg, err := json.Marshal(payload)
		if err != nil {
			log.Fatalf("Error: %s", err)
		}

		nc.Publish("greeting", msg)
		if i == 3 {
			// Flush ensure that the server has received the data
			// before sending more
			// nc.Publish is non blocking
			nc.Flush()
		}
	}

	nc.Subscribe("help", func(m *nats.Msg) {
		log.Printf("[Received]: %s", string(m.Data))
		nc.Publish(m.Reply, []byte("I can help!!!"))
	})

	// Request will block until timeout or response
	response, err := nc.Request("help", []byte("help!!"), 1*time.Second)
	if err != nil {
		log.Fatalf("Error: %s", err)
	}

	log.Println("[Response]: " + string(response.Data))

	// Close will flush any pending data in the buffer
	// nc.Close()
	runtime.Goexit()

}
