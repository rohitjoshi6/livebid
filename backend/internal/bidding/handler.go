package bidding

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rohitjoshi6/livebid/backend/internal/auction"
	"github.com/rohitjoshi6/livebid/backend/internal/auth"
	"github.com/rohitjoshi6/livebid/backend/internal/httpx"
	"github.com/rohitjoshi6/livebid/backend/internal/realtime"
)

type Handler struct {
	service   *Service
	publisher realtime.Publisher
}

func NewHandler(service *Service, publisher realtime.Publisher) *Handler {
	return &Handler{service: service, publisher: publisher}
}

type placeBidRequest struct {
	AmountCents    int64   `json:"amount_cents"`
	IdempotencyKey *string `json:"idempotency_key"`
}

func (h *Handler) Place(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized", "Authentication is required.")
		return
	}
	var req placeBidRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}
	result, err := h.service.Place(r.Context(), PlaceInput{
		AuctionID:      chi.URLParam(r, "auctionID"),
		BidderID:       claims.UserID,
		AmountCents:    req.AmountCents,
		IdempotencyKey: req.IdempotencyKey,
	})
	if err != nil {
		writeBidError(w, err)
		return
	}
	_ = h.publisher.Publish(r.Context(), realtime.Event{
		Type:      realtime.EventBidPlaced,
		AuctionID: result.Auction.ID,
		Data:      result,
	})
	_ = h.publisher.Publish(r.Context(), realtime.Event{
		Type:      realtime.EventPriceUpdated,
		AuctionID: result.Auction.ID,
		Data: map[string]any{
			"current_price_cents": result.Auction.CurrentPriceCents,
			"version":             result.Auction.Version,
		},
	})
	httpx.WriteJSON(w, http.StatusCreated, result)
}

func writeBidError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrInvalidBid):
		httpx.WriteError(w, http.StatusBadRequest, "invalid_bid", "Bid amount and auction are required.")
	case errors.Is(err, ErrMissingIdempotencyKey):
		httpx.WriteError(w, http.StatusBadRequest, "missing_idempotency_key", "Bid requests require an idempotency key.")
	case errors.Is(err, ErrIdempotencyConflict):
		httpx.WriteError(w, http.StatusConflict, "idempotency_conflict", "Idempotency key was already used for a different bid.")
	case errors.Is(err, auction.ErrNotFound):
		httpx.WriteError(w, http.StatusNotFound, "auction_not_found", "Auction was not found.")
	case errors.Is(err, ErrAuctionNotLive):
		httpx.WriteError(w, http.StatusConflict, "auction_not_live", "Auction is not accepting bids.")
	case errors.Is(err, ErrAuctionExpired):
		httpx.WriteError(w, http.StatusConflict, "auction_expired", "Auction has already expired.")
	case errors.Is(err, ErrSellerCannotBid):
		httpx.WriteError(w, http.StatusForbidden, "seller_cannot_bid", "Seller cannot bid on their own auction.")
	case errors.Is(err, ErrBidTooLow):
		httpx.WriteError(w, http.StatusUnprocessableEntity, "bid_too_low", "Bid must exceed the current price.")
	default:
		httpx.WriteError(w, http.StatusInternalServerError, "bid_error", "Bid could not be placed.")
	}
}
