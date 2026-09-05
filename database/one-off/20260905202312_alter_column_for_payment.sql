BEGIN;

-- This migration assumes the payments table is still empty.
-- Remove this check if you already have a data-migration strategy.
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM payments LIMIT 1) THEN
        RAISE EXCEPTION
            'payments table contains data; run a data migration before applying this change';
    END IF;
END $$;

-- Drop the old merchant foreign key and index first.
ALTER TABLE payments
DROP CONSTRAINT IF EXISTS fk_payments_merchant;

DROP INDEX IF EXISTS idx_payments_merchant_id;

-- Replace merchant_id with sender/receiver merchant IDs.
ALTER TABLE payments
DROP COLUMN IF EXISTS merchant_id;

ALTER TABLE payments
ADD COLUMN sender_merchant_id UUID NOT NULL,
ADD COLUMN receiver_merchant_id UUID NOT NULL;

-- Add settlement status.
ALTER TABLE payments
ADD COLUMN settlement_status VARCHAR(50) NOT NULL DEFAULT 'pending';

-- Foreign keys.
ALTER TABLE payments
ADD CONSTRAINT fk_payments_sender_merchant
    FOREIGN KEY (sender_merchant_id)
    REFERENCES merchants(id),

ADD CONSTRAINT fk_payments_receiver_merchant
    FOREIGN KEY (receiver_merchant_id)
    REFERENCES merchants(id);

-- Prevent a merchant from paying itself.
ALTER TABLE payments
ADD CONSTRAINT chk_payments_different_merchants
    CHECK (sender_merchant_id <> receiver_merchant_id);

-- Preserve your amount validation.
-- The existing constraint remains, so no need to recreate it.

-- Indexes for merchant-scoped payment queries.
CREATE INDEX idx_payments_sender_merchant_id
    ON payments(sender_merchant_id);

CREATE INDEX idx_payments_receiver_merchant_id
    ON payments(receiver_merchant_id);

COMMIT;