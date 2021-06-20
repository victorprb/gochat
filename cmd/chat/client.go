package main

import (
	"time"

	"github.com/gorilla/websocket"
)

type UserData struct {
	Name string `json:"name"`
}

// client represents a single chatting user.
type client struct {
	socket   *websocket.Conn
	send     chan *message
	room     *room
	userData UserData
}

func (c *client) read() {
	defer c.socket.Close()
	for {
		var msg *message
		err := c.socket.ReadJSON(&msg)
		if err != nil {
			return
		}
		msg.When = time.Now()
		msg.Name = c.userData.Name
		c.room.forward <- msg
	}
}

func (c *client) write() {
	defer c.socket.Close()
	for msg := range c.send {
		err := c.socket.WriteJSON(msg)
		if err != nil {
			break
		}
	}
}
