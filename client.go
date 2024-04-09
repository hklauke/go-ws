package main

import (
	"log"

	"github.com/gorilla/websocket"
)

type ClientList map[*Client]bool

type Client struct {
	// one on one with ws connection
	connection *websocket.Conn
	manager    *Manager

	// egress eused to avoid concurrent writes on ws connection
	egress chan []byte
}

func NewClient(conn *websocket.Conn, manager *Manager) *Client {
	return &Client{
		connection: conn,
		manager:    manager,
		egress:     make(chan []byte),
	}
}

func (c *Client) readMessages() {
	//defer func used to cleanup any unused clients or clients having issues
	defer func() {
		//cleanup connection
		c.manager.removeClient(c)
	}()

	for {
		messageType, payload, err := c.connection.ReadMessage()

		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error reading message: %v", err)
			}
			break
		}
		log.Println(messageType)
		log.Println(string(payload))

		// Hack to test that WriteMessages works as intended
		// Will be replaced soon
		for wsclient := range c.manager.clients {
			wsclient.egress <- payload
		}

	}
}

func (c *Client) writeMessages() {
	defer func() {
		c.manager.removeClient(c)
	}()

	for {
		// https://go.dev/tour/concurrency/5 : select stateement lets a gouroutine wait on multiple communication operations
		select {
		case message, ok := <-c.egress:
			if !ok {
				if err := c.connection.WriteMessage(websocket.CloseMessage, nil); err != nil {
					log.Println("connection closed: ", err)
				}
				return //breaks for loop, triggers cleanup
			}

			if err := c.connection.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Printf("failed to send message: %v", err)
			}
			log.Println("message sent")
		}
	}
}
