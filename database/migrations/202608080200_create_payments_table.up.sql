BEGIN;

CREATE TABLE payments (
    id UUID PRIMARY KEY,
    merchant_id UUID NOT NULL,
    payment_reference VARCHAR(255) UNIQUE NOT NULL,
    amount NUMERIC(20, 2) NOT NULL,
    currency VARCHAR(10) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    description TEXT,
    customer_reference VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_payments_merchant
        FOREIGN KEY (merchant_id)
        REFERENCES merchants(id),

    CONSTRAINT chk_payments_amount
        CHECK (amount > 0)
);

CREATE INDEX idx_payments_merchant_id
    ON payments(merchant_id);

COMMIT;