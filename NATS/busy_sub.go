package main

import (
	"log"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
)

func BusySub() {
	var busy bool
	var l sync.Mutex
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		log.Fatal(err)
	}
	myInbox := nats.NewInbox()
	nc.QueueSubscribe("very.long.request", "workers",
		func(m *nats.Msg) {
			l.Lock()
			shouldSkip := busy
			l.Unlock()
			// Only reply when not busy
			if shouldSkip {
				// Reply with empty inbox to signal that
				// was not available to process request.
				nc.PublishRequest(m.Reply, "", []byte(""))
				return
			}
			log.Println("[Processing] Announcing owninbox...")
			nc.PublishRequest(m.Reply, myInbox, []byte(""))
		})
	nc.Subscribe(myInbox, func(m *nats.Msg) {
		log.Println("[Processing] Message:", string(m.Data))
		l.Lock()
		busy = true
		l.Unlock()
		time.Sleep(20 * time.Second)
		l.Lock()
		busy = false
		l.Unlock()
		nc.Publish(m.Reply, []byte("done!"))
	})
	log.Println("[Started]")
	select {}
}

func HelpSeeker() {
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		log.Fatal(err)
	}
	var i int
	var inbox string
	for ; i < 5; i++ {
		log.Println("[Inbox Request]")
		reply, err := nc.Request("very.long.request",
			[]byte(""), 5*time.Second)
		if err != nil {
			log.Println("Retrying...")
			continue
		}
		if reply.Reply == "" {
			log.Println("Node replied with empty inbox, retry again later...")
			time.Sleep(1 * time.Second)
			continue
		}
		inbox = reply.Reply
		break
	}
	if i == 5 {
		log.Fatalf("No nodes available to reply!")
	}
	log.Println("[Detected node]", inbox)
	payload := []byte("hi...")
	response, err := nc.Request(inbox, payload, 30*time.Second)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("[Response]", string(response.Data))
}
