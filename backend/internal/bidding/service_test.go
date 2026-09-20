package bidding

import (
	"context"
	"testing"
	"time"
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
	service := NewService(NewRepository(nil), 10*time.Second, 10*time.Second)
	_, err := service.Place(context.Background(), PlaceInput{
		AuctionID:   "auction-1",
		BidderID:    "bidder-1",
		AmountCents: 100,
	})
	if err != ErrMissingIdempotencyKey {
		t.Fatalf("err = %v, want ErrMissingIdempotencyKey", err)
	}
}

func TestAntiSnipingExtensionExtendsInsideWindow(t *testing.T) {
	end := time.Date(2026, 1, 1, 12, 0, 10, 0, time.UTC)
	now := time.Date(2026, 1, 1, 12, 0, 1, 0, time.UTC)

	previous, extended, ok := antiSnipingExtension(&end, now, 10*time.Second, 10*time.Second)
	if !ok {
		t.Fatal("expected extension")
	}
	if !previous.Equal(end) {
		t.Fatalf("previous = %v, want %v", previous, end)
	}
	if !extended.Equal(end.Add(10 * time.Second)) {
		t.Fatalf("extended = %v, want %v", extended, end.Add(10*time.Second))
	}
}

func TestAntiSnipingExtensionDoesNotExtendOutsideWindow(t *testing.T) {
	end := time.Date(2026, 1, 1, 12, 1, 0, 0, time.UTC)
	now := time.Date(2026, 1, 1, 12, 0, 1, 0, time.UTC)

	_, _, ok := antiSnipingExtension(&end, now, 10*time.Second, 10*time.Second)
	if ok {
		t.Fatal("did not expect extension")
	}
}
