# Compliance Domain

## Overview

The Compliance domain is responsible for validating Merchants, Users and financial transactions against SamPay's internal compliance policies.

Compliance acts as a decision engine that determines whether a request is allowed to proceed.

It does not modify financial data, move money, or persist compliance records. Its sole responsibility is to evaluate predefined rules and return a decision.

---

# Responsibilities

The Compliance domain is responsible for:

- Validating Merchant onboarding requests.
- Validating User onboarding requests.
- Validating Payments.
- Validating Refunds.
- Validating Payouts.
- Evaluating internal compliance rules.
- Returning an approval decision.

The Compliance domain is **not** responsible for:

- Processing Payments.
- Processing Refunds.
- Processing Payouts.
- Managing Wallets.
- Managing Ledger entries.
- Managing Settlements.
- Managing Merchant or User state.

---

# Compliance Decision

Compliance always returns one of the following decisions:

```text
APPROVED

or

REJECTED
```

No additional financial actions are performed by the Compliance domain.

If a request is rejected, the calling domain immediately terminates the workflow.

---

# Supported Requests

Compliance may be invoked by:

- Merchant
- User
- Payment
- Refund
- Payout

Every domain remains responsible for its own business logic after receiving the compliance decision.

---

# Compliance Workflow

```text
Incoming Request
        │
        ▼
Build Compliance Context
        │
        ▼
Execute Compliance Rules
        │
        ├───────────────┐
        │               │
    APPROVED       REJECTED
        │               │
        ▼               ▼
Continue         Stop Processing
Business Flow
```

---

# Compliance Rules (Version 1)

The following rules are supported in Version 1:

## Merchant Blacklist

Reject requests from Merchants that have been blacklisted.

---

## User Blacklist

Reject requests initiated by blacklisted Users.

---

## Recipient Blacklist

Reject transactions involving blacklisted Recipients.

---

## Sanctioned Country Check

Reject requests involving sanctioned countries.

---

## Supported Country Validation

Ensure the request originates from or targets a country supported by SamPay.

---

## Supported Currency Validation

Validate that the transaction uses a supported currency.

Version 1 supports INR only.

---

## Internal Policy Checks

Additional internal policies may be evaluated where applicable.

These policies remain configurable within the Compliance domain.

---

# Business Rules

## Stateless Service

Compliance maintains no financial state.

Every evaluation is independent.

---

## No Financial Operations

Compliance never:

- Reserves Wallet balances.
- Creates Ledger entries.
- Creates Settlements.
- Updates financial records.

---

## No Database Tables

Compliance does not own any database tables.

If a domain wishes to persist compliance-related information, it may store it within its own metadata.

Example:

```json
{
  "compliance": {
    "result": "APPROVED",
    "version": "v1"
  }
}
```

---

## Rule Execution

Compliance evaluates rules sequentially.

Processing stops immediately when a rule returns **REJECTED**.

If all rules pass, the request is **APPROVED**.

---

## Extensibility

Compliance is designed as a rule engine.

Each compliance rule is implemented independently.

New rules can be introduced without modifying existing business domains.

Examples of future rules include:

- AML
- PEP Screening
- Fraud Detection
- Risk Scoring
- Velocity Monitoring
- External Compliance Providers

---

# APIs

Compliance is an internal domain.

It exposes internal validation interfaces only.

Typical operations include:

```text
ValidateMerchant()

ValidateUser()

ValidatePayment()

ValidateRefund()

ValidatePayout()
```

These interfaces are consumed internally by the corresponding business domains.

---

# Relationships

| Entity | Relationship |
|----------|--------------|
| Merchant | Validates onboarding and updates |
| User | Validates onboarding and updates |
| Payment | Validates payment requests |
| Refund | Validates refund requests |
| Payout | Validates payout requests |

---

# Notes

- Compliance is a stateless internal service.
- Compliance returns only **APPROVED** or **REJECTED**.
- Compliance owns no financial data.
- Compliance performs no balance updates.
- Compliance performs no ledger operations.
- Compliance performs no settlement operations.
- Business domains remain responsible for handling approved or rejected requests.
- The Compliance rule set is designed to evolve without impacting existing business workflows.