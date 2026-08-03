# Refund Table

## Overview

The `refund` table stores all refund transactions initiated by a Merchant.

A Refund represents returning funds from the Merchant's Wallet to the original payer.

The table captures the business state of the refund, while all financial movements are recorded separately in the Ledger.

Business rules related to Refunds are documented in `docs/architecture/refund.md`.

---

# Table Name

```text
refund
```

---

# Columns

| Column | Type | Nullable | Default | Description |
|---------|------|----------|---------|-------------|
| id | UUID | No | gen_random_uuid() | Internal primary key |
| refund_reference | VARCHAR(30) | No | - | Public refund identifier |
| payment_id | UUID | No | - | Original payment being refunded |
| merchant_id | UUID | No | - | Merchant initiating the refund |
| refund_type | refund_type | No | 'PAYMENT_REFUND' | Type of refund |
| amount | BIGINT | No | - | Refund amount (smallest currency unit) |
| currency | currency | No | 'INR' | Refund currency |
| status | refund_status | No | 'PENDING' | Current refund status |
| compliance_metadata | JSONB | Yes | NULL | Compliance response metadata |
| is_notified | BOOLEAN | No | false | Indicates whether notification has been sent |
| created_by | UUID | No | - | User who initiated the refund |
| created_at | TIMESTAMPTZ | No | now() | Record creation timestamp |
| updated_at | TIMESTAMPTZ | No | now() | Record last update timestamp |

---

# Primary Key

```sql
PRIMARY KEY (id)
```

---

# Foreign Keys

```sql
payment_id
REFERENCES payment(id)

merchant_id
REFERENCES merchant(id)

created_by
REFERENCES "user"(id)
```

---

# Unique Constraints

```sql
UNIQUE (refund_reference)
```

---

# Check Constraints

```sql
CHECK (amount > 0)
```

---

# Indexes

No additional indexes are defined.

---

# Enums Used

## refund_status

```text
PENDING
PROCESSING
COMPLETED
FAILED
```

---

## refund_type

```text
PAYMENT_REFUND
PARTIAL_REFUND
```

---

## currency

```text
INR
```

---

# Notes

- Refund amounts are stored as `BIGINT` in the smallest currency unit (paise).
- Every Refund references exactly one original Payment.
- A Payment may have multiple Refunds (partial refunds).
- Financial movements are recorded in the Ledger.
- Compliance decisions are stored in `compliance_metadata`.
- Notification delivery is tracked using `is_notified`.