CREATE TABLE outbox_events (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    aggregate_id UUID NOT NULL,                -- Payment ID
    event_type   VARCHAR(100) NOT NULL,        -- "payment.initiated", "payment.captured"
    payload      JSONB NOT NULL,               -- copy of event
    published    BOOLEAN NOT NULL DEFAULT FALSE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    published_at TIMESTAMPTZ                   -- null until published
);

CREATE INDEX idx_outbox_unpublished ON outbox_events (created_at)
WHERE published = false;
```