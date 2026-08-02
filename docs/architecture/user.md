# User Domain

## Overview

A User is a human identity that belongs to exactly one Merchant and performs actions on behalf of that Merchant.

A User never owns financial resources. All financial resources belong to the Merchant.

Authentication is handled separately by the Authentication domain, while the User domain is responsible for managing the business identity of the user within a Merchant.

---

# Responsibilities

The User domain is responsible for:

- Managing User profiles
- Managing Merchant association
- Managing assigned Roles
- Managing User lifecycle
- Managing Merchant ownership

The User domain is **not** responsible for:

- Authentication
- Authorization
- Password Management
- Session Management
- Payment Processing
- Wallet Management
- Bank Account Management
- Financial Operations

---

# User Owns

```text
User
│
├── Profile
├── Merchant Association
└── Assigned Role
```

---

# User Lifecycle

```text
INVITED
    │
    ▼
PENDING_ACTIVATION
    │
    ▼
ACTIVE
   │
   ├──────────────► LOCKED
   │
   └──────────────► DEACTIVATED
```

---

# User Attributes

| Field | Description |
|---------|-------------|
| id | Internal UUID |
| user_reference | Public User Identifier |
| merchant_id | Merchant to which the User belongs |
| role_id | Assigned Role |
| is_owner | Indicates whether the User is the Merchant Owner |
| first_name | User first name |
| last_name | User last name |
| email | Primary business email |
| status | User lifecycle status |
| created_at | Record creation timestamp |
| updated_at | Record update timestamp |

---

# Predefined Roles

SamPay provides a fixed set of system-defined roles.

Roles are implemented as predefined application constants and are not stored as a separate domain in Version 1.

## OWNER

Represents the Merchant Owner.

- Every Merchant must have exactly one OWNER.
- Ownership is additionally represented by the `is_owner` attribute.
- Ownership can only be transferred to another ACTIVE User.

---

## ADMIN

Responsible for Merchant administration.

Typical responsibilities include:

- Manage Merchant profile
- Invite and manage Users
- Assign Roles
- Manage linked Bank Accounts
- Configure Webhooks

---

## FINANCE

Responsible for financial operations.

Typical responsibilities include:

- Create Payments
- Create Payouts
- Create Refunds
- View Wallet
- View Ledger

---

## OPERATIONS

Responsible for day-to-day operational work.

Typical responsibilities include:

- Create Payments
- Create Refunds
- View Transactions
- View Wallet

---

# Role Business Rules

- Every User must have exactly one Role.
- Roles are predefined by SamPay.
- Roles cannot be created.
- Roles cannot be modified.
- Roles cannot be deleted.
- Merchants can only assign predefined Roles.
- Authorization is enforced by the application based on the assigned Role.

---

# Business Rules

## Identity

- Every User belongs to exactly one Merchant.
- A User cannot exist without a Merchant.
- A User cannot belong to multiple Merchants.
- A User cannot be transferred to another Merchant.

---

## Ownership

- Every Merchant must always have exactly one Owner.
- Ownership is represented by the `is_owner` attribute.
- Ownership can only be transferred to an ACTIVE User.
- Ownership transfer is atomic, ensuring that a Merchant never has zero or multiple Owners.
- Ownership is independent of the assigned Role.

---

## Roles

- Every User must have exactly one assigned Role.
- Roles are predefined by SamPay.
- Merchants may assign predefined Roles to their Users.
- Merchants cannot create, modify or delete Roles in Version 1.
- Users inherit permissions from their assigned Role.

---

## Status

- Only ACTIVE Users may perform financial operations.
- INVITED Users cannot access the platform.
- PENDING_ACTIVATION Users must complete onboarding before accessing the platform.
- LOCKED Users cannot authenticate or perform operations.
- DEACTIVATED Users are permanently inactive.

---

## Email

- Every User must have a unique email address.
- Email is used for authentication.
- Email is the primary business contact for the User.
- Transaction notifications are delivered to this email.
- Email verification is required before the User becomes ACTIVE.

---

# Domain Events

- UserInvited
- UserActivated
- UserLocked
- UserDeactivated
- UserProfileUpdated
- UserRoleAssigned
- UserRoleChanged
- MerchantOwnershipTransferred

---

# APIs

## User Management

```http
POST   /merchants/{merchantId}/users
GET    /users/{userId}
PATCH  /users/{userId}
DELETE /users/{userId}
```

---

# Business Scenarios

## Invite User

```text
Merchant Owner
        │
        ▼
Invite User
        │
        ▼
Compliance Validation
        │
        ▼
User Created (INVITED)
        │
        ▼
Invitation Email Sent
        │
        ▼
User Accepts Invitation
        │
        ▼
Authentication Setup
        │
        ▼
Email Verification
        │
        ▼
User Activated
```

---

## Update User Role

```text
Merchant Owner
        │
        ▼
Select User
        │
        ▼
Assign Predefined Role
        │
        ▼
Permissions Updated
```

---

## Transfer Merchant Ownership

```text
Current Owner
        │
        ▼
Select ACTIVE User
        │
        ▼
Transfer Ownership
        │
        ▼
Previous Owner
is_owner = false
        │
        ▼
New Owner
is_owner = true
```

Ownership transfer is performed atomically to ensure only one Owner exists at any point in time.

---

# Relationships

| Entity | Relationship |
|----------|--------------|
| Merchant | N : 1 |
| Role | N : 1 |
| Authentication | 1 : 1 |

---

# Notes

- Every User belongs to exactly one Merchant.
- A User never owns financial resources.
- Financial operations are always performed on behalf of the Merchant.
- Authentication is delegated to the Authentication domain.
- Authorization is determined by the assigned Role.
- Merchant ownership is represented by the `is_owner` attribute and is independent of the assigned Role.

---

# Version

**Architecture Status:** ✅ Frozen (V1)

**Last Updated:** 2026-08-03