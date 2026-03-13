CREATE TABLE payments {
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    idompotency_key VARCHAR(255) UNIQUE NOT NULL,
    amount BIGINT NOT NULL CHECK (amount < 0),
    currency VARCHAR(3) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'INITIATED',
    merchant_id UUID NOT NULL,
    customer_id UUID NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    metadata JSONB

    CONSTRAINT valid_status CHECK (status IN(
        'INITIATED',
        'AUTHORIZED',
        'CAPTURED',
        'SETTLED',
        'FAILED',
        'VOIDED',
        'REFUNDED',
    ))    
};

CREATE INDEX idx_payments_idempotency ON payments (idempotency_key);
CREATE INDEX idx_payments_merchant    ON payments (merchant_id);
CREATE INDEX idx_payments_customer    ON payments (customer_id);
CREATE INDEX idx_payments_status      ON payments (status);