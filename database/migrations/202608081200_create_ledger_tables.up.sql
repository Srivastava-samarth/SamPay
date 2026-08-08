BEGIN;

CREATE TABLE ledger_transactions (
    id UUID PRIMARY KEY,
    transaction_ref VARCHAR(255) UNIQUE NOT NULL,
    type VARCHAR(50) NOT NULL,
    reference_id UUID NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'posted',
    settlement_status VARCHAR(50) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_ledger_transactions_reference
    ON ledger_transactions(type, reference_id);

CREATE INDEX idx_ledger_transactions_settlement_status
    ON ledger_transactions(settlement_status);

CREATE TABLE ledger_entries (
    id UUID PRIMARY KEY,
    ledger_transaction_id UUID NOT NULL,
    account_type VARCHAR(50) NOT NULL,
    account_id UUID NOT NULL,
    entry_type VARCHAR(20) NOT NULL,
    amount NUMERIC(20, 2) NOT NULL,
    currency VARCHAR(10) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_ledger_entries_transaction
        FOREIGN KEY (ledger_transaction_id)
        REFERENCES ledger_transactions(id)
        ON DELETE RESTRICT,

    CONSTRAINT chk_ledger_entries_type
        CHECK (entry_type IN ('debit', 'credit')),

    CONSTRAINT chk_ledger_entries_amount
        CHECK (amount > 0),

    CONSTRAINT chk_ledger_entries_account_type
        CHECK (
            account_type IN (
                'wallet',
                'bank_account',
                'vault'
            )
        )
);

CREATE INDEX idx_ledger_entries_transaction_id
    ON ledger_entries(ledger_transaction_id);

CREATE INDEX idx_ledger_entries_account
    ON ledger_entries(account_type, account_id);

COMMIT;