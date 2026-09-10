CREATE TABLE refunds (
    id UUID PRIMARY KEY,

    payment_id UUID NOT NULL,
    merchant_id UUID NOT NULL,

    refund_reference VARCHAR(255) NOT NULL UNIQUE,
    external_reference VARCHAR(255),

    amount NUMERIC(30, 18) NOT NULL,
    currency VARCHAR(20) NOT NULL,

    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    reason TEXT,

    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_refunds_payment
        FOREIGN KEY (payment_id)
        REFERENCES payments(id),

    CONSTRAINT fk_refunds_merchant
        FOREIGN KEY (merchant_id)
        REFERENCES merchants(id)
);

CREATE INDEX idx_refunds_payment_id
    ON refunds(payment_id);

CREATE INDEX idx_refunds_merchant_id
    ON refunds(merchant_id);