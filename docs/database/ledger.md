# Ledger Table

## Overview

The `ledger` table stores the immutable financial records for every movement of money within SamPay.

Every financial transaction performed by the platform must create one or more Ledger entries. The Ledger serves as the financial source of truth and follows the principles of double-entry accounting.

Business rules related to the Ledger are documented in `docs/architecture/ledger.md`.

---

# Table Name

```text
ledger
```

---

# Columns

| Column | Type | Nullable | Default | Description |
|---------|------|----------|---------|-------------|
| id | UUID | No | gen_random_uuid() | Internal primary key |
| debit_account_id | UUID | No | - | Ledger account being debited |
| credit_account_id | UUID | No | - | Ledger account being credited |
| amount | BIGINT | No | - | Amount transferred (smallest currency unit) |
| currency | currency | No | 'INR' | Transaction currency |
| transaction_type | ledger_transaction_type | No | - | PAYMENT / PAYOUT / REFUND / SETTLEMENT / FEE |
| transaction_id | UUID | No | - | Business transaction identifier |
| created_at | TIMESTAMPTZ | No | now() | Record creation timestamp |

---

# Primary Key

```sql
PRIMARY KEY (id)
```

---

# Foreign Keys

```sql
debit_account_id
REFERENCES ledger_account(id)

credit_account_id
REFERENCES ledger_account(id)
```

---

# Unique Constraints

None.

Ledger is append-only and multiple entries may belong to the same business transaction.

---

# Check Constraints

```sql
CHECK (amount > 0)
```

---

# Indexes

No additional indexes are defined.

The Primary Key index is automatically created by PostgreSQL.

Additional indexes should only be introduced when justified by production query patterns.

---

# Enums Used

## ledger_transaction_type

```text
PAYMENT
PAYOUT
REFUND
SETTLEMENT
FEE
```

---

## currency

```text
INR
```

---

# Referenced By

Ledger entries are automatically created by the following workflows:

- Payment
- Payout
- Refund
- Settlement

---

# Notes

- Ledger entries are immutable.
- Ledger is append-only.
- Ledger entries can never be updated or deleted.
- Money is stored in the smallest currency unit (`BIGINT`).
- Every financial operation creates one or more Ledger entries.
- Every Ledger entry is created within the same database transaction as the corresponding Wallet and Vault updates.
- `transaction_id` references the business transaction that created the Ledger entry. Its actual table is determined by `transaction_type`.
- The Ledger represents the financial history of the platform, while Wallets and Vaults represent the current balances.