package internal

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 5 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 30 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 5) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

type handlers struct {
	cacheRepo *cacheRepo
	upgrader  websocket.Upgrader
}

func newHandlers() *handlers {
	return &handlers{
		cacheRepo: newCacheRepo(),
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				origin := r.Header.Get("Origin")
				return origin == Envs.WEB_URL
			},
		},
	}
}

func (h *handlers) handleWS(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	if username == "" {
		http.Error(w, "Username is required", http.StatusBadRequest)
		Log.ErrorContext(r.Context(), "Username is required")
		return
	}

	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		Log.ErrorContext(r.Context(), "WebSocket upgrade failed", "error", err)
		http.Error(w, "Failed to upgrade connection", http.StatusInternalServerError)
		return
	}

	Log.InfoContext(r.Context(), "WebSocket connection established", "username", username)

	ctx := context.Background()
	go h.writePump(ctx, conn)
	go h.readPump(ctx, conn, username)
}

func (h *handlers) readPump(ctx context.Context, conn *websocket.Conn, username string) {
	defer func() {
		payload, err := GetJSONMessage(username, "LEAVE", fmt.Sprintf("%v left the room", username))
		if err == nil {
			h.cacheRepo.publish(ctx, Envs.CHAT_CHANNEL, string(payload))
		}
		conn.Close()
	}()

	conn.SetReadLimit(maxMessageSize)
	conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(appData string) error {
		conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	payload, err := GetJSONMessage(username, "JOIN", fmt.Sprintf("%v joined the room", username))
	if err == nil {
		h.cacheRepo.publish(ctx, Envs.CHAT_CHANNEL, string(payload))
	}

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				Log.ErrorContext(ctx, "error on ws", "error", err)
			}
			break
		}
		if err := h.cacheRepo.publish(ctx, Envs.CHAT_CHANNEL, string(msg)); err != nil {
			Log.ErrorContext(ctx, "error publishing message to redis", "error", err)
		}
	}
}

func (h *handlers) writePump(ctx context.Context, conn *websocket.Conn) {
	ticker := time.NewTicker(pingPeriod)
	pubsub := h.cacheRepo.redis.Subscribe(ctx, Envs.CHAT_CHANNEL)

	defer func() {
		pubsub.Close()
		ticker.Stop()
		conn.Close()
	}()

	ch := pubsub.Channel()

	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-ch:
			if !ok {
				conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			// Expliciting checking message size, even though checkOrigin check is added
			// so that any manuall attempt to direct conn on ws with the same origin
			// doesn't cause cause conn break
			if len(msg.Payload) > maxMessageSize {
				Log.ErrorContext(ctx, "message size exceeds limit", "size", len(msg.Payload))
				continue
			}

			// if message writing takes more time than writeWait,
			// which can means client as slow internet
			// and then just hang the conn and UI will restablish it
			conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := conn.WriteMessage(websocket.TextMessage, []byte(msg.Payload)); err != nil {
				Log.ErrorContext(ctx, "error sending messsage to websocket", "error", err)
				return
			}
		case <-ticker.C:
			// Here we are pinging the client on pingPeriod to check whether conn exists
			// this is done to remove any stale conns
			conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				Log.ErrorContext(ctx, "error sending ping message", "error", err)
				return
			}
		}
	}
}
