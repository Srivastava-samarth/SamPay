# User Domain

## Overview

A User is a human identity that belongs to exactly one Merchant and performs actions on behalf of that Merchant.

A User does not own financial resources. All financial resources belong to the Merchant.

---

# Responsibilities

The User domain is responsible for:

- Managing user profile
- Managing merchant association
- Managing user lifecycle
- Managing user roles

The User domain is **not** responsible for:

- Authentication
- Authorization
- Password Management
- Session Management
- Payment Processing
- Wallet Management
- Bank Account Management

---

# User Owns

```text
User
│
├── Profile
├── Merchant Association
├── Role
└── Audit Information
```

---

# User Lifecycle

```text
INVITED
    │
    ▼
ACTIVE
   │
   ▼
LOCKED
   │
   ▼
DEACTIVATED
```

---

# User Attributes

| Field | Description |
|---------|-------------|
| id | Internal UUID |
| merchant_id | Merchant to which the user belongs |
| role_id | Assigned role |
| user_reference | Public User Identifier |
| first_name | User first name |
| last_name | User last name |
| email | Primary business email |
| status | User lifecycle status |
| created_at | Record creation timestamp |
| updated_at | Record update timestamp |

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
- Ownership must be transferred before the current Owner can be removed.

---

## Roles

- Every User must have exactly one Role.
- Roles are predefined by SamPay.
- Merchants can assign predefined roles to their Users.
- Merchants cannot create, update or delete Roles in Version 1.
- Users inherit permissions from their assigned Role.

---

## Status

- Only ACTIVE users can perform financial operations.
- INVITED users cannot access the platform until onboarding is completed.
- LOCKED users cannot access the platform.
- DEACTIVATED users are permanently inactive.

---

## Email

- Every User must have a unique email address.
- The email acts as the primary business contact for the User.
- Transaction notifications are sent to this email.
- Authentication is handled separately by the Authentication domain.

---

# Domain Events

- UserInvited
- UserActivated
- UserLocked
- UserDeactivated
- UserProfileUpdated
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

1. Merchant Owner invites a new User.
2. User is created in the INVITED state.
3. Authentication onboarding starts separately.
4. User becomes ACTIVE after completing onboarding.

---

## Update User Role

1. Merchant Owner selects a User.
2. Assigns a predefined Role.
3. New permissions take effect immediately.

---

## Transfer Merchant Ownership

1. Current Owner selects another ACTIVE User.
2. Ownership is transferred.
3. Previous Owner is assigned another predefined Role.
4. New Owner receives full ownership privileges.

---

# Relationships

| Entity | Relationship |
|----------|--------------|
| Merchant | N : 1 |
| Role | N : 1 |
| Authentication | 1 : 1 |
| Audit Logs | 1 : N |