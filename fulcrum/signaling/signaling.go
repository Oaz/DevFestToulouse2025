package signaling

import (
	"encoding/json"
	"fulcrum/ports"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // For development, restrict in production
	},
}

type Hub struct {
	clients    map[*Client]bool
	register   chan *Client
	unregister chan *Client
	mutex      sync.Mutex
}

type Client struct {
	hub  *Hub
	conn *websocket.Conn
	send chan []byte
	id   string // Player/game ID for targeted messages
}

// Initialize and run the WebSocket hub
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// Run starts the hub's main processing loop
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mutex.Lock()
			h.clients[client] = true
			h.mutex.Unlock()
			//log.Printf("Client %s registered, total clients: %d", client.id, len(h.clients))
		case client := <-h.unregister:
			h.mutex.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				log.Println("Client unregistered, remaining clients:", len(h.clients))
			}
			h.mutex.Unlock()
		}
	}
}

// HandleWebSocket upgrades HTTP connection to WebSocket
func (h *Hub) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	client := &Client{
		hub:  h,
		conn: conn,
		send: make(chan []byte, 256),
		id:   r.URL.Query().Get("player_id"), // Extract player ID from query
	}

	client.hub.register <- client

	// For downstream-only, we only need the writePump
	// Optional minimal readPump to detect disconnections
	go client.readPump()
	go client.writePump()
}

// Simplified readPump that only detects disconnections
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(512)
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	// Just read messages to detect disconnection, but don't do anything with them
	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		// Ignore any messages from client
	}
}

// writePump pumps messages from the hub to the websocket connection
func (c *Client) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				// The hub closed the channel
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages to the current websocket message
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// Broadcast sends a message to all clients
func (h *Hub) Broadcast(message []byte) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if len(h.clients) == 0 {
		log.Println("Warning: Broadcasting message but no clients connected")
		return
	}

	for client := range h.clients {
		select {
		case client.send <- message:
			//log.Printf("Broadcast to client %s", client.id)
			// Message sent successfully
		default:
			// Client's send buffer is full, unregister client
			//log.Printf("closing client: %s", client.id)
			close(client.send)
			delete(h.clients, client)
		}
	}
}

func (h *Hub) BroadcastJson(msg any) error {
	jsonMsg, err := json.Marshal(msg)
	if err != nil {
		log.Printf("error marshaling message: %v", err)
		return err
	}
	h.Broadcast(jsonMsg)
	return nil
}

var _ ports.Notifier = (*Hub)(nil)
