package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type room struct {

	// clients in the room
	clients map[*client]bool

	// join channel
	join chan *client

	// leave channel
	leave chan *client

	// forward channel
	forward chan []byte
}

func newRoom() *room {
	return &room{
		forward: make(chan []byte),
		join:    make(chan *client),
		leave:   make(chan *client),
		clients: make(map[*client]bool),
	}

}

func (r *room) run() {
	for {
		select {
		//add client to room
		case clinet := <-r.join:
			r.clients[clinet] = true
			//remove client from room
		case clinet := <-r.leave:
			delete(r.clients, clinet)
			close(clinet.recive)
			//forward message to all clients
		case msg := <-r.forward:
			for client := range r.clients {
				client.recive <- msg
			}
		}
	}
}

// update room
const (
	socketBufferSize  = 1024
	messageBufferSize = 256
)

var upgrader = &websocket.Upgrader{ReadBufferSize: socketBufferSize, WriteBufferSize: messageBufferSize}

func (r *room) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	socket, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Fatal("ServerHTTP:", err)
		return
	}
	client := &client{
		socket: socket,
		recive: make(chan []byte, messageBufferSize),
		room:   r,
	}
	r.join <- client
	defer func() { r.leave <- client }()

	go client.write()
	client.read()
}
