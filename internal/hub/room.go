package hub

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = &websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 256,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

type room struct {
	name    string
	join    chan *client
	leave   chan *client
	forward chan []byte
	clients map[*client]bool
}

func newRoom(name string) *room {
	return &room{
		name:    name,
		join:    make(chan *client),
		leave:   make(chan *client),
		forward: make(chan []byte),
		clients: make(map[*client]bool),
	}
}

func (r *room) run() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

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
		case <-ticker.C:
			r.broadcastUsers()
		}
	}
}

func (r *room) removeClient(c *client) {
	if _, ok := r.clients[c]; ok {
		delete(r.clients, c)
		close(c.receive)
	}
}

// отправка списка активных пользователей
func (r *room) broadcastUsers() {
	users := []string{}
	for c := range r.clients {
		users = append(users, c.name)
	}
	msg := map[string]interface{}{
		"type": "users",
		"list": users,
	}
	for c := range r.clients {
		c.sendJSON(msg)
	}
}

// serveHTTP с ником пользователя
func (r *room) serveHTTP(w http.ResponseWriter, req *http.Request, nick string) {
	socket, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}

	client := newClient(socket, r, nick)
	r.join <- client
	defer func() { r.leave <- client }()

	go client.write()
	client.read()
}
