	BEGIN;
    
    CREATE TABLE merchants (
        id UUID PRIMARY KEY,
        merchant_reference VARCHAR(255) UNIQUE NOT NULL,
        merchant_name VARCHAR(255) NOT NULL,
        email VARCHAR(255) UNIQUE NOT NULL,
        phone_number VARCHAR(20),
        status VARCHAR(50) DEFAULT 'pending' NOT NULL,
        created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
    );

    COMMIT;