package websocket

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"

	"notification-service/internal/models"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow connections from any origin
		return true
	},
}

type Client struct {
	hub      *Hub
	conn     *websocket.Conn
	send     chan []byte
	userID   string
	clientID string
}

type Hub struct {
	clients     map[*Client]bool
	register    chan *Client
	unregister  chan *Client
	broadcast   chan []byte
	userClients map[string][]*Client
	mutex       sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		clients:     make(map[*Client]bool),
		register:    make(chan *Client),
		unregister:  make(chan *Client),
		broadcast:   make(chan []byte),
		userClients: make(map[string][]*Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mutex.Lock()
			h.clients[client] = true
			if client.userID != "" {
				h.userClients[client.userID] = append(h.userClients[client.userID], client)
			}
			h.mutex.Unlock()

			zap.L().Info("Client registered",
				zap.String("clientID", client.clientID),
				zap.String("userID", client.userID),
				zap.Int("totalClients", len(h.clients)),
			)

		case client := <-h.unregister:
			h.mutex.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)

				// Remove from user clients
				if client.userID != "" {
					userClients := h.userClients[client.userID]
					for i, c := range userClients {
						if c == client {
							h.userClients[client.userID] = append(userClients[:i], userClients[i+1:]...)
							break
						}
					}
					if len(h.userClients[client.userID]) == 0 {
						delete(h.userClients, client.userID)
					}
				}
			}
			h.mutex.Unlock()

			zap.L().Info("Client unregistered",
				zap.String("clientID", client.clientID),
				zap.String("userID", client.userID),
				zap.Int("totalClients", len(h.clients)),
			)

		case message := <-h.broadcast:
			h.mutex.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.mutex.RUnlock()
		}
	}
}

func (h *Hub) BroadcastToAll(notification *models.Notification) {
	message, err := json.Marshal(notification)
	if err != nil {
		zap.L().Error("Failed to marshal notification", zap.Error(err))
		return
	}

	h.broadcast <- message
	zap.L().Info("Broadcast notification to all clients",
		zap.String("type", notification.Type),
		zap.String("title", notification.Title),
	)
}

func (h *Hub) BroadcastToUser(userID string, notification *models.Notification) {
	message, err := json.Marshal(notification)
	if err != nil {
		zap.L().Error("Failed to marshal notification", zap.Error(err))
		return
	}

	h.mutex.RLock()
	userClients, exists := h.userClients[userID]
	h.mutex.RUnlock()

	if !exists || len(userClients) == 0 {
		zap.L().Warn("No clients found for user", zap.String("userID", userID))
		return
	}

	for _, client := range userClients {
		select {
		case client.send <- message:
		default:
			close(client.send)
			h.mutex.Lock()
			delete(h.clients, client)
			h.mutex.Unlock()
		}
	}

	zap.L().Info("Broadcast notification to user",
		zap.String("userID", userID),
		zap.String("type", notification.Type),
		zap.String("title", notification.Title),
		zap.Int("clientCount", len(userClients)),
	)
}

func (h *Hub) GetClients() map[*Client]bool {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	return h.clients
}

func (h *Hub) GetUserConnectionCount(userID string) int {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	if clients, exists := h.userClients[userID]; exists {
		return len(clients)
	}
	return 0
}

func (h *Hub) ServeWS(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		zap.L().Error("Failed to upgrade connection", zap.Error(err))
		return
	}

	userID := c.Query("userId")
	clientID := c.Query("clientId")
	if clientID == "" {
		clientID = "anonymous"
	}

	client := &Client{
		hub:      h,
		conn:     conn,
		send:     make(chan []byte, 256),
		userID:   userID,
		clientID: clientID,
	}

	client.hub.register <- client

	go client.writePump()
	go client.readPump()
}
