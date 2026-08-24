# Savio — API Contract

## Related Documents

- [README.md](../../README.md) — project overview, setup, and documentation index.
- [Business Requirements](../product/business-requirements.md) — business rules encoded by each endpoint.
- [Database Design](../database/database-design.md) — entities and fields behind request/response schemas.
- [Security Architecture](../engineering/security.md) — auth, CSRF, and error-handling requirements.
- [Testing Strategy](../engineering/testing-strategy.md) — how endpoint contracts are verified.
- [System Architecture](../architecture/system-architecture.md) — handler/service/repository flow this contract sits on.

## 1. Document Purpose

This document defines the REST API contract for Savio.

The purpose of this document is to translate Savio's:

- product foundation,
- business requirements,
- user flows,
- and database design

into a concrete HTTP API contract that can be implemented by the backend and consumed consistently by the frontend.

This document defines:

- API conventions,
- authentication behavior,
- CSRF behavior,
- request and response envelopes,
- validation behavior,
- error semantics,
- pagination,
- filtering,
- sorting,
- financial resource endpoints,
- workflow/action endpoints,
- forecast endpoints,
- scenario simulation endpoints,
- AI endpoints,
- notification endpoints,
- and administrative/supporting endpoints.

The API must preserve the central Savio principle:

> **Finance Engine calculates. AI interprets. User decides.**

The backend remains the authoritative source for:

- authentication,
- authorization,
- financial state,
- validation,
- business rules,
- calculations,
- and state transitions.

---

# 2. API Style

Savio uses:

```text
REST
+
JSON
+
Versioned API
```

Base path:

```text
/api/v1
```

Example:

```text
GET /api/v1/accounts
POST /api/v1/transactions
GET /api/v1/forecast
```

---

# 3. Content Type

Default request content type:

```http
Content-Type: application/json
```

Default response content type:

```http
Content-Type: application/json
```

File upload endpoints may use:

```http
multipart/form-data
```

---

# 4. Authentication Model

Authentication is cookie-based.

Savio should not store authentication tokens in:

```text
localStorage
sessionStorage
```

Recommended authentication model:

```text
Short-lived access token
+
Rotating refresh token
+
Server-side refresh session
```

Cookies:

```text
access_token
refresh_token
csrf_token
```

Exact cookie names may be configuration-driven.

---

# 5. Authentication Cookie Behavior

Recommended access cookie:

```text
HttpOnly = true
Secure = true in production
SameSite = Lax
Path = /api
Max-Age = short-lived
```

Recommended refresh cookie:

```text
HttpOnly = true
Secure = true in production
SameSite = Lax
Path = /api/v1/auth/refresh
Max-Age = refresh session lifetime
```

CSRF token cookie:

```text
HttpOnly = false
Secure = true in production
SameSite = Lax
Path = /
```

The frontend must be able to read the CSRF token and send it in a request header.

---

# 6. CSRF Protection

State-changing requests require CSRF protection.

Protected methods:

```text
POST
PUT
PATCH
DELETE
```

Recommended header:

```http
X-CSRF-Token: <token>
```

Conceptual validation:

```text
CSRF Cookie
+
CSRF Header
+
Server Validation
```

A signed double-submit token or equivalent session-bound design is preferred.

Invalid CSRF:

```http
403 Forbidden
```

Error code:

```text
CSRF_TOKEN_INVALID
```

---

# 7. Standard Success Response

For resource responses:

```json
{
  "success": true,
  "data": {},
  "message": "Success"
}
```

The `message` field may be omitted for simple reads if the project chooses that convention.

Consistency is more important than including unnecessary text.

---

# 8. Standard Collection Response

Example:

```json
{
  "success": true,
  "data": [
    {}
  ],
  "meta": {
    "page": 1,
    "limit": 20,
    "total": 142,
    "total_pages": 8
  }
}
```

---

# 9. Standard Error Response

All application errors should follow a consistent structure.

```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "details": {}
  },
  "message": "The request contains invalid data."
}
```

---

# 10. Field Validation Error

Example:

```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "details": {
      "amount": [
        "Amount must be greater than zero."
      ],
      "category_id": [
        "Category is required."
      ]
    }
  },
  "message": "Please correct the highlighted fields."
}
```

The frontend should map these errors to form fields.

---

# 11. Request ID

Every API response should expose a request identifier.

Recommended response header:

```http
X-Request-ID: req_abc123
```

The same request ID should appear in structured backend logs.

For unexpected errors, the frontend may display:

```text
Reference:
req_abc123
```

without exposing technical details.

---

# 12. HTTP Status Codes

Savio should use HTTP status codes consistently.

```text
200 OK
→ successful read or action

201 Created
→ resource created

204 No Content
→ successful operation with no response body

400 Bad Request
→ malformed request

401 Unauthorized
→ authentication required or expired

403 Forbidden
→ authenticated but not permitted / CSRF failure

404 Not Found
→ requested resource unavailable

409 Conflict
→ state conflict, duplicate, optimistic lock conflict

422 Unprocessable Entity
→ validation or business rule violation

429 Too Many Requests
→ rate limit reached

500 Internal Server Error
→ unexpected server failure

503 Service Unavailable
→ optional use for unavailable external dependency
```

---

# 13. Common Error Codes

Authentication:

```text
AUTHENTICATION_REQUIRED
INVALID_CREDENTIALS
SESSION_EXPIRED
SESSION_REVOKED
REFRESH_TOKEN_INVALID
CSRF_TOKEN_INVALID
```

Authorization:

```text
PERMISSION_DENIED
RESOURCE_ACCESS_DENIED
```

Validation:

```text
VALIDATION_ERROR
INVALID_UUID
INVALID_DATE_RANGE
INVALID_ENUM_VALUE
```

Financial:

```text
ACCOUNT_ARCHIVED
ACCOUNT_HAS_FINANCIAL_HISTORY
INVALID_CATEGORY_TYPE
TRANSFER_SAME_ACCOUNT
DUPLICATE_BUDGET
INVALID_RECURRING_DATE
INVALID_SCENARIO_MODIFICATION
VERSION_CONFLICT
```

AI:

```text
AI_PROVIDER_UNAVAILABLE
AI_TIMEOUT
AI_OUTPUT_INVALID
AI_FEATURE_DISABLED
INSUFFICIENT_CONTEXT
```

General:

```text
RESOURCE_NOT_FOUND
RESOURCE_CONFLICT
RATE_LIMIT_EXCEEDED
INTERNAL_SERVER_ERROR
```

---

# 14. Pagination

Default query:

```http
?page=1&limit=20
```

Rules:

```text
page >= 1
limit >= 1
limit <= 100
```

Default:

```text
page = 1
limit = 20
```

Response metadata:

```json
{
  "meta": {
    "page": 1,
    "limit": 20,
    "total": 142,
    "total_pages": 8
  }
}
```

---

# 15. Sorting

Sorting should use allowlisted fields.

Example:

```http
?sort=transaction_date&order=desc
```

Possible order:

```text
asc
desc
```

Invalid sort fields must return validation error rather than being passed directly to SQL.

---

# 16. Date Format

API date-only fields use:

```text
YYYY-MM-DD
```

Example:

```text
2026-08-24
```

Timestamp fields use ISO 8601.

Example:

```text
2026-08-24T15:30:00Z
```

---

# 17. Monetary Representation

API monetary fields should use decimal-safe representation.

Storage uses BIGINT integer minor units; the API transports decimal strings converted to/from minor units (2-decimal scale by default).

Recommended JSON representation:

```json
{
  "amount": "1250000.00"
}
```

Using strings prevents accidental floating-point precision issues across frontend/backend boundaries.

Frontend may format values as:

```text
Rp1.250.000
```

but should preserve decimal-safe internal handling.

---

# 18. API Resource Ownership

All financial endpoints are scoped to the authenticated user.

Example:

```http
GET /api/v1/accounts/:id
```

Backend logically resolves:

```text
account.id = :id
AND
account.user_id = authenticated_user.id
```

A valid UUID alone does not grant access.

---

# 19. Authentication Endpoints

Base:

```text
/api/v1/auth
```

Endpoints:

```text
POST /api/v1/auth/register
POST /api/v1/auth/login
POST /api/v1/auth/refresh
POST /api/v1/auth/logout
POST /api/v1/auth/logout-all
GET  /api/v1/auth/me
GET  /api/v1/auth/csrf
```

---

# 20. Register

```http
POST /api/v1/auth/register
```

Request:

```json
{
  "name": "Alex Wijaya",
  "email": "alex@example.com",
  "password": "StrongPassword123!",
  "password_confirmation": "StrongPassword123!"
}
```

Success:

```http
201 Created
```

Response:

```json
{
  "success": true,
  "data": {
    "user": {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "Alex Wijaya",
      "email": "alex@example.com",
      "timezone": "Asia/Jakarta",
      "default_currency": "IDR",
      "status": "ACTIVE"
    }
  },
  "message": "Account created successfully."
}
```

Cookies may be issued automatically.

Possible errors:

```text
VALIDATION_ERROR
EMAIL_ALREADY_EXISTS
RATE_LIMIT_EXCEEDED
```

---

# 21. Login

```http
POST /api/v1/auth/login
```

Request:

```json
{
  "email": "alex@example.com",
  "password": "StrongPassword123!"
}
```

Success:

```http
200 OK
```

Response:

```json
{
  "success": true,
  "data": {
    "user": {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "Alex Wijaya",
      "email": "alex@example.com",
      "timezone": "Asia/Jakarta",
      "default_currency": "IDR"
    }
  },
  "message": "Login successful."
}
```

Possible errors:

```text
INVALID_CREDENTIALS
RATE_LIMIT_EXCEEDED
```

---

# 22. Current User

```http
GET /api/v1/auth/me
```

Success:

```json
{
  "success": true,
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "Alex Wijaya",
    "email": "alex@example.com",
    "timezone": "Asia/Jakarta",
    "default_currency": "IDR",
    "locale": "id-ID"
  }
}
```

Unauthenticated:

```http
401 Unauthorized
```

---

# 23. CSRF Bootstrap

```http
GET /api/v1/auth/csrf
```

Purpose:

Issue or refresh the CSRF token required for state-changing requests.

Response:

```json
{
  "success": true,
  "data": {
    "csrf_token": "..."
  }
}
```

The token may also be provided through a readable cookie.

---

# 24. Refresh

```http
POST /api/v1/auth/refresh
```

No refresh token should be passed in JSON if it is stored in HttpOnly cookie.

Request:

```json
{}
```

Success:

```json
{
  "success": true,
  "data": null,
  "message": "Session refreshed."
}
```

Behavior:

```text
validate refresh cookie
↓
validate server session
↓
rotate refresh token
↓
issue new access token
↓
issue new refresh token
```

Possible errors:

```text
REFRESH_TOKEN_INVALID
SESSION_EXPIRED
SESSION_REVOKED
```

Failed refresh should cause the frontend to logout.

---

# 25. Logout

```http
POST /api/v1/auth/logout
```

Success:

```http
204 No Content
```

Behavior:

```text
revoke current refresh session
clear auth cookies
```

---

# 26. Logout All

```http
POST /api/v1/auth/logout-all
```

Success:

```json
{
  "success": true,
  "data": null,
  "message": "All sessions have been revoked."
}
```

---

# 27. Session Endpoints

Base:

```text
/api/v1/sessions
```

Endpoints:

```text
GET    /api/v1/sessions
DELETE /api/v1/sessions/:id
DELETE /api/v1/sessions
```

---

# 28. List Sessions

```http
GET /api/v1/sessions
```

Response:

```json
{
  "success": true,
  "data": [
    {
      "id": "session-uuid",
      "device_name": "MacBook Chrome",
      "ip_address": "127.0.0.1",
      "last_used_at": "2026-08-24T14:00:00Z",
      "created_at": "2026-08-20T10:00:00Z",
      "expires_at": "2026-08-31T10:00:00Z",
      "is_current": true
    }
  ]
}
```

---

# 29. Revoke Session

```http
DELETE /api/v1/sessions/:id
```

Response:

```http
204 No Content
```

The session must belong to the authenticated user.

---

# 30. User Profile Endpoints

Base:

```text
/api/v1/profile
```

Endpoints:

```text
GET   /api/v1/profile
PATCH /api/v1/profile
```

---

# 31. Update Profile

```http
PATCH /api/v1/profile
```

Request:

```json
{
  "name": "Alex Wijaya",
  "timezone": "Asia/Jakarta",
  "locale": "id-ID"
}
```

Response:

```json
{
  "success": true,
  "data": {
    "id": "user-uuid",
    "name": "Alex Wijaya",
    "email": "alex@example.com",
    "timezone": "Asia/Jakarta",
    "default_currency": "IDR",
    "locale": "id-ID"
  }
}
```

---

# 32. Settings Endpoints

Base:

```text
/api/v1/settings
```

Endpoints:

```text
GET   /api/v1/settings
PATCH /api/v1/settings
```

---

# 33. Get Settings

```http
GET /api/v1/settings
```

Response:

```json
{
  "success": true,
  "data": {
    "ai_insights_enabled": true,
    "ai_copilot_enabled": true,
    "notifications_enabled": true,
    "budget_warning_threshold": "80.00",
    "low_balance_threshold": "1000000.00"
  }
}
```

---

# 34. Update Settings

```http
PATCH /api/v1/settings
```

Request:

```json
{
  "ai_insights_enabled": true,
  "notifications_enabled": true,
  "budget_warning_threshold": "85.00",
  "low_balance_threshold": "1500000.00"
}
```

Response:

```json
{
  "success": true,
  "data": {
    "ai_insights_enabled": true,
    "notifications_enabled": true,
    "budget_warning_threshold": "85.00",
    "low_balance_threshold": "1500000.00"
  }
}
```

---

# 35. Accounts API

Base:

```text
/api/v1/accounts
```

Endpoints:

```text
GET    /api/v1/accounts
POST   /api/v1/accounts
GET    /api/v1/accounts/:id
PATCH  /api/v1/accounts/:id
POST   /api/v1/accounts/:id/archive
POST   /api/v1/accounts/:id/restore
POST   /api/v1/accounts/:id/reconcile
DELETE /api/v1/accounts/:id
```

---

# 36. List Accounts

```http
GET /api/v1/accounts
```

Query parameters:

```text
status
type
sort
order
page
limit
```

Example:

```http
GET /api/v1/accounts?status=ACTIVE&type=BANK&page=1&limit=20
```

Response:

```json
{
  "success": true,
  "data": [
    {
      "id": "account-uuid",
      "name": "BCA Main",
      "type": "BANK",
      "currency": "IDR",
      "initial_balance": "10000000.00",
      "current_balance": "14750000.00",
      "institution_name": "BCA",
      "status": "ACTIVE",
      "version": 3
    }
  ],
  "meta": {
    "page": 1,
    "limit": 20,
    "total": 1,
    "total_pages": 1
  }
}
```

---

# 37. Create Account

```http
POST /api/v1/accounts
```

Request:

```json
{
  "name": "BCA Main",
  "type": "BANK",
  "currency": "IDR",
  "initial_balance": "10000000.00",
  "institution_name": "BCA",
  "description": "Primary salary account"
}
```

Success:

```http
201 Created
```

Response:

```json
{
  "success": true,
  "data": {
    "id": "account-uuid",
    "name": "BCA Main",
    "type": "BANK",
    "currency": "IDR",
    "initial_balance": "10000000.00",
    "current_balance": "10000000.00",
    "status": "ACTIVE",
    "version": 1
  },
  "message": "Account created successfully."
}
```

---

# 38. Get Account

```http
GET /api/v1/accounts/:id
```

Response:

```json
{
  "success": true,
  "data": {
    "id": "account-uuid",
    "name": "BCA Main",
    "type": "BANK",
    "currency": "IDR",
    "initial_balance": "10000000.00",
    "current_balance": "14750000.00",
    "status": "ACTIVE",
    "version": 3,
    "summary": {
      "income_this_month": "12000000.00",
      "expense_this_month": "7250000.00",
      "incoming_transfer_this_month": "0.00",
      "outgoing_transfer_this_month": "0.00"
    }
  }
}
```

---

# 39. Update Account

```http
PATCH /api/v1/accounts/:id
```

Request:

```json
{
  "name": "BCA Personal",
  "institution_name": "BCA",
  "description": "Primary personal account",
  "version": 3
}
```

Success response includes incremented version:

```json
{
  "success": true,
  "data": {
    "id": "account-uuid",
    "name": "BCA Personal",
    "version": 4
  }
}
```

Version conflict:

```http
409 Conflict
```

```json
{
  "success": false,
  "error": {
    "code": "VERSION_CONFLICT"
  },
  "message": "The account was modified by another request."
}
```

---

# 40. Archive Account

```http
POST /api/v1/accounts/:id/archive
```

Request:

```json
{
  "version": 4
}
```

Response:

```json
{
  "success": true,
  "data": {
    "id": "account-uuid",
    "status": "ARCHIVED",
    "version": 5
  },
  "message": "Account archived."
}
```

---

# 41. Restore Account

```http
POST /api/v1/accounts/:id/restore
```

Request:

```json
{
  "version": 5
}
```

Response:

```json
{
  "success": true,
  "data": {
    "id": "account-uuid",
    "status": "ACTIVE",
    "version": 6
  }
}
```

---

# 42. Reconcile Account

```http
POST /api/v1/accounts/:id/reconcile
```

Request:

```json
{
  "actual_balance": "5000000.00",
  "reason": "Matched balance with bank application.",
  "version": 6
}
```

Backend calculates:

```text
current tracked balance
vs
actual balance
```

Response:

```json
{
  "success": true,
  "data": {
    "account": {
      "id": "account-uuid",
      "current_balance": "5000000.00",
      "version": 7
    },
    "adjustment": {
      "id": "transaction-uuid",
      "type": "ADJUSTMENT",
      "direction": "IN",
      "amount": "200000.00",
      "reason": "Matched balance with bank application."
    }
  }
}
```

The client does not provide the adjustment amount directly.

---

# 43. Delete Account

```http
DELETE /api/v1/accounts/:id
```

Allowed only when there is no financial history.

Success:

```http
204 No Content
```

Conflict:

```json
{
  "success": false,
  "error": {
    "code": "ACCOUNT_HAS_FINANCIAL_HISTORY"
  },
  "message": "Accounts with financial history must be archived instead."
}
```

---

# 44. Categories API

Base:

```text
/api/v1/categories
```

Endpoints:

```text
GET    /api/v1/categories
POST   /api/v1/categories
GET    /api/v1/categories/:id
PATCH  /api/v1/categories/:id
POST   /api/v1/categories/:id/archive
POST   /api/v1/categories/:id/restore
DELETE /api/v1/categories/:id
```

---

# 45. List Categories

```http
GET /api/v1/categories
```

Query:

```text
type=INCOME|EXPENSE
status=ACTIVE|ARCHIVED
include_system=true|false
search=
```

Example:

```http
GET /api/v1/categories?type=EXPENSE&status=ACTIVE
```

Response:

```json
{
  "success": true,
  "data": [
    {
      "id": "category-uuid",
      "name": "Food & Dining",
      "type": "EXPENSE",
      "is_system": true,
      "status": "ACTIVE"
    }
  ]
}
```

---

# 46. Create Custom Category

```http
POST /api/v1/categories
```

Request:

```json
{
  "name": "Gym",
  "type": "EXPENSE",
  "icon": "dumbbell"
}
```

Response:

```http
201 Created
```

```json
{
  "success": true,
  "data": {
    "id": "category-uuid",
    "name": "Gym",
    "type": "EXPENSE",
    "is_system": false,
    "status": "ACTIVE"
  }
}
```

---

# 47. Transactions API

Base:

```text
/api/v1/transactions
```

Endpoints:

```text
GET    /api/v1/transactions
POST   /api/v1/transactions
GET    /api/v1/transactions/:id
PATCH  /api/v1/transactions/:id
POST   /api/v1/transactions/:id/void
```

PATCH is limited to still-pending DRAFT transactions.

Posted transactions are financially immutable.

Correction is performed by voiding the original and creating a replacement transaction.

Hard delete is not recommended for posted financial records.

Voiding is preferred.

---

# 48. List Transactions

```http
GET /api/v1/transactions
```

Query parameters:

```text
search
type
account_id
category_id
date_from
date_to
min_amount
max_amount
status
sort
order
page
limit
```

Example:

```http
GET /api/v1/transactions?type=EXPENSE&category_id=abc&date_from=2026-08-01&date_to=2026-08-31&sort=transaction_date&order=desc&page=1&limit=20
```

Response:

```json
{
  "success": true,
  "data": [
    {
      "id": "transaction-uuid",
      "type": "EXPENSE",
      "amount": "87500.00",
      "transaction_date": "2026-08-24",
      "description": "Lunch",
      "merchant": "GrabFood",
      "status": "POSTED",
      "account": {
        "id": "account-uuid",
        "name": "GoPay"
      },
      "category": {
        "id": "category-uuid",
        "name": "Food & Dining"
      },
      "version": 1
    }
  ],
  "meta": {
    "page": 1,
    "limit": 20,
    "total": 72,
    "total_pages": 4
  }
}
```

---

# 49. Create Income Transaction

```http
POST /api/v1/transactions
```

Request:

```json
{
  "account_id": "account-uuid",
  "category_id": "salary-category-uuid",
  "type": "INCOME",
  "amount": "12000000.00",
  "transaction_date": "2026-08-25",
  "description": "August salary"
}
```

Success:

```http
201 Created
```

Response:

```json
{
  "success": true,
  "data": {
    "transaction": {
      "id": "transaction-uuid",
      "type": "INCOME",
      "amount": "12000000.00",
      "transaction_date": "2026-08-25",
      "status": "POSTED",
      "version": 1
    },
    "account": {
      "id": "account-uuid",
      "current_balance": "16000000.00"
    }
  }
}
```

---

# 50. Create Expense Transaction

```http
POST /api/v1/transactions
```

Request:

```json
{
  "account_id": "account-uuid",
  "category_id": "food-category-uuid",
  "type": "EXPENSE",
  "amount": "87500.00",
  "transaction_date": "2026-08-24",
  "description": "Dinner",
  "merchant": "GrabFood"
}
```

Response:

```json
{
  "success": true,
  "data": {
    "transaction": {
      "id": "transaction-uuid",
      "type": "EXPENSE",
      "amount": "87500.00",
      "status": "POSTED",
      "version": 1
    },
    "account": {
      "id": "account-uuid",
      "current_balance": "4912500.00"
    }
  }
}
```

---

# 51. Invalid Category Example

Request:

```text
type = EXPENSE
category = Salary
```

Response:

```http
422 Unprocessable Entity
```

```json
{
  "success": false,
  "error": {
    "code": "INVALID_CATEGORY_TYPE",
    "details": {
      "category_id": [
        "Expense transactions require an expense category."
      ]
    }
  },
  "message": "The selected category cannot be used for this transaction."
}
```

---

# 52. Get Transaction

```http
GET /api/v1/transactions/:id
```

Response includes:

```json
{
  "success": true,
  "data": {
    "id": "transaction-uuid",
    "type": "EXPENSE",
    "amount": "87500.00",
    "transaction_date": "2026-08-24",
    "description": "Dinner",
    "merchant": "GrabFood",
    "notes": null,
    "source": "MANUAL",
    "status": "POSTED",
    "version": 1,
    "account": {
      "id": "account-uuid",
      "name": "GoPay"
    },
    "category": {
      "id": "category-uuid",
      "name": "Food & Dining"
    }
  }
}
```

---

# 53. Update Transaction

```http
PATCH /api/v1/transactions/:id
```

Request:

```json
{
  "account_id": "account-uuid",
  "category_id": "food-category-uuid",
  "amount": "120000.00",
  "transaction_date": "2026-08-24",
  "description": "Dinner",
  "version": 1
}
```

Backend must:

```text
load existing transaction
↓
verify ownership
↓
verify transaction is DRAFT (still-pending)
↓
verify version
↓
apply new balance effect
↓
update transaction
↓
increment version
↓
commit
```

PATCH on a POSTED transaction is rejected with `409 TRANSACTION_IMMUTABLE`; correct the record by voiding it and creating a replacement.

Response:

```json
{
  "success": true,
  "data": {
    "transaction": {
      "id": "transaction-uuid",
      "amount": "120000.00",
      "version": 2
    },
    "account": {
      "id": "account-uuid",
      "current_balance": "4880000.00"
    }
  }
}
```

---

# 54. Void Transaction

```http
POST /api/v1/transactions/:id/void
```

Request:

```json
{
  "reason": "Duplicate transaction.",
  "version": 2
}
```

Response:

```json
{
  "success": true,
  "data": {
    "id": "transaction-uuid",
    "status": "VOIDED",
    "voided_at": "2026-08-24T15:30:00Z",
    "void_reason": "Duplicate transaction.",
    "version": 3
  }
}
```

The financial effect must be voided atomically and the historical record preserved.

---

# 55. Transfers API

Base:

```text
/api/v1/transfers
```

Endpoints:

```text
GET  /api/v1/transfers
POST /api/v1/transfers
GET  /api/v1/transfers/:id
POST /api/v1/transfers/:id/void
```

---

# 56. Create Transfer

```http
POST /api/v1/transfers
```

Request:

```json
{
  "source_account_id": "bca-uuid",
  "destination_account_id": "gopay-uuid",
  "amount": "1000000.00",
  "transfer_date": "2026-08-24",
  "description": "Top up wallet"
}
```

Success:

```http
201 Created
```

Response:

```json
{
  "success": true,
  "data": {
    "transfer": {
      "id": "transfer-uuid",
      "amount": "1000000.00",
      "transfer_date": "2026-08-24",
      "status": "POSTED",
      "version": 1
    },
    "source_account": {
      "id": "bca-uuid",
      "current_balance": "4000000.00"
    },
    "destination_account": {
      "id": "gopay-uuid",
      "current_balance": "1300000.00"
    }
  }
}
```

---

# 57. Invalid Same-Account Transfer

Response:

```http
422 Unprocessable Entity
```

```json
{
  "success": false,
  "error": {
    "code": "TRANSFER_SAME_ACCOUNT"
  },
  "message": "Source and destination accounts must be different."
}
```

---

# 58. Void Transfer

```http
POST /api/v1/transfers/:id/void
```

Request:

```json
{
  "reason": "Transfer recorded by mistake.",
  "version": 1
}
```

Backend must atomically void both account effects.

---

# 59. Recurring Transactions API

Base:

```text
/api/v1/recurring-transactions
```

Endpoints:

```text
GET   /api/v1/recurring-transactions
POST  /api/v1/recurring-transactions
GET   /api/v1/recurring-transactions/:id
PATCH /api/v1/recurring-transactions/:id

POST /api/v1/recurring-transactions/:id/pause
POST /api/v1/recurring-transactions/:id/resume
POST /api/v1/recurring-transactions/:id/end

GET  /api/v1/recurring-transactions/:id/occurrences
```

---

# 60. Create Recurring Transaction

```http
POST /api/v1/recurring-transactions
```

Request:

```json
{
  "account_id": "account-uuid",
  "category_id": "salary-category-uuid",
  "type": "INCOME",
  "amount": "12000000.00",
  "frequency": "MONTHLY",
  "interval_value": 1,
  "start_date": "2026-08-25",
  "day_of_month": 25,
  "auto_post": true,
  "description": "Monthly Salary"
}
```

Success:

```json
{
  "success": true,
  "data": {
    "id": "recurring-uuid",
    "type": "INCOME",
    "amount": "12000000.00",
    "frequency": "MONTHLY",
    "next_occurrence_date": "2026-08-25",
    "auto_post": true,
    "status": "ACTIVE",
    "version": 1
  }
}
```

---

# 61. Pause Recurring

```http
POST /api/v1/recurring-transactions/:id/pause
```

Request:

```json
{
  "version": 1
}
```

Response:

```json
{
  "success": true,
  "data": {
    "id": "recurring-uuid",
    "status": "PAUSED",
    "version": 2
  }
}
```

---

# 62. Resume Recurring

```http
POST /api/v1/recurring-transactions/:id/resume
```

Request:

```json
{
  "version": 2
}
```

Response:

```json
{
  "success": true,
  "data": {
    "id": "recurring-uuid",
    "status": "ACTIVE",
    "next_occurrence_date": "2026-09-25",
    "version": 3
  }
}
```

---

# 63. End Recurring

```http
POST /api/v1/recurring-transactions/:id/end
```

Request:

```json
{
  "version": 3
}
```

Response:

```json
{
  "success": true,
  "data": {
    "id": "recurring-uuid",
    "status": "ENDED",
    "version": 4
  }
}
```

Ending a recurring rule does not alter already posted occurrences.

---

# 64. Recurring Occurrences

```http
GET /api/v1/recurring-transactions/:id/occurrences
```

Response:

```json
{
  "success": true,
  "data": [
    {
      "id": "occurrence-uuid",
      "occurrence_date": "2026-08-25",
      "status": "CONFIRMED",
      "transaction_id": "transaction-uuid",
      "generated_at": "2026-08-25T00:05:00Z"
    }
  ]
}
```

---

# 65. Budgets API

Base:

```text
/api/v1/budgets
```

Endpoints:

```text
GET   /api/v1/budgets
POST  /api/v1/budgets
GET   /api/v1/budgets/:id
PATCH /api/v1/budgets/:id
POST  /api/v1/budgets/:id/archive
```

---

# 66. List Budgets

```http
GET /api/v1/budgets
```

Query:

```text
period
category_id
status
page
limit
```

Response:

```json
{
  "success": true,
  "data": [
    {
      "id": "budget-uuid",
      "category": {
        "id": "food-category-uuid",
        "name": "Food & Dining"
      },
      "amount": "2000000.00",
      "spent": "1450000.00",
      "remaining": "550000.00",
      "utilization_percent": "72.50",
      "computed_status": "ON_TRACK",
      "projected_spend": "2350000.00",
      "projected_overspend": "350000.00",
      "start_date": "2026-08-01",
      "end_date": "2026-08-31",
      "status": "ACTIVE",
      "version": 1
    }
  ]
}
```

Derived fields are calculated by the finance engine.

---

# 67. Create Budget

```http
POST /api/v1/budgets
```

Request:

```json
{
  "category_id": "food-category-uuid",
  "amount": "2000000.00",
  "period_type": "MONTHLY",
  "start_date": "2026-08-01"
}
```

Backend may derive:

```text
end_date = 2026-08-31
```

Response:

```http
201 Created
```

```json
{
  "success": true,
  "data": {
    "id": "budget-uuid",
    "amount": "2000000.00",
    "start_date": "2026-08-01",
    "end_date": "2026-08-31",
    "status": "ACTIVE",
    "version": 1
  }
}
```

---

# 68. Duplicate Budget

```http
409 Conflict
```

```json
{
  "success": false,
  "error": {
    "code": "DUPLICATE_BUDGET"
  },
  "message": "An active budget already exists for this category and period."
}
```

---

# 69. Budget Detail

```http
GET /api/v1/budgets/:id
```

Response:

```json
{
  "success": true,
  "data": {
    "id": "budget-uuid",
    "amount": "2000000.00",
    "spent": "1450000.00",
    "remaining": "550000.00",
    "utilization_percent": "72.50",
    "computed_status": "ON_TRACK",
    "projected": {
      "spend": "2350000.00",
      "overspend": "350000.00",
      "risk": "WARNING"
    },
    "category": {
      "id": "category-uuid",
      "name": "Food & Dining"
    },
    "version": 1
  }
}
```

---

# 70. Financial Goals API

Base:

```text
/api/v1/goals
```

Endpoints:

```text
GET   /api/v1/goals
POST  /api/v1/goals
GET   /api/v1/goals/:id
PATCH /api/v1/goals/:id

POST /api/v1/goals/:id/pause
POST /api/v1/goals/:id/resume
POST /api/v1/goals/:id/cancel
POST /api/v1/goals/:id/mark-achieved
```

---

# 71. Create Goal

```http
POST /api/v1/goals
```

Request:

```json
{
  "name": "Emergency Fund",
  "target_amount": "30000000.00",
  "current_amount": "12000000.00",
  "target_date": "2027-02-28",
  "priority": "HIGH"
}
```

Response:

```json
{
  "success": true,
  "data": {
    "id": "goal-uuid",
    "name": "Emergency Fund",
    "target_amount": "30000000.00",
    "current_amount": "12000000.00",
    "progress_percent": "40.00",
    "remaining_amount": "18000000.00",
    "required_monthly_contribution": "3000000.00",
    "feasibility": "AT_RISK",
    "status": "ACTIVE",
    "version": 1
  }
}
```

---

# 72. Goal Detail

```http
GET /api/v1/goals/:id
```

Response may include:

```json
{
  "success": true,
  "data": {
    "id": "goal-uuid",
    "name": "Emergency Fund",
    "target_amount": "30000000.00",
    "current_amount": "12000000.00",
    "progress_percent": "40.00",
    "remaining_amount": "18000000.00",
    "target_date": "2027-02-28",
    "required_monthly_contribution": "3000000.00",
    "estimated_free_cashflow": "2400000.00",
    "feasibility": "AT_RISK",
    "status": "ACTIVE",
    "version": 1
  }
}
```

---

# 73. Update Goal

```http
PATCH /api/v1/goals/:id
```

Request:

```json
{
  "current_amount": "14000000.00",
  "target_date": "2027-03-31",
  "version": 1
}
```

Finance engine recalculates all derived values.

---

# 74. Dashboard API

```http
GET /api/v1/dashboard
```

Recommended as a composite endpoint.

Query:

```text
period=2026-08
```

Response:

```json
{
  "success": true,
  "data": {
    "period": {
      "start": "2026-08-01",
      "end": "2026-08-31"
    },
    "balance": {
      "total": "16250000.00",
      "currency": "IDR"
    },
    "cashflow": {
      "income": "12000000.00",
      "expense": "8400000.00",
      "net": "3600000.00",
      "savings_rate_percent": "30.00"
    },
    "budgets": {
      "active_count": 4,
      "warning_count": 1,
      "exceeded_count": 0
    },
    "goals": {
      "active_count": 2,
      "at_risk_count": 1
    },
    "upcoming": [
      {
        "date": "2026-08-25",
        "name": "Salary",
        "direction": "IN",
        "amount": "12000000.00"
      }
    ],
    "forecast": {
      "minimum_balance": "3200000.00",
      "ending_balance": "12000000.00",
      "confidence": "MEDIUM"
    },
    "insights": [
      {
        "id": "insight-uuid",
        "type": "SPENDING_ANOMALY",
        "severity": "MEDIUM",
        "title": "Dining spending increased"
      }
    ]
  }
}
```

---

# 75. Analytics API

Base:

```text
/api/v1/analytics
```

Endpoints:

```text
GET /api/v1/analytics/cashflow
GET /api/v1/analytics/categories
GET /api/v1/analytics/period-comparison
GET /api/v1/analytics/recurring-expenses
GET /api/v1/analytics/spending-changes
```

---

# 76. Cashflow Analytics

```http
GET /api/v1/analytics/cashflow
```

Query:

```text
date_from
date_to
```

Example:

```http
GET /api/v1/analytics/cashflow?date_from=2026-08-01&date_to=2026-08-31
```

Response:

```json
{
  "success": true,
  "data": {
    "income": "12000000.00",
    "expense": "8400000.00",
    "net_cashflow": "3600000.00",
    "savings_rate_percent": "30.00"
  }
}
```

---

# 77. Category Breakdown

```http
GET /api/v1/analytics/categories
```

Query:

```text
type=EXPENSE
date_from
date_to
```

Response:

```json
{
  "success": true,
  "data": [
    {
      "category": {
        "id": "food-uuid",
        "name": "Food & Dining"
      },
      "amount": "2400000.00",
      "percentage": "28.57"
    }
  ]
}
```

---

# 78. Period Comparison

```http
GET /api/v1/analytics/period-comparison
```

Example:

```http
GET /api/v1/analytics/period-comparison?period=current_month&baseline=previous_3_month_average
```

Response:

```json
{
  "success": true,
  "data": {
    "current": {
      "expense": "8400000.00"
    },
    "baseline": {
      "expense": "6800000.00"
    },
    "difference": {
      "amount": "1600000.00",
      "percent": "23.53"
    },
    "drivers": [
      {
        "category": "Food & Dining",
        "difference": "700000.00"
      },
      {
        "category": "Shopping",
        "difference": "550000.00"
      }
    ]
  }
}
```

---

# 79. Forecast API

Base:

```text
/api/v1/forecast
```

Endpoints:

```text
GET  /api/v1/forecast
POST /api/v1/forecast/calculate
GET  /api/v1/forecast/history
GET  /api/v1/forecast/:snapshotId
```

---

# 80. Calculate Forecast

```http
POST /api/v1/forecast/calculate
```

Request:

```json
{
  "horizon_days": 90,
  "persist": true,
  "assumptions": {
    "variable_spending_method": "HISTORICAL_AVERAGE"
  }
}
```

Backend performs deterministic calculation.

Response:

```json
{
  "success": true,
  "data": {
    "snapshot_id": "forecast-uuid",
    "generated_at": "2026-08-24T15:00:00Z",
    "data_through_date": "2026-08-24",
    "horizon_days": 90,
    "confidence": "MEDIUM",
    "calculation_version": "finance-engine-v1",
    "summary": {
      "opening_balance": "16250000.00",
      "projected_income": "36000000.00",
      "projected_expense": "30800000.00",
      "ending_balance": "21450000.00",
      "minimum_balance": "3200000.00"
    },
    "assumptions": [
      {
        "type": "VARIABLE_SPENDING",
        "description": "Estimated using recent historical average."
      }
    ],
    "timeline": [
      {
        "date": "2026-08-25",
        "type": "SCHEDULED",
        "source": "Monthly Salary",
        "direction": "IN",
        "amount": "12000000.00",
        "projected_balance_after": "28250000.00"
      },
      {
        "date": "2026-09-01",
        "type": "SCHEDULED",
        "source": "Rent",
        "direction": "OUT",
        "amount": "3000000.00",
        "projected_balance_after": "25250000.00"
      }
    ]
  }
}
```

---

# 81. Get Latest Forecast

```http
GET /api/v1/forecast
```

Response:

```json
{
  "success": true,
  "data": {
    "snapshot_id": "forecast-uuid",
    "status": "FRESH",
    "generated_at": "2026-08-24T15:00:00Z",
    "confidence": "MEDIUM",
    "summary": {
      "ending_balance": "21450000.00",
      "minimum_balance": "3200000.00"
    }
  }
}
```

If none exists:

```http
404 Not Found
```

or return a documented empty state.

Consistency should be chosen during implementation.

---

# 82. Stale Forecast

Example response:

```json
{
  "success": true,
  "data": {
    "snapshot_id": "forecast-uuid",
    "status": "STALE",
    "generated_at": "2026-08-24T12:00:00Z",
    "stale_reason": "Financial data changed after forecast generation."
  }
}
```

---

# 83. Scenario API

Base:

```text
/api/v1/scenarios
```

Endpoints:

```text
GET    /api/v1/scenarios
POST   /api/v1/scenarios
GET    /api/v1/scenarios/:id
PATCH  /api/v1/scenarios/:id
DELETE /api/v1/scenarios/:id

POST   /api/v1/scenarios/:id/modifications
PATCH  /api/v1/scenarios/:id/modifications/:modificationId
DELETE /api/v1/scenarios/:id/modifications/:modificationId

POST   /api/v1/scenarios/:id/calculate
GET    /api/v1/scenarios/:id/snapshots
GET    /api/v1/scenarios/:id/snapshots/:snapshotId

POST   /api/v1/scenarios/:id/archive
```

---

# 84. Create Scenario

```http
POST /api/v1/scenarios
```

Request:

```json
{
  "name": "Buy MacBook",
  "description": "Evaluate buying a laptop this month.",
  "horizon_days": 180
}
```

Response:

```http
201 Created
```

```json
{
  "success": true,
  "data": {
    "id": "scenario-uuid",
    "name": "Buy MacBook",
    "horizon_days": 180,
    "status": "DRAFT",
    "is_stale": false,
    "version": 1
  }
}
```

---

# 85. Add One-Time Expense Modification

```http
POST /api/v1/scenarios/:id/modifications
```

Request:

```json
{
  "type": "ONE_TIME_EXPENSE",
  "name": "MacBook Purchase",
  "amount": "15000000.00",
  "effective_date": "2026-09-10"
}
```

Response:

```json
{
  "success": true,
  "data": {
    "id": "modification-uuid",
    "type": "ONE_TIME_EXPENSE",
    "name": "MacBook Purchase",
    "amount": "15000000.00",
    "effective_date": "2026-09-10"
  }
}
```

---

# 86. Add Income Reduction Modification

```http
POST /api/v1/scenarios/:id/modifications
```

Request:

```json
{
  "type": "INCOME_REDUCTION",
  "name": "Salary Reduction",
  "percentage": "30.00",
  "effective_date": "2026-10-01",
  "source_recurring_transaction_id": "salary-recurring-uuid"
}
```

---

# 87. Add Income Removal Modification

```http
POST /api/v1/scenarios/:id/modifications
```

Request:

```json
{
  "type": "INCOME_REMOVAL",
  "name": "Resign",
  "effective_date": "2026-10-01",
  "source_recurring_transaction_id": "salary-recurring-uuid"
}
```

---

# 88. Add Recurring Expense Modification

```http
POST /api/v1/scenarios/:id/modifications
```

Request:

```json
{
  "type": "RECURRING_EXPENSE",
  "name": "Motorcycle Installment",
  "amount": "1800000.00",
  "effective_date": "2026-10-01",
  "frequency": "MONTHLY",
  "duration_months": 24
}
```

---

# 89. Calculate Scenario

```http
POST /api/v1/scenarios/:id/calculate
```

Request:

```json
{
  "persist_snapshot": true
}
```

Response:

```json
{
  "success": true,
  "data": {
    "scenario_id": "scenario-uuid",
    "snapshot_id": "scenario-snapshot-uuid",
    "calculated_at": "2026-08-24T15:20:00Z",
    "calculation_version": "finance-engine-v1",
    "data_through_date": "2026-08-24",
    "baseline": {
      "ending_balance": "18400000.00",
      "minimum_balance": "8200000.00",
      "savings_rate_percent": "27.00",
      "cash_runway_months": "4.10"
    },
    "scenario": {
      "ending_balance": "3400000.00",
      "minimum_balance": "1100000.00",
      "savings_rate_percent": "8.00",
      "cash_runway_months": "1.90"
    },
    "difference": {
      "ending_balance": "-15000000.00",
      "minimum_balance": "-7100000.00",
      "savings_rate_percentage_points": "-19.00",
      "cash_runway_months": "-2.20"
    },
    "goal_impacts": [
      {
        "goal_id": "goal-uuid",
        "goal_name": "Emergency Fund",
        "baseline_completion_date": "2026-12-31",
        "scenario_completion_date": "2027-03-31",
        "delay_months": 3
      }
    ],
    "status": "CALCULATED"
  }
}
```

These values are deterministic.

---

# 90. Scenario Stale Response

```json
{
  "success": true,
  "data": {
    "id": "scenario-uuid",
    "status": "CALCULATED",
    "is_stale": true,
    "last_calculated_at": "2026-08-24T15:20:00Z",
    "stale_reason": "Financial records changed after the last scenario calculation."
  }
}
```

---

# 91. AI API

Base:

```text
/api/v1/ai
```

Endpoints:

```text
POST /api/v1/ai/categorize-transaction
GET  /api/v1/ai/insights
GET  /api/v1/ai/insights/:id
POST /api/v1/ai/insights/:id/view
POST /api/v1/ai/insights/:id/acknowledge
POST /api/v1/ai/insights/:id/dismiss
POST /api/v1/ai/insights/:id/feedback

POST /api/v1/ai/copilot
POST /api/v1/ai/explain-scenario
```

Automatic insight generation may run through background jobs rather than public HTTP endpoints.

---

# 92. AI Transaction Categorization

```http
POST /api/v1/ai/categorize-transaction
```

Request:

```json
{
  "description": "GRAB*FOOD 83219",
  "merchant": "GrabFood",
  "transaction_type": "EXPENSE",
  "amount": "87500.00"
}
```

Response:

```json
{
  "success": true,
  "data": {
    "suggested_category": {
      "id": "food-category-uuid",
      "name": "Food & Dining"
    },
    "confidence": "0.9600",
    "reason": "The transaction description and merchant are associated with food delivery."
  }
}
```

This endpoint does not create or modify a transaction.

---

# 93. List AI Insights

```http
GET /api/v1/ai/insights
```

Query:

```text
type
severity
status
date_from
date_to
page
limit
```

Response:

```json
{
  "success": true,
  "data": [
    {
      "id": "insight-uuid",
      "type": "SPENDING_ANOMALY",
      "severity": "MEDIUM",
      "title": "Dining spending increased",
      "summary": "Dining spending is 60% above your recent baseline.",
      "status": "NEW",
      "generated_at": "2026-08-24T14:00:00Z"
    }
  ],
  "meta": {
    "page": 1,
    "limit": 20,
    "total": 4,
    "total_pages": 1
  }
}
```

---

# 94. AI Insight Detail

```http
GET /api/v1/ai/insights/:id
```

Response:

```json
{
  "success": true,
  "data": {
    "id": "insight-uuid",
    "type": "SPENDING_ANOMALY",
    "severity": "MEDIUM",
    "title": "Dining spending increased",
    "summary": "Dining spending is significantly above your recent baseline.",
    "drivers": [
      {
        "label": "Food Delivery",
        "difference": "720000.00"
      }
    ],
    "context": {
      "current_amount": "2400000.00",
      "baseline_amount": "1500000.00",
      "change_percent": "60.00"
    },
    "suggested_actions": [
      {
        "type": "REVIEW_BUDGET",
        "label": "Review Food & Dining budget"
      }
    ],
    "status": "NEW"
  }
}
```

---

# 95. Dismiss Insight

```http
POST /api/v1/ai/insights/:id/dismiss
```

Response:

```json
{
  "success": true,
  "data": {
    "id": "insight-uuid",
    "status": "DISMISSED"
  }
}
```

---

# 96. AI Insight Feedback

```http
POST /api/v1/ai/insights/:id/feedback
```

Request:

```json
{
  "feedback": "HELPFUL"
}
```

or:

```json
{
  "feedback": "NOT_HELPFUL",
  "comment": "This spending increase was intentional because of travel."
}
```

Response:

```http
201 Created
```

---

# 97. AI Copilot

```http
POST /api/v1/ai/copilot
```

Request:

```json
{
  "message": "Why did I spend more this month?"
}
```

The backend must not simply forward this message to the LLM.

Flow:

```text
Authenticate
↓
Rate Limit
↓
Classify Intent
↓
Select Deterministic Tools
↓
Execute Finance Tools
↓
Build Minimal Structured Context
↓
Call LLM
↓
Validate Output
↓
Return Response
```

Response:

```json
{
  "success": true,
  "data": {
    "answer": "Your spending increased by Rp1.6M compared with your recent baseline. Dining and shopping were the main contributors.",
    "facts": [
      {
        "label": "Current expenses",
        "value": "8400000.00"
      },
      {
        "label": "Baseline expenses",
        "value": "6800000.00"
      },
      {
        "label": "Difference",
        "value": "1600000.00"
      }
    ],
    "sources": [
      {
        "type": "ANALYTICS_TOOL",
        "name": "compare_periods"
      },
      {
        "type": "ANALYTICS_TOOL",
        "name": "get_spending_changes"
      }
    ]
  }
}
```

---

# 98. AI Copilot Tool Contract

The AI orchestration layer may expose internal tools such as:

```text
get_cashflow_summary
get_account_summary
get_category_breakdown
compare_periods
get_spending_changes
get_recurring_expenses
get_budget_status
get_goal_status
calculate_forecast
calculate_scenario
get_scenario_comparison
```

These are internal backend tools.

They are not necessarily public HTTP endpoints.

The LLM consumes their structured results.

---

# 99. Copilot Scenario Question

User:

```text
Can I afford a Rp12M laptop this month?
```

If date is missing, AI may request clarification:

```json
{
  "success": true,
  "data": {
    "type": "CLARIFICATION_REQUIRED",
    "answer": "When should I assume the purchase happens?",
    "options": [
      {
        "value": "TODAY",
        "label": "Today"
      },
      {
        "value": "NEXT_PAYDAY",
        "label": "Next payday"
      },
      {
        "value": "CUSTOM",
        "label": "Choose date"
      }
    ]
  }
}
```

After clarification, backend invokes deterministic scenario engine.

---

# 100. Scenario AI Explanation

```http
POST /api/v1/ai/explain-scenario
```

Request:

```json
{
  "scenario_id": "scenario-uuid",
  "snapshot_id": "scenario-snapshot-uuid"
}
```

The endpoint must load authoritative scenario results itself.

The frontend must not send arbitrary baseline/scenario values for AI to trust.

Response:

```json
{
  "success": true,
  "data": {
    "summary": "The purchase is possible without creating an immediate negative balance, but it significantly reduces your financial buffer.",
    "key_impacts": [
      "Lowest projected balance falls from Rp8.2M to Rp1.1M.",
      "Emergency fund completion is delayed by approximately three months.",
      "Estimated cash runway falls from 4.1 to 1.9 months."
    ]
  }
}
```

---

# 101. AI Provider Failure

Example:

```http
503 Service Unavailable
```

```json
{
  "success": false,
  "error": {
    "code": "AI_PROVIDER_UNAVAILABLE"
  },
  "message": "AI features are temporarily unavailable. Your financial data and calculations are unaffected."
}
```

Deterministic endpoints continue to function.

---

# 102. AI Timeout

```http
503 Service Unavailable
```

or documented `500`/`503` policy.

Response:

```json
{
  "success": false,
  "error": {
    "code": "AI_TIMEOUT"
  },
  "message": "The AI request took too long. Please try again."
}
```

---

# 103. AI Rate Limit

```http
429 Too Many Requests
```

Headers may include:

```http
Retry-After: 30
```

Response:

```json
{
  "success": false,
  "error": {
    "code": "RATE_LIMIT_EXCEEDED"
  },
  "message": "Please wait before sending another AI request."
}
```

---

# 104. Notifications API

Base:

```text
/api/v1/notifications
```

Endpoints:

```text
GET  /api/v1/notifications
GET  /api/v1/notifications/unread-count
POST /api/v1/notifications/:id/read
POST /api/v1/notifications/read-all
POST /api/v1/notifications/:id/archive
```

---

# 105. List Notifications

```http
GET /api/v1/notifications
```

Query:

```text
status
type
page
limit
```

Response:

```json
{
  "success": true,
  "data": [
    {
      "id": "notification-uuid",
      "type": "BUDGET_WARNING",
      "title": "Food budget is almost reached",
      "message": "You have used 82% of your Food & Dining budget.",
      "status": "UNREAD",
      "created_at": "2026-08-24T10:00:00Z"
    }
  ],
  "meta": {
    "page": 1,
    "limit": 20,
    "total": 8,
    "total_pages": 1
  }
}
```

---

# 106. Mark Notification Read

```http
POST /api/v1/notifications/:id/read
```

Response:

```json
{
  "success": true,
  "data": {
    "id": "notification-uuid",
    "status": "READ",
    "read_at": "2026-08-24T15:30:00Z"
  }
}
```

---

# 107. Reports API

Base:

```text
/api/v1/reports
```

Endpoints:

```text
GET /api/v1/reports/summary
GET /api/v1/reports/income-expense
GET /api/v1/reports/categories
GET /api/v1/reports/budgets
GET /api/v1/reports/goals
```

Reports are derived data.

They do not create financial state.

---

# 108. Financial Summary Report

```http
GET /api/v1/reports/summary
```

Query:

```text
date_from
date_to
```

Response:

```json
{
  "success": true,
  "data": {
    "period": {
      "start": "2026-08-01",
      "end": "2026-08-31"
    },
    "income": "12000000.00",
    "expense": "8400000.00",
    "net_cashflow": "3600000.00",
    "savings_rate_percent": "30.00",
    "largest_expense_category": {
      "name": "Food & Dining",
      "amount": "2400000.00"
    }
  }
}
```

---

# 109. Audit API

For normal users, broad audit access may not be necessary.

Security/session activity can be exposed through dedicated UX.

If audit log access is implemented:

```text
GET /api/v1/audit-logs
```

must return only the authenticated user's safe audit events.

Never expose:

```text
refresh hashes
password information
secret data
sensitive raw AI context
```

---

# 110. Health Endpoints

Infrastructure endpoints:

```text
GET /health
GET /ready
```

These may exist outside `/api/v1`.

---

# 111. Liveness

```http
GET /health
```

Response:

```json
{
  "status": "ok"
}
```

Purpose:

```text
application process is alive
```

---

# 112. Readiness

```http
GET /ready
```

Response:

```json
{
  "status": "ready",
  "dependencies": {
    "postgres": "up",
    "redis": "up",
    "ai_provider": "degraded"
  }
}
```

PostgreSQL should normally be required for readiness.

Optional dependencies may be reported as degraded without necessarily making the API unavailable.

---

# 113. Version Endpoint

Optional:

```http
GET /version
```

Response:

```json
{
  "version": "1.0.0",
  "commit": "abc1234"
}
```

Useful for deployment diagnostics.

---

# 114. Idempotency

Important financial create endpoints may support:

```http
Idempotency-Key: <unique-key>
```

Candidate endpoints:

```text
POST /transactions
POST /transfers
POST recurring auto-post internal command
```

Behavior:

```text
same user
+
same idempotency key
+
same endpoint
```

must not create duplicate financial effects.

---

# 115. Idempotency Conflict

If the same idempotency key is reused with different payload:

```http
409 Conflict
```

Example code:

```text
IDEMPOTENCY_KEY_CONFLICT
```

---

# 116. Optimistic Locking Contract

Mutable resources requiring version protection accept:

```json
{
  "version": 4
}
```

or version may be supplied through:

```http
If-Match
```

For simplicity and explicitness in this project, request-body version is acceptable.

Failure:

```http
409 Conflict
```

```json
{
  "success": false,
  "error": {
    "code": "VERSION_CONFLICT"
  },
  "message": "The resource changed since it was loaded."
}
```

---

# 117. Search Contract

Search query:

```text
search
```

Example:

```http
GET /api/v1/transactions?search=grab
```

Searchable fields may include:

```text
description
merchant
notes
```

Search must remain scoped by authenticated user.

---

# 118. Validation Contract

Validation occurs for:

```text
JSON body
path parameters
query parameters
date ranges
UUIDs
enums
pagination
amounts
cross-resource ownership
business state
AI output
```

Backend is the source of truth.

---

# 119. UUID Path Validation

Example:

```http
GET /api/v1/accounts/not-a-uuid
```

Response:

```http
400 Bad Request
```

```json
{
  "success": false,
  "error": {
    "code": "INVALID_UUID"
  },
  "message": "The supplied resource identifier is invalid."
}
```

---

# 120. Resource Not Found Policy

For user-owned resources, Savio may return:

```http
404 Not Found
```

for both:

```text
resource does not exist
```

and:

```text
resource belongs to another user
```

This reduces resource enumeration risk.

Example:

```json
{
  "success": false,
  "error": {
    "code": "RESOURCE_NOT_FOUND"
  },
  "message": "The requested resource was not found."
}
```

---

# 121. 401 Frontend Contract

When frontend receives:

```http
401
```

the Axios interceptor should:

```text
check retry marker
↓
single-flight refresh
↓
queue concurrent failed requests
↓
refresh success?
├── Yes
│   ↓
│  retry original request once
│
└── No
    ↓
   clear auth
    ↓
   redirect login
```

---

# 122. Refresh Single-Flight Requirement

If these requests fail simultaneously:

```text
GET /dashboard
GET /accounts
GET /notifications
```

frontend must perform:

```text
ONE refresh request
```

not:

```text
THREE refresh requests
```

All waiting requests reuse the same refresh result.

---

# 123. Refresh Infinite Loop Prevention

Requests should carry internal retry state.

Concept:

```text
_retry = true
```

A retried request receiving another `401` must not trigger unlimited refresh attempts.

---

# 124. 403 Frontend Contract

```http
403 Forbidden
```

Frontend behavior:

```text
do not logout automatically
```

Display:

```text
You do not have permission to perform this action.
```

For CSRF failures, the frontend may first refresh CSRF state only if the implemented security design safely supports that behavior.

---

# 125. 422 Frontend Contract

Frontend maps:

```text
error.details
```

to form fields.

Example backend:

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "details": {
      "amount": [
        "Amount must be greater than zero."
      ]
    }
  }
}
```

Frontend:

```text
Amount
[ -100 ]

Amount must be greater than zero.
```

---

# 126. 429 Frontend Contract

Frontend must:

- display a clear rate-limit message,
- avoid repeated automatic retries,
- respect `Retry-After` when available.

Particularly important for:

```text
login
refresh
AI Copilot
AI categorization
```

---

# 127. 500 Frontend Contract

Unexpected server errors:

```http
500 Internal Server Error
```

Frontend should display:

```text
Something went wrong.

Please try again.
```

May include request reference.

Never display:

```text
stack trace
SQL query
database error
secret
```

---

# 128. Composite API Principle

Composite APIs are appropriate where a screen requires a coherent aggregate.

Examples:

```text
GET /dashboard
GET /budgets/:id
GET /goals/:id
GET /scenarios/:id
```

Avoid forcing frontend to reconstruct core business calculations from many independent responses.

The backend should return domain-level aggregates where that improves consistency.

---

# 129. Finance Calculation Boundary

Frontend must never independently calculate authoritative:

```text
account balance
budget status
savings rate
forecast
cash runway
scenario result
goal feasibility
financial health
```

Frontend may calculate purely visual transformations.

Authoritative values come from backend finance services.

---

# 130. AI Calculation Boundary

AI endpoints must not be treated as calculator endpoints.

Bad design:

```text
prompt:
"Calculate my savings rate from this data."
```

Preferred:

```text
Finance Engine
→ savings_rate = 30%

AI
→ explains what that means
```

---

# 131. AI Tool Security

Internal AI tools must execute using authenticated user context.

The model must not choose arbitrary:

```text
user_id
account_id
transaction owner
```

outside the authorized context.

Example internal execution context:

```json
{
  "authenticated_user_id": "user-uuid",
  "request_id": "req-123"
}
```

The orchestrator injects ownership context.

The LLM does not control it.

---

# 132. AI Tool Arguments

AI tool arguments must be validated exactly like external request input.

Example tool:

```text
get_category_breakdown
```

Arguments:

```json
{
  "date_from": "2026-08-01",
  "date_to": "2026-08-31",
  "type": "EXPENSE"
}
```

Invalid date ranges must be rejected before tool execution.

---

# 133. AI Structured Output Contract

Application-integrated AI output should use strict schemas.

Example insight:

```json
{
  "type": "SPENDING_ANOMALY",
  "severity": "MEDIUM",
  "title": "Dining spending increased",
  "summary": "Dining spending is above your recent baseline.",
  "drivers": [
    {
      "label": "Food Delivery",
      "difference": "720000.00"
    }
  ]
}
```

Backend validates:

```text
type
severity
length
drivers
allowed action types
```

before persistence.

---

# 134. AI Suggested Action Contract

AI suggestions must use allowlisted action types.

Example:

```json
{
  "type": "REVIEW_BUDGET",
  "resource_id": "budget-uuid"
}
```

Allowed examples:

```text
REVIEW_BUDGET
VIEW_FORECAST
OPEN_SCENARIO
REVIEW_GOAL
VIEW_RECURRING
```

Avoid accepting arbitrary client routes generated by AI.

---

# 135. AI Write Proposal Contract

Future AI write action:

```json
{
  "type": "CREATE_BUDGET",
  "payload": {
    "category_id": "category-uuid",
    "amount": "1500000.00",
    "period_type": "MONTHLY"
  },
  "requires_confirmation": true
}
```

Frontend shows confirmation.

Only after user confirmation:

```text
normal budget API
```

executes.

---

# 136. AI Copilot Conversation Contract

If persistent conversation is implemented:

```text
POST /api/v1/ai/conversations
GET  /api/v1/ai/conversations
GET  /api/v1/ai/conversations/:id
POST /api/v1/ai/conversations/:id/messages
DELETE /api/v1/ai/conversations/:id
```

This can remain P1.

Core AI Copilot may initially use stateless request/response.

---

# 137. Data Import API — Future

Potential future base:

```text
/api/v1/imports
```

Endpoints:

```text
POST /api/v1/imports/transactions
GET  /api/v1/imports/:id
GET  /api/v1/imports/:id/rows
POST /api/v1/imports/:id/confirm
POST /api/v1/imports/:id/cancel
```

Lifecycle:

```text
UPLOAD
→ PARSE
→ VALIDATE
→ REVIEW
→ IMPORT
```

No invalid row should silently become an authoritative financial record.

---

# 138. File Upload API — Future

Potential receipt upload:

```http
POST /api/v1/receipts
```

Content type:

```http
multipart/form-data
```

Validation:

```text
authentication
file size
MIME type
extension
ownership
safe object key
```

Binary storage:

```text
MinIO
```

Metadata:

```text
PostgreSQL
```

---

# 139. API Rate Limiting Strategy

Suggested categories:

## Authentication

```text
strict
```

Examples:

```text
login
register
refresh
```

## AI

```text
strict / cost-aware
```

Examples:

```text
copilot
categorization
scenario explanation
```

## Ordinary Reads

```text
more permissive
```

## Financial Writes

Protection may include:

```text
reasonable rate limits
idempotency
CSRF
authentication
```

---

# 140. API Security Headers

HTTP layer should consider:

```text
Content-Security-Policy
X-Content-Type-Options
Referrer-Policy
Permissions-Policy
Strict-Transport-Security
```

depending on deployment architecture.

If frontend and API are served separately, responsibility may be shared with the reverse proxy.

---

# 141. CORS

CORS must explicitly allow approved frontend origins.

Do not use:

```text
Access-Control-Allow-Origin: *
```

with credentialed authentication.

Example environment:

```text
FRONTEND_ORIGIN=https://savio.example.com
```

Requests use:

```text
credentials: include
```

---

# 142. Cache Control

Sensitive financial API responses should avoid inappropriate shared caching.

Recommended for authenticated financial data:

```http
Cache-Control: private, no-store
```

where appropriate.

---

# 143. Swagger / OpenAPI

Savio must maintain an OpenAPI document.

Recommended file:

```text
docs/api/openapi.yaml
```

Swagger UI may be served at:

```text
/api/docs
```

OpenAPI should document:

```text
authentication
cookies
CSRF header
requests
responses
errors
pagination
schemas
examples
```

---

# 144. OpenAPI Source of Truth

Ideally:

```text
API implementation
↔
OpenAPI contract
```

remain synchronized.

Possible strategies:

```text
spec-first
```

or:

```text
code annotations + generated spec
```

For this project, either is acceptable if documentation is accurate.

---

# 145. API Module Mapping

Recommended backend modules:

```text
auth
sessions
profile
settings

accounts
categories
transactions
transfers
recurring

budgets
goals

analytics
forecast
scenarios

ai
notifications
audit
```

---

# 146. Handler Responsibilities

HTTP handlers should handle:

```text
request binding
path/query extraction
validation invocation
auth context extraction
service invocation
response serialization
```

Handlers should not contain complex financial calculations.

---

# 147. Service Responsibilities

Services own:

```text
business rules
ownership validation
financial calculations
transactions
state transitions
cross-module orchestration
```

Example:

```text
TransactionService.CreateExpense()
```

rather than performing balance changes directly in HTTP handler.

---

# 148. Repository Responsibilities

Repositories own persistence.

Examples:

```text
AccountRepository
TransactionRepository
BudgetRepository
ScenarioRepository
```

Repository should not become the main home for business policy.

---

# 149. API Transaction Boundary Example

Create expense:

```text
POST /transactions
     ↓
Transaction Handler
     ↓
Transaction Service
     ↓
DB Transaction
     ├── lock/update account
     ├── insert transaction
     ├── mark forecast stale
     └── audit
     ↓
COMMIT
```

Only then should success be returned.

---

# 150. Background Job Triggering

When a transaction creates async follow-up work:

```text
transaction commit
↓
enqueue non-critical job
```

Potential jobs:

```text
recalculate signals
generate AI insight
send notification
```

Financial transaction correctness must not depend on the queue succeeding.

---

# 151. Queue Failure Contract

Example:

```text
Expense created successfully
but AI insight queue unavailable
```

The financial write should normally still succeed.

Backend logs queue failure.

Depending on implementation, retry or recovery mechanism handles the async side effect.

The client should not receive:

```text
500 transaction failed
```

when authoritative financial state already committed successfully.

---

# 152. Transactional Outbox — Future

If higher async reliability is needed:

```text
financial transaction
+
outbox event
```

can be written in the same PostgreSQL transaction.

Worker later publishes the event.

This is a P2 reliability improvement, not required before core behavior is correct.

---

# 153. Performance Rules

Endpoints returning collections must avoid:

```text
N+1 queries
unbounded result sets
SELECT *
```

where unnecessary.

Analytics should prefer database aggregation.

Important queries should be measurable using:

```text
EXPLAIN ANALYZE
```

when performance becomes relevant.

---

# 154. API Timeouts

External dependencies require bounded timeout.

Especially:

```text
AI provider
object storage
external future integrations
```

Database queries should also have request-scoped context cancellation.

---

# 155. Context Cancellation

When HTTP client disconnects or request timeout occurs:

```text
Go context
```

should propagate through:

```text
handler
service
repository
external AI request
```

where supported.

---

# 156. API Logging

Structured request logs should include:

```text
request_id
method
path
status
duration
user_id when authenticated
```

Do not log:

```text
password
auth cookie
refresh token
CSRF secret
full financial body indiscriminately
```

---

# 157. AI Logging

AI observability may include:

```text
request_id
feature
model
provider
latency
status
token usage
```

Do not automatically log:

```text
full financial context
full user transaction history
sensitive raw prompt
```

---

# 158. API Testing Contract

Every endpoint category should have tests for:

```text
success
authentication
ownership
validation
business errors
version conflict where applicable
```

Financial endpoints additionally need:

```text
balance correctness
transaction atomicity
concurrency behavior
```

---

# 159. Authentication API Tests

Required examples:

```text
register success
duplicate email
login success
wrong password
disabled user
refresh success
refresh rotation
revoked refresh
logout
logout all
CSRF failure
```

---

# 160. Transaction API Tests

Required:

```text
create income
create expense
category mismatch
unknown account
other-user account
archived account
update amount
update account
void transaction
version conflict
concurrent writes
```

---

# 161. Transfer API Tests

Required:

```text
successful transfer
same account rejected
other-user account rejected
archived account rejected
source/destination atomicity
voiding
version conflict
```

---

# 162. Forecast API Tests

Required:

```text
known events
scheduled recurring events
estimated spending
event ordering
minimum balance
ending balance
confidence classification
insufficient history
snapshot persistence
stale status
```

---

# 163. Scenario API Tests

Required:

```text
one-time expense
income reduction
income removal
recurring expense
multiple modifications
baseline isolation
real data unchanged
snapshot persistence
goal impact
scenario stale detection
```

---

# 164. AI API Tests

AI should be testable without live external model calls.

Use mock AI provider.

Required:

```text
valid structured output
invalid structured output
timeout
provider failure
AI disabled
rate limit
minimal context
tool execution
ownership protection
```

---

# 165. API MVP Scope

## P0

Required endpoints:

```text
Auth
Profile
Settings

Accounts
Categories

Transactions
Transfers

Recurring Transactions

Budgets
Goals

Dashboard
Analytics

Forecast

Scenarios

AI Categorization
AI Insights
AI Copilot
AI Scenario Explanation

Health
Readiness
```

---

## P1

High value:

```text
Sessions UI APIs

Notifications

Insight Feedback

AI Conversation Persistence

Advanced Reports
```

---

## P2

Future:

```text
Transaction Import
Receipt Upload
Receipt Extraction
Household Finance
Exports
Advanced Integration APIs
```

---

# 166. Core Endpoint Summary

```text
AUTH

POST /api/v1/auth/register
POST /api/v1/auth/login
POST /api/v1/auth/refresh
POST /api/v1/auth/logout
POST /api/v1/auth/logout-all
GET  /api/v1/auth/me
GET  /api/v1/auth/csrf
```

```text
PROFILE

GET   /api/v1/profile
PATCH /api/v1/profile
```

```text
SETTINGS

GET   /api/v1/settings
PATCH /api/v1/settings
```

```text
ACCOUNTS

GET    /api/v1/accounts
POST   /api/v1/accounts
GET    /api/v1/accounts/:id
PATCH  /api/v1/accounts/:id
POST   /api/v1/accounts/:id/archive
POST   /api/v1/accounts/:id/restore
POST   /api/v1/accounts/:id/reconcile
DELETE /api/v1/accounts/:id
```

```text
CATEGORIES

GET    /api/v1/categories
POST   /api/v1/categories
GET    /api/v1/categories/:id
PATCH  /api/v1/categories/:id
POST   /api/v1/categories/:id/archive
POST   /api/v1/categories/:id/restore
DELETE /api/v1/categories/:id
```

```text
TRANSACTIONS

GET   /api/v1/transactions
POST  /api/v1/transactions
GET   /api/v1/transactions/:id
PATCH /api/v1/transactions/:id
POST  /api/v1/transactions/:id/void
```

```text
TRANSFERS

GET  /api/v1/transfers
POST /api/v1/transfers
GET  /api/v1/transfers/:id
POST /api/v1/transfers/:id/void
```

```text
RECURRING

GET   /api/v1/recurring-transactions
POST  /api/v1/recurring-transactions
GET   /api/v1/recurring-transactions/:id
PATCH /api/v1/recurring-transactions/:id

POST /api/v1/recurring-transactions/:id/pause
POST /api/v1/recurring-transactions/:id/resume
POST /api/v1/recurring-transactions/:id/end

GET /api/v1/recurring-transactions/:id/occurrences
```

```text
BUDGETS

GET  /api/v1/budgets
POST /api/v1/budgets
GET  /api/v1/budgets/:id
PATCH /api/v1/budgets/:id
POST /api/v1/budgets/:id/archive
```

```text
GOALS

GET  /api/v1/goals
POST /api/v1/goals
GET  /api/v1/goals/:id
PATCH /api/v1/goals/:id

POST /api/v1/goals/:id/pause
POST /api/v1/goals/:id/resume
POST /api/v1/goals/:id/cancel
POST /api/v1/goals/:id/mark-achieved
```

```text
ANALYTICS

GET /api/v1/analytics/cashflow
GET /api/v1/analytics/categories
GET /api/v1/analytics/period-comparison
GET /api/v1/analytics/recurring-expenses
GET /api/v1/analytics/spending-changes
```

```text
FORECAST

GET  /api/v1/forecast
POST /api/v1/forecast/calculate
GET  /api/v1/forecast/history
GET  /api/v1/forecast/:snapshotId
```

```text
SCENARIOS

GET    /api/v1/scenarios
POST   /api/v1/scenarios
GET    /api/v1/scenarios/:id
PATCH  /api/v1/scenarios/:id
DELETE /api/v1/scenarios/:id

POST   /api/v1/scenarios/:id/modifications
PATCH  /api/v1/scenarios/:id/modifications/:modificationId
DELETE /api/v1/scenarios/:id/modifications/:modificationId

POST /api/v1/scenarios/:id/calculate

GET /api/v1/scenarios/:id/snapshots
GET /api/v1/scenarios/:id/snapshots/:snapshotId

POST /api/v1/scenarios/:id/archive
```

```text
AI

POST /api/v1/ai/categorize-transaction

GET  /api/v1/ai/insights
GET  /api/v1/ai/insights/:id
POST /api/v1/ai/insights/:id/view
POST /api/v1/ai/insights/:id/acknowledge
POST /api/v1/ai/insights/:id/dismiss
POST /api/v1/ai/insights/:id/feedback

POST /api/v1/ai/copilot
POST /api/v1/ai/explain-scenario
```

```text
DASHBOARD

GET /api/v1/dashboard
```

```text
SESSIONS

GET    /api/v1/sessions
DELETE /api/v1/sessions/:id
DELETE /api/v1/sessions
```

```text
NOTIFICATIONS

GET  /api/v1/notifications
GET  /api/v1/notifications/unread-count
POST /api/v1/notifications/:id/read
POST /api/v1/notifications/read-all
POST /api/v1/notifications/:id/archive
```

```text
INFRASTRUCTURE

GET /health
GET /ready
```

---

# 167. Core Request Flow

The canonical request flow is:

```text
HTTP Request
    ↓
Request ID Middleware
    ↓
Security Middleware
    ↓
Authentication
    ↓
CSRF Validation if required
    ↓
Rate Limiting
    ↓
Request Validation
    ↓
Handler
    ↓
Service
    ↓
Authorization / Ownership
    ↓
Business Rules
    ↓
Finance Engine if required
    ↓
Repository
    ↓
PostgreSQL
    ↓
Response
```

For AI:

```text
HTTP Request
    ↓
Authentication
    ↓
Rate Limit
    ↓
AI Orchestrator
    ↓
Finance Tools
    ↓
Deterministic Results
    ↓
Context Builder
    ↓
AI Provider
    ↓
Schema Validation
    ↓
Response
```

---

# 168. Financial API Authority Rule

The following backend modules are authoritative:

```text
Account Service
Transaction Service
Transfer Service
Recurring Service
Budget Service
Goal Service
Analytics Service
Forecast Engine
Scenario Engine
```

AI APIs consume output from these modules.

AI APIs do not replace them.

---

# 169. API Design Rule

Whenever a new Savio endpoint is proposed, ask:

1. Is the endpoint resource-oriented or a real business action?
2. Does the backend enforce ownership?
3. Are all inputs validated?
4. Is the status code correct?
5. Is the financial result deterministic?
6. Does the operation require a database transaction?
7. Is optimistic locking needed?
8. Is idempotency needed?
9. Does the endpoint invalidate forecast/scenario freshness?
10. Does it require audit logging?
11. If AI is involved, is the source data authoritative?
12. Can AI failure occur without corrupting financial state?

---

# 170. Final API Principle

The Savio API must make this impossible:

```text
Frontend guesses financial state
AI invents financial state
Client bypasses business rules
Resource ownership is inferred only from IDs
Concurrent writes silently overwrite each other
```

The intended hierarchy is:

```text
USER REQUEST
     ↓
AUTHENTICATION
     ↓
AUTHORIZATION
     ↓
VALIDATION
     ↓
BUSINESS RULE
     ↓
DETERMINISTIC FINANCE ENGINE
     ↓
AUTHORITATIVE RESULT
     ↓
OPTIONAL AI INTERPRETATION
     ↓
CLIENT RESPONSE
```

The final rule remains:

> **Finance Engine calculates. AI interprets. User decides.**