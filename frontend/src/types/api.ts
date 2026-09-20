export type User = {
  id: string;
  email: string;
  username: string;
  created_at: string;
};

export type Auction = {
  id: string;
  seller_id: string;
  title: string;
  description: string;
  image_url?: string;
  starting_price_cents: number;
  current_price_cents: number;
  status: 'draft' | 'scheduled' | 'live' | 'completed' | 'cancelled';
  duration_seconds: number;
  start_time?: string;
  end_time?: string;
  winner_id?: string;
  version: number;
  created_at: string;
  updated_at: string;
};

export type Bid = {
  id: string;
  auction_id: string;
  bidder_id: string;
  amount_cents: number;
  idempotency_key?: string;
  created_at: string;
};

export type AuthResult = {
  access_token: string;
  user: User;
};

export type AuctionList = {
  auctions: Auction[];
  page: number;
  limit: number;
};

export type BidResult = {
  bid: Bid;
  auction: Auction;
  extended: boolean;
  previous_end_time?: string;
};

export type RealtimeEvent<T = unknown> = {
  type: string;
  auction_id: string;
  data: T;
  sent_at: string;
};

