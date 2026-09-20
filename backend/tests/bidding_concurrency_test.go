package tests

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/rohitjoshi6/livebid/backend/internal/bidding"
	"github.com/rohitjoshi6/livebid/backend/internal/database"
)

func TestConcurrentBidsSerializeOnAuctionRow(t *testing.T) {
	databaseURL := os.Getenv("LIVEBID_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("set LIVEBID_DATABASE_URL to run PostgreSQL-backed concurrency test")
	}

	ctx := context.Background()
	db, err := database.Open(ctx, databaseURL)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	defer db.Close()
	if err := database.RunMigrations(ctx, db, "../migrations"); err != nil {
		t.Fatalf("run migrations: %v", err)
	}

	runID := time.Now().UnixNano()
	sellerID := insertTestUser(t, ctx, db, fmt.Sprintf("seller-%d@example.com", runID), fmt.Sprintf("seller-%d", runID))
	auctionID := insertLiveAuction(t, ctx, db, sellerID)

	repo := bidding.NewRepository(db)
	const bidderCount = 50
	var wg sync.WaitGroup
	var mu sync.Mutex
	accepted := []int64{}

	for i := 0; i < bidderCount; i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			bidderID := insertTestUser(t, ctx, db, fmt.Sprintf("buyer-%d-%d@example.com", runID, i), fmt.Sprintf("buyer-%d-%d", runID, i))
			amount := int64(1001 + i)
			_, err := repo.Place(ctx, bidding.PlaceParams{
				AuctionID:   auctionID,
				BidderID:    bidderID,
				AmountCents: amount,
				Now:         time.Now().UTC(),
			})
			if err == nil {
				mu.Lock()
				accepted = append(accepted, amount)
				mu.Unlock()
				return
			}
			if !errors.Is(err, bidding.ErrBidTooLow) {
				t.Errorf("unexpected bid error: %v", err)
			}
		}()
	}
	wg.Wait()

	if len(accepted) == 0 {
		t.Fatal("expected at least one accepted bid")
	}
	var currentPrice int64
	if err := db.QueryRowContext(ctx, "SELECT current_price_cents FROM auctions WHERE id = $1", auctionID).Scan(&currentPrice); err != nil {
		t.Fatalf("read final price: %v", err)
	}
	maxAccepted := accepted[0]
	for _, amount := range accepted {
		if amount > maxAccepted {
			maxAccepted = amount
		}
	}
	if currentPrice != maxAccepted {
		t.Fatalf("final price = %d, want max accepted bid %d", currentPrice, maxAccepted)
	}
}

func insertTestUser(t *testing.T, ctx context.Context, db *sql.DB, email, username string) string {
	t.Helper()
	var id string
	err := db.QueryRowContext(ctx, `
		INSERT INTO users (email, username, password_hash)
		VALUES ($1, $2, 'test-hash')
		RETURNING id::text
	`, email, username).Scan(&id)
	if err != nil {
		t.Fatalf("insert test user: %v", err)
	}
	return id
}

func insertLiveAuction(t *testing.T, ctx context.Context, db *sql.DB, sellerID string) string {
	t.Helper()
	var id string
	err := db.QueryRowContext(ctx, `
		INSERT INTO auctions (
			seller_id, title, description, starting_price_cents, current_price_cents,
			status, duration_seconds, start_time, end_time
		)
		VALUES ($1, 'Concurrency Camera', 'A test auction for concurrent bids.', 1000, 1000, 'live', 120, now(), now() + interval '2 minutes')
		RETURNING id::text
	`, sellerID).Scan(&id)
	if err != nil {
		t.Fatalf("insert live auction: %v", err)
	}
	return id
}
