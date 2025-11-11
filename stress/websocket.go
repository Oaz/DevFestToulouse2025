package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/gorilla/websocket"
)

type WebSocketClient struct {
	conn     *websocket.Conn
	url      string
	done     chan struct{}
	messages chan []byte
	errors   chan error
}

var websocketUrl = os.Getenv("WEBSOCKET_URL")

func CreateWebSocket(clientId string) *WebSocketClient {
	ws := &WebSocketClient{
		url:      fmt.Sprintf("%s?player_id=%s", websocketUrl, clientId),
		done:     make(chan struct{}),
		messages: make(chan []byte, 100), // Buffered channel
		errors:   make(chan error, 10),   // Buffered channel for errors
	}
	var err error
	ws.conn, _, err = websocket.DefaultDialer.Dial(ws.url, nil)
	if err != nil {
		log.Fatal(fmt.Errorf("failed to connect: %v", err))
	}
	go ws.readMessages()
	return ws
}

func (c *WebSocketClient) readMessages() {
	defer close(c.messages)
	defer close(c.errors)

	for {
		select {
		case <-c.done:
			return
		default:
			messageType, message, err := c.conn.ReadMessage()
			//log.Printf("MESSAGE TYPE %v", messageType)

			if err != nil {
				c.errors <- fmt.Errorf("read error: %v", err)
				return
			}

			if messageType == websocket.TextMessage {
				c.messages <- message
			}
		}
	}
}

func (c *WebSocketClient) Close() error {
	close(c.done)
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *WebSocketClient) IsConnected() bool {
	return c.conn != nil
}

type Message struct {
	Type string
	Data json.RawMessage
}

func ParseMessageStream(data []byte) <-chan Message {
	ch := make(chan Message)
	go func() {
		defer close(ch)
		content := string(data)
		messages := strings.Split(content, "}{")

		for i, msg := range messages {
			if i > 0 {
				msg = "{" + msg
			}
			if i < len(messages)-1 {
				msg = msg + "}"
			}

			msgBytes := []byte(msg)
			var typeMsg struct {
				Type string `json:"type"`
			}
			if err := json.Unmarshal(msgBytes, &typeMsg); err != nil {
				continue
			}

			ch <- Message{
				Type: typeMsg.Type,
				Data: json.RawMessage(msgBytes),
			}
		}
	}()
	return ch
}

func (c *WebSocketClient) handleWebsocket(log func(text string), change func(phaseID int)) bool {
	select {
	case bytes, ok := <-c.messages:
		if !ok {
			log("Message channel closed")
			return false
		}
		for msg := range ParseMessageStream(bytes) {
			switch msg.Type {
			case "phase_change":
				var result struct {
					PhaseID int `json:"phase_id"`
				}
				err := json.Unmarshal(msg.Data, &result)
				if err != nil {
					log(fmt.Sprintf("failed to get phase id: %v", err))
				}
				change(result.PhaseID)
			}
		}

	case err, ok := <-c.errors:
		if !ok {
			log("Error channel closed")
			return false
		}
		log(fmt.Sprintf("Error: %v", err))
	default:
	}
	return true
}
