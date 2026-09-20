CREATE TABLE IF NOT EXISTS auctions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    seller_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    image_url TEXT,
    starting_price_cents BIGINT NOT NULL CHECK (starting_price_cents > 0),
    current_price_cents BIGINT NOT NULL CHECK (current_price_cents >= starting_price_cents),
    status TEXT NOT NULL CHECK (status IN ('draft', 'scheduled', 'live', 'completed', 'cancelled')),
    duration_seconds INTEGER NOT NULL CHECK (duration_seconds > 0),
    start_time TIMESTAMPTZ,
    end_time TIMESTAMPTZ,
    winner_id UUID REFERENCES users(id),
    version BIGINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS auctions_status_end_time_idx ON auctions (status, end_time);
CREATE INDEX IF NOT EXISTS auctions_seller_id_idx ON auctions (seller_id);
CREATE INDEX IF NOT EXISTS auctions_created_at_idx ON auctions (created_at DESC);

