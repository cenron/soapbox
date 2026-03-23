package ws

import (
	"log/slog"
	"net/http"

	"github.com/coder/websocket"
	"github.com/radni/soapbox/internal/core/types"
)

// TokenValidator validates a JWT token string and returns the user ID.
// Provided by the users module at composition time.
type TokenValidator func(token string) (types.ID, error)

// UpgradeHandler returns an http.HandlerFunc that upgrades to WebSocket,
// authenticates via ?token= query parameter, and keeps the connection alive.
func UpgradeHandler(hub *Hub, validate TokenValidator, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.URL.Query().Get("token")
		if token == "" {
			http.Error(w, "missing token", http.StatusUnauthorized)
			return
		}

		userID, err := validate(token)
		if err != nil {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		ws, err := websocket.Accept(w, r, &websocket.AcceptOptions{
			InsecureSkipVerify: true, // Allow connections from any origin in dev.
		})
		if err != nil {
			logger.Error("ws: accept failed", "error", err)
			return
		}

		conn := NewConn(ws)
		hub.Register(userID, conn)
		defer func() {
			hub.Deregister(userID, conn)
			conn.Close()
		}()

		logger.Debug("ws: client connected", "user_id", userID)

		// Read pump — keeps the connection alive and detects client disconnect.
		// We don't process client messages; this just blocks until close.
		for {
			_, _, err := ws.Read(r.Context())
			if err != nil {
				break
			}
		}
	}
}
