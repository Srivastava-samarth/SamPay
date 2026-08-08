BEGIN;

CREATE TABLE linked_bank_accounts (
    id UUID PRIMARY KEY,
    merchant_id UUID NOT NULL,
    bank_account_id UUID NOT NULL,
    type VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_linked_bank_accounts_merchant
        FOREIGN KEY (merchant_id)
        REFERENCES merchants(id),

    CONSTRAINT fk_linked_bank_accounts_bank_account
        FOREIGN KEY (bank_account_id)
        REFERENCES bank_accounts(id),

    CONSTRAINT uq_linked_bank_accounts_merchant_bank_account
        UNIQUE (merchant_id, bank_account_id),

    CONSTRAINT chk_linked_bank_accounts_type
        CHECK (type IN ('primary', 'secondary'))
);

CREATE INDEX idx_linked_bank_accounts_bank_account_id
    ON linked_bank_accounts(bank_account_id);

CREATE UNIQUE INDEX idx_one_primary_bank_account_per_merchant
    ON linked_bank_accounts(merchant_id)
    WHERE type = 'primary';

COMMIT;