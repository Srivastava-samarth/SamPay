# Wallet Table

## Overview

The `wallet` table stores the current financial balances for every Merchant.

Each Merchant owns exactly one Wallet.

The Wallet maintains the Merchant's Available, Reserved and Pending balances.

Business rules related to Wallets are documented in `docs/architecture/wallet.md`.

---

# Table Name

```text
wallet
```

---

# Columns

| Column | Type | Nullable | Default | Description |
|---------|------|----------|---------|-------------|
| id | UUID | No | gen_random_uuid() | Internal primary key |
| wallet_reference | VARCHAR(30) | No | - | Public wallet identifier |
| merchant_id | UUID | No | - | Wallet owner |
| currency | currency | No | 'INR' | Wallet currency |
| available_balance | BIGINT | No | 0 | Spendable balance |
| reserved_balance | BIGINT | No | 0 | Funds reserved for outgoing transactions |
| pending_balance | BIGINT | No | 0 | Incoming funds awaiting settlement |
| status | wallet_status | No | 'ACTIVE' | Wallet status |
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

ON DELETE RESTRICT

---

# Unique Constraints

```sql
UNIQUE (wallet_reference)

UNIQUE (merchant_id)
```

Every Merchant owns exactly one Wallet.

---

# Check Constraints

```sql
CHECK (available_balance >= 0)

CHECK (reserved_balance >= 0)

CHECK (pending_balance >= 0)
```

Wallet balances can never become negative.

---

# Indexes

| Index | Purpose |
|---------|----------|
| merchant_id | Fetch Merchant Wallet |
| status | Operational queries |
| created_at | Reporting and ordering |

---

# Enums Used

## wallet_status

```text
ACTIVE
SUSPENDED
CLOSED
```

---

## currency

```text
INR
```

---

# Referenced By

The following tables interact with Wallets:

- payment
- payout
- refund
- settlement
- ledger

---

# Notes

- Every Merchant owns exactly one Wallet.
- Wallet balances are stored as `BIGINT` in the smallest currency unit (paise).
- Version 1 supports INR only.
- Wallet balances are modified only through financial workflows.
- Direct balance updates are not permitted.
- Wallet updates must use row-level locking to prevent concurrent modifications and double spending.
- Every Wallet balance modification must create the corresponding Ledger entry within the same database transaction.
- `available_balance` represents spendable funds.
- `reserved_balance` represents funds locked for in-flight outgoing transactions.
- `pending_balance` represents incoming funds awaiting settlement.