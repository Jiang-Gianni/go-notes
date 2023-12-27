package main

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/nats-io/nats.go"
)

type RequestWithKeepAlive struct {
	HeartbeatsInbox string `json:"hb_inbox"`
	Data            []byte `json:"data"`
}

func HeartBeatRequestor() {
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		log.Fatal(err)
	}
	hbInbox := nats.NewInbox()
	req := &RequestWithKeepAlive{
		HeartbeatsInbox: hbInbox,
		Data:            []byte("hello world"),
	}
	payload, err := json.Marshal(req)
	if err != nil {
		log.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	t := time.AfterFunc(10*time.Second, func() {
		cancel()
	})
	nc.Subscribe(hbInbox, func(m *nats.Msg) {
		log.Println("[Heartbeat] extendingdeadline...")
		t.Reset(10 * time.Second)
	})
	response, err := nc.RequestWithContext(ctx, "long.request", payload)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("[Response]", string(response.Data))
}

func HeartBeatSubscriber() {
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		log.Fatal(err)
	}
	nc.Subscribe("long.request", func(m *nats.Msg) {
		log.Println("[Processing]", string(m.Data))
		var req RequestWithKeepAlive
		err := json.Unmarshal(m.Data, &req)
		if err != nil {
			log.Printf("Error: %s", err)
			nc.Publish(m.Reply, []byte("error!"))
			return
		}
		log.Printf("[Heartbeats] %+v", req)
		t := time.NewTicker(5 * time.Second)

		defer t.Stop()
		go func() {
			for range t.C {
				log.Println("[Heartbeat]")
				nc.Publish(req.HeartbeatsInbox,
					[]byte("OK"))
			}
		}()
		// Long processing time...
		time.Sleep(20 * time.Second)
		nc.Publish(m.Reply, []byte("done!"))
	})
	log.Println("[Started]")
	select {}
}
