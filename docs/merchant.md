# Merchant Domain

## Overview

A Merchant is a legal entity registered on SamPay that owns financial resources and authorizes users to perform financial operations on its behalf.

Every financial resource in the system belongs to exactly one Merchant.

---

## Merchant Types

- INDIVIDUAL
- CORPORATE

---

## Responsibilities

The Merchant domain is responsible for:

- Registering a Merchant
- Activating / Deactivating a Merchant
- Managing Merchant Users
- Managing Merchant Configuration
- Managing Linked Bank Accounts
- Owning Financial Resources

The Merchant domain is **not** responsible for:

- Processing Payments
- Processing Payouts
- Processing Refunds
- Settlement Execution
- Ledger Balance Calculation

---

## Merchant Owns

```text
Merchant
│
├── Users
├── Wallet
├── Bank Accounts
├── Payments
├── Payouts
├── Refunds
├── Settlements
├── Ledger Account
├── Compliance Profile
└── Webhook Configuration
```

---

## Merchant Lifecycle

```text
REGISTERED
      │
      ▼
PENDING_KYC
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

## Merchant Attributes

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
| daily_transaction_limit | Maximum transaction limit per day |
| created_at | Created timestamp |
| updated_at | Updated timestamp |

---

## Business Rules

### Identity

- Merchant ID is immutable.
- Merchant Reference is immutable.
- Entity Type cannot be changed.

### Users

- Every Merchant must always have one Owner.
- Ownership must be transferred before removing the current Owner.

### Wallet

- Every Merchant owns exactly one Wallet in Version 1.
- Wallet cannot exist without a Merchant.

### Bank Accounts

- Maximum three linked bank accounts.
- Exactly one Primary Bank Account.
- Primary account cannot be removed until another account is marked as Primary.

### Merchant Status

- Merchant cannot become ACTIVE until KYC is approved.
- Suspended or Blocked Merchants cannot initiate new Payments, Payouts or Refunds.

---

## Domain Events

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

## APIs

### Merchant

```
POST   /merchants
GET    /merchants/{merchantId}
PATCH  /merchants/{merchantId}
POST   /merchants/{merchantId}/activate
POST   /merchants/{merchantId}/deactivate
```

### Users

```
POST   /merchants/{merchantId}/users
DELETE /merchants/{merchantId}/users/{userId}
```

### Bank Accounts

```
POST   /merchants/{merchantId}/bank-accounts
DELETE /merchants/{merchantId}/bank-accounts/{bankAccountId}
```

### Webhooks

```
PATCH /merchants/{merchantId}/webhook
```

---

## Business Scenarios

### Merchant Registration

1. Merchant registers.
2. Merchant enters `REGISTERED` state.
3. KYC process starts.
4. Merchant cannot perform financial operations.
5. Once KYC is approved, Merchant becomes `ACTIVE`.
6. Wallet is created automatically.

---

### Add Merchant User

1. Merchant Owner invites a User.
2. User accepts the invitation.
3. Role is assigned.
4. User can now access Merchant resources.

---

### Link Bank Account

1. Merchant submits bank account details.
2. System validates ownership.
3. Bank account is linked.
4. Merchant can link a maximum of three bank accounts.

---

## Relationships

| Entity | Relationship |
|----------|--------------|
| User | 1:N |
| Wallet | 1:1 |
| Bank Account | 1:N |
| Payment | 1:N |
| Payout | 1:N |
| Refund | 1:N |
| Settlement | 1:N |
| Ledger Account | 1:1 |
| Compliance Profile | 1:1 |
| Webhook Configuration | 1:1 |