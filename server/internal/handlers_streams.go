package internal

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

type handlersStreams struct {
	repo     *cacheRepo
	upgrader websocket.Upgrader
}

func newHandlersStreams() *handlersStreams {
	return &handlersStreams{
		repo: newCacheRepo(),
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

func (h *handlersStreams) handleWS(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	h.repo.initStream(ctx, Envs.STREAM_KEY)

	username := r.URL.Query().Get("username")
	if username == "" {
		http.Error(w, "Username is required", http.StatusBadRequest)
		Log.ErrorContext(ctx, "Username is required")
		return
	}

	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		Log.ErrorContext(ctx, "WebSocket upgrade failed", "error", err)
		http.Error(w, "Failed to upgrade connection", http.StatusInternalServerError)
		return
	}

	Log.InfoContext(ctx, "WebSocket connection established", "username", username)

	err = h.sendChatHistory(ctx, conn)
	if err != nil {
		Log.ErrorContext(ctx, "error sending chat history", "error", err)
	}

	go h.writePump(ctx, conn)
	go h.readPump(ctx, conn, username)
}

func (h *handlersStreams) sendChatHistory(ctx context.Context, conn *websocket.Conn) error {
	msgs, err := h.repo.getMessagesFromStream(ctx, Envs.STREAM_KEY)
	if err == nil {
		for i := len(msgs) - 1; i >= 0; i-- {
			if err := conn.WriteMessage(websocket.TextMessage, []byte(msgs[i])); err != nil {
				return err
			}
		}
	}
	return nil
}

func (h *handlersStreams) readPump(ctx context.Context, conn *websocket.Conn, username string) {
	defer func() {
		payload, err := GetJSONMessage(username, "LEAVE", fmt.Sprintf("%v left the room", username))
		if err == nil {
			h.repo.writeToStream(ctx, Envs.STREAM_KEY, string(payload))
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
		h.repo.writeToStream(ctx, Envs.STREAM_KEY, string(payload))
	}

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				Log.ErrorContext(ctx, "error on ws", "error", err)
			}
			break
		}
		if err := h.repo.writeToStream(ctx, Envs.STREAM_KEY, string(msg)); err != nil {
			Log.ErrorContext(ctx, "error publishing message to redis", "error", err)
		}
	}
}

func (h *handlersStreams) writePump(ctx context.Context, conn *websocket.Conn) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		conn.Close()
	}()

	lastMsgID := "$"
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Here we are pinging the client on pingPeriod to check whether conn exists
			// this is done to remove any stale conns
			conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				Log.ErrorContext(ctx, "error sending ping message", "error", err)
				return
			}
		default:
			streams, err := h.repo.redis.XRead(ctx, &redis.XReadArgs{
				Streams: []string{Envs.STREAM_KEY, lastMsgID},
				Block:   0,
			}).Result()

			if err != nil {
				Log.ErrorContext(ctx, "error reading from stream", "error", err)
				continue
			}

			for _, stream := range streams {
				for _, msg := range stream.Messages {
					lastMsgID = msg.ID
					payload := msg.Values["message"].(string)

					// if message writing takes more time than writeWait,
					// which can means client as slow internet
					// and then just hang the conn and UI will restablish it
					conn.SetWriteDeadline(time.Now().Add(writeWait))
					if err := conn.WriteMessage(websocket.TextMessage, []byte(payload)); err != nil {
						Log.ErrorContext(ctx, "error sending messsage to websocket", "error", err)
						return
					}
				}
			}
		}
	}
}
