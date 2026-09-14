BEGIN;

CREATE TABLE payouts (
    id UUID PRIMARY KEY,
    merchant_id UUID NOT NULL,
    source_wallet_id UUID,
    source_bank_account_id UUID,
    destination_bank_account_id  UUID NOT NULL,
    payout_reference VARCHAR(255) UNIQUE NOT NULL,
    external_reference VARCHAR(255),
    amount NUMERIC(20, 2) NOT NULL,
    currency VARCHAR(10) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_payouts_merchant
        FOREIGN KEY (merchant_id)
        REFERENCES merchants(id),

    CONSTRAINT fk_payouts_source_wallet
        FOREIGN KEY (source_wallet_id)
        REFERENCES wallets(id),

    CONSTRAINT fk_payouts_source_bank_account
        FOREIGN KEY (source_bank_account_id)
        REFERENCES bank_accounts(id),

    CONSTRAINT fk_payouts_destination_bank_account
    FOREIGN KEY (destination_bank_account_id)
    REFERENCES bank_account(id),

    CONSTRAINT chk_payouts_amount
        CHECK (amount > 0),

    CONSTRAINT chk_payouts_source
        CHECK (
            (
                source_wallet_id IS NOT NULL
                AND source_bank_account_id IS NULL
            )
            OR
            (
                source_wallet_id IS NULL
                AND source_bank_account_id IS NOT NULL
            )
        )
);

CREATE INDEX idx_payouts_merchant_id
    ON payouts(merchant_id);

CREATE INDEX idx_payouts_source_wallet_id
    ON payouts(source_wallet_id);

CREATE INDEX idx_payouts_source_bank_account_id
    ON payouts(source_bank_account_id);

CREATE INDEX idx_payouts_destination_linked_bank_account_id
    ON payouts(destination_linked_bank_account_id);

COMMIT;