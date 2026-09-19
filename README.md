# LiveBid

LiveBid is a production-oriented real-time auction marketplace built to demonstrate full-stack ownership, concurrent bid processing, WebSocket updates, durable state management, observability, and practical distributed systems tradeoffs.

This repository is being developed incrementally. The first milestone establishes the monorepo structure, local development conventions, and infrastructure foundation before feature implementation.

## Planned Stack

- Frontend: React, TypeScript, Vite, Tailwind CSS
- Backend: Go, Chi or Gin, REST APIs, WebSockets
- Data: PostgreSQL as durable source of truth
- Real-time state: Redis for pub/sub, cache, coordination, and rate limiting
- Infrastructure: Docker Compose
- Observability: structured logging, Prometheus-compatible metrics, OpenTelemetry traces
- Testing: Go unit/integration/concurrency tests, frontend tests where useful, load tests

## Repository Structure

```text
backend/
  cmd/api/
  internal/
    auth/
    auction/
    bidding/
    middleware/
    observability/
    repository/
    service/
    users/
    websocket/
  migrations/
  tests/
frontend/
  src/
    components/
    context/
    hooks/
    pages/
    services/
    types/
loadtest/
docs/
```

## Current Status

Foundation commit only. Feature work will be added through meaningful incremental commits.

