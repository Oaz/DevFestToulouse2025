package signaling

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

// setupTestHub creates a hub, starts it in a goroutine, and sets up a test server.
// Returns the hub, the test server, and a cleanup function that should be deferred.
func setupTestHub(t *testing.T) (*Hub, *httptest.Server, func()) {
	// Create a new hub
	hub := NewHub()

	// Start the hub in a goroutine
	go hub.Run()

	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hub.HandleWebSocket(w, r)
	}))

	// Return cleanup function
	cleanup := func() {
		server.Close()
	}

	return hub, server, cleanup
}

// createWsURL converts HTTP URL to WebSocket URL with query parameters
func createWsURL(serverURL string, path string, queryParams map[string]string) string {
	u, _ := url.Parse(serverURL)
	u.Scheme = "ws"
	u.Path = path

	if queryParams != nil {
		q := u.Query()
		for key, value := range queryParams {
			q.Set(key, value)
		}
		u.RawQuery = q.Encode()
	}

	return u.String()
}

// connectTestClients connects multiple test clients to the WebSocket server
func connectTestClients(t *testing.T, serverURL string, numClients int) []*websocket.Conn {
	clients := make([]*websocket.Conn, numClients)

	for i := 0; i < numClients; i++ {
		wsURL := createWsURL(serverURL, "/ws", map[string]string{
			"player_id": "player_" + string(rune('A'+i)),
		})

		ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		assert.NoError(t, err)
		clients[i] = ws
	}

	// Wait for connections to establish
	time.Sleep(100 * time.Millisecond)

	return clients
}

// TestHubBroadcastMultipleClients tests that the Broadcast method sends to all clients
func TestHubBroadcastMultipleClients(t *testing.T) {
	hub, server, cleanup := setupTestHub(t)
	defer cleanup()

	// Connect multiple clients
	const numClients = 3
	clients := connectTestClients(t, server.URL, numClients)

	// Ensure all clients are closed when the test finishes
	for i := range clients {
		defer clients[i].Close()
	}

	err := hub.BroadcastJson(map[string]interface{}{
		"type":  "foo",
		"value": "bar",
	})
	assert.NoError(t, err)

	// All clients should receive the message
	for i := 0; i < numClients; i++ {
		_, receivedMsg, err := clients[i].ReadMessage()
		assert.NoError(t, err)
		assert.Equal(t, "{\"type\":\"foo\",\"value\":\"bar\"}", string(receivedMsg))
	}
}

// TestConnectionParameters tests that client IDs are properly extracted from URL
func TestConnectionParameters(t *testing.T) {
	hub, server, cleanup := setupTestHub(t)
	defer cleanup()

	// Create channel to signal when client is registered with expected ID
	clientRegistered := make(chan struct{})

	// Create a spy function to check clients after they're registered
	go func() {
		// Check every 10ms for new clients with our expected ID
		ticker := time.NewTicker(10 * time.Millisecond)
		defer ticker.Stop()

		expectedPlayerID := "player_xyz_123"

		for {
			select {
			case <-ticker.C:
				hub.mutex.Lock()
				// Look for a client with our expected ID
				for client := range hub.clients {
					if client.id == expectedPlayerID {
						hub.mutex.Unlock()
						close(clientRegistered)
						return
					}
				}
				hub.mutex.Unlock()
			case <-time.After(2 * time.Second):
				// Give up after 2 seconds
				return
			}
		}
	}()

	// Create URL with specific player ID
	expectedPlayerID := "player_xyz_123"
	wsURL := createWsURL(server.URL, "/ws", map[string]string{
		"player_id": expectedPlayerID,
	})

	// Connect a test client
	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	assert.NoError(t, err)
	defer ws.Close()

	// Wait for client registration with a timeout
	select {
	case <-clientRegistered:
		// Client was registered with the expected ID, test passed
	case <-time.After(2 * time.Second):
		t.Fatal("Timed out waiting for client to be registered with expected ID")
	}
}

// TestClientDisconnect tests that clients are properly removed from the hub when they disconnect
func TestClientDisconnect(t *testing.T) {
	hub, server, cleanup := setupTestHub(t)
	defer cleanup()

	// Create WebSocket URL
	wsURL := createWsURL(server.URL, "/ws", map[string]string{
		"player_id": "test_player",
	})

	// Connect a test client
	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	assert.NoError(t, err)

	// Wait for the connection to establish
	time.Sleep(100 * time.Millisecond)

	// Count clients before disconnect
	hub.mutex.Lock()
	clientCountBefore := len(hub.clients)
	hub.mutex.Unlock()
	assert.Equal(t, 1, clientCountBefore)

	// Disconnect the client
	ws.Close()

	// Wait for the disconnect to propagate
	time.Sleep(100 * time.Millisecond)

	// Count clients after disconnect
	hub.mutex.Lock()
	clientCountAfter := len(hub.clients)
	hub.mutex.Unlock()
	assert.Equal(t, 0, clientCountAfter)
}
