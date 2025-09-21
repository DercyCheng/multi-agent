package websocket

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

// Message represents a WebSocket message
type Message struct {
	Type      string                 `json:"type"`
	Payload   interface{}            `json:"payload"`
	Timestamp time.Time              `json:"timestamp"`
	ClientID  string                 `json:"client_id,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// Client represents a connected WebSocket client
type Client struct {
	ID       string
	Hub      *Hub
	Conn     interface{} // WebSocket connection interface
	Send     chan Message
	TenantID string
	UserID   string
	Groups   []string
}

// Hub maintains the set of active clients and broadcasts messages
type Hub struct {
	clients    map[*Client]bool
	broadcast  chan Message
	register   chan *Client
	unregister chan *Client
	logger     interface{}
	mu         sync.RWMutex
}

// NewHub creates a new WebSocket hub
func NewHub(logger interface{}) *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan Message),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		logger:     logger,
	}
}

// Run starts the hub
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()

			// Send welcome message
			welcome := Message{
				Type:      "connection",
				Payload:   map[string]interface{}{"status": "connected", "client_id": client.ID},
				Timestamp: time.Now(),
			}

			select {
			case client.Send <- welcome:
			default:
				close(client.Send)
				h.mu.Lock()
				delete(h.clients, client)
				h.mu.Unlock()
			}

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.Send)
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				// Check if client should receive this message
				if h.shouldReceiveMessage(client, message) {
					select {
					case client.Send <- message:
					default:
						close(client.Send)
						delete(h.clients, client)
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

// shouldReceiveMessage determines if a client should receive a message
func (h *Hub) shouldReceiveMessage(client *Client, message Message) bool {
	// Check tenant isolation
	if tenantID, exists := message.Metadata["tenant_id"]; exists {
		if tenantID != client.TenantID {
			return false
		}
	}

	// Check user targeting
	if userID, exists := message.Metadata["user_id"]; exists {
		if userID != client.UserID {
			return false
		}
	}

	// Check group targeting
	if groups, exists := message.Metadata["groups"]; exists {
		if groupList, ok := groups.([]string); ok {
			hasGroup := false
			for _, group := range groupList {
				for _, clientGroup := range client.Groups {
					if group == clientGroup {
						hasGroup = true
						break
					}
				}
				if hasGroup {
					break
				}
			}
			if !hasGroup {
				return false
			}
		}
	}

	return true
}

// BroadcastFlagChange broadcasts feature flag changes
func (h *Hub) BroadcastFlagChange(flagID, action string, flag interface{}, tenantID string) {
	message := Message{
		Type: "flag_change",
		Payload: map[string]interface{}{
			"action": action,
			"flag":   flag,
		},
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"flag_id":   flagID,
			"tenant_id": tenantID,
		},
	}

	h.broadcast <- message
}

// BroadcastConfigChange broadcasts configuration changes
func (h *Hub) BroadcastConfigChange(configKey, action string, config interface{}, tenantID string) {
	message := Message{
		Type: "config_change",
		Payload: map[string]interface{}{
			"action": action,
			"config": config,
		},
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"config_key": configKey,
			"tenant_id":  tenantID,
		},
	}

	h.broadcast <- message
}

// BroadcastServiceChange broadcasts service discovery changes
func (h *Hub) BroadcastServiceChange(serviceID, action string, service interface{}) {
	message := Message{
		Type: "service_change",
		Payload: map[string]interface{}{
			"action":  action,
			"service": service,
		},
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"service_id": serviceID,
		},
	}

	h.broadcast <- message
}

// BroadcastCronJobUpdate broadcasts cron job status updates
func (h *Hub) BroadcastCronJobUpdate(jobID, status string, execution interface{}, tenantID string) {
	message := Message{
		Type: "cron_job_update",
		Payload: map[string]interface{}{
			"status":    status,
			"execution": execution,
		},
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"job_id":    jobID,
			"tenant_id": tenantID,
		},
	}

	h.broadcast <- message
}

// RegisterClient registers a new WebSocket client
func (h *Hub) RegisterClient(client *Client) {
	h.register <- client
}

// UnregisterClient unregisters a WebSocket client
func (h *Hub) UnregisterClient(client *Client) {
	h.unregister <- client
}

// GetClientCount returns the number of connected clients
func (h *Hub) GetClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// HandleWebSocket handles WebSocket connections (mock implementation)
func (h *Hub) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// In a real implementation, this would:
	// 1. Upgrade HTTP connection to WebSocket
	// 2. Create a new Client
	// 3. Register the client with the hub
	// 4. Start goroutines for reading/writing messages

	// Mock response for now
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]interface{}{
		"message": "WebSocket endpoint - would upgrade connection in real implementation",
		"clients": h.GetClientCount(),
	}

	json.NewEncoder(w).Encode(response)
}

// SendToClient sends a message to a specific client
func (h *Hub) SendToClient(clientID string, message Message) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for client := range h.clients {
		if client.ID == clientID {
			select {
			case client.Send <- message:
			default:
				// Client's send channel is full, remove client
				close(client.Send)
				delete(h.clients, client)
			}
			break
		}
	}
}

// SendToTenant sends a message to all clients in a tenant
func (h *Hub) SendToTenant(tenantID string, message Message) {
	message.Metadata = map[string]interface{}{
		"tenant_id": tenantID,
	}
	h.broadcast <- message
}

// SendToUser sends a message to a specific user
func (h *Hub) SendToUser(userID, tenantID string, message Message) {
	message.Metadata = map[string]interface{}{
		"user_id":   userID,
		"tenant_id": tenantID,
	}
	h.broadcast <- message
}

// SendToGroups sends a message to specific user groups
func (h *Hub) SendToGroups(groups []string, tenantID string, message Message) {
	message.Metadata = map[string]interface{}{
		"groups":    groups,
		"tenant_id": tenantID,
	}
	h.broadcast <- message
}
