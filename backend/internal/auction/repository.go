package auction

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

var ErrNotFound = errors.New("auction not found")

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

type CreateParams struct {
	SellerID           string
	Title              string
	Description        string
	ImageURL           *string
	StartingPriceCents int64
	DurationSeconds    int
}

type UpdateDraftParams struct {
	Title              string
	Description        string
	ImageURL           *string
	StartingPriceCents int64
	DurationSeconds    int
}

func (r *Repository) Create(ctx context.Context, params CreateParams) (Auction, error) {
	var item Auction
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO auctions (
			seller_id, title, description, image_url, starting_price_cents,
			current_price_cents, status, duration_seconds
		)
		VALUES ($1, $2, $3, $4, $5, $5, 'draft', $6)
		RETURNING
			id::text, seller_id::text, title, description, image_url,
			starting_price_cents, current_price_cents, status, duration_seconds,
			start_time, end_time, winner_id::text, version, created_at, updated_at
	`, params.SellerID, params.Title, params.Description, params.ImageURL, params.StartingPriceCents, params.DurationSeconds).Scan(scanAuction(&item)...)
	if err != nil {
		return Auction{}, fmt.Errorf("create auction: %w", err)
	}
	return item, nil
}

func (r *Repository) FindByID(ctx context.Context, id string) (Auction, error) {
	var item Auction
	err := r.db.QueryRowContext(ctx, selectAuctionSQL()+` WHERE id = $1`, id).Scan(scanAuction(&item)...)
	if errors.Is(err, sql.ErrNoRows) {
		return Auction{}, ErrNotFound
	}
	if err != nil {
		return Auction{}, fmt.Errorf("find auction: %w", err)
	}
	return item, nil
}

func (r *Repository) List(ctx context.Context, status string, limit, offset int) ([]Auction, error) {
	query := selectAuctionSQL()
	args := []any{}
	if status != "" {
		query += " WHERE status = $1"
		args = append(args, status)
	}
	query += " ORDER BY created_at DESC LIMIT $" + fmt.Sprint(len(args)+1) + " OFFSET $" + fmt.Sprint(len(args)+2)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list auctions: %w", err)
	}
	defer rows.Close()

	auctions := []Auction{}
	for rows.Next() {
		var item Auction
		if err := rows.Scan(scanAuction(&item)...); err != nil {
			return nil, fmt.Errorf("scan auction: %w", err)
		}
		auctions = append(auctions, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate auctions: %w", err)
	}
	return auctions, nil
}

func (r *Repository) UpdateDraft(ctx context.Context, id, sellerID string, params UpdateDraftParams) (Auction, error) {
	var item Auction
	err := r.db.QueryRowContext(ctx, `
		UPDATE auctions
		SET title = $3,
			description = $4,
			image_url = $5,
			starting_price_cents = $6,
			current_price_cents = $6,
			duration_seconds = $7,
			version = version + 1,
			updated_at = now()
		WHERE id = $1 AND seller_id = $2 AND status = 'draft'
		RETURNING
			id::text, seller_id::text, title, description, image_url,
			starting_price_cents, current_price_cents, status, duration_seconds,
			start_time, end_time, winner_id::text, version, created_at, updated_at
	`, id, sellerID, params.Title, params.Description, params.ImageURL, params.StartingPriceCents, params.DurationSeconds).Scan(scanAuction(&item)...)
	if errors.Is(err, sql.ErrNoRows) {
		return Auction{}, ErrNotFound
	}
	if err != nil {
		return Auction{}, fmt.Errorf("update draft auction: %w", err)
	}
	return item, nil
}

func (r *Repository) Start(ctx context.Context, id, sellerID string, now time.Time) (Auction, error) {
	var item Auction
	err := r.db.QueryRowContext(ctx, `
		UPDATE auctions
		SET status = 'live',
			start_time = $3,
			end_time = $3 + (duration_seconds || ' seconds')::interval,
			version = version + 1,
			updated_at = now()
		WHERE id = $1 AND seller_id = $2 AND status = 'draft'
		RETURNING
			id::text, seller_id::text, title, description, image_url,
			starting_price_cents, current_price_cents, status, duration_seconds,
			start_time, end_time, winner_id::text, version, created_at, updated_at
	`, id, sellerID, now).Scan(scanAuction(&item)...)
	if errors.Is(err, sql.ErrNoRows) {
		return Auction{}, ErrNotFound
	}
	if err != nil {
		return Auction{}, fmt.Errorf("start auction: %w", err)
	}
	return item, nil
}

func (r *Repository) Cancel(ctx context.Context, id, sellerID string) (Auction, error) {
	var item Auction
	err := r.db.QueryRowContext(ctx, `
		UPDATE auctions
		SET status = 'cancelled',
			version = version + 1,
			updated_at = now()
		WHERE id = $1 AND seller_id = $2 AND status IN ('draft', 'scheduled', 'live')
		RETURNING
			id::text, seller_id::text, title, description, image_url,
			starting_price_cents, current_price_cents, status, duration_seconds,
			start_time, end_time, winner_id::text, version, created_at, updated_at
	`, id, sellerID).Scan(scanAuction(&item)...)
	if errors.Is(err, sql.ErrNoRows) {
		return Auction{}, ErrNotFound
	}
	if err != nil {
		return Auction{}, fmt.Errorf("cancel auction: %w", err)
	}
	return item, nil
}

func (r *Repository) CompleteExpired(ctx context.Context, now time.Time, batchSize int) ([]Auction, error) {
	rows, err := r.db.QueryContext(ctx, `
		UPDATE auctions
		SET status = 'completed',
			version = version + 1,
			updated_at = now()
		WHERE id IN (
			SELECT id
			FROM auctions
			WHERE status = 'live' AND end_time <= $1
			ORDER BY end_time ASC
			FOR UPDATE SKIP LOCKED
			LIMIT $2
		)
		RETURNING
			id::text, seller_id::text, title, description, image_url,
			starting_price_cents, current_price_cents, status, duration_seconds,
			start_time, end_time, winner_id::text, version, created_at, updated_at
	`, now, batchSize)
	if err != nil {
		return nil, fmt.Errorf("complete expired auctions: %w", err)
	}
	defer rows.Close()

	completed := []Auction{}
	for rows.Next() {
		var item Auction
		if err := rows.Scan(scanAuction(&item)...); err != nil {
			return nil, fmt.Errorf("scan completed auction: %w", err)
		}
		completed = append(completed, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate completed auctions: %w", err)
	}
	return completed, nil
}

func selectAuctionSQL() string {
	return `
		SELECT
			id::text, seller_id::text, title, description, image_url,
			starting_price_cents, current_price_cents, status, duration_seconds,
			start_time, end_time, winner_id::text, version, created_at, updated_at
		FROM auctions
	`
}

func scanAuction(item *Auction) []any {
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
