# User Internal API

## Overview

The User Internal API is used by SamPay internal systems to manage Merchant Users.

A User represents a human identity associated with exactly one Merchant.

These APIs are consumed by the SamPay Admin Portal and Operations Team.

These APIs are **not** exposed to Merchants.

---

# Authentication

All APIs require an authenticated SamPay Internal User.

```http
Authorization: Bearer <internal_access_token>
```

---

# Endpoints

| Method | Endpoint | Description |
|----------|----------|-------------|
| POST | `/api/v1/internal/users` | Create User |
| GET | `/api/v1/internal/users` | List Users |
| GET | `/api/v1/internal/users/{userId}` | Get User Details |
| PATCH | `/api/v1/internal/users/{userId}` | Update User |
| DELETE | `/api/v1/internal/users/{userId}` | Deactivate User |

---

# Create User

## Endpoint

```http
POST /api/v1/internal/users
```

---

## Headers

```http
Authorization: Bearer <internal_access_token>

Idempotency-Key: <uuid>

Content-Type: application/json
```

---

## Request Body

```json
{
    "merchant_id": "uuid",
    "first_name": "John",
    "last_name": "Doe",
    "email": "john.doe@merchant.com",
    "role": "OWNER"
}
```

---

## Success Response

```http
201 Created
```

```json
{
    "success": true,
    "request_id": "uuid",
    "data": {
        "user_id": "uuid",
        "merchant_id": "uuid",
        "status": "INVITED"
    }
}
```

---

## Business Validations

- Merchant must exist.
- Email must be unique.
- Role must be valid.
- User is created in the `INVITED` state.
- Authentication onboarding is initiated separately.

---

## Possible Errors

```text
400 Bad Request

404 Merchant Not Found

409 User Already Exists

500 Internal Server Error
```

---

# List Users

## Endpoint

```http
GET /api/v1/internal/users
```

---

## Query Parameters

```text
?page=1

&page_size=20

&merchant_id=<merchant_id>

&status=ACTIVE

&sort=created_at

&order=desc
```

---

## Success Response

```json
{
    "success": true,
    "request_id": "uuid",
    "data": [
        {
            "user_id": "uuid",
            "merchant_id": "uuid",
            "first_name": "John",
            "last_name": "Doe",
            "email": "john.doe@merchant.com",
            "role": "OWNER",
            "status": "ACTIVE"
        }
    ],
    "meta": {
        "page": 1,
        "page_size": 20,
        "total_records": 50,
        "total_pages": 3
    }
}
```

---

# Get User

## Endpoint

```http
GET /api/v1/internal/users/{userId}
```

---

## Success Response

```json
{
    "success": true,
    "request_id": "uuid",
    "data": {
        "user_id": "uuid",
        "merchant_id": "uuid",
        "first_name": "John",
        "last_name": "Doe",
        "email": "john.doe@merchant.com",
        "role": "OWNER",
        "status": "ACTIVE",
        "created_at": "2026-08-03T10:30:45Z",
        "updated_at": "2026-08-03T10:30:45Z"
    }
}
```

---

# Update User

## Endpoint

```http
PATCH /api/v1/internal/users/{userId}
```

---

## Request Body

```json
{
    "first_name": "Johnny",
    "last_name": "Doe",
    "role": "ADMIN"
}
```

---

## Editable Fields

- first_name
- last_name
- role

The following fields cannot be modified:

- merchant_id
- email

---

## Business Rules

- Merchant ownership cannot be changed.
- Role must be one of the supported system roles.
- If changing the OWNER role, ownership transfer rules must be enforced.

---

# Deactivate User

## Endpoint

```http
DELETE /api/v1/internal/users/{userId}
```

---

## Success Response

```http
204 No Content
```

---

## Business Rules

- Users are soft deleted by changing their status to `DEACTIVATED`.
- The last active OWNER of a Merchant cannot be deactivated.
- An OWNER must transfer ownership before being deactivated.

---

# Notes

- These APIs are intended for SamPay internal use only.
- Merchants cannot access these endpoints.
- Authentication onboarding is handled by the Authentication domain.
- Password management is outside the scope of the User domain.
- All APIs follow the standards defined in `common.md`.