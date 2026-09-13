BEGIN;

ALTER TABLE merchants
    ADD COLUMN compliance_details JSONB,
    ADD COLUMN country VARCHAR(50);

COMMIT;