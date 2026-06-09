# Security Audit

This audit is based on the current source code only.

## High-Risk Findings

| Severity | Issue | Source | Impact | Recommendation |
| --- | --- | --- | --- | --- |
| Critical | JWT secret mismatch between token generation and validation | `internal/usecase/jwt.go`, `internal/usecase/validate_jwt.go`, `cmd/server/main.go` | If only `JWT_SECRET` is configured, login can issue tokens that auth middleware later rejects | Make token generation and validation read the same fallback sequence |
| High | Auth cookies are not marked `Secure` | `internal/handler/auth_handler.go` | Tokens can be exposed over non-HTTPS traffic | Set `Secure: true` in production and serve only over HTTPS |
| High | No rate limiting or abuse throttling | `internal/handler/auth_handler.go`, `internal/handler/ai_handler.go` | Brute-force, OTP spam, and AI endpoint abuse are possible | Add request throttling per IP and per account |
| High | Cookie-based auth has no CSRF defense layer | `internal/middleware/auth_middleware.go`, `internal/handler/auth_handler.go` | State-changing routes can be triggered with browser cookies in some scenarios | Add CSRF protection or move to header-only auth for browser sessions |
| Medium | Account enumeration is possible | `internal/handler/auth_handler.go` | Login, register, and password reset reveal whether an email exists or is verified | Return less specific responses for sensitive auth flows |
| Medium | Booking review eligibility is weaker than a true completed-booking check | `internal/usecase/review_usecase.go` | Users with any non-cancelled booking can review after the trip ends | Require a completed or settled booking state if that is the intended policy |
| Medium | Slot schema includes guide and driver IDs without matching entities or FKs | `internal/entity/trip_slot.go` | Data integrity relies on raw integers rather than enforced relations | Add guide/driver models or remove the columns until implemented |
| Medium | Pending payment state has no payment subsystem | `internal/usecase/booking_usecase.go`, `internal/entity/booking.go` | Bookings can remain in an unresolved intermediate state | Implement payment confirmation and reconciliation flow |

## Token Issues

- `GenerateAccessToken` accepts `JWT_SECRET` first, then `JWT_SECRETKEY`.
- `ValidateJwt` reads only `JWT_SECRETKEY`.
- Access tokens expire in 1 hour, but the cookie is set to 15 minutes.
- Refresh tokens are stored hashed, which is good, but the code relies on cookie delivery for browser sessions.

## Validation Gaps

- Most DTOs are parsed manually without a centralized validator.
- Several handlers check only for presence of required fields and do not enforce length or format rules consistently.
- `TripSlot` updates recalculate seat counts from totals and booked seats, so the inbound `available_seats` field is not authoritative.

## Authorization Risks

- The permission middleware gives admins a global bypass.
- `RoleMiddleware("admin")` is used for admin-only routes, but some operational routes rely solely on permissions.
- The auth middleware accepts both bearer tokens and cookies, so browser sessions and API clients share the same auth surface.
- Notification read access is role-aware, but route ownership still depends on the calling path being correct.

## SQL and ORM Risks

- Most database access uses GORM with parameterized queries.
- The only dynamic SQL fragment identified is the slot assignment overlap helper, and its column names are hard-coded internally.
- No direct SQL injection path was found in the inspected repository.

## Missing Operational Protections

- No rate limiter middleware
- No audit log table or append-only audit trail
- No webhook signature verification because no payment or webhook system exists yet
- No CSRF middleware
- No websocket authentication because no websocket server exists

## Source References

- `internal/usecase/jwt.go`
- `internal/usecase/validate_jwt.go`
- `internal/handler/auth_handler.go`
- `internal/middleware/auth_middleware.go`
- `internal/usecase/review_usecase.go`
- `internal/usecase/booking_usecase.go`
- `internal/entity/trip_slot.go`
