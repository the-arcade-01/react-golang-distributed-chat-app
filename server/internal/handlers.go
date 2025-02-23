package internal

import (
	"context"
	"encoding/json"
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

func (h *handlers) greet(w http.ResponseWriter, r *http.Request) {
	ResponseWithJSON(w, http.StatusOK, Response{Status: http.StatusOK, Message: "Hello, World"})
}

func (h *handlers) checkExists(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	result := h.cacheRepo.exists(r.Context(), key)
	if result {
		ResponseWithJSON(w, http.StatusOK, Response{Status: http.StatusOK, Message: "Exists"})
		return
	}
	ResponseWithJSON(w, http.StatusNotFound, Response{Status: http.StatusNotFound, Message: "Not Exists"})
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

	go h.writePump(context.Background(), conn)
	go h.readPump(context.Background(), conn, username)
}

func (h *handlers) readPump(ctx context.Context, conn *websocket.Conn, username string) {
	defer func() {
		payload, err := json.Marshal(&Message{
			Timestamp: time.Now().UnixMilli(),
			Username:  username,
			Type:      "LEAVE",
			Content:   fmt.Sprintf("%v left the room", username),
		})
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

	payload, err := json.Marshal(&Message{
		Timestamp: time.Now().UnixMilli(),
		Username:  username,
		Type:      "JOIN",
		Content:   fmt.Sprintf("%v joined the room", username),
	})
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

/**
TODO:
	use mysql table to store message, use batch insert and cron and clean up or limit only last 10 message insert, something like that, send last 10 message on ws conn
	mysql table might look like this
	timestamp, jsontext -> &Message{}
*/
