# Merchant Domain

## Overview

A Merchant is a legal entity registered on SamPay that owns financial resources and authorizes users to perform financial operations on its behalf.

Every financial resource in the system belongs to exactly one Merchant.

---

# Merchant Types

- INDIVIDUAL
- CORPORATE

---

# Responsibilities

The Merchant domain is responsible for:

- Registering a Merchant
- Activating / Deactivating a Merchant
- Managing Merchant Users
- Managing Merchant Configuration
- Managing Linked Bank Accounts
- Owning Financial Resources
- Provisioning a Wallet during Merchant onboarding

The Merchant domain is **not** responsible for:

- Authentication
- Authorization
- Processing Payments
- Processing Payouts
- Processing Refunds
- Settlement Execution
- Ledger Balance Calculation
- Compliance Validation
- Notification Delivery

---

# Merchant Owns

```text
Merchant
│
├── Users
├── Wallet
├── Linked Bank Accounts
├── Payments
├── Payouts
├── Refunds
├── Ledger
└── Webhook Configuration
```

---

# Merchant Lifecycle

```text
REGISTERED
      │
      ▼
KYC_PENDING
      │
      ▼
ACTIVE
   │      │
   ▼      ▼
SUSPENDED BLOCKED
      │
      ▼
    CLOSED
```

---

# Merchant Attributes

| Field | Description |
|--------|-------------|
| id | Internal UUID |
| merchant_reference | Public Merchant Identifier |
| entity_type | INDIVIDUAL / CORPORATE |
| legal_name | Official registered name |
| display_name | Merchant display name |
| merchant_notification_email | Business notification email |
| status | Merchant status |
| kyc_status | KYC status |
| created_at | Created timestamp |
| updated_at | Updated timestamp |

---

# Business Rules

## Identity

- Merchant ID is immutable.
- Merchant Reference is immutable.
- Entity Type cannot be changed.

---

## Users

- Every Merchant must always have one Owner.
- Ownership must be transferred before removing the current Owner.

---

## Wallet

- Every Merchant owns exactly one Wallet in Version 1.
- Wallet is automatically provisioned during Merchant onboarding.
- Wallet cannot exist without a Merchant.

---

## Bank Accounts

- Maximum three linked bank accounts.
- Exactly one Primary Bank Account.
- Primary Bank Account cannot be deleted until another linked account is promoted as Primary.

---

## Merchant Status

- Merchant cannot become ACTIVE until KYC is approved.
- Suspended or Blocked Merchants cannot initiate new Payments, Payouts or Refunds.

---

# Domain Events

- MerchantRegistered
- MerchantKYCSubmitted
- MerchantKYCApproved
- MerchantActivated
- MerchantSuspended
- MerchantBlocked
- MerchantClosed
- MerchantUserAdded
- MerchantUserRemoved
- MerchantBankLinked
- MerchantBankUnlinked
- MerchantWebhookUpdated

---

# APIs

## Merchant

```http
POST   /merchants
GET    /merchants/{merchantId}
PATCH  /merchants/{merchantId}
POST   /merchants/{merchantId}/activate
POST   /merchants/{merchantId}/deactivate
```

---

## Users

```http
POST   /merchants/{merchantId}/users
DELETE /merchants/{merchantId}/users/{userId}
```

---

## Linked Bank Accounts

```http
POST   /merchants/{merchantId}/bank-accounts
DELETE /merchants/{merchantId}/bank-accounts/{bankAccountId}
```

---

## Webhooks

```http
PATCH /merchants/{merchantId}/webhook
```

---

# Business Scenarios

## Merchant Registration

```text
Merchant Registration
        │
        ▼
Compliance Validation
        │
        ▼
Merchant Created
        │
        ▼
Wallet Provisioned
        │
        ▼
Notification Sent
```

Initially, the Merchant is created in the **REGISTERED** state.

The Merchant progresses to **ACTIVE** only after successful KYC approval.

---

## Add Merchant User

```text
Merchant Owner
        │
        ▼
Invite User
        │
        ▼
User Accepts Invitation
        │
        ▼
Role Assigned
        │
        ▼
User Gains Access
```

---

## Link Bank Account

```text
Merchant
        │
        ▼
Submit Bank Account
        │
        ▼
Validate Ownership
        │
        ▼
Link Account
        │
        ▼
Mark Primary (if applicable)
```

A Merchant may link a maximum of three bank accounts.

---

# Relationships

| Entity | Relationship |
|----------|--------------|
| User | 1:N |
| Wallet | 1:1 |
| Linked Bank Account | 1:N |
| Payment | 1:N |
| Payout | 1:N |
| Refund | 1:N |
| Ledger | 1:N |
| Webhook Configuration | 1:1 |

---

# Notes

- Every Merchant owns exactly one Wallet.
- Wallet is automatically provisioned during onboarding.
- Authentication is delegated to the Authentication domain.
- Compliance validation is performed before Merchant creation.
- Notification is handled asynchronously after successful business events.
- Merchant owns business resources but does not execute financial operations directly.

---

# Version

**Architecture Status:** ✅ Frozen (V1)

**Last Updated:** 2026-08-02