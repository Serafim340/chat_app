package hub

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"

	"github.com/gorilla/websocket"
)

const messageBufferSize = 256

var upgrader = &websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: messageBufferSize,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

type room struct {
	join    chan *client
	leave   chan *client
	forward chan []byte
	clients map[*client]bool
}

func newRoom() *room {
	return &room{
		join:    make(chan *client),
		leave:   make(chan *client),
		forward: make(chan []byte),
		clients: make(map[*client]bool),
	}
}

func (r *room) run() {
	for {
		select {
		case c := <-r.join:
			r.clients[c] = true
		case c := <-r.leave:
			r.removeClient(c)
		case msg := <-r.forward:
			for c := range r.clients {
				select {
				case c.receive <- msg:
				default:
					r.removeClient(c)
				}
			}
		}
	}
}

func (r *room) removeClient(c *client) {
	if _, ok := r.clients[c]; ok {
		delete(r.clients, c)
		close(c.receive)
	}
}

// serveHTTP — приватный метод
func (r *room) serveHTTP(w http.ResponseWriter, req *http.Request) {
	socket, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}

	client := newClient(socket, r, fmt.Sprintf("User_%d", rand.Intn(1000)))
	r.join <- client
	defer func() { r.leave <- client }()

	go client.write()
	client.read()
}
