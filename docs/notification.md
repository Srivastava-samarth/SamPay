# Notification Domain

## Overview

The Notification domain is responsible for sending email notifications in response to business events occurring within SamPay.

It listens to events published by other domains, prepares the appropriate email template and delivers the notification to the intended recipient.

Notification is an asynchronous service and does not participate in any business workflow or financial transaction.

---

# Responsibilities

The Notification domain is responsible for:

- Listening to domain events.
- Selecting the appropriate email template.
- Populating template variables.
- Sending email notifications.
- Retrying failed email deliveries.
- Logging notification failures.

The Notification domain is **not** responsible for:

- Processing Payments.
- Processing Refunds.
- Processing Payouts.
- Managing Users.
- Managing Merchants.
- Managing Wallets.
- Managing Ledger entries.
- Managing Settlements.
- Updating business data.

---

# Supported Channels

Version 1 supports:

- Email

Future versions may support:

- SMS
- Push Notifications
- WhatsApp
- In-App Notifications

---

# Notification Workflow

```text
Business Event
        │
        ▼
Notification Service
        │
        ▼
Select Email Template
        │
        ▼
Populate Template Variables
        │
        ▼
Send Email
        │
        ├───────────────┐
        │               │
    Success         Failure
        │               │
        ▼               ▼
      Done         Retry (Max 3 Times)
                        │
                        ▼
                 Log Failure
```

---

# Supported Events

Notification listens to events published by other domains.

Examples include:

## Merchant

- MerchantCreated
- MerchantUpdated

---

## User

- UserCreated
- UserInvited
- PasswordResetRequested

---

## Payment

- PaymentCompleted
- PaymentFailed

---

## Refund

- RefundCompleted
- RefundFailed

---

## Payout

- PayoutCompleted
- PayoutFailed

---

## Settlement

- SettlementCompleted
- SettlementFailed

---

# Business Rules

## Asynchronous Processing

Notification processing is asynchronous.

Business operations never wait for email delivery.

---

## Failure Handling

Failure to send an email must never affect the originating business transaction.

For example:

```text
Payment Completed

↓

Email Failed

↓

Payment remains COMPLETED
```

---

## Retry Policy

Failed email deliveries are automatically retried.

Maximum retry attempts:

```text
3
```

If all retries fail:

- Stop retrying.
- Log the failure.

---

## Stateless Service

Notification owns no business state.

It does not maintain notification records or update business entities.

---

## No Database Tables

Version 1 does not require dedicated notification tables.

Notification consumes events and performs email delivery only.

---

# Email Templates

Each business event maps to a predefined email template.

Examples include:

- Welcome Email
- User Invitation
- Payment Receipt
- Payment Failure
- Refund Confirmation
- Refund Failure
- Payout Confirmation
- Payout Failure
- Settlement Confirmation

---

# APIs

Notification is an internal service.

No public APIs are exposed.

Notification operates exclusively through domain events.

---

# Relationships

| Domain | Event Source |
|----------|-------------|
| Merchant | Merchant events |
| User | User events |
| Payment | Payment events |
| Refund | Refund events |
| Payout | Payout events |
| Settlement | Settlement events |

---

# Notes

- Notification is an internal asynchronous service.
- Version 1 supports Email only.
- Business operations never depend on successful email delivery.
- Failed notifications are retried up to three times.
- Notification owns no business data and maintains no database tables.
- Notification reacts only to events published by other domains.