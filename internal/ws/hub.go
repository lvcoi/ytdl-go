package ws

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for now
	},
}

// WSMessage matches the JSON data contract.
type WSMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// ProgressPayload matches the JSON contract A.
type ProgressPayload struct {
	ID       string  `json:"id"`
	Filename string  `json:"filename,omitempty"`
	Percent  float64 `json:"percent"`
	Status   string  `json:"status"`
	ETA      string  `json:"eta,omitempty"`
}

// ErrorPayload matches the JSON contract B.
type ErrorPayload struct {
	ID      string `json:"id"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}

// Client represents a connected WebSocket user.
type Client struct {
	hub  *Hub
	conn *websocket.Conn
	send chan WSMessage // Unbuffered for strict backpressure
}

// Hub maintains the set of active clients and broadcasts messages to them.
type Hub struct {
	clients    map[*Client]bool
	broadcast  chan WSMessage
	register   chan *Client
	unregister chan *Client
	mu         sync.Mutex
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan WSMessage, 1024), // Buffered to prevent blocking the sender
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()
		case message := <-h.broadcast:
			h.mu.Lock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					log.Printf("ws: client send buffer full or blocked, disconnecting client")
					client.conn.Close()
					delete(h.clients, client)
					close(client.send)
				}
			}
			h.mu.Unlock()
		}
	}
}

func (h *Hub) HandleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("ws upgrade error:", err)
		return
	}
	client := &Client{hub: h, conn: conn, send: make(chan WSMessage)}
	h.register <- client

	go client.writePump()
	go client.readPump()
}

func (c *Client) writePump() {
	defer func() {
		c.conn.Close()
	}()
	for message := range c.send {
		if err := c.conn.WriteJSON(message); err != nil {
			log.Printf("ws write error: %v", err)
			break
		}
	}
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

func (h *Hub) Broadcast(msg WSMessage) {
	select {
	case h.broadcast <- msg:
	default:
		// Drop if buffer is full to prevent blocking the caller (Pool)
		log.Printf("ws: broadcast buffer full, dropping message")
	}
}
