# Testing Guide

## Existing Tests

The repository currently has one meaningful test file:

- `internal/usecase/security_flows_test.go`

It uses an in-memory SQLite database and auto-migrates the core models needed for the tested flows.

### Covered flows

- OTP generation and verification
- Refresh token rotation
- Refresh token rejection for blocked users
- User blocking revokes refresh tokens
- Role assignment keeps `user_roles` and `users.role` aligned
- AI trip request approval is transactional
- Trip booking rounds money correctly and creates notifications

## Coverage Areas

The current tests focus on business rules in `internal/usecase/`.

That means the following layers are only lightly covered or not covered directly:

- HTTP handlers
- Route registration
- Middleware
- Repository query behavior
- Slot booking flow
- Vehicle assignment flow
- Offer CRUD
- Reviews
- Complaints
- Tracking
- Notification read permissions

## Missing Tests

Important gaps in the current suite:

- Auth handler validation and error mapping
- JWT validation using `JWT_SECRET` vs `JWT_SECRETKEY`
- Public trip list and search endpoints
- Admin trip CRUD
- Trip slot CRUD and overlap validation
- Slot booking path and pending-payment state
- Legacy direct-trip booking compatibility path
- Vehicle permissions for agency users
- Offer coupon validation and expiry handling
- Review eligibility enforcement
- Complaint ownership enforcement
- Tracking authorization for driver/admin vs user
- RBAC endpoints and permission middleware
- Notification visibility and read-marking authorization

## Recommended Test Cases

### Auth

- Register with duplicate email
- Login with blocked account
- Login with unverified account
- Refresh rotation invalidates old token
- Password reset revokes all refresh tokens

### Booking and slots

- Slot booking with shared vs private booking types
- Seat underflow and overflow
- Duplicate booking for same slot
- Booking with inactive or expired offer
- Trip slot overlap for the same vehicle
- Trip slot overlap for guide/driver assignment IDs

### RBAC and admin

- Role creation rejects non-canonical names
- Default roles cannot be deleted
- Permission assignment and removal
- Admin user deletion restrictions
- User block/unblock revokes sessions

### Reviews, complaints, tracking

- Review before trip end should fail
- Review without a booking should fail
- Complaint can only be filed by booking owner
- Tracking can only be updated by a user with `manage_tracking`
- Driver and admin can read tracking for a booking

## Suggested Test Strategy

- Keep the current usecase tests for business rule coverage.
- Add handler tests for request parsing and status mapping.
- Add repository tests for preload and query behavior.
- Add integration tests for key transactional flows.

## Source References

- `internal/usecase/security_flows_test.go`
- `internal/usecase/*`
- `internal/handler/*`
- `internal/repository/*`
