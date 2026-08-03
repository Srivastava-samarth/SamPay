# Payout Table

## Overview

The `payout` table stores all payout transactions initiated by a Merchant.

A Payout represents the transfer of funds from a Merchant's Wallet to a linked Bank Account.

The table captures the business state of the payout, while all financial movements are recorded separately in the Ledger.

Business rules related to Payouts are documented in `docs/architecture/payout.md`.

---

# Table Name

```text
payout
```

---

# Columns

| Column | Type | Nullable | Default | Description |
|---------|------|----------|---------|-------------|
| id | UUID | No | gen_random_uuid() | Internal primary key |
| payout_reference | VARCHAR(30) | No | - | Public payout identifier |
| merchant_id | UUID | No | - | Merchant initiating the payout |
| linked_bank_account_id | UUID | No | - | Destination bank account |
| payout_type | payout_type | No | 'WALLET_TO_BANK' | Type of payout |
| amount | BIGINT | No | - | Payout amount (smallest currency unit) |
| currency | currency | No | 'INR' | Payout currency |
| status | payout_status | No | 'PENDING' | Current payout status |
| compliance_metadata | JSONB | Yes | NULL | Compliance response metadata |
| is_notified | BOOLEAN | No | false | Indicates whether notification has been sent |
| created_by | UUID | No | - | User who initiated the payout |
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

linked_bank_account_id
REFERENCES linked_bank_account(id)

created_by
REFERENCES "user"(id)
```

---

# Unique Constraints

```sql
UNIQUE (payout_reference)
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

## payout_status

```text
PENDING
PROCESSING
COMPLETED
FAILED
```

---

## payout_type

```text
WALLET_TO_BANK
```

---

## currency

```text
INR
```

---

# Notes

- Payout amounts are stored as `BIGINT` in the smallest currency unit (paise).
- Every Payout belongs to exactly one Merchant.
- Every Payout transfers funds to one Linked Bank Account.
- Financial movements are recorded in the Ledger.
- Compliance decisions are stored in `compliance_metadata`.
- Notification delivery is tracked using `is_notified`.