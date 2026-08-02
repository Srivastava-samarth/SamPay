# Authentication Domain

## Overview

The Authentication domain is responsible for verifying user identity and issuing secure access credentials for the SamPay platform.

Authentication determines **who the user is**.

Authorization determines **what the user is allowed to do**.

Authentication and Authorization are separate concerns.

---

# Responsibilities

The Authentication domain is responsible for:

- User Login.
- User Logout.
- Access Token generation.
- Refresh Token generation.
- Refresh Token rotation.
- Password verification.
- Password hashing.
- Password reset.
- Session management.

The Authentication domain is **not** responsible for:

- User Management.
- Role Management.
- Permission Management.
- Merchant Management.
- Business logic.
- Financial operations.

---

# Authentication Flow

```text
Email
        │
        ▼
Password
        │
        ▼
Verify Credentials
        │
        ▼
Generate Access Token
        │
        ▼
Generate Refresh Token
        │
        ▼
Store Refresh Token Hash
        │
        ▼
Return Tokens
```

---

# Access Token

Access Tokens are JWTs used to authenticate API requests.

Properties:

- Valid for 1 hour.
- Stateless.
- Not stored in the database.
- Signed by SamPay.

Every authenticated API request must include a valid Access Token.

---

# Refresh Token

Refresh Tokens are used to obtain a new Access Token after the previous one expires.

Properties:

- Single-use only.
- Rotated after every successful refresh.
- Stored as a hash.
- Revocable.
- Managed per user session.

A Refresh Token becomes permanently invalid immediately after it has been used.

---

# Refresh Flow

```text
Refresh Token
        │
        ▼
Hash Token
        │
        ▼
Lookup Refresh Token
        │
        ▼
Validate
        │
        ▼
Revoke Old Token
        │
        ▼
Generate New Access Token
        │
        ▼
Generate New Refresh Token
        │
        ▼
Store New Refresh Token Hash
        │
        ▼
Return New Tokens
```

---

# Logout Flow

```text
Logout Request
        │
        ▼
Locate Refresh Token
        │
        ▼
Mark Revoked
        │
        ▼
Logout Complete
```

---

# Logout From All Devices

All active sessions belonging to the user are revoked.

Subsequent refresh attempts require the user to authenticate again.

---

# Password Management

Passwords are never stored in plain text.

Passwords are hashed using bcrypt before storage.

Authentication compares the incoming password with the stored password hash.

---

# Password Reset

Forgot Password flow:

```text
Request Password Reset
        │
        ▼
Generate Reset Token / OTP
        │
        ▼
Send Email
        │
        ▼
Verify Token
        │
        ▼
Reset Password
```

Password reset invalidates all active refresh tokens.

The user must log in again after resetting the password.

---

# Authorization

Authentication identifies the user.

Authorization verifies whether the authenticated user may perform the requested action.

Authorization is based on the User's assigned Role.

Permissions are implemented in application code.

No permission tables are maintained.

Typical request flow:

```text
Incoming Request
        │
        ▼
Authenticate JWT
        │
        ▼
Load User
        │
        ▼
Validate Role
        │
        ▼
Execute Business Logic
```

---

# User Refresh Token Structure

Each active login session owns one Refresh Token.

| Field | Description |
|--------|-------------|
| id | Internal UUID |
| user_id | Owner User |
| refresh_token_hash | Hashed Refresh Token |
| expires_at | Expiration timestamp |
| last_used_at | Last successful refresh |
| revoked_at | Revocation timestamp |
| created_at | Record creation timestamp |
| updated_at | Record update timestamp |

---

# Business Rules

## Access Token

- Valid for 1 hour.
- Never stored in the database.
- Used to authenticate API requests.

---

## Refresh Token

- Single-use.
- Rotated after every refresh.
- Stored only as a hash.
- Can be revoked.
- One refresh generates a completely new refresh token.

---

## Multiple Sessions

A User may have multiple active sessions.

Each session owns an independent Refresh Token.

Revoking one session does not affect the others.

---

## Logout

Logout revokes the corresponding Refresh Token.

Access Tokens naturally expire after one hour.

---

## Logout From All Devices

Revokes every active Refresh Token belonging to the User.

---

## Password Security

- Passwords are hashed using bcrypt.
- Plain-text passwords are never stored.
- Password hashes are never returned.

---

## Session Security

- Refresh Tokens are stored only as hashes.
- Used Refresh Tokens cannot be reused.
- Revoked Refresh Tokens cannot be used again.

---

# APIs

```http
POST /auth/login

POST /auth/refresh

POST /auth/logout

POST /auth/logout-all

POST /auth/forgot-password

POST /auth/reset-password
```

---

# Relationships

| Entity | Relationship |
|----------|--------------|
| User | Authenticated entity |
| Role | Used for Authorization |

---

# Notes

- Authentication identifies the user.
- Authorization determines what the user can access.
- Access Tokens are valid for one hour.
- Refresh Tokens are single-use and rotated after every successful refresh.
- Refresh Tokens are stored only as hashes.
- Authentication owns session management.
- Authorization is implemented using Roles and application code.
- Version 1 uses Email and Password authentication.