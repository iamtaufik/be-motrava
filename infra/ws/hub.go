package ws

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"

	"github.com/redis/go-redis/v9"

	"motrava/core/models"
	"motrava/core/repository"
)

// PositionUpdate is the payload sent from mobile and broadcast to dashboard.
type PositionUpdate struct {
	TripID    string  `json:"trip_id"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Speed     float64 `json:"speed"`
	Heading   float64 `json:"heading"`
	Accuracy  float64 `json:"accuracy"`
	Altitude  float64 `json:"altitude"`
	Battery   int     `json:"battery"`
	Timestamp string  `json:"timestamp"`
}

// Hub manages WebSocket connections grouped by trip_id for broadcasting.
type Hub struct {
	mu          sync.RWMutex
	connections map[string]map[*Client]bool // trip_id -> set of clients
	log         *slog.Logger
	rdb         *redis.Client
	tripPointRepo repository.TripPointRepository
}

type Client struct {
	hub    *Hub
	TripID string
	UserID string
	Send   chan []byte
	close  chan struct{}
}

func NewClient(hub *Hub, tripID, userID string) *Client {
	return &Client{
		hub:    hub,
		TripID: tripID,
		UserID: userID,
		Send:   make(chan []byte, 256),
	}
}

func NewHub(logger *slog.Logger, rdb *redis.Client, tripPointRepo repository.TripPointRepository) *Hub {
	return &Hub{
		connections:   make(map[string]map[*Client]bool),
		log:           logger,
		rdb:           rdb,
		tripPointRepo: tripPointRepo,
	}
}

func (h *Hub) Register(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.connections[client.TripID] == nil {
		h.connections[client.TripID] = make(map[*Client]bool)
	}
	h.connections[client.TripID][client] = true
	h.log.Info("ws client connected", "module", "ws_hub", "trip_id", client.TripID, "user_id", client.UserID)
}

func (h *Hub) Unregister(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if clients, ok := h.connections[client.TripID]; ok {
		delete(clients, client)
		if len(clients) == 0 {
			delete(h.connections, client.TripID)
		}
	}
	h.log.Info("ws client disconnected", "module", "ws_hub", "trip_id", client.TripID, "user_id", client.UserID)
}

func (h *Hub) BroadcastToTrip(tripID string, data []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if clients, ok := h.connections[tripID]; ok {
		for client := range clients {
			select {
			case client.Send <- data:
			default:
				h.log.Warn("ws client send buffer full, dropping message", "module", "ws_hub", "trip_id", tripID)
			}
		}
	}
}

func (h *Hub) HandlePosition(payload PositionUpdate, speed float64) {
	payload.Speed = speed

	data, _ := json.Marshal(payload)

	h.rdb.Set(context.Background(), "trip:"+payload.TripID+":current", string(data), 0)

	h.BroadcastToTrip(payload.TripID, data)
}

func (h *Hub) UpdateLastPoint(tripID string, point models.TripPoint) {
	data, _ := json.Marshal(point)
	h.rdb.Set(context.Background(), "trip:"+tripID+":last_point", string(data), 0)
}

func (h *Hub) Redis() *redis.Client {
	return h.rdb
}
