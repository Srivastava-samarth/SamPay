# Idempotency Key Table

## Overview

The `idempotency_key` table stores idempotency records for write operations performed on the SamPay platform.

It ensures that retrying the same request with the same idempotency key does not result in duplicate resource creation or duplicate financial transactions.

The original API response is stored and returned directly for subsequent requests using the same idempotency key.

---

# Table Name

```text
idempotency_key
```

---

# Columns

| Column | Type | Nullable | Default | Description |
|---------|------|----------|---------|-------------|
| id | UUID | No | gen_random_uuid() | Internal primary key |
| merchant_id | UUID | No | - | Merchant making the request |
| endpoint | VARCHAR(100) | No | - | API endpoint (e.g. `/payments`) |
| idempotency_key | VARCHAR(255) | No | - | Client supplied idempotency key |
| response_data | JSONB | No | - | Complete API response returned to the client |
| expires_at | TIMESTAMPTZ | No | - | Expiration timestamp |
| created_at | TIMESTAMPTZ | No | now() | Record creation timestamp |

---

# Primary Key

```sql
PRIMARY KEY (id)
```

---

# Foreign Keys

```sql
merchant_id
REFERENCES merchant(id)
```

---

# Unique Constraints

```sql
UNIQUE (merchant_id, endpoint, idempotency_key)
```

---

# Check Constraints

None.

---

# Indexes

No additional indexes are defined.

The following indexes are automatically created by PostgreSQL:

- Primary Key (`id`)
- Unique Constraint (`merchant_id`, `endpoint`, `idempotency_key`)

Additional indexes should only be introduced when justified by production query patterns.

---

# Referenced By

The following APIs use the Idempotency Key table:

- Payment API
- Payout API
- Refund API
- Merchant API
- User API
- Linked Bank Account API

---

# Notes

- Every idempotent request must provide an `Idempotency-Key` header.
- The combination of `(merchant_id, endpoint, idempotency_key)` must be unique.
- The complete API response is stored in `response_data`.
- If a request is retried with the same idempotency key, the stored response is returned without executing the business operation again.
- Idempotency records expire after a configurable retention period and may be cleaned up by a background job.