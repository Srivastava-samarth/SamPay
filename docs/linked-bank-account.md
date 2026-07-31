# Linked Bank Account Domain

## Overview

A Linked Bank Account represents an external bank account associated with a Merchant.

Linked Bank Accounts are used to transfer funds between the Merchant and the SamPay platform. They act as the source or destination for financial operations such as Wallet Top-ups, Payouts and Settlements.

SamPay does not own or maintain the balance of a Linked Bank Account. It only stores the information required to securely identify and interact with the account.

---

# Responsibilities

The Linked Bank Account domain is responsible for:

- Maintaining bank account information.
- Verifying bank account ownership.
- Managing the Merchant's linked bank accounts.
- Maintaining the primary settlement account.
- Providing verified bank accounts for financial operations.

The Linked Bank Account domain is **not** responsible for:

- Maintaining account balances.
- Payment Processing.
- Payout Processing.
- Settlement Processing.
- Wallet Management.
- Compliance Decision Making.
- Notification Delivery.

---

# Linked Bank Account Structure

A Linked Bank Account owns the following information:

- Merchant
- Bank Details
- Verification Status
- Primary Account Flag
- Status
- Audit Information

---

# Linked Bank Account Attributes

| Field | Description |
|--------|-------------|
| id | Internal UUID |
| merchant_id | Merchant owning the bank account |
| bank_name | Name of the bank |
| account_holder_name | Name registered with the bank |
| account_number | Encrypted bank account number |
| ifsc_code | IFSC Code |
| currency | Supported currency (INR) |
| account_type | Savings or Current |
| is_primary | Indicates whether this is the default settlement account |
| status | Current bank account status |
| verification_status | Verification state of the account |
| created_at | Record creation timestamp |
| updated_at | Record last update timestamp |

---

# Bank Account Status

## ACTIVE

The bank account can participate in all supported financial operations.

---

## INACTIVE

The bank account has been disabled or unlinked and cannot be used for any financial operation.

---

# Verification Status

## PENDING

The bank account has been linked but verification has not yet completed.

---

## VERIFIED

The ownership of the bank account has been successfully verified.

Only verified bank accounts may participate in financial operations.

---

## FAILED

The verification process failed.

The Merchant must update the account details or link a different bank account.

---

# Business Rules

## Ownership

- Every Linked Bank Account belongs to exactly one Merchant.
- A Merchant can link a maximum of three bank accounts.
- A Merchant must have at least one verified bank account before activation.

---

## Primary Bank Account

- Exactly one bank account must be marked as the Primary account.
- The Primary account is used as the default settlement destination.
- The Primary account cannot be unlinked.
- To unlink the current Primary account, the Merchant must first designate another linked account as Primary.

---

## Verification

- Every newly linked bank account starts in the **PENDING** verification state.
- Only **VERIFIED** bank accounts may be used for Wallet Top-ups, Payouts or Settlements.

---

## Security

- Bank account numbers must be encrypted at rest.
- APIs should expose only masked account numbers.
- Sensitive information must never appear in logs.

---

# APIs

## Linked Bank Account

```http
POST   /bank-accounts

GET    /bank-accounts

GET    /bank-accounts/{bankAccountId}

PATCH  /bank-accounts/{bankAccountId}

POST   /bank-accounts/{bankAccountId}/unlink

PATCH  /bank-accounts/{bankAccountId}/primary
```

---

# Relationships

| Entity | Relationship |
|----------|--------------|
| Merchant | N : 1 |
| Wallet | Referenced by financial workflows |
| Payment | Used for Wallet Top-up |
| Payout | Destination account |
| Settlement | Destination account |

---

# Notes

- Linked Bank Accounts represent external banking infrastructure.
- SamPay never stores or maintains the balance of a Linked Bank Account.
- All modifications are recorded in the append-only `bank_account_history` table.
- Financial operations may only use verified and active bank accounts.

---

# Version 2 Considerations

As the platform evolves, the Linked Bank Account domain may support:

- Multiple currencies.
- Shared corporate bank accounts.
- Automatic bank account verification.
- Bank account nicknames.
- External banking provider integrations.
- Webhook-based verification updates.

These enhancements are intentionally deferred from Version 1 to keep the onboarding process simple and maintainable.