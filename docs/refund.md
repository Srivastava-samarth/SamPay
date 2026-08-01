# Refund Domain

## Overview

A Refund represents the reversal of a previously completed Payment.

Unlike Payments and Payouts, a Refund cannot exist independently. Every Refund is associated with exactly one Payment and transfers funds back to the original payment source.

Refunds may be full or partial but can never exceed the original Payment amount.

---

# Responsibilities

The Refund domain is responsible for:

- Processing Merchant refunds.
- Validating refund eligibility.
- Validating refundable amount.
- Reserving Wallet funds.
- Coordinating Compliance validation.
- Moving funds from Wallet to the Refund Vault.
- Recording refund lifecycle.
- Publishing refund events.

The Refund domain is **not** responsible for:

- Payment Management.
- Wallet Management.
- Recipient Management.
- Settlement Execution.
- Ledger Management.
- Notification Delivery.
- Compliance Decision Making.

---

# Refund Structure

A Refund owns the following information:

- Original Payment
- Merchant
- Initiating User
- Refund Amount
- Refund Reason
- Status
- Audit Information

---

# Refund Attributes

| Field | Description |
|--------|-------------|
| id | Internal UUID |
| payment_id | Original Payment being refunded |
| merchant_id | Merchant initiating the refund |
| initiated_by_user_id | User initiating the refund |
| amount | Refund amount |
| reason | Refund reason (optional) |
| currency | Supported currency (INR) |
| idempotency_key | Prevents duplicate refund creation |
| status | Current refund status |
| created_at | Record creation timestamp |
| updated_at | Record last update timestamp |

---

# Refund Status

## CREATED

Refund request has been accepted.

---

## PROCESSING

Refund is currently being processed.

---

## COMPLETED

Funds have been successfully refunded.

---

## FAILED

Refund processing failed.

Reserved Wallet balance is immediately released back to the Merchant Wallet.

---

# Refund Workflow

```text
User Initiates Refund
        │
        ▼
Locate Original Payment
        │
        ▼
Validate Refund Eligibility
        │
        ▼
Validate Refund Amount
        │
        ▼
Authenticate User
        │
        ▼
Authorize User Role
        │
        ▼
Check Wallet Balance
        │
        ├───────────────┐
        │               │
Enough Balance    Insufficient Balance
        │               │
        │               ▼
        │        Refund Rejected
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
Credit Refund Vault      │
        │                ▼
        ▼             FAILED
COMPLETED
        │
        ▼
Publish RefundCompleted Event
```

---

# Business Rules

## Original Payment

- Every Refund must reference an existing Payment.
- Only COMPLETED Payments can be refunded.

---

## Refund Amount

- Refund amount must be greater than zero.
- Refund amount cannot exceed the remaining refundable balance.

Validation:

```
Total Refunded + Current Refund <= Original Payment Amount
```

---

## Full & Partial Refunds

Both full and partial refunds are supported.

Example:

Original Payment: ₹10,000

Refund 1: ₹3,000

Refund 2: ₹2,000

Remaining Refundable Amount: ₹5,000

---

## Destination

- Refunds always follow the original Payment destination.
- Merchants cannot select a different destination.
- Recipient details are inherited from the original Payment.

---

## Wallet Balance

- Merchant Wallet must contain sufficient available balance.
- Automatic Wallet funding is not performed during Refunds.
- If sufficient balance is unavailable, the Refund request is rejected.

---

## Wallet Reservation

- Wallet funds are reserved before Compliance validation.
- Row-level locking is used during reservation.
- Reserved funds prevent concurrent overspending.

---

## Failed Refund

If a Refund fails after Wallet reservation:

- Reserved funds are immediately released.
- Merchant Wallet balance is restored.

---

## Currency

Version 1 supports INR only.

---

## Idempotency

- Every Refund request must include an Idempotency Key.
- Duplicate requests return the existing Refund.

---

## Immutability

Refunds are immutable financial records.

The following fields cannot be modified after creation:

- Payment
- Amount
- Currency

---

## Security

- Only authenticated Users may initiate Refunds.
- User authorization is determined by the assigned Role.
- Transaction limits are enforced before processing.
- High-value Refunds require Email OTP verification.

---

# Domain Events

- RefundCreated
- RefundProcessing
- RefundCompleted
- RefundFailed

These events are consumed by:

- Notification
- Ledger
- Audit

---

# APIs

```http
POST /payments/{paymentId}/refund

GET /refunds

GET /refunds/{refundId}
```

Refunds cannot be updated or deleted.

---

# Relationships

| Entity | Relationship |
|----------|--------------|
| Payment | N : 1 |
| Merchant | N : 1 |
| User | N : 1 |
| Wallet | Source of Funds |
| Refund Vault | Temporary holding before refund |
| Compliance | Validates Refund |
| Ledger | Records financial movements |
| Notification | Consumes Refund events |

---

# Notes

- Refunds always reference an existing Payment.
- A Payment may have multiple partial Refunds.
- The sum of all Refunds can never exceed the original Payment amount.
- Refunds always return funds using the original Payment details.
- Every Refund belongs to a Merchant and is initiated by a User acting on behalf of that Merchant.
- Every balance modification creates corresponding History and Ledger records.