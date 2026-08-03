# Linked Bank Account Table

## Overview

The `linked_bank_account` table stores external bank accounts linked to a Merchant.

Linked Bank Accounts are used for:

- Automatic Wallet funding.
- Receiving Payouts.

Business rules related to Linked Bank Accounts are documented in `docs/architecture/linked-bank-account.md`.

---

# Table Name

```text
linked_bank_account
```

---

# Columns

| Column | Type | Nullable | Default | Description |
|---------|------|----------|---------|-------------|
| id | UUID | No | gen_random_uuid() | Internal primary key |
| merchant_id | UUID | No | - | Merchant owning the bank account |
| bank_name | VARCHAR(100) | No | - | Bank name |
| account_holder_name | VARCHAR(255) | No | - | Name registered with the bank |
| account_number | TEXT | No | - | Encrypted bank account number |
| ifsc_code | VARCHAR(11) | No | - | IFSC code |
| currency | currency | No | 'INR' | Supported currency |
| is_primary | BOOLEAN | No | false | Indicates the Merchant's default bank account |
| is_verified | BOOLEAN | No | false | Indicates whether the account has been verified |
| status | bank_account_status | No | 'ACTIVE' | Bank account status |
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
ON DELETE RESTRICT
```

---

# Unique Constraints

None.

Business rules regarding duplicate bank accounts are enforced by the application layer.

---

# Check Constraints

None.

The following business rules are enforced by the application:

- A Merchant may link a maximum of three bank accounts.
- Exactly one bank account must be marked as Primary.

---

# Indexes

| Index | Purpose |
|---------|----------|
| merchant_id | Fetch all bank accounts for a Merchant |
| is_primary | Quickly locate the Primary bank account |
| status | Filter active bank accounts |
| created_at | Reporting and ordering |

---

# Enums Used

## bank_account_status

```text
ACTIVE
UNLINKED
```

---

## currency

```text
INR
```

---

# Referenced By

The following workflows use Linked Bank Accounts:

- Wallet Funding
- Payout

---

# Notes

- Every Linked Bank Account belongs to exactly one Merchant.
- A Merchant may link a maximum of three bank accounts.
- Only one Linked Bank Account may be marked as Primary.
- Account numbers are stored encrypted.
- APIs expose only masked account numbers.
- Only verified and active bank accounts may participate in financial operations.
- SamPay never stores or manages bank account balances.
- The database stores only the information required to identify and interact with the external bank account.