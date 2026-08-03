# Vault Table

## Overview

The `vault` table stores SamPay's internal platform-owned financial accounts.

Vaults temporarily hold funds during financial workflows such as Payments, Payouts, Refunds and Company operations.

Business rules related to Vaults are documented in `docs/architecture/vault.md`.

---

# Table Name

```text
vault
```

---

# Columns

| Column | Type | Nullable | Default | Description |
|---------|------|----------|---------|-------------|
| id | UUID | No | gen_random_uuid() | Internal primary key |
| vault_reference | VARCHAR(30) | No | - | Public vault identifier |
| vault_type | vault_type | No | - | PAYMENT / PAYOUT / REFUND / COMPANY |
| currency | currency | No | 'INR' | Supported currency |
| balance | BIGINT | No | 0 | Current vault balance (smallest currency unit) |
| status | vault_status | No | 'ACTIVE' | Vault status |
| created_at | TIMESTAMPTZ | No | now() | Record creation timestamp |
| updated_at | TIMESTAMPTZ | No | now() | Record last update timestamp |

---

# Primary Key

```sql
PRIMARY KEY (id)
```

---

# Foreign Keys

None.

Vaults are platform-owned resources.

---

# Unique Constraints

```sql
UNIQUE (vault_reference)

UNIQUE (vault_type)
```

Since Version 1 contains exactly one Vault for each Vault Type.

---

# Check Constraints

```sql
CHECK (balance >= 0)
```

Vault balances can never become negative.

---

# Indexes

| Index | Purpose |
|---------|----------|
| vault_type | Fast vault lookup |
| status | Operational queries |
| created_at | Reporting and ordering |

---

# Enums Used

## vault_type

```text
PAYMENT
PAYOUT
REFUND
COMPANY
```

---

## vault_status

```text
ACTIVE
INACTIVE
```

---

## currency

```text
INR
```

---

# Referenced By

The following tables interact with Vaults:

- payment
- payout
- refund
- settlement
- ledger

---

# Notes

- Vaults are owned exclusively by SamPay.
- Version 1 maintains exactly four Vaults:
  - PAYMENT
  - PAYOUT
  - REFUND
  - COMPANY
- Vault balances are stored as `BIGINT` in the smallest currency unit.
- Vault balances are updated only through internal financial workflows.
- Vault balance modifications are performed using row-level locking.
- Every Vault balance modification must create a corresponding Ledger entry within the same database transaction.