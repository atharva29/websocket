package main

import (
	"fmt"
	"sync"

	"github.com/gorilla/websocket"
)

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Registered clients.
	clients map[*Client]bool

	// broadcast messages to this channel
	broadcast chan string

	// Register connection
	register chan *Client

	// Unregister connection
	unregister chan *Client

	mutex sync.Mutex
}

func newHub() *Hub {
	return &Hub{
		broadcast:  make(chan string),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			// h.mutex.Lock()
			h.clients[client] = true
			// h.mutex.Unlock()
		case client := <-h.unregister:
			// h.mutex.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			// h.mutex.Unlock()
		case message := <-h.broadcast:
			fmt.Println("broadcasting message ", len(h.clients))
			for client := range h.clients {
				w, err := client.conn.NextWriter(websocket.TextMessage)
				if err != nil {
					return
				}
				fmt.Println("writing jin message", message)
				_, err = w.Write([]byte(message))
				if err != nil {
					fmt.Println(err)
				}
			}
		}
	}
}
