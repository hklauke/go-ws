package main

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var (
	websocketUpgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

// Manager is used to hold references to all Clients Registered, and Broadcasting etc
type Manager struct {
	clients ClientList
	sync.RWMutex
}

// NewManager is used to initalize all the values inside the manager
func NewManager() *Manager {
	return &Manager{
		clients: make(ClientList),
	}
}

// serveWS is a HTTP Handler that has the manager that handles connections
func (m *Manager) serveWs(w http.ResponseWriter, r *http.Request) {
	log.Println("new connection")

	// Upgraede the http request to WS
	conn, err := websocketUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	// Create client
	client := NewClient(conn, m)
	m.addClient(client)

	// start the read/write process
	go client.readMessages()
	//go client.writeMessages()
}

// addClient will add clients to our clientlist
func (m *Manager) addClient(client *Client) {
	// Lock so we can manipulate
	m.Lock()
	defer m.Unlock()

	// Add client
	m.clients[client] = true
}

func (m *Manager) removeClient(client *Client) {
	m.Lock()
	defer m.Unlock()

	// Check if client exists and then delete it

	if _, ok := m.clients[client]; ok {
		client.connection.Close()
		delete(m.clients, client)
	}
}
