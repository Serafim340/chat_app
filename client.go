package main

import (
	"encoding/json"
	"fmt"

	"github.com/gorilla/websocket"
)

type client struct {

	//websocket for the client
	socket *websocket.Conn

	// recieved messages
	recive chan []byte

	//room chat
	room *room

	name string
}

// sent message
func (c *client) read() {

	defer c.socket.Close()

	for {
		_, msg, err := c.socket.ReadMessage()
		if err != nil {
			return
		}
		outgoing := map[string]string{
			"name":    c.name,
			"message": string(msg),
		}

		jsMessage, err := json.Marshal(outgoing)
		if err != nil {
			fmt.Println("Encoding failed", err)
			continue
		}
		c.room.forward <- jsMessage

	}
}

func (c *client) write() {

	defer c.socket.Close()

	for msg := range c.recive {
		err := c.socket.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			return
		}
	}
}
