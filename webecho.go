package main

import (
	"flag"
	"log"
	"net/http"

	"gopkg.in/igm/sockjs-go.v2/sockjs"
)

var (
	websocket = flag.Bool("websocket", true, "enable/disable websocket protocol")
)

var p *PubSub

func init() {
	flag.Parse()
}

func main() {
	p = NewPubSub()

	opts := sockjs.DefaultOptions
	opts.Websocket = *websocket
	handler := sockjs.NewHandler("/echo", opts, echoHandler)
	http.Handle("/echo/", handler)
	http.Handle("/", http.FileServer(http.Dir("web/")))
	log.Println("Server started on port: 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

type PubSub struct {
	in   chan<- interface{}
	subs *[]Callback
}

type Callback func(string)

func NewPubSub() *PubSub {
	var subs []Callback
	ch := make(chan interface{}, 5)
	go func() {
		for {
			select {
			case c := <-ch:
				for _, v := range subs {
					v(c.(string))
				}
			}
		}
	}()

	return &PubSub{
		in:   ch,
		subs: &subs,
	}
}

func (p *PubSub) Publish(thing interface{}) {
	p.in <- thing
}

func (p *PubSub) Subscribe(f Callback) {
	*p.subs = append(*p.subs, f)
}

func echoHandler(session sockjs.Session) {
	log.Println("new sockjs session established")
	p.Subscribe(func(words string) {
		session.Send(words)
	})
	for {
		log.Println("Looping...")
		if msg, err := session.Recv(); err == nil {
			p.Publish(msg)
			continue
		}
		break
	}
	log.Println("sockjs session closed")
}
