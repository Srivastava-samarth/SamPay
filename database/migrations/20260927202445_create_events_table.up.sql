BEGIN;

CREATE TABLE events (
    id UUID PRIMARY KEY,
    event_id UUID NOT NULL UNIQUE,

    event_type VARCHAR(100) NOT NULL,
    direction VARCHAR(20) NOT NULL,

    payload JSONB NOT NULL,

    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    processed_at TIMESTAMPTZ,

    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT chk_events_direction
        CHECK (direction IN ('inbound', 'outbound')),

    CONSTRAINT chk_events_status
        CHECK (
            status IN (
                'pending',
                'processing',
                'processed',
                'failed'
            )
        )
);

CREATE INDEX idx_events_event_type_direction
    ON events(event_type, direction);

CREATE INDEX idx_events_status
    ON events(status);

CREATE INDEX idx_events_created_at
    ON events(created_at);

COMMIT;