import React from 'react';
import ReactDOM from 'react-dom/client';
import './styles.css';

function App() {
  return (
    <main className="min-h-screen bg-slate-50 text-ink">
      <section className="mx-auto flex min-h-screen w-full max-w-6xl flex-col justify-center px-6 py-12">
        <p className="text-sm font-semibold uppercase tracking-wide text-market">LiveBid</p>
        <h1 className="mt-4 max-w-3xl text-5xl font-bold leading-tight">
          Real-time auctions built for correctness under pressure.
        </h1>
        <p className="mt-5 max-w-2xl text-lg text-slate-600">
          The MVP is being built incrementally: durable auction state, atomic bidding,
          WebSocket updates, anti-sniping extensions, and observable request flows.
        </p>
        <div className="mt-8 grid gap-4 sm:grid-cols-3">
          {['Atomic bids', 'Live updates', 'Fair endings'].map((label) => (
            <div key={label} className="rounded-lg border border-slate-200 bg-white p-5 shadow-sm">
              <p className="font-semibold">{label}</p>
            </div>
          ))}
        </div>
      </section>
    </main>
  );
}

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>,
);

