# Merchant Internal API

## Overview

The Merchant Internal API is used exclusively by SamPay internal systems to manage Merchant onboarding and lifecycle.

These APIs are consumed by the SamPay Admin Portal, Operations Team and Customer Support.

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
| POST | `/api/v1/internal/merchants` | Create Merchant |
| GET | `/api/v1/internal/merchants` | List Merchants |
| GET | `/api/v1/internal/merchants/{merchantId}` | Get Merchant Details |
| PATCH | `/api/v1/internal/merchants/{merchantId}` | Update Merchant |
| PATCH | `/api/v1/internal/merchants/{merchantId}/status` | Update Merchant Status |
| PATCH | `/api/v1/internal/merchants/{merchantId}/kyc` | Update KYC Status |

---

# Create Merchant

## Endpoint

```http
POST /api/v1/internal/merchants
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
    "legal_name": "ABC Technologies Pvt Ltd",
    "notification_email": "finance@abc.com",
    "entity_type": "CORPORATE"
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
        "merchant_id": "uuid",
        "merchant_reference": "MER000001",
        "status": "REGISTERED",
        "kyc_status": "PENDING"
    }
}
```

---

## Business Validations

- Legal name is mandatory.
- Notification email must be unique.
- Entity type must be valid.
- Merchant Reference is generated automatically.
- Wallet is automatically created.
- Merchant starts in the `REGISTERED` state.

---

## Possible Errors

```text
400 Bad Request

409 Merchant Already Exists

500 Internal Server Error
```

---

# List Merchants

## Endpoint

```http
GET /api/v1/internal/merchants
```

---

## Query Parameters

```text
?page=1

&page_size=20

&status=ACTIVE

&kyc_status=APPROVED

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
            "merchant_id": "uuid",
            "merchant_reference": "MER000001",
            "legal_name": "ABC Technologies Pvt Ltd",
            "status": "ACTIVE",
            "kyc_status": "APPROVED"
        }
    ],
    "meta": {
        "page": 1,
        "page_size": 20,
        "total_records": 100,
        "total_pages": 5
    }
}
```

---

# Get Merchant

## Endpoint

```http
GET /api/v1/internal/merchants/{merchantId}
```

---

## Success Response

```json
{
    "success": true,
    "request_id": "uuid",
    "data": {
        "merchant_id": "uuid",
        "merchant_reference": "MER000001",
        "legal_name": "ABC Technologies Pvt Ltd",
        "notification_email": "finance@abc.com",
        "entity_type": "CORPORATE",
        "status": "ACTIVE",
        "kyc_status": "APPROVED",
        "created_at": "2026-08-03T10:30:45Z",
        "updated_at": "2026-08-03T10:30:45Z"
    }
}
```

---

# Update Merchant

## Endpoint

```http
PATCH /api/v1/internal/merchants/{merchantId}
```

---

## Request Body

```json
{
    "legal_name": "ABC Technologies Limited",
    "notification_email": "accounts@abc.com"
}
```

---

## Editable Fields

- legal_name
- notification_email

The following fields cannot be modified:

- merchant_reference
- entity_type

---

# Update Merchant Status

## Endpoint

```http
PATCH /api/v1/internal/merchants/{merchantId}/status
```

---

## Request Body

```json
{
    "status": "SUSPENDED"
}
```

---

## Allowed Status Values

```text
REGISTERED

PENDING_KYC

ACTIVE

SUSPENDED

BLOCKED

CLOSED
```

---

## Business Rules

- Status transitions must follow the Merchant lifecycle.
- Only authorized SamPay internal users can update merchant status.

---

# Update KYC Status

## Endpoint

```http
PATCH /api/v1/internal/merchants/{merchantId}/kyc
```

---

## Request Body

```json
{
    "kyc_status": "APPROVED"
}
```

---

## Allowed Values

```text
PENDING

APPROVED

REJECTED
```

---

## Business Rules

- Merchant cannot become `ACTIVE` until KYC is `APPROVED`.
- Updating KYC status may trigger Merchant lifecycle transitions.

---

# Notes

- These APIs are intended for SamPay internal use only.
- Merchants cannot access these endpoints.
- Merchant onboarding is initiated through the Create Merchant API.
- Wallet creation is automatic during Merchant creation.
- All APIs follow the standards defined in `common.md`.