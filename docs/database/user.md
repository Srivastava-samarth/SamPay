# User Table

## Overview

The `user` table stores every user associated with a Merchant.

Users perform actions on behalf of their Merchant but do not own any financial resources.

Business rules related to User lifecycle, permissions and ownership are documented in `docs/architecture/user.md`.

---

# Table Name

```text
user
```

---

# Columns

| Column | Type | Nullable | Default | Description |
|---------|------|----------|---------|-------------|
| id | UUID | No | gen_random_uuid() | Internal primary key |
| user_reference | VARCHAR(30) | No | - | Public user identifier |
| merchant_id | UUID | No | - | Merchant owning the user |
| first_name | VARCHAR(100) | No | - | User first name |
| last_name | VARCHAR(100) | No | - | User last name |
| email | VARCHAR(255) | No | - | Business email address |
| role | user_role | No | 'OPERATIONS' | Assigned role |
| status | user_status | No | 'INVITED' | User lifecycle status |
| created_at | TIMESTAMPTZ | No | now() | Record creation timestamp |
| updated_at | TIMESTAMPTZ | No | now() | Record last update timestamp |

---

# Primary Key

```sql
PRIMARY KEY (id)
```

---

# Foreign Keys

```sql
merchant_id
REFERENCES merchant(id)
```

ON DELETE RESTRICT

---

# Unique Constraints

```sql
UNIQUE (user_reference)

UNIQUE (email)
```

---

# Indexes

| Index | Purpose |
|---------|----------|
| merchant_id | Fetch Merchant users |
| email | Fast user lookup |
| status | Active user filtering |
| role | Role-based filtering |
| created_at | Reporting and ordering |

---

# Enums Used

## user_role

```text
OWNER
ADMIN
FINANCE
OPERATIONS
```

---

## user_status

```text
INVITED
ACTIVE
LOCKED
DEACTIVATED
```

---

# Referenced By

The following tables reference `user.id`:

- authentication

Future versions may additionally reference:

- notification
- audit_log

---

# Notes

- Every User belongs to exactly one Merchant.
- Users cannot be transferred between Merchants.
- Every User must have exactly one Role.
- Roles are stored as PostgreSQL enums in Version 1.
- Email addresses should be stored in lowercase.
- User lifecycle transitions are enforced by the application layer.
- Ownership transfer is handled by updating the `role` of the involved users within the same database transaction.