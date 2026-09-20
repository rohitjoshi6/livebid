package bidding

import (
	"time"

	"github.com/rohitjoshi6/livebid/backend/internal/auction"
)

type Bid struct {
	ID             string    `json:"id"`
	AuctionID      string    `json:"auction_id"`
	BidderID       string    `json:"bidder_id"`
	AmountCents    int64     `json:"amount_cents"`
	IdempotencyKey *string   `json:"idempotency_key,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

type PlaceResult struct {
	Bid             Bid             `json:"bid"`
	Auction         auction.Auction `json:"auction"`
	Extended        bool            `json:"extended"`
	PreviousEndTime *time.Time      `json:"previous_end_time,omitempty"`
}
