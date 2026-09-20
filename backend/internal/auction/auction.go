package auction

import "time"

const (
	StatusDraft     = "draft"
	StatusScheduled = "scheduled"
	StatusLive      = "live"
	StatusCompleted = "completed"
	StatusCancelled = "cancelled"
)

type Auction struct {
	ID                 string     `json:"id"`
	SellerID           string     `json:"seller_id"`
	Title              string     `json:"title"`
	Description        string     `json:"description"`
	ImageURL           *string    `json:"image_url,omitempty"`
	StartingPriceCents int64      `json:"starting_price_cents"`
	CurrentPriceCents  int64      `json:"current_price_cents"`
	Status             string     `json:"status"`
	DurationSeconds    int        `json:"duration_seconds"`
	StartTime          *time.Time `json:"start_time,omitempty"`
	EndTime            *time.Time `json:"end_time,omitempty"`
	WinnerID           *string    `json:"winner_id,omitempty"`
	Version            int64      `json:"version"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

type ListResult struct {
	Auctions []Auction `json:"auctions"`
	Page     int       `json:"page"`
	Limit    int       `json:"limit"`
}
