import React, { useEffect, useMemo, useState } from 'react';
import ReactDOM from 'react-dom/client';
import './styles.css';
import { cancelAuction, createAuction, getAuction, listAuctions, login, placeBid, register, startAuction } from './services/api';
import type { Auction, AuthResult, BidResult, RealtimeEvent } from './types/api';

const tokenStorageKey = 'livebid.token';
const userStorageKey = 'livebid.user';

function money(cents: number) {
  return new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(cents / 100);
}

function secondsRemaining(endTime?: string) {
  if (!endTime) return 0;
  return Math.max(0, Math.ceil((new Date(endTime).getTime() - Date.now()) / 1000));
}

function App() {
  const [token, setToken] = useState(() => localStorage.getItem(tokenStorageKey) ?? '');
  const [user, setUser] = useState<AuthResult['user'] | null>(() => {
    const raw = localStorage.getItem(userStorageKey);
    return raw ? JSON.parse(raw) : null;
  });
  const [authMode, setAuthMode] = useState<'login' | 'register'>('login');
  const [auctions, setAuctions] = useState<Auction[]>([]);
  const [selectedAuction, setSelectedAuction] = useState<Auction | null>(null);
  const [status, setStatus] = useState('Loading auctions...');
  const [bidAmount, setBidAmount] = useState('');
  const [recentBids, setRecentBids] = useState<BidResult[]>([]);
  const [connectionState, setConnectionState] = useState('idle');
  const [tick, setTick] = useState(Date.now());

  useEffect(() => {
    const timer = window.setInterval(() => setTick(Date.now()), 1000);
    return () => window.clearInterval(timer);
  }, []);

  useEffect(() => {
    void refreshAuctions();
  }, []);

  useEffect(() => {
    if (!selectedAuction) return;
    setConnectionState('connecting');
    const socket = new WebSocket(import.meta.env.VITE_WS_BASE_URL ?? websocketURLFromAuction(selectedAuction.id));
    socket.onopen = () => setConnectionState('connected');
    socket.onclose = () => setConnectionState('disconnected');
    socket.onerror = () => setConnectionState('error');
    socket.onmessage = (message) => {
      const event = JSON.parse(message.data) as RealtimeEvent;
      if (event.type === 'bid_placed') {
        const result = event.data as BidResult;
        setSelectedAuction(result.auction);
        setRecentBids((items) => [result, ...items].slice(0, 10));
      }
      if (event.type === 'price_updated' || event.type === 'auction_extended') {
        void getAuction(selectedAuction.id).then(setSelectedAuction);
      }
    };
    return () => socket.close();
  }, [selectedAuction?.id]);

  async function refreshAuctions() {
    try {
      const result = await listAuctions();
      setAuctions(result.auctions);
      setStatus(result.auctions.length ? 'Auctions loaded.' : 'No auctions yet.');
    } catch (error) {
      setStatus(error instanceof Error ? error.message : 'Could not load auctions.');
    }
  }

  function applyAuth(result: AuthResult) {
    localStorage.setItem(tokenStorageKey, result.access_token);
    localStorage.setItem(userStorageKey, JSON.stringify(result.user));
    setToken(result.access_token);
    setUser(result.user);
  }

  async function submitAuth(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    const email = String(form.get('email') ?? '');
    const username = String(form.get('username') ?? '');
    const password = String(form.get('password') ?? '');
    try {
      const result = authMode === 'login' ? await login(email, password) : await register(email, username, password);
      applyAuth(result);
      setStatus(`Signed in as ${result.user.username}.`);
    } catch (error) {
      setStatus(error instanceof Error ? error.message : 'Authentication failed.');
    }
  }

  async function openAuction(auction: Auction) {
    const fresh = await getAuction(auction.id);
    setSelectedAuction(fresh);
    setRecentBids([]);
  }

  async function submitBid(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!token || !selectedAuction) {
      setStatus('Log in before bidding.');
      return;
    }
    try {
      const result = await placeBid(token, selectedAuction.id, Math.round(Number(bidAmount) * 100));
      setSelectedAuction(result.auction);
      setRecentBids((items) => [result, ...items].slice(0, 10));
      setBidAmount('');
      setStatus(result.extended ? 'Bid accepted and auction extended.' : 'Bid accepted.');
    } catch (error) {
      setStatus(error instanceof Error ? error.message : 'Bid failed.');
    }
  }

  async function submitAuction(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!token) {
      setStatus('Log in before creating an auction.');
      return;
    }
    const form = new FormData(event.currentTarget);
    try {
      const auction = await createAuction(token, {
        title: String(form.get('title') ?? ''),
        description: String(form.get('description') ?? ''),
        image_url: String(form.get('image_url') ?? '') || undefined,
        starting_price_cents: Math.round(Number(form.get('starting_price')) * 100),
        duration_seconds: Math.round(Number(form.get('duration_minutes')) * 60),
      });
      setAuctions((items) => [auction, ...items]);
      setSelectedAuction(auction);
      event.currentTarget.reset();
      setStatus('Auction draft created.');
    } catch (error) {
      setStatus(error instanceof Error ? error.message : 'Could not create auction.');
    }
  }

  async function sellerAction(action: 'start' | 'cancel') {
    if (!token || !selectedAuction) return;
    try {
      const updated = action === 'start' ? await startAuction(token, selectedAuction.id) : await cancelAuction(token, selectedAuction.id);
      setSelectedAuction(updated);
      setAuctions((items) => items.map((item) => (item.id === updated.id ? updated : item)));
      setStatus(action === 'start' ? 'Auction started.' : 'Auction cancelled.');
    } catch (error) {
      setStatus(error instanceof Error ? error.message : `Could not ${action} auction.`);
    }
  }

  const remaining = useMemo(() => secondsRemaining(selectedAuction?.end_time), [selectedAuction?.end_time, tick]);
  const nextBid = selectedAuction ? selectedAuction.current_price_cents + 100 : 0;

  return (
    <main className="min-h-screen bg-slate-50 text-ink">
      <header className="border-b border-slate-200 bg-white">
        <div className="mx-auto flex max-w-7xl items-center justify-between px-6 py-4">
          <button className="text-left text-2xl font-bold text-market" onClick={() => setSelectedAuction(null)}>
            LiveBid
          </button>
          <div className="text-sm text-slate-600">{user ? `${user.username}` : 'Guest buyer'}</div>
        </div>
      </header>

      <div className="mx-auto grid max-w-7xl gap-6 px-6 py-6 lg:grid-cols-[340px_1fr]">
        <aside className="space-y-5">
          <section className="rounded-lg border border-slate-200 bg-white p-5 shadow-sm">
            <div className="mb-4 flex gap-2">
              <button className={authMode === 'login' ? 'tab-active' : 'tab'} onClick={() => setAuthMode('login')}>
                Login
              </button>
              <button className={authMode === 'register' ? 'tab-active' : 'tab'} onClick={() => setAuthMode('register')}>
                Register
              </button>
            </div>
            <form className="space-y-3" onSubmit={submitAuth}>
              <input className="input" name="email" placeholder="email" type="email" required />
              {authMode === 'register' && <input className="input" name="username" placeholder="username" required />}
              <input className="input" name="password" placeholder="password" type="password" required />
              <button className="primary-button" type="submit">
                {authMode === 'login' ? 'Sign in' : 'Create account'}
              </button>
            </form>
          </section>

          <section className="rounded-lg border border-slate-200 bg-white p-5 shadow-sm">
            <h2 className="mb-3 font-semibold">Create auction</h2>
            <form className="space-y-3" onSubmit={submitAuction}>
              <input className="input" name="title" placeholder="title" required />
              <textarea className="textarea" name="description" placeholder="description" required />
              <input className="input" name="image_url" placeholder="image URL" />
              <div className="grid grid-cols-2 gap-3">
                <input className="input" min="0.01" name="starting_price" placeholder="starting $" step="0.01" type="number" required />
                <input className="input" min="1" name="duration_minutes" placeholder="minutes" type="number" required />
              </div>
              <button className="primary-button" type="submit">
                Save draft
              </button>
            </form>
          </section>

          <section className="rounded-lg border border-slate-200 bg-white p-5 shadow-sm">
            <div className="mb-3 flex items-center justify-between">
              <h2 className="font-semibold">Auctions</h2>
              <button className="secondary-button" onClick={refreshAuctions}>
                Refresh
              </button>
            </div>
            <div className="space-y-2">
              {auctions.map((auction) => (
                <button className="auction-row" key={auction.id} onClick={() => openAuction(auction)}>
                  <span className="font-medium">{auction.title}</span>
                  <span className="text-sm text-slate-500">
                    {auction.status} · {money(auction.current_price_cents)}
                  </span>
                </button>
              ))}
            </div>
          </section>
        </aside>

        <section className="min-h-[640px] rounded-lg border border-slate-200 bg-white p-6 shadow-sm">
          {!selectedAuction ? (
            <div className="flex h-full flex-col justify-center">
              <p className="text-sm font-semibold uppercase tracking-wide text-market">Auction discovery</p>
              <h1 className="mt-3 max-w-3xl text-4xl font-bold">Open an auction to watch live price movement.</h1>
              <p className="mt-4 max-w-2xl text-slate-600">{status}</p>
            </div>
          ) : (
            <div className="grid gap-6 lg:grid-cols-[1fr_320px]">
              <div>
                {selectedAuction.image_url && (
                  <img className="mb-5 aspect-video w-full rounded-lg object-cover" src={selectedAuction.image_url} alt="" />
                )}
                <div className="flex flex-wrap items-start justify-between gap-4">
                  <div>
                    <p className="text-sm font-semibold uppercase tracking-wide text-market">{selectedAuction.status}</p>
                    <h1 className="mt-2 text-4xl font-bold">{selectedAuction.title}</h1>
                    <p className="mt-3 max-w-3xl text-slate-600">{selectedAuction.description}</p>
                  </div>
                  <div className="rounded-lg bg-slate-100 px-4 py-3 text-right">
                    <p className="text-sm text-slate-500">Countdown</p>
                    <p className="text-3xl font-bold">{remaining}s</p>
                  </div>
                </div>

                <div className="mt-8 grid gap-4 sm:grid-cols-3">
                  <Metric label="Current price" value={money(selectedAuction.current_price_cents)} />
                  <Metric label="Next valid bid" value={money(nextBid)} />
                  <Metric label="Connection" value={connectionState} />
                </div>

                {user?.id === selectedAuction.seller_id && (
                  <div className="mt-5 flex flex-wrap gap-3">
                    <button className="secondary-button" onClick={() => sellerAction('start')}>
                      Start auction
                    </button>
                    <button className="secondary-button" onClick={() => sellerAction('cancel')}>
                      Cancel auction
                    </button>
                  </div>
                )}

                <form className="mt-8 flex gap-3" onSubmit={submitBid}>
                  <input
                    className="input"
                    min={nextBid / 100}
                    step="0.01"
                    placeholder={(nextBid / 100).toFixed(2)}
                    type="number"
                    value={bidAmount}
                    onChange={(event) => setBidAmount(event.target.value)}
                  />
                  <button className="primary-button max-w-36" type="submit">
                    Place bid
                  </button>
                </form>
                <p className="mt-3 text-sm text-slate-500">{status}</p>
              </div>

              <aside>
                <h2 className="mb-3 font-semibold">Recent bids</h2>
                <div className="space-y-2">
                  {recentBids.map((result) => (
                    <div className="rounded-lg border border-slate-200 p-3" key={result.bid.id}>
                      <div className="flex justify-between gap-3">
                        <span className="font-medium">{money(result.bid.amount_cents)}</span>
                        <span className="text-xs text-slate-500">{new Date(result.bid.created_at).toLocaleTimeString()}</span>
                      </div>
                      {result.extended && <p className="mt-1 text-sm text-signal">Auction extended</p>}
                    </div>
                  ))}
                </div>
              </aside>
            </div>
          )}
        </section>
      </div>
    </main>
  );
}

function Metric({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-lg border border-slate-200 bg-slate-50 p-4">
      <p className="text-sm text-slate-500">{label}</p>
      <p className="mt-1 text-xl font-semibold">{value}</p>
    </div>
  );
}

function websocketURLFromAuction(auctionId: string) {
  const apiBase = import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:8080/api/v1';
  return `${apiBase.replace(/^http/, 'ws')}/auctions/${auctionId}/ws`;
}

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>,
);
