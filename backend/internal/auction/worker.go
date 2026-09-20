package auction

import (
	"context"
	"log/slog"
	"time"
)

type ExpirationWorker struct {
	service  *Service
	logger   *slog.Logger
	interval time.Duration
	batch    int
}

func NewExpirationWorker(service *Service, logger *slog.Logger, interval time.Duration, batch int) *ExpirationWorker {
	return &ExpirationWorker{
		service:  service,
		logger:   logger,
		interval: interval,
		batch:    batch,
	}
}

func (w *ExpirationWorker) Run(ctx context.Context) {
	if w.interval <= 0 {
		w.interval = 2 * time.Second
	}
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	w.completeOnce(ctx)
	for {
		select {
		case <-ctx.Done():
			w.logger.Info("auction expiration worker stopped")
			return
		case <-ticker.C:
			w.completeOnce(ctx)
		}
	}
}

func (w *ExpirationWorker) completeOnce(ctx context.Context) {
	completed, err := w.service.CompleteExpired(ctx, w.batch)
	if err != nil {
		w.logger.Error("auction expiration pass failed", "error", err)
		return
	}
	for _, item := range completed {
		w.logger.Info("auction completed by expiration worker", "auction_id", item.ID, "final_price_cents", item.CurrentPriceCents)
	}
}
