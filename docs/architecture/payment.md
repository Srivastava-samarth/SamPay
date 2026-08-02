# Payment Domain

## Overview

A Payment represents a financial transaction initiated by a Merchant User to transfer funds owned by the Merchant to a Recipient.

A Payment is funded using the Merchant Wallet. If the Wallet has insufficient funds and automatic top-up is enabled, the Payment workflow debits the Merchant's linked Bank Account, credits the Wallet and then continues processing.

The Payment domain is responsible for orchestrating the complete payment lifecycle from validation to successful completion.

---

# Responsibilities

The Payment domain is responsible for:

- Processing Merchant payments.
- Validating payment requests.
- Validating transaction limits.
- Automatically funding the Wallet when enabled.
- Reserving Wallet funds.
- Coordinating Compliance validation.
- Moving funds from Wallet to the Payment Vault.
- Recording payment lifecycle.
- Publishing payment events.

The Payment domain is **not** responsible for:

- Wallet Management.
- Vault Management.
- Recipient Management.
- Settlement Execution.
- Ledger Management.
- Notification Delivery.
- Compliance Decision Making.

---

# Payment Structure

A Payment owns the following information:

- Merchant
- Initiating User
- Recipient
- Amount
- Fee
- Net Amount
- Currency
- Funding Source
- Status
- Settlement Status
- Merchant Reference
- Description
- Audit Information

---

# Payment Attributes

| Field | Description |
|--------|-------------|
| id | Internal UUID |
| merchant_id | Merchant initiating the payment |
| initiated_by_user_id | User initiating the payment |
| recipient_id | Recipient receiving the payment |
| funding_source | WALLET / AUTO_TOPUP |
| amount | Payment amount |
| fee | Platform fee |
| net_amount | Amount after fee deduction |
| currency | Supported currency (INR) |
| merchant_reference | Merchant supplied reference (optional) |
| description | Payment description (optional) |
| idempotency_key | Prevents duplicate payment creation |
| status | Current payment status |
| settlement_status | Settlement lifecycle |
| created_at | Record creation timestamp |
| updated_at | Record last update timestamp |

---

# Payment Status

## CREATED

Payment request has been accepted.

---

## PROCESSING

Payment is currently being processed.

---

## COMPLETED

Funds have successfully moved to the Payment Vault.

---

## FAILED

Payment processing failed.

Reserved Wallet balance is immediately released back to the Merchant Wallet.

If an automatic Wallet funding occurred before failure, those funds remain inside the Merchant Wallet for future Payments or Payouts.

---

## SETTLED

Funds have been successfully settled to the recipient.

---

# Settlement Status

## PENDING

Waiting for settlement.

---

## PROCESSING

Settlement is currently running.

---

## COMPLETED

Settlement completed successfully.

---

## FAILED

Settlement failed.

---

# Funding Sources

## WALLET

The Merchant Wallet contained sufficient balance.

---

## AUTO_TOPUP

The Merchant Wallet had insufficient balance.

The Payment workflow automatically:

- Debits the linked Bank Account.
- Credits the Merchant Wallet.
- Continues payment processing.

---

# Payment Workflow

```text
User Initiates Payment
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
Validate Recipient
        │
        ▼
Validate Transaction Limits
        │
        ▼
Check Wallet Balance
        │
        ├────────────────────────────┐
        │                            │
Enough Balance             Insufficient Balance
        │                            │
        │                   Auto Top-up Enabled?
        │                            │
        │                ├───────────┴───────────┐
        │                │                       │
        │               Yes                     No
        │                │                       │
        │                ▼                       ▼
        │      Debit Linked Bank          Payment Failed
        │      Credit Merchant Wallet
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
Credit Payment Vault      │
        │                 ▼
        ▼              FAILED
COMPLETED
        │
        ▼
Publish PaymentCompleted Event
        │
        ▼
Settlement Queue
```

---

# Business Rules

## Ownership

- Every Payment belongs to exactly one Merchant.
- Every Payment is initiated by exactly one User.
- Merchant funds are always used for payment processing.

---

## Recipient

- Every Payment references exactly one Recipient.
- Recipient must be ACTIVE.
- Recipient must be VERIFIED.

---

## Funding

- Payments always execute using Merchant Wallet funds.
- If Wallet balance is insufficient and Auto Top-up is enabled, the Payment workflow automatically funds the Wallet before continuing.
- Payment is the only workflow responsible for this automatic funding behaviour.

---

## Wallet Reservation

- Wallet funds are reserved before Compliance validation.
- Row-level locking is used during reservation.
- Reserved funds prevent concurrent overspending.

---

## Failed Payment

If a Payment fails after Wallet reservation:

- Reserved funds are immediately released.
- Any successful automatic Wallet funding is not reversed.
- Funds remain inside the Merchant Wallet.
- Merchants may later transfer those funds back to a linked Bank Account through the Payout workflow.

---

## Fees

- Platform fees are calculated during Payment creation.
- Amount, Fee and Net Amount are permanently stored.
- Financial values are never recalculated.

---

## Currency

Version 1 supports INR only.

---

## Idempotency

- Every Payment request must include an Idempotency Key.
- Duplicate requests return the existing Payment.

---

## Immutability

Payments are immutable financial records.

The following fields cannot be modified after creation:

- Amount
- Recipient
- Currency

---

## Security

- Only authenticated Users may initiate Payments.
- User authorization is determined by the assigned Role.
- Transaction limits are enforced before processing.
- High-value transactions require Email OTP verification.

---

# Domain Events

- PaymentCreated
- PaymentProcessing
- PaymentCompleted
- PaymentFailed
- PaymentSettled

These events are consumed by:

- Settlement
- Notification
- Ledger
- Audit

---

# APIs

```http
POST   /payments

GET    /payments

GET    /payments/{paymentId}

POST   /payments/{paymentId}/cancel
```

Payments cannot be updated or deleted.

---

# Relationships

| Entity | Relationship |
|----------|--------------|
| Merchant | N : 1 |
| User | N : 1 |
| Recipient | N : 1 |
| Wallet | Source of Funds |
| Linked Bank Account | Auto-top-up funding source |
| Payment Vault | Destination of Funds |
| Settlement | Settles completed Payments |
| Compliance | Validates Payments |
| Ledger | Records financial movements |
| Notification | Consumes Payment events |

---

# Notes

- Payment is a business workflow, not merely a database entity.
- Every Payment belongs to a Merchant and is initiated by a User acting on behalf of that Merchant.
- Wallet funding is handled internally by the Payment workflow when required.
- Payment publishes events instead of directly invoking downstream domains.
- Every balance modification creates corresponding History and Ledger records.

---

# Version 2 Considerations

As SamPay evolves, the Payment domain may support:

- Scheduled Payments.
- Batch Payments.
- Recurring Payments.
- Multi-currency support.
- External Payment Providers.
- Risk Scoring.
- Fraud Detection.
- Real-time Payment Analytics.