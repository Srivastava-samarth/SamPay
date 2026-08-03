# OAuth Token Table

## Overview

The `oauth_token` table stores authentication tokens issued to Users for accessing the SamPay platform.

For security reasons, SamPay never stores the raw access token. Instead, a cryptographic hash of the token is stored and used during authentication.

Business rules related to authentication are documented in `docs/architecture/authentication.md`.

---

# Table Name

```text
oauth_token
```

---

# Columns

| Column | Type | Nullable | Default | Description |
|---------|------|----------|---------|-------------|
| id | UUID | No | gen_random_uuid() | Internal primary key |
| user_id | UUID | No | - | User owning the token |
| token_hash | TEXT | No | - | SHA-256 hash of the issued access token |
| expires_at | TIMESTAMPTZ | No | - | Token expiration timestamp |
| is_used | BOOLEAN | No | false | Indicates whether the token has already been consumed |
| created_at | TIMESTAMPTZ | No | now() | Record creation timestamp |

---

# Primary Key

```sql
PRIMARY KEY (id)
```

---

# Foreign Keys

```sql
user_id
REFERENCES "user"(id)
```

---

# Unique Constraints

```sql
UNIQUE (token_hash)
```

---

# Check Constraints

None.

---

# Indexes

No additional indexes are defined.

The following indexes are automatically created by PostgreSQL:

- Primary Key (`id`)
- Unique Constraint (`token_hash`)

Additional indexes should only be introduced when justified by production query patterns.

---

# Referenced By

The Authentication service uses this table to:

- Validate access tokens
- Check token expiration
- Prevent token reuse

---

# Authentication Flow

```text
User Login
      │
      ▼
Generate JWT
      │
      ▼
Generate SHA-256(Token)
      │
      ▼
Store token_hash
      │
      ▼
Return JWT to User
```

During authentication:

```text
Incoming JWT
      │
      ▼
Generate SHA-256(Token)
      │
      ▼
Find matching token_hash
      │
      ▼
Validate:
    • Token exists
    • Not expired
    • is_used = false
      │
      ▼
Authenticate User
      │
      ▼
Mark token as used
```

---

# Notes

- Every token belongs to exactly one User.
- Tokens are valid for **1 hour** from issuance.
- Tokens are **single-use** and become invalid immediately after successful authentication.
- Expired tokens are invalid even if they have not been used.
- Only the SHA-256 hash of the token is stored in the database.
- The raw JWT is never persisted.
- Authentication is handled by the Authentication service and is independent of financial domains.