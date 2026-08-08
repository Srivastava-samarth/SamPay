BEGIN;

CREATE TABLE idempotency_keys (
    id UUID PRIMARY KEY,
    merchant_id UUID NOT NULL,
    idempotency_key VARCHAR(255) NOT NULL,
    request_hash VARCHAR(255) NOT NULL,
    response_status INTEGER,
    response_body JSONB,
    status VARCHAR(50) NOT NULL DEFAULT 'in_progress',
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_idempotency_keys_merchant
        FOREIGN KEY (merchant_id)
        REFERENCES merchants(id),

    CONSTRAINT uq_idempotency_keys_merchant_key
        UNIQUE (merchant_id, idempotency_key),

    CONSTRAINT chk_idempotency_keys_status
        CHECK (status IN ('in_progress', 'completed', 'failed'))
);

CREATE INDEX idx_idempotency_keys_expires_at
    ON idempotency_keys(expires_at);

COMMIT;