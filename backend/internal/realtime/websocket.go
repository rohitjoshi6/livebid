package realtime

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

type WebSocketHandler struct {
	client   *redis.Client
	logger   *slog.Logger
	upgrader websocket.Upgrader
}

func NewWebSocketHandler(client *redis.Client, logger *slog.Logger, allowedOrigins []string) *WebSocketHandler {
	allowed := map[string]struct{}{}
	for _, origin := range allowedOrigins {
		allowed[origin] = struct{}{}
	}
	return &WebSocketHandler{
		client: client,
		logger: logger,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				if len(allowed) == 0 {
					return false
				}
				_, ok := allowed[r.Header.Get("Origin")]
				return ok
			},
		},
	}
}

func (h *WebSocketHandler) Subscribe(w http.ResponseWriter, r *http.Request) {
	auctionID := chi.URLParam(r, "auctionID")
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Warn("websocket upgrade failed", "auction_id", auctionID, "error", err)
		return
	}
	defer conn.Close()

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	pubsub := h.client.Subscribe(ctx, Channel(auctionID))
	defer pubsub.Close()
	if _, err := pubsub.Receive(ctx); err != nil {
		h.logger.Error("redis subscribe failed", "auction_id", auctionID, "error", err)
		_ = conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseInternalServerErr, "subscription failed"), time.Now().Add(time.Second))
		return
	}

	restored := Event{
		Type:      EventConnectionRestored,
		AuctionID: auctionID,
		Data:      map[string]string{"status": "subscribed"},
		SentAt:    time.Now().UTC(),
	}
	if err := conn.WriteJSON(restored); err != nil {
		return
	}

	go drainClientMessages(ctx, cancel, conn)

	for msg := range pubsub.Channel() {
		if !json.Valid([]byte(msg.Payload)) {
			h.logger.Warn("discarding malformed realtime payload", "auction_id", auctionID)
			continue
		}
		if err := conn.WriteMessage(websocket.TextMessage, []byte(msg.Payload)); err != nil {
			if !errors.Is(err, context.Canceled) {
				h.logger.Info("websocket write stopped", "auction_id", auctionID, "error", err)
			}
			return
		}
	}
}

func drainClientMessages(ctx context.Context, cancel context.CancelFunc, conn *websocket.Conn) {
	defer cancel()
	for {
		select {
		case <-ctx.Done():
			return
		default:
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
		}
	}
}
