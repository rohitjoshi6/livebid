package auction

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"time"
)

var ErrInvalidAuction = errors.New("invalid auction")
var ErrForbidden = errors.New("forbidden")

type Service struct {
	repo *Repository
	now  func() time.Time
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo, now: time.Now}
}

type CreateInput struct {
	SellerID           string
	Title              string
	Description        string
	ImageURL           *string
	StartingPriceCents int64
	DurationSeconds    int
}

type UpdateInput struct {
	SellerID           string
	Title              string
	Description        string
	ImageURL           *string
	StartingPriceCents int64
	DurationSeconds    int
}

func (s *Service) Create(ctx context.Context, input CreateInput) (Auction, error) {
	params, err := validateAuctionInput(input.Title, input.Description, input.ImageURL, input.StartingPriceCents, input.DurationSeconds)
	if err != nil {
		return Auction{}, err
	}
	return s.repo.Create(ctx, CreateParams{
		SellerID:           input.SellerID,
		Title:              params.title,
		Description:        params.description,
		ImageURL:           params.imageURL,
		StartingPriceCents: params.startingPriceCents,
		DurationSeconds:    params.durationSeconds,
	})
}

func (s *Service) Get(ctx context.Context, id string) (Auction, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *Service) List(ctx context.Context, status string, page, limit int) (ListResult, error) {
	if status != "" && !validStatus(status) {
		return ListResult{}, ErrInvalidAuction
	}
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit
	items, err := s.repo.List(ctx, status, limit, offset)
	if err != nil {
		return ListResult{}, err
	}
	return ListResult{Auctions: items, Page: page, Limit: limit}, nil
}

func (s *Service) UpdateDraft(ctx context.Context, id string, input UpdateInput) (Auction, error) {
	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return Auction{}, err
	}
	if existing.SellerID != input.SellerID {
		return Auction{}, ErrForbidden
	}
	if existing.Status != StatusDraft {
		return Auction{}, ErrInvalidAuction
	}

	params, err := validateAuctionInput(input.Title, input.Description, input.ImageURL, input.StartingPriceCents, input.DurationSeconds)
	if err != nil {
		return Auction{}, err
	}
	return s.repo.UpdateDraft(ctx, id, input.SellerID, UpdateDraftParams{
		Title:              params.title,
		Description:        params.description,
		ImageURL:           params.imageURL,
		StartingPriceCents: params.startingPriceCents,
		DurationSeconds:    params.durationSeconds,
	})
}

func (s *Service) Start(ctx context.Context, id, sellerID string) (Auction, error) {
	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return Auction{}, err
	}
	if existing.SellerID != sellerID {
		return Auction{}, ErrForbidden
	}
	if existing.Status != StatusDraft {
		return Auction{}, ErrInvalidAuction
	}
	return s.repo.Start(ctx, id, sellerID, s.now().UTC())
}

func (s *Service) Cancel(ctx context.Context, id, sellerID string) (Auction, error) {
	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return Auction{}, err
	}
	if existing.SellerID != sellerID {
		return Auction{}, ErrForbidden
	}
	if existing.Status == StatusCompleted || existing.Status == StatusCancelled {
		return Auction{}, ErrInvalidAuction
	}
	return s.repo.Cancel(ctx, id, sellerID)
}

func (s *Service) CompleteExpired(ctx context.Context, batchSize int) ([]Auction, error) {
	if batchSize < 1 {
		batchSize = 50
	}
	return s.repo.CompleteExpired(ctx, s.now().UTC(), batchSize)
}

type normalizedInput struct {
	title              string
	description        string
	imageURL           *string
	startingPriceCents int64
	durationSeconds    int
}

func validateAuctionInput(title, description string, imageURL *string, startingPriceCents int64, durationSeconds int) (normalizedInput, error) {
	title = strings.TrimSpace(title)
	description = strings.TrimSpace(description)
	if len(title) < 3 || len(title) > 120 {
		return normalizedInput{}, ErrInvalidAuction
	}
	if len(description) < 10 || len(description) > 2000 {
		return normalizedInput{}, ErrInvalidAuction
	}
	if startingPriceCents <= 0 {
		return normalizedInput{}, ErrInvalidAuction
	}
	if durationSeconds < 30 || durationSeconds > 30*24*60*60 {
		return normalizedInput{}, ErrInvalidAuction
	}
	normalizedImageURL := normalizeOptionalURL(imageURL)
	if normalizedImageURL != nil {
		parsed, err := url.ParseRequestURI(*normalizedImageURL)
		if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") {
			return normalizedInput{}, ErrInvalidAuction
		}
	}
	return normalizedInput{
		title:              title,
		description:        description,
		imageURL:           normalizedImageURL,
		startingPriceCents: startingPriceCents,
		durationSeconds:    durationSeconds,
	}, nil
}

func normalizeOptionalURL(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func validStatus(status string) bool {
	switch status {
	case StatusDraft, StatusScheduled, StatusLive, StatusCompleted, StatusCancelled:
		return true
	default:
		return false
	}
}
