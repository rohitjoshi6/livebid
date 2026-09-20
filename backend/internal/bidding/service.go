package bidding

import (
	"context"
	"errors"
	"strings"
	"time"
)

var ErrInvalidBid = errors.New("invalid bid")
var ErrMissingIdempotencyKey = errors.New("missing idempotency key")

type Service struct {
	repo         *Repository
	now          func() time.Time
	extendWindow time.Duration
	extendBy     time.Duration
}

func NewService(repo *Repository, extendWindow, extendBy time.Duration) *Service {
	return &Service{repo: repo, now: time.Now, extendWindow: extendWindow, extendBy: extendBy}
}

type PlaceInput struct {
	AuctionID      string
	BidderID       string
	AmountCents    int64
	IdempotencyKey *string
}

func (s *Service) Place(ctx context.Context, input PlaceInput) (PlaceResult, error) {
	if input.AuctionID == "" || input.BidderID == "" || input.AmountCents <= 0 {
		return PlaceResult{}, ErrInvalidBid
	}
	key := normalizeKey(input.IdempotencyKey)
	if key == nil {
		return PlaceResult{}, ErrMissingIdempotencyKey
	}
	return s.repo.Place(ctx, PlaceParams{
		AuctionID:      input.AuctionID,
		BidderID:       input.BidderID,
		AmountCents:    input.AmountCents,
		IdempotencyKey: key,
		Now:            s.now().UTC(),
		ExtendWindow:   s.extendWindow,
		ExtendBy:       s.extendBy,
	})
}

func normalizeKey(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}
