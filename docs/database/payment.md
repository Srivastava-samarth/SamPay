# Payment Table

## Overview

The `payment` table stores all payment transactions initiated on SamPay.

A Payment represents the transfer of funds from a Merchant's Wallet to a Recipient. The table captures the business state of the payment, while all financial movements are recorded separately in the Ledger.

Business rules related to Payments are documented in `docs/architecture/payment.md`.

---

# Table Name

```text
payment
```

---

# Columns

| Column | Type | Nullable | Default | Description |
|---------|------|----------|---------|-------------|
| id | UUID | No | gen_random_uuid() | Internal primary key |
| payment_reference | VARCHAR(30) | No | - | Public payment identifier |
| merchant_id | UUID | No | - | Merchant initiating the payment |
| recipient_id | UUID | No | - | Recipient receiving the payment |
| payment_type | payment_type | No | 'WALLET_TO_WALLET' | Type of payment |
| amount | BIGINT | No | - | Payment amount (smallest currency unit) |
| currency | currency | No | 'INR' | Payment currency |
| status | payment_status | No | 'PENDING' | Current payment status |
| compliance_metadata | JSONB | Yes | NULL | Compliance response metadata |
| is_notified | BOOLEAN | No | false | Indicates whether notification has been sent |
| created_by | UUID | No | - | User who initiated the payment |
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
merchant_id
REFERENCES merchant(id)

recipient_id
REFERENCES recipient(id)

created_by
REFERENCES "user"(id)
```

---

# Unique Constraints

```sql
UNIQUE (payment_reference)
```

---

# Check Constraints

```sql
CHECK (amount > 0)
```

---

# Indexes

No additional indexes are defined.

The following indexes are automatically created by PostgreSQL:

- Primary Key (`id`)
- Unique Constraint (`payment_reference`)

Additional indexes should only be introduced when justified by production query patterns.

---

# Enums Used

## payment_status

```text
PENDING
PROCESSING
COMPLETED
FAILED
```

---

## payment_type

```text
WALLET_TO_WALLET
BANK_TO_WALLET
WALLET_TO_BANK
BANK_TO_BANK
```

---

## currency

```text
INR
```

---

# Referenced By

The following workflows reference Payments:

- settlement
- refund
- ledger
- notification

---

# Notes

- Payment amounts are stored as `BIGINT` in the smallest currency unit (paise).
- Every Payment belongs to exactly one Merchant.
- Every Payment is initiated by exactly one User.
- Every Payment is made to exactly one Recipient.
- The Payment table stores only the business state of the transaction.
- Financial movements are recorded in the Ledger.
- Compliance decisions are stored in `compliance_metadata`.
- Notification delivery is tracked using `is_notified`.
- Version 1 supports only `WALLET_TO_WALLET` payments. Other payment types are reserved for future platform capabilities.
- Every Payment creates one or more corresponding Ledger entries as part of the same database transaction.