# Linked Bank Account Domain

## Overview

A Linked Bank Account represents an external bank account associated with a Merchant.

Linked Bank Accounts are used to move funds between the Merchant and the SamPay platform.

In Version 1, they are primarily used for:

- Automatic Wallet funding
- Receiving Payouts

SamPay does not own or maintain the balance of a Linked Bank Account. It only stores the information required to securely identify and interact with the account.

---

# Responsibilities

The Linked Bank Account domain is responsible for:

- Maintaining bank account information.
- Verifying bank account ownership.
- Managing the Merchant's linked bank accounts.
- Maintaining the Primary Bank Account.
- Providing verified bank accounts for financial operations.

The Linked Bank Account domain is **not** responsible for:

- Maintaining account balances.
- Processing Payments.
- Processing Payouts.
- Executing Settlements.
- Managing Wallet balances.
- Performing Compliance decisions.
- Sending Notifications.

---

# Linked Bank Account Structure

Every Linked Bank Account maintains:

- Merchant
- Bank Details
- Verification Status
- Primary Account Flag
- Status

---

# Linked Bank Account Attributes

| Field | Description |
|--------|-------------|
| id | Internal UUID |
| account_reference | Public Bank Account Identifier |
| merchant_id | Merchant owning the bank account |
| bank_name | Name of the bank |
| account_holder_name | Name registered with the bank |
| account_number | Encrypted bank account number |
| ifsc_code | IFSC Code |
| currency | Supported currency (INR) |
| account_type | SAVINGS / CURRENT |
| is_primary | Indicates whether this is the default bank account |
| status | ACTIVE / UNLINKED |
| verification_status | PENDING / VERIFIED / FAILED |
| created_at | Record creation timestamp |
| updated_at | Record last update timestamp |

---

# Bank Account Status

## ACTIVE

The bank account can participate in supported financial operations.

---

## UNLINKED

The bank account has been removed from the Merchant and can no longer be used.

---

# Verification Status

## PENDING

The account has been linked but verification has not yet completed.

---

## VERIFIED

Ownership has been successfully verified.

Only VERIFIED bank accounts may participate in financial operations.

---

## FAILED

Verification failed.

The Merchant must update the account details or link another bank account.

---

# Business Rules

## Ownership

- Every Linked Bank Account belongs to exactly one Merchant.
- A Merchant may link a maximum of three bank accounts.

---

## Primary Bank Account

- Exactly one Linked Bank Account must be marked as Primary.
- The Primary Bank Account is used by default for:
  - Automatic Wallet funding.
  - Receiving Payouts.
- The Primary Bank Account cannot be unlinked.
- Before unlinking the current Primary account, another linked account must first be designated as Primary.

---

## Verification

- Every newly linked bank account starts in the **PENDING** state.
- Only **VERIFIED** bank accounts may participate in financial operations.

---

## Security

- Account numbers must be encrypted at rest.
- APIs expose only masked account numbers.
- Sensitive banking information must never appear in logs.

---

# APIs

```http
POST   /merchants/{merchantId}/bank-accounts

GET    /merchants/{merchantId}/bank-accounts

GET    /merchants/{merchantId}/bank-accounts/{bankAccountId}

PATCH  /merchants/{merchantId}/bank-accounts/{bankAccountId}

PATCH  /merchants/{merchantId}/bank-accounts/{bankAccountId}/primary

DELETE /merchants/{merchantId}/bank-accounts/{bankAccountId}
```

---

# Relationships

| Entity | Relationship |
|----------|--------------|
| Merchant | N : 1 |
| Wallet | Source for automatic Wallet funding |
| Payment | May fund Wallet when required |
| Payout | Destination bank account |

---

# Notes

- Linked Bank Accounts represent external banking infrastructure.
- SamPay never stores or maintains bank account balances.
- Only VERIFIED and ACTIVE bank accounts may participate in financial operations.
- Sensitive banking information is encrypted at rest and masked in API responses.
- Financial workflows reference Linked Bank Accounts but never modify their balances.

---

# Version 2 Considerations

Future enhancements may include:

- Multiple currencies.
- Shared corporate bank accounts.
- Automatic bank account verification.
- Bank account nicknames.
- External banking provider integrations.
- Webhook-based verification updates.

These enhancements are intentionally deferred from Version 1 to keep the platform simple and maintainable.

---

# Version

**Architecture Status:** ✅ Frozen (V1)

**Last Updated:** 2026-08-03