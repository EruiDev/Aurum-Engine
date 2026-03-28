# Aurum — Payment Processing Engine

A production-ready payment processing backend written in Go. Aurum handles the full payment lifecycle (initiation → authorization → capture → settlement/void/refund) and guarantees reliable event delivery to downstream systems via the **Transactional Outbox Pattern**.

## Architecture

```
┌──────────────┐    ┌───────────────┐    ┌────────────────┐
│ HTTP Client  │--->│    Handler    │--->│    Service     │
└──────────────┘    └───────────────┘    └────────┬───────┘
                            │                     │
                    ┌───────▼────┐        ┌───────┴────────────┐
                    │  /metrics  │        │                    │
                    └──────┬─────┘        ▼                    ▼
                           │        ┌─────────────┐    ┌───────────────────┐
                    ┌──────▼─────┐  │  Payments   │    │   Outbox Events   │
                    │ Prometheus │  │  (Postgres) │    │   (Postgres)      │
                    └──────┬─────┘  └─────────────┘    └─────────┬─────────┘
                           │                                     │
                    ┌──────▼──────┐                     ┌────────▼────────┐
                    │   Grafana   │                     │  Outbox Worker  │
                    └─────────────┘                     └────────┬────────┘
                                                                 │
                                                          ┌──────▼───────┐
                                                          │    Kafka     │
                                                          └──────────────┘
```

**Packages:**

| Package | Responsibility |
|---|---|
| `cmd/api` | Entry point, HTTP server setup, graceful shutdown |
| `internal/domain` | Payment model, state machine, validation |
| `internal/service` | Business logic, transaction orchestration, idempotency |
| `internal/handler` | HTTP request parsing, response formatting |
| `internal/repository` | PostgreSQL queries for payments and outbox events |
| `internal/publisher` | Kafka event publishing |
| `internal/worker` | Background outbox poller — publishes pending events |
| `internal/db` | Connection pool, migrations, transaction helper |
| `internal/metrics` | Prometheus counters and histograms |
| `internal/middleware` | HTTP middleware for request duration tracking |

## Payment State Machine

```
                    ┌──────────────┐
          ┌────────>│  AUTHORIZED  │────────┐
          │         └──────────────┘        │
          │                                 ▼
┌─────────┴──┐                      ┌──────────────┐
│  INITIATED │                      │   CAPTURED   │────────┐
└─────────┬──┘                      └──────┬───────┘        │
          │                                │                ▼
          │                                ▼          ┌──────────────┐
          │                         ┌──────────────┐  │   REFUNDED   │
          │                         │   SETTLED    │─>│  (terminal)  │
          │                         └──────────────┘  └──────────────┘
          │
          ▼
     ┌──────────┐     ┌──────────┐     ┌──────────┐
     │  FAILED  │     │  VOIDED  │     │ REFUNDED │
     │(terminal)│     │(terminal)│     │(terminal)│
     └──────────┘     └──────────┘     └──────────┘
```

## API

### Create Payment
```
POST /payments
Idempotency-Key: <unique-key>
Content-Type: application/json

{
  "amount": 1000,
  "currency": "USD",
  "merchant_id": "<uuid>",
  "customer_id": "<uuid>"
}
```
- `amount` is in the smallest currency unit (e.g. cents)
- `currency` must be a valid ISO 4217 code
- Repeated requests with the same `Idempotency-Key` return the original payment

**Responses:** `201 Created` · `400 Bad Request` · `422 Unprocessable Entity`

---

### Get Payment
```
GET /payments/{id}
```
**Responses:** `200 OK` · `400 Bad Request` (invalid UUID) · `404 Not Found`

---

### Transition Payment
```
POST /payments/{id}/{action}
```
Valid actions: `authorize`, `capture`, `void`

**Responses:** `200 OK` · `404 Not Found` · `422 Unprocessable Entity` (invalid transition)

---

### Health Check
```
GET /health
```
Returns `{"status":"ok","db":"ok"}` or `{"status":"degraded","db":"unreachable"}` with `503`.

---

### Metrics
```
GET /metrics
```
Prometheus-formatted metrics.

## Getting Started

### Prerequisites
- Docker & Docker Compose

### Run with Docker Compose

```bash
docker compose up
```

This starts:
| Service                     | Port                   |
|-----------------------------|------------------------|
| API                         | `8080`                 |
| PostgreSQL                  | `5432`                 |
| Redpanda (Kafka-compatible) | `19092` (external)     |
| Prometheus                  | `9090`                 |
| Grafana                     | `3000` (admin / admin) |

### Run locally

1. Copy the environment file and set your database URL:
```bash
cp .env.example .env
# Edit DATABASE_URL and KAFKA_BROKERS
```

2. Start dependencies:
```bash
docker compose up postgres redpanda -d
```

3. Run the API:
```bash
go run ./cmd/api
```

### Environment Variables

| Variable        | Default | Description                            |
|-----------------|---------|----------------------------------------|
| `DATABASE_URL`  |    —    | PostgreSQL connection string           |
| `KAFKA_BROKERS` |    —    | Comma-separated Kafka broker addresses |
| `PORT`          |  `8080` | HTTP server port                       |

## Observability

Prometheus metrics exposed at `/metrics`:

| Metric                          | Type      | Description                                   |
|---------------------------------|-----------|-----------------------------------------------|
| `payments_created_total`        | Counter   | Payments created, labelled by `currency`      |
| `payment_transition_total`      | Counter   | State transitions, labelled by `from`/`to`    |
| `outbox_pending_events`         | Gauge     | Unpublished outbox events                     |
| `http_request_duration_seconds` | Histogram | Request latency by `method`, `path`, `status` |
| `db_query_duration_seconds`     | Histogram | DB query latency by `operation`               |

## Tech Stack

- **Go** 1.25
- **PostgreSQL** 16 — payments and outbox event storage
- **Kafka** (Redpanda) — event streaming
- **Prometheus** + **Grafana** — metrics and dashboards
