package ws

import (
	"log/slog"
	"sync"

	"github.com/radni/soapbox/internal/core/types"
)

// Message is the envelope pushed to WebSocket clients.
type Message struct {
	Type string `json:"type"`
	Data any    `json:"data"`
}

// Hub manages WebSocket connections grouped by user ID.
// Multiple connections per user are supported (multi-device).
type Hub struct {
	mu    sync.RWMutex
	conns map[types.ID]map[*Conn]struct{}

	logger *slog.Logger
}

func NewHub(logger *slog.Logger) *Hub {
	return &Hub{
		conns:  make(map[types.ID]map[*Conn]struct{}),
		logger: logger,
	}
}

// Register adds a connection for the given user.
func (h *Hub) Register(userID types.ID, conn *Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.conns[userID] == nil {
		h.conns[userID] = make(map[*Conn]struct{})
	}
	h.conns[userID][conn] = struct{}{}
}

// Deregister removes a connection for the given user.
func (h *Hub) Deregister(userID types.ID, conn *Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	set := h.conns[userID]
	if set == nil {
		return
	}

	delete(set, conn)
	if len(set) == 0 {
		delete(h.conns, userID)
	}
}

// Send pushes a message to all connections for the given user.
// Connections that fail to write are closed and removed.
func (h *Hub) Send(userID types.ID, msg Message) {
	h.mu.RLock()
	set := h.conns[userID]
	if len(set) == 0 {
		h.mu.RUnlock()
		return
	}

	// Copy to avoid holding lock during writes.
	targets := make([]*Conn, 0, len(set))
	for c := range set {
		targets = append(targets, c)
	}
	h.mu.RUnlock()

	for _, c := range targets {
		if err := c.WriteJSON(msg); err != nil {
			h.logger.Debug("ws: send failed, closing connection",
				"user_id", userID,
				"error", err,
			)
			c.Close()
			h.Deregister(userID, c)
		}
	}
}
