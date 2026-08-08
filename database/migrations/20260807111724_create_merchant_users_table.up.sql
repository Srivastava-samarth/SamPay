BEGIN;

CREATE TABLE merchant_users (
    id UUID PRIMARY KEY,
    merchant_id UUID NOT NULL,
    user_id UUID NOT NULL,
    role VARCHAR(50) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_merchant_users_merchant
        FOREIGN KEY (merchant_id)
        REFERENCES merchants(id),

    CONSTRAINT fk_merchant_users_user
        FOREIGN KEY (user_id)
        REFERENCES users(id),

    CONSTRAINT uq_merchant_users_merchant_user
        UNIQUE (merchant_id, user_id)
);

CREATE INDEX idx_merchant_users_user_id
    ON merchant_users(user_id);

COMMIT;