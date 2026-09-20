CREATE UNIQUE INDEX IF NOT EXISTS bids_auction_bidder_idempotency_key_idx
    ON bids (auction_id, bidder_id, idempotency_key)
    WHERE idempotency_key IS NOT NULL;

