# Role Domain

## Overview

A Role defines the responsibilities of a User within a Merchant.

Every User must be assigned exactly one Role.

Roles are predefined by SamPay and can be assigned to Users by Merchant Administrators.

Permissions associated with each role are enforced by the application and are not stored in the database.

---

# Responsibilities

The Role domain is responsible for:

- Defining predefined system roles
- Providing role information
- Assigning roles to users

The Role domain is **not** responsible for:

- User Management
- Authentication
- Authorization Logic
- Merchant Management

---

# Role Attributes

| Field | Description |
|---------|-------------|
| id | Internal identifier |
| name | Unique role name |
| description | Business description of the role |
| created_at | Record creation timestamp |
| updated_at | Record update timestamp |

---

# Predefined Roles

## ADMIN

Merchant administrator responsible for managing the merchant and its resources.

Responsibilities:

- Manage merchant profile
- Invite and remove users
- Assign user roles
- Manage wallets
- Link and remove bank accounts
- Configure webhooks
- Create payments
- Create payouts
- Create refunds
- Perform settlements

---

## FINANCE

Responsible for managing the merchant's financial operations.

Responsibilities:

- Create payments
- Create payouts
- Create refunds
- Perform settlements
- View wallet
- View ledger
- Monitor balances

---

## OPERATIONS

Responsible for handling day-to-day merchant operations.

Responsibilities:

- Create payments
- Create refunds
- View transactions
- View wallet information

---

## AUDITOR

Responsible for auditing and monitoring merchant activities.

Responsibilities:

- View merchant information
- View transactions
- View wallet
- View ledger
- Read-only access to merchant resources

---

# Business Rules

- Every Role has a unique name.
- Roles are predefined by SamPay.
- Merchants can assign predefined Roles to Users.
- Merchants cannot create Roles.
- Merchants cannot update Roles.
- Merchants cannot delete Roles.
- Every User must have exactly one Role.
- Permissions are enforced by the application based on the assigned Role.

---

# Domain Events

- RoleAssigned
- RoleChanged

---

# APIs

## Role Management

```http
GET /roles
GET /roles/{roleId}
```

## User Role Management

```http
PATCH /users/{userId}/role
```

---

# Relationships

| Entity | Relationship |
|----------|--------------|
| User | 1 : N |