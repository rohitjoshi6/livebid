import type { Auction, AuctionList, AuthResult, BidResult, User } from '../types/api';

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:8080/api/v1';

type RequestOptions = {
  token?: string;
  method?: string;
  body?: unknown;
};

export class ApiError extends Error {
  status: number;
  code: string;

  constructor(status: number, code: string, message: string) {
    super(message);
    this.status = status;
    this.code = code;
  }
}

async function request<T>(path: string, options: RequestOptions = {}): Promise<T> {
  const response = await fetch(`${API_BASE_URL}${path}`, {
    method: options.method ?? 'GET',
    headers: {
      'Content-Type': 'application/json',
      ...(options.token ? { Authorization: `Bearer ${options.token}` } : {}),
    },
    body: options.body ? JSON.stringify(options.body) : undefined,
  });

  const payload = await response.json().catch(() => null);
  if (!response.ok) {
    const error = payload?.error;
    throw new ApiError(response.status, error?.code ?? 'request_failed', error?.message ?? 'Request failed');
  }
  return payload as T;
}

export function register(email: string, username: string, password: string) {
  return request<AuthResult>('/auth/register', {
    method: 'POST',
    body: { email, username, password },
  });
}

export function login(email: string, password: string) {
  return request<AuthResult>('/auth/login', {
    method: 'POST',
    body: { email, password },
  });
}

export function me(token: string) {
  return request<User>('/auth/me', { token });
}

export function listAuctions(status?: string) {
  const query = status ? `?status=${encodeURIComponent(status)}` : '';
  return request<AuctionList>(`/auctions/${query}`);
}

export function getAuction(id: string) {
  return request<Auction>(`/auctions/${id}/`);
}

export function placeBid(token: string, auctionId: string, amountCents: number) {
  return request<BidResult>(`/auctions/${auctionId}/bids`, {
    token,
    method: 'POST',
    body: {
      amount_cents: amountCents,
      idempotency_key: crypto.randomUUID(),
    },
  });
}

export function websocketURL(auctionId: string) {
  const base = API_BASE_URL.replace(/^http/, 'ws').replace(/\/api\/v1$/, '');
  return `${base}/api/v1/auctions/${auctionId}/ws`;
}

