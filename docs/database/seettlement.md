# Settlement Table

## Overview

The `settlement` table stores all settlement transactions executed by SamPay.

A Settlement represents the movement of funds from a Merchant's Pending Balance to their Available Balance after a successful incoming Payment.

The table captures the business state of the settlement, while all financial movements are recorded separately in the Ledger.

Business rules related to Settlements are documented in `docs/architecture/settlement.md`.

---

# Table Name

```text
settlement
```

---

# Columns

| Column | Type | Nullable | Default | Description |
|---------|------|----------|---------|-------------|
| id | UUID | No | gen_random_uuid() | Internal primary key |
| settlement_reference | VARCHAR(30) | No | - | Public settlement identifier |
| merchant_id | UUID | No | - | Merchant receiving the settlement |
| amount | BIGINT | No | - | Settlement amount (smallest currency unit) |
| currency | currency | No | 'INR' | Settlement currency |
| status | settlement_status | No | 'PENDING' | Current settlement status |
| is_notified | BOOLEAN | No | false | Indicates whether notification has been sent |
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
```

---

# Unique Constraints

```sql
UNIQUE (settlement_reference)
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
- Unique Constraint (`settlement_reference`)

Additional indexes should only be introduced when justified by production query patterns.

---

# Enums Used

## settlement_status

```text
PENDING
PROCESSING
COMPLETED
FAILED
```

---

## currency

```text
INR
```

---

# Referenced By

The following workflows reference Settlements:

- ledger
- notification

---

# Notes

- Settlement amounts are stored as `BIGINT` in the smallest currency unit (paise).
- Every Settlement belongs to exactly one Merchant.
- Settlements are system-generated and are not created directly by Users.
- A Settlement moves funds from the Merchant's Pending Balance to Available Balance.
- Financial movements are recorded in the Ledger.
- Notification delivery is tracked using `is_notified`.
- Every Settlement creates one or more corresponding Ledger entries as part of the same database transaction.