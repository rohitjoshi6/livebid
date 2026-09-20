package bidding

import "testing"

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
