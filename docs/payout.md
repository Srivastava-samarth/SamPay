# Payout Domain

## Overview

A Payout represents a financial transaction initiated by a Merchant User to transfer funds from the Merchant Wallet to one of the Merchant's linked Bank Accounts.

Payouts allow Merchants to withdraw funds from SamPay back into their own bank accounts.

Unlike Payments and Refunds, Payouts do not pass through the Settlement workflow. Once a Payout request is validated, SamPay directly initiates the transfer to the selected linked Bank Account.

---

# Responsibilities

The Payout domain is responsible for:

- Processing Merchant payouts.
- Validating payout requests.
- Validating destination Bank Accounts.
- Reserving Wallet funds.
- Initiating bank transfers.
- Recording payout lifecycle.
- Publishing payout events.

The Payout domain is **not** responsible for:

- Wallet Management.
- Bank Account Management.
- Settlement.
- Ledger Management.
- Notification Delivery.
- Compliance Decision Making.

---

# Payout Structure

A Payout owns the following information:

- Merchant
- Initiating User
- Destination Bank Account
- Amount
- Fee
- Net Amount
- Currency
- Status
- Description
- Audit Information

---

# Payout Attributes

| Field | Description |
|--------|-------------|
| id | Internal UUID |
| merchant_id | Merchant initiating the payout |
| initiated_by_user_id | User initiating the payout |
| bank_account_id | Destination linked Bank Account |
| amount | Payout amount |
| fee | Platform fee |
| net_amount | Amount transferred after fee deduction |
| currency | Supported currency (INR) |
| merchant_reference | Merchant supplied reference (optional) |
| description | Optional payout description |
| idempotency_key | Prevents duplicate payout creation |
| status | Current payout status |
| created_at | Record creation timestamp |
| updated_at | Record last update timestamp |

---

# Payout Status

## CREATED

Payout request has been accepted.

---

## PROCESSING

Payout is currently being processed.

---

## COMPLETED

Funds have been successfully transferred to the Merchant's linked Bank Account.

---

## FAILED

Payout processing failed.

Any reserved Wallet balance is immediately released.

---

# Payout Workflow

```text
User Initiates Payout
        │
        ▼
Authenticate User
        │
        ▼
Authorize User Role
        │
        ▼
Validate Merchant
        │
        ▼
Validate Linked Bank Account
        │
        ▼
Validate Transaction Limits
        │
        ▼
Check Wallet Balance
        │
        ├───────────────┐
        │               │
Enough Balance    Insufficient Balance
        │               │
        │               ▼
        │        Reject Payout
        │
        ▼
Reserve Wallet Balance
        │
        ▼
Compliance Validation
        │
        ├───────────────┐
        │               │
      Pass            Fail
        │               │
        ▼               ▼
Call Bank API     Release Reserved Funds
        │               │
        ├───────┐       ▼
        │       │    FAILED
     Success   Failure
        │       │
        ▼       ▼
Debit Wallet  Release Reserved Funds
Create Ledger Entry
        │
        ▼
COMPLETED
        │
        ▼
Publish PayoutCompleted Event
```

---

# Business Rules

## Ownership

- Every Payout belongs to exactly one Merchant.
- Every Payout is initiated by exactly one User.

---

## Destination Bank Account

- Payouts can only be made to Bank Accounts owned by the Merchant.
- Destination Bank Account must be ACTIVE.
- Destination Bank Account must be VERIFIED (if verification is enabled).

---

## Wallet Balance

- Merchant Wallet must contain sufficient available balance.
- Wallet funds are reserved before initiating the bank transfer.
- Automatic Wallet funding is never performed during Payouts.

---

## Wallet Reservation

- Wallet funds are reserved before contacting the bank.
- Row-level locking is used while reserving funds.
- Reserved funds prevent concurrent overspending.

---

## Successful Payout

After the bank transfer succeeds:

- Reserved funds are permanently debited from the Wallet.
- Wallet History is created.
- Ledger entries are created.
- Payout is marked COMPLETED.

---

## Failed Payout

If the bank transfer fails:

- Reserved Wallet funds are immediately released.
- Wallet balance is restored.
- Payout is marked FAILED.

---

## Primary Bank Account

### Automatic Wallet Funding

Automatic Wallet funding for Payments always debits the Merchant's Primary Linked Bank Account.

---

### Payout Destination

A Merchant may transfer funds to any ACTIVE linked Bank Account.

If no destination is provided, the Primary Linked Bank Account is used by default.

---

## Fees

- Platform fees are calculated during Payout creation.
- Amount, Fee and Net Amount are permanently stored.
- Financial values are never recalculated.

---

## Currency

Version 1 supports INR only.

---

## Idempotency

- Every Payout request must include an Idempotency Key.
- Duplicate requests return the existing Payout.

---

## Immutability

Payouts are immutable financial records.

The following fields cannot be modified after creation:

- Amount
- Destination Bank Account
- Currency

---

## Security

- Only authenticated Users may initiate Payouts.
- User authorization is determined by the assigned Role.
- Transaction limits are enforced before processing.
- High-value Payouts require Email OTP verification.

---

# Domain Events

- PayoutCreated
- PayoutProcessing
- PayoutCompleted
- PayoutFailed

These events are consumed by:

- Notification
- Ledger
- Audit

---

# APIs

```http
POST /payouts

GET /payouts

GET /payouts/{payoutId}

POST /payouts/{payoutId}/cancel
```

Payouts cannot be updated or deleted.

---

# Relationships

| Entity | Relationship |
|----------|--------------|
| Merchant | N : 1 |
| User | N : 1 |
| Wallet | Source of Funds |
| Linked Bank Account | Destination |
| Compliance | Validates Payout |
| Ledger | Records financial movements |
| Notification | Consumes Payout events |

---

# Notes

- Payouts transfer funds directly from the Merchant Wallet to a linked Bank Account.
- Payouts do not participate in the Settlement workflow.
- Wallet balance is reserved before initiating the bank transfer.
- Reserved funds are released immediately if the transfer fails.
- Every successful Wallet debit creates Wallet History and Ledger entries within the same database transaction.
- Every Payout belongs to a Merchant and is initiated by a User acting on behalf of that Merchant.