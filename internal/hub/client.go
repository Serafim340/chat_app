package hub

import (
	"encoding/json"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = 50 * time.Second
	maxMessageSize = 512
)

type client struct {
	socket  *websocket.Conn
	receive chan []byte
	room    *room
	name    string
}

func newClient(socket *websocket.Conn, room *room, name string) *client {
	return &client{
		socket:  socket,
		receive: make(chan []byte, 256),
		room:    room,
		name:    name,
	}
}

func (c *client) read() {
	defer func() {
		c.room.leave <- c
		c.socket.Close()
	}()

	c.socket.SetReadLimit(maxMessageSize)
	c.socket.SetReadDeadline(time.Now().Add(pongWait))
	c.socket.SetPongHandler(func(string) error { c.socket.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		_, msg, err := c.socket.ReadMessage()
		if err != nil {
			return
		}
		outgoing := map[string]string{
			"name":    c.name,
			"message": string(msg),
		}
		jsMsg, err := json.Marshal(outgoing)
		if err != nil {
			continue
		}
		c.room.forward <- jsMsg
	}
}

func (c *client) write() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.socket.Close()
	}()

	for {
		select {
		case msg, ok := <-c.receive:
			c.socket.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.socket.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.socket.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}
		case <-ticker.C:
			c.socket.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.socket.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
