package bidding

import (
	"context"
	"testing"
)

func TestNormalizeKeyTrimsBlankToNil(t *testing.T) {
	value := "   "
	if normalizeKey(&value) != nil {
		t.Fatal("blank idempotency key should normalize to nil")
	}
}

func TestNormalizeKeyTrimsValue(t *testing.T) {
	value := " bid-123 "
	key := normalizeKey(&value)
	if key == nil || *key != "bid-123" {
		t.Fatalf("unexpected key: %#v", key)
	}
}

func TestPlaceRequiresIdempotencyKey(t *testing.T) {
	service := NewService(NewRepository(nil))
	_, err := service.Place(context.Background(), PlaceInput{
		AuctionID:   "auction-1",
		BidderID:    "bidder-1",
		AmountCents: 100,
	})
	if err != ErrMissingIdempotencyKey {
		t.Fatalf("err = %v, want ErrMissingIdempotencyKey", err)
	}
}
