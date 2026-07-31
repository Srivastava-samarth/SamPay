# Payout Domain

## Overview

A Payout represents a financial transaction initiated by a Merchant User to transfer funds from the Merchant Wallet to one of the Merchant's linked Bank Accounts.

Payouts are used by Merchants to withdraw funds from SamPay back into their own bank accounts.

Unlike Payments, Payouts never transfer funds to external recipients. They can only be made to Bank Accounts owned by the Merchant.

---

# Responsibilities

The Payout domain is responsible for:

- Processing Merchant payouts.
- Validating payout requests.
- Validating linked Bank Accounts.
- Reserving Wallet funds.
- Coordinating Compliance validation.
- Moving funds from Wallet to the Payout Vault.
- Recording payout lifecycle.
- Publishing payout events.

The Payout domain is **not** responsible for:

- Wallet Management.
- Bank Account Management.
- Settlement Execution.
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
| description | Payout description (optional) |
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

Reserved Wallet balance is immediately released back to the Merchant Wallet.

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
        │        Payout Rejected
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
Debit Wallet      Release Reserved Funds
Credit Payout Vault      │
        │                ▼
        ▼             FAILED
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
- Merchant funds are always used for payout processing.

---

## Destination Bank Account

- A Payout can only be made to a Bank Account owned by the Merchant.
- Destination Bank Account must be ACTIVE.
- Destination Bank Account must be VERIFIED (if verification is enabled).

---

## Wallet Balance

- Wallet must contain sufficient available balance before processing.
- Automatic Wallet funding is **not** performed during Payouts.
- If sufficient balance is unavailable, the Payout request is rejected.

---

## Wallet Reservation

- Wallet funds are reserved before Compliance validation.
- Row-level locking is used during reservation.
- Reserved funds prevent concurrent overspending.

---

## Failed Payout

If a Payout fails after Wallet reservation:

- Reserved funds are immediately released.
- Merchant Wallet balance is restored.

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

## Primary Bank Account Rules

### Automatic Wallet Funding (Payments)

When a Payment requires automatic Wallet funding due to insufficient balance:

- Funds are always debited from the Merchant's **Primary Linked Bank Account**.
- Secondary linked Bank Accounts are never considered for automatic funding.
- Merchants can change which account is used by changing their Primary Bank Account.

### Payout Destination

- A Merchant may transfer funds to **any ACTIVE linked Bank Account** they own.
- If no destination Bank Account is specified, the Merchant's **Primary Linked Bank Account** is used by default.

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
POST   /payouts

GET    /payouts

GET    /payouts/{payoutId}

POST   /payouts/{payoutId}/cancel
```

Payouts cannot be updated or deleted.

---

# Relationships

| Entity | Relationship |
|----------|--------------|
| Merchant | N : 1 |
| User | N : 1 |
| Wallet | Source of Funds |
| Linked Bank Account | Destination of Funds |
| Payout Vault | Temporary holding before bank transfer |
| Compliance | Validates Payout |
| Ledger | Records financial movements |
| Notification | Consumes Payout events |

---

# Notes

- Payouts always transfer funds from the Merchant Wallet to one of the Merchant's own linked Bank Accounts.
- Automatic Wallet funding is never performed during Payouts.
- The Payment workflow always uses the Primary Linked Bank Account for automatic Wallet funding.
- Every Payout belongs to a Merchant and is initiated by a User acting on behalf of that Merchant.
- Every balance modification creates corresponding History and Ledger records.