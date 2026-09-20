package auction

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/rohitjoshi6/livebid/backend/internal/auth"
	"github.com/rohitjoshi6/livebid/backend/internal/httpx"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

type auctionRequest struct {
	Title              string  `json:"title"`
	Description        string  `json:"description"`
	ImageURL           *string `json:"image_url"`
	StartingPriceCents int64   `json:"starting_price_cents"`
	DurationSeconds    int     `json:"duration_seconds"`
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized", "Authentication is required.")
		return
	}
	var req auctionRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}
	item, err := h.service.Create(r.Context(), CreateInput{
		SellerID:           claims.UserID,
		Title:              req.Title,
		Description:        req.Description,
		ImageURL:           req.ImageURL,
		StartingPriceCents: req.StartingPriceCents,
		DurationSeconds:    req.DurationSeconds,
	})
	if err != nil {
		writeAuctionError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, item)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	item, err := h.service.Get(r.Context(), chi.URLParam(r, "auctionID"))
	if err != nil {
		writeAuctionError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	page := parseIntQuery(r, "page", 1)
	limit := parseIntQuery(r, "limit", 20)
	result, err := h.service.List(r.Context(), r.URL.Query().Get("status"), page, limit)
	if err != nil {
		writeAuctionError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result)
}

func (h *Handler) UpdateDraft(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized", "Authentication is required.")
		return
	}
	var req auctionRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}
	item, err := h.service.UpdateDraft(r.Context(), chi.URLParam(r, "auctionID"), UpdateInput{
		SellerID:           claims.UserID,
		Title:              req.Title,
		Description:        req.Description,
		ImageURL:           req.ImageURL,
		StartingPriceCents: req.StartingPriceCents,
		DurationSeconds:    req.DurationSeconds,
	})
	if err != nil {
		writeAuctionError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) Start(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized", "Authentication is required.")
		return
	}
	item, err := h.service.Start(r.Context(), chi.URLParam(r, "auctionID"), claims.UserID)
	if err != nil {
		writeAuctionError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) Cancel(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized", "Authentication is required.")
		return
	}
	item, err := h.service.Cancel(r.Context(), chi.URLParam(r, "auctionID"), claims.UserID)
	if err != nil {
		writeAuctionError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}

func parseIntQuery(r *http.Request, key string, fallback int) int {
	value := r.URL.Query().Get(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func writeAuctionError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrInvalidAuction):
		httpx.WriteError(w, http.StatusBadRequest, "invalid_auction", "Auction fields are invalid.")
	case errors.Is(err, ErrForbidden):
		httpx.WriteError(w, http.StatusForbidden, "forbidden", "You do not have access to this auction.")
	case errors.Is(err, ErrNotFound):
		httpx.WriteError(w, http.StatusNotFound, "auction_not_found", "Auction was not found.")
	default:
		httpx.WriteError(w, http.StatusInternalServerError, "auction_error", "Auction request failed.")
	}
}
