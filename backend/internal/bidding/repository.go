package bidding

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/rohitjoshi6/livebid/backend/internal/auction"
)

var ErrAuctionNotLive = errors.New("auction is not live")
var ErrBidTooLow = errors.New("bid must exceed current price")
var ErrAuctionExpired = errors.New("auction has expired")
var ErrSellerCannotBid = errors.New("seller cannot bid on own auction")

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

type PlaceParams struct {
	AuctionID      string
	BidderID       string
	AmountCents    int64
	IdempotencyKey *string
	Now            time.Time
}

func (r *Repository) Place(ctx context.Context, params PlaceParams) (PlaceResult, error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return PlaceResult{}, fmt.Errorf("begin bid transaction: %w", err)
	}
	defer tx.Rollback()

	item, err := lockAuction(ctx, tx, params.AuctionID)
	if err != nil {
		return PlaceResult{}, err
	}
	if item.Status != auction.StatusLive {
		return PlaceResult{}, ErrAuctionNotLive
	}
	if item.EndTime == nil || !params.Now.Before(*item.EndTime) {
		return PlaceResult{}, ErrAuctionExpired
	}
	if item.SellerID == params.BidderID {
		return PlaceResult{}, ErrSellerCannotBid
	}
	if params.AmountCents <= item.CurrentPriceCents {
		return PlaceResult{}, ErrBidTooLow
	}

	bid, err := insertBid(ctx, tx, params)
	if err != nil {
		return PlaceResult{}, err
	}
	updated, err := updateAuctionPrice(ctx, tx, params.AuctionID, params.AmountCents)
	if err != nil {
		return PlaceResult{}, err
	}
	if err := tx.Commit(); err != nil {
		return PlaceResult{}, fmt.Errorf("commit bid transaction: %w", err)
	}
	return PlaceResult{Bid: bid, Auction: updated}, nil
}

func lockAuction(ctx context.Context, tx *sql.Tx, auctionID string) (auction.Auction, error) {
	var item auction.Auction
	err := tx.QueryRowContext(ctx, `
		SELECT
			id::text, seller_id::text, title, description, image_url,
			starting_price_cents, current_price_cents, status, duration_seconds,
			start_time, end_time, winner_id::text, version, created_at, updated_at
		FROM auctions
		WHERE id = $1
		FOR UPDATE
	`, auctionID).Scan(scanAuction(&item)...)
	if errors.Is(err, sql.ErrNoRows) {
		return auction.Auction{}, auction.ErrNotFound
	}
	if err != nil {
		return auction.Auction{}, fmt.Errorf("lock auction: %w", err)
	}
	return item, nil
}

func insertBid(ctx context.Context, tx *sql.Tx, params PlaceParams) (Bid, error) {
	var bid Bid
	err := tx.QueryRowContext(ctx, `
		INSERT INTO bids (auction_id, bidder_id, amount_cents, idempotency_key)
		VALUES ($1, $2, $3, $4)
		RETURNING id::text, auction_id::text, bidder_id::text, amount_cents, idempotency_key, created_at
	`, params.AuctionID, params.BidderID, params.AmountCents, params.IdempotencyKey).Scan(
		&bid.ID,
		&bid.AuctionID,
		&bid.BidderID,
		&bid.AmountCents,
		&bid.IdempotencyKey,
		&bid.CreatedAt,
	)
	if err != nil {
		return Bid{}, fmt.Errorf("insert bid: %w", err)
	}
	return bid, nil
}

func updateAuctionPrice(ctx context.Context, tx *sql.Tx, auctionID string, amountCents int64) (auction.Auction, error) {
	var item auction.Auction
	err := tx.QueryRowContext(ctx, `
		UPDATE auctions
		SET current_price_cents = $2,
			version = version + 1,
			updated_at = now()
		WHERE id = $1
		RETURNING
			id::text, seller_id::text, title, description, image_url,
			starting_price_cents, current_price_cents, status, duration_seconds,
			start_time, end_time, winner_id::text, version, created_at, updated_at
	`, auctionID, amountCents).Scan(scanAuction(&item)...)
	if err != nil {
		return auction.Auction{}, fmt.Errorf("update auction price: %w", err)
	}
	return item, nil
}

func scanAuction(item *auction.Auction) []any {
	return []any{
		&item.ID,
		&item.SellerID,
		&item.Title,
		&item.Description,
		&item.ImageURL,
		&item.StartingPriceCents,
		&item.CurrentPriceCents,
		&item.Status,
		&item.DurationSeconds,
		&item.StartTime,
		&item.EndTime,
		&item.WinnerID,
		&item.Version,
		&item.CreatedAt,
		&item.UpdatedAt,
	}
}
