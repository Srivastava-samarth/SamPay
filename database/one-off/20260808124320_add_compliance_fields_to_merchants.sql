BEGIN;

ALTER TABLE merchants
    ADD COLUMN compliance_status VARCHAR(50) NOT NULL DEFAULT 'pending',
    ADD COLUMN compliance_date TIMESTAMPTZ,
    ADD COLUMN compliance_reason TEXT;

COMMIT;