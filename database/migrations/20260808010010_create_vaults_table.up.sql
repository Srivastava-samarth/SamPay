BEGIN;

CREATE TABLE vaults (
    id UUID PRIMARY KEY,
    type VARCHAR(50) NOT NULL,
    balance NUMERIC(20, 2) NOT NULL DEFAULT 0,
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT uq_vaults_type
        UNIQUE (type),

    CONSTRAINT chk_vaults_type
        CHECK (type IN ('payment', 'payout', 'company')),

    CONSTRAINT chk_vaults_balance
        CHECK (balance >= 0)
);

COMMIT;