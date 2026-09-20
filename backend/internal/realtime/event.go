package realtime

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

const (
	EventAuctionStarted     = "auction_started"
	EventBidPlaced          = "bid_placed"
	EventPriceUpdated       = "price_updated"
	EventAuctionCancelled   = "auction_cancelled"
	EventConnectionRestored = "connection_restored"
)

type Event struct {
	Type      string    `json:"type"`
	AuctionID string    `json:"auction_id"`
	Data      any       `json:"data"`
	SentAt    time.Time `json:"sent_at"`
}

type Publisher interface {
	Publish(ctx context.Context, event Event) error
}

func Channel(auctionID string) string {
	return fmt.Sprintf("auction:%s", auctionID)
}

func Encode(event Event) ([]byte, error) {
	return json.Marshal(event)
}
