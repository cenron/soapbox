package ws

import (
	"context"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

const writeTimeout = 5 * time.Second

// Conn wraps a websocket connection with a write mutex for safe concurrent writes.
type Conn struct {
	ws *websocket.Conn
	mu sync.Mutex
}

func NewConn(ws *websocket.Conn) *Conn {
	return &Conn{ws: ws}
}

// WriteJSON serializes v as JSON and writes it to the connection.
func (c *Conn) WriteJSON(v any) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), writeTimeout)
	defer cancel()

	return wsjson.Write(ctx, c.ws, v)
}

// Close closes the underlying WebSocket connection.
func (c *Conn) Close() {
	_ = c.ws.CloseNow()
}
