package main

import "github.com/gorilla/websocket"

type client struct {

	//websocket for the client
	socket *websocket.Conn

	// recieved messages
	recive chan []byte

	//room chat
	room *room
}

//sent message
func (c *client) read() {

	defer c.socket.Close()

	for {
		_, msg, err := c.socket.ReadMessage()
		if err != nil {
			return
		}
		c.room.forward <- msg
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
