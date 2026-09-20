package auction

import "testing"

func TestValidateAuctionInputRequiresPositivePrice(t *testing.T) {
	_, err := validateAuctionInput("Vintage Camera", "A working film camera.", nil, 0, 60)
	if err == nil {
		t.Fatal("expected zero starting price to be invalid")
	}
}

func TestValidateAuctionInputNormalizesOptionalImageURL(t *testing.T) {
	raw := " https://example.com/camera.jpg "
	input, err := validateAuctionInput("Vintage Camera", "A working film camera.", &raw, 1000, 60)
	if err != nil {
		t.Fatalf("expected valid auction input: %v", err)
	}
	if input.imageURL == nil || *input.imageURL != "https://example.com/camera.jpg" {
		t.Fatalf("unexpected normalized image url: %#v", input.imageURL)
	}
}
