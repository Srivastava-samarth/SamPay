# Recipient Domain

## Overview

A Recipient represents a beneficiary to whom a Merchant can transfer funds.

Recipients are owned by Merchants and are reused across multiple Payments. Every Payment references a Recipient instead of storing beneficiary details directly.

Recipients are immutable financial entities. Any material change to recipient information creates a new Recipient while the previous Recipient is marked as INACTIVE.

---

# Responsibilities

The Recipient domain is responsible for:

- Managing Merchant recipients.
- Storing beneficiary information.
- Validating recipient status before Payments.
- Maintaining immutable recipient records.

The Recipient domain is **not** responsible for:

- Processing Payments.
- Wallet Management.
- Settlement.
- Compliance Decisions.
- Notification Delivery.

---

# Recipient Structure

A Recipient owns the following information:

- Merchant
- Recipient Type
- Recipient Details
- Verification Status
- Status
- Audit Information

---

# Recipient Attributes

| Field | Description |
|--------|-------------|
| id | Internal UUID |
| merchant_id | Merchant who owns the recipient |
| recipient_type | MERCHANT / BANK_ACCOUNT |
| recipient_merchant_id | Merchant ID if the recipient is another SamPay Merchant (nullable) |
| name | Friendly display name |
| account_holder_name | Bank account holder name (nullable) |
| account_number | Encrypted bank account number (nullable) |
| ifsc_code | Bank IFSC code (nullable) |
| bank_name | Bank name (nullable) |
| currency | INR |
| verification_status | PENDING / VERIFIED / FAILED |
| status | ACTIVE / INACTIVE |
| created_at | Record creation timestamp |
| updated_at | Record last update timestamp |

---

# Recipient Types

## MERCHANT

Represents another Merchant registered on SamPay.

---

## BANK_ACCOUNT

Represents an external bank account outside SamPay.

---

# Verification Status

## PENDING

Recipient has been created but not yet verified.

---

## VERIFIED

Recipient is eligible to receive Payments.

---

## FAILED

Recipient verification failed.

Payments are not allowed.

---

# Recipient Status

## ACTIVE

Recipient can receive new Payments.

---

## INACTIVE

Recipient cannot receive new Payments.

Historical Payments remain unaffected.

---

# Business Rules

## Ownership

- Every Recipient belongs to exactly one Merchant.
- Different Merchants cannot share Recipient records.

---

## Immutability

Recipients are immutable.

The following fields cannot be modified:

- Recipient Type
- Merchant
- Account Holder Name
- Account Number
- IFSC Code
- Bank Name

If any of these values need to change:

1. Create a new Recipient.
2. Mark the previous Recipient as INACTIVE.
3. Future Payments use the new Recipient.

Historical Payments continue referencing the original Recipient.

---

## Verification

- Only VERIFIED recipients may receive Payments.
- Payments to PENDING or FAILED recipients are rejected.

---

## Security

- Bank account numbers are encrypted at rest.
- APIs return masked account numbers.
- Sensitive banking information must never be written to logs.

---

# APIs

```http
POST   /recipients

GET    /recipients

GET    /recipients/{recipientId}

POST   /recipients/{recipientId}/deactivate
```

Recipients cannot be updated or deleted.

---

# Relationships

| Entity | Relationship |
|----------|--------------|
| Merchant | N : 1 |
| Payment | 1 : N |

---

# Notes

- Recipients act as reusable beneficiaries for Payments.
- A Recipient may represent either another SamPay Merchant or an external Bank Account.
- Payments reference Recipient IDs instead of storing beneficiary details.
- Recipient history is preserved through immutability.