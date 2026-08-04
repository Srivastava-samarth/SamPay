# Bank Service Database Schema

## Overview

The Bank Service is responsible for managing merchant bank accounts, bank transfers, and banking ledger entries.

It is completely independent from the Payment Service.

The Bank Service **does not** know about:

- Wallets
- Payments
- Refunds
- Settlements
- Payment Ledger

The Payment Service will interact with the Bank Service only through its APIs.

---

# Tables

The Bank Service consists of four core tables:

1. `bank_accounts`
2. `linked_bank_accounts`
3. `bank_transfers`
4. `bank_ledger`

---

# 1. bank_accounts

## Purpose

This table represents actual bank accounts maintained by the SamPay Mock Bank.

It is the source of truth for:

- Account Number
- Current Balance
- Currency
- Account Status

All money movement inside the bank references this table.

## Schema

| Column | Type | Description |
|----------|------|-------------|
| id | UUID | Primary Key |
| account_number | VARCHAR | Unique account number |
| currency | VARCHAR | Currency (INR initially) |
| balance | DECIMAL | Current account balance |
| status | ENUM | ACTIVE, BLOCKED, CLOSED |
| created_at | TIMESTAMP | Creation timestamp |
| updated_at | TIMESTAMP | Last update timestamp |

---

# 2. linked_bank_accounts

## Purpose

Represents merchant-owned bank accounts inside SamPay Bank.

This table connects merchants with their corresponding bank accounts while allowing merchants to create multiple accounts.

The bank account itself remains the primary financial entity.

## Schema

| Column | Type | Description |
|----------|------|-------------|
| id | UUID | Primary Key |
| merchant_id | UUID | Merchant owning the account |
| bank_account_id | UUID | FK → bank_accounts.id |
| account_name | VARCHAR | User-defined name (Business, Savings, Payroll etc.) |
| account_type | ENUM | PRIMARY, SECONDARY |
| is_primary | BOOLEAN | Indicates default account |
| status | ENUM | ACTIVE, BLOCKED, CLOSED |
| created_at | TIMESTAMP | Creation timestamp |
| updated_at | TIMESTAMP | Last update timestamp |

---

# 3. bank_transfers

## Purpose

Represents the lifecycle of a bank transfer.

Unlike the ledger, transfers are mutable.

A transfer can move through different states before completion.

## Supported States

- CREATED
- PROCESSING
- SUCCESS
- FAILED

## Schema

| Column | Type | Description |
|----------|------|-------------|
| id | UUID | Primary Key |
| reference_type | VARCHAR | FUNDING, INTERNAL_TRANSFER, PAYOUT, etc. |
| reference_id | UUID | Business reference |
| from_bank_account_id | UUID | Source bank account |
| to_bank_account_id | UUID | Destination bank account |
| amount | DECIMAL | Transfer amount |
| currency | VARCHAR | Currency |
| status | ENUM | CREATED, PROCESSING, SUCCESS, FAILED |
| failure_reason | TEXT | Failure reason (if any) |
| created_at | TIMESTAMP | Creation timestamp |
| updated_at | TIMESTAMP | Last update timestamp |
| completed_at | TIMESTAMP | Completion timestamp |

---

# 4. bank_ledger

## Purpose

Immutable accounting records.

Every successful movement of money creates one or more ledger entries.

Ledger entries are never updated.

## Schema

| Column | Type | Description |
|----------|------|-------------|
| id | UUID | Primary Key |
| reference_type | VARCHAR | FUNDING, TRANSFER, FEE, etc. |
| reference_id | UUID | Business reference |
| from_bank_account_id | UUID | Source bank account |
| to_bank_account_id | UUID | Destination bank account |
| amount | DECIMAL | Transfer amount |
| currency | VARCHAR | Currency |
| description | TEXT | Optional description |
| created_at | TIMESTAMP | Creation timestamp |

---

# Relationship Diagram

```text
Merchant
    │
    ▼
linked_bank_accounts
    │
    ▼
bank_accounts
    │
    ├──────────────► bank_transfers
    │
    └──────────────► bank_ledger
```

---

# Funding Flow

Merchant initiates a funding request.

Example:

```
Funding Amount = ₹100,000
Funding Fee    = ₹1,000 (1%)
Credited       = ₹99,000
```

Flow:

```text
Merchant
      │
      ▼
Fund Bank Account
      │
      ▼
Funding Fee Calculated
      │
      ├────────────► Company Bank Account (+₹1,000)
      │
      ▼
Merchant Bank Account (+₹99,000)
```

Ledger Entries:

```
External
        │
        ▼
Merchant Bank Account
₹99,000
```

```
External
        │
        ▼
Company Bank Account
₹1,000
```

---

# Internal Bank Transfer Flow

```text
Merchant A Bank Account

↓

Merchant B Bank Account
```

Creates:

- Bank Transfer
- Bank Ledger Entry

---

# Design Principles

- `bank_accounts` is the financial source of truth.
- `linked_bank_accounts` stores merchant ownership and metadata.
- `bank_transfers` manages transfer lifecycle.
- `bank_ledger` stores immutable accounting entries.
- Only successful transfers generate ledger entries.
- All balances are maintained in `bank_accounts`.
- All money movement references `bank_account_id`.
- Merchant-specific information is isolated in `linked_bank_accounts`.