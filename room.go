package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync"

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

var rooms = make(map[string]*room)

var mu sync.Mutex

func getRoom(name string) *room {
	mu.Lock()
	defer mu.Unlock()
	if r, ok := rooms[name]; ok {
		return r
	}
	r := newRoom()
	rooms[name] = r
	go r.run()
	return r
}

// update room
const (
	socketBufferSize  = 1024
	messageBufferSize = 256
)

var upgrader = &websocket.Upgrader{ReadBufferSize: socketBufferSize, WriteBufferSize: messageBufferSize}

func (r *room) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	roomName := req.URL.Query().Get("room")
	if roomName == "" {
		http.Error(w, "No room specified", http.StatusBadRequest)
		return
	}

	realRoom := getRoom(roomName)

	socket, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Fatal("ServerHTTP:", err)
		return
	}
	client := &client{
		socket: socket,
		recive: make(chan []byte, messageBufferSize),
		room:   r,
		name:   fmt.Sprintf("User_%d", rand.Intn(1000)),
	}
	realRoom.join <- client
	defer func() { r.leave <- client }()

	go client.write()
	client.read()
}
