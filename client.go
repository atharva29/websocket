package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	hub *Hub

	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte
}

// Payload to read the message
type Payload struct {
	Message string
}

// read reads messages from the clients and send events to hub on client's connection closing
// for every client writePump go routine is made
func (c *Client) read() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
	}
}

// write pumps messages from the hub to the websocket connection.
// for every client write go routine is made
func (c *Client) write() {
	defer func() {
		c.conn.Close()
	}()

	for message := range c.send {

		c.conn.SetWriteDeadline(time.Now().Add(writeWait))
		w, err := c.conn.NextWriter(websocket.TextMessage)
		if err != nil {
			return
		}
		w.Write(message)

		// Add queued chat messages to the current websocket message.
		n := len(c.send)
		for i := 0; i < n; i++ {
			w.Write(newline)
			w.Write(<-c.send)
		}

		// close the writer
		if err := w.Close(); err != nil {
			log.Println(err)
			return
		}
	}
}

// serveWs handles websocket requests from the peer.
func serveWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	client := &Client{hub: hub, conn: conn, send: make(chan []byte, 256)}
	// register client to the hubF
	client.hub.register <- client

	// Start a new reader go routine and writer go routine for the client
	go client.write()
	go client.read()
}

// servePublish reads data and publishes to broadcasting channel
func servePublish(hub *Hub, w http.ResponseWriter, r *http.Request) {
	var p Payload
	err := json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	fmt.Println("Message recieved", p)
	// broadcast the message only if clients are present
	if p.Message != "" && len(hub.clients) > 0 {
		hub.broadcast <- []byte(p.Message)
	}
}
