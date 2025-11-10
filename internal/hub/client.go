package hub

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

type client struct {
	socket  *websocket.Conn
	receive chan []byte
	room    *room
	name    string
}

func newClient(socket *websocket.Conn, room *room, nick string) *client {
	return &client{
		socket:  socket,
		receive: make(chan []byte, 256),
		room:    room,
		name:    nick,
	}
}

const (
	writeWait  = 10 * time.Second
	pongWait   = 60 * time.Second
	pingPeriod = 50 * time.Second
)

func (c *client) read() {
	defer func() {
		c.room.leave <- c
		c.socket.Close()
	}()

	c.socket.SetReadLimit(512)
	c.socket.SetReadDeadline(time.Now().Add(pongWait))
	c.socket.SetPongHandler(func(string) error {
		c.socket.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, raw, err := c.socket.ReadMessage()
		if err != nil {
			return
		}

		// формируем сообщение клиента
		msg := map[string]interface{}{
			"type":    "message",
			"name":    c.name,
			"message": string(raw),
		}

		jsMsg, err := json.Marshal(msg)
		if err != nil {
			log.Println("JSON marshal error:", err)
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
			c.socket.WriteMessage(websocket.TextMessage, msg)
		case <-ticker.C:
			c.socket.SetWriteDeadline(time.Now().Add(writeWait))
			c.socket.WriteMessage(websocket.PingMessage, nil)
		}
	}
}

func (c *client) sendJSON(v interface{}) {
	data, err := json.Marshal(v)
	if err != nil {
		log.Println("JSON marshal error:", err)
		return
	}
	c.receive <- data
}
