BEGIN;

CREATE TABLE wallets (
    id UUID PRIMARY KEY,
    merchant_id UUID UNIQUE NOT NULL,
    available_balance NUMERIC(20, 2) NOT NULL DEFAULT 0,
    reserved_balance NUMERIC(20, 2) NOT NULL DEFAULT 0,
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_wallets_merchant
        FOREIGN KEY (merchant_id)
        REFERENCES merchants(id),

    CONSTRAINT chk_wallets_available_balance
        CHECK (available_balance >= 0),

    CONSTRAINT chk_wallets_reserved_balance
        CHECK (reserved_balance >= 0)
);

COMMIT;