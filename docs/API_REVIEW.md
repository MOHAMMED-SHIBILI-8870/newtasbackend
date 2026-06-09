# API Review

This review is based on the current code in `internal/routes`, `internal/handler`, `internal/usecase`, `internal/repository`, `internal/entity`, and middleware.

## Findings

| Severity | Finding | Evidence | Impact | Recommendation |
| --- | --- | --- | --- | --- |
| Critical | JWT signing and validation do not read the same secret fallback sequence | `internal/usecase/jwt.go`, `internal/usecase/validate_jwt.go`, `cmd/server/main.go` | Access tokens can be issued successfully and then rejected by auth middleware if only `JWT_SECRET` is configured | Make token generation and validation share one secret lookup helper |
| High | Cookie-based auth has no CSRF defense and cookies are not marked secure | `internal/handler/auth_handler.go`, `internal/middleware/auth_middleware.go` | Browser sessions can be abused on state-changing routes in unsafe deployments | Add CSRF protection or move browser sessions to a safer auth pattern; set `Secure: true` in production |
| High | The booking flow still allows a confirmed booking path without any payment module | `internal/usecase/booking_usecase.go`, `internal/handler/booking_handler.go`, `internal/entity/booking.go` | `POST /bookings/trip/:id` can confirm a booking immediately, while the newer slot path stops at `pending_payment` | Introduce the payment workflow and retire the legacy direct-confirm path once the new flow is stable |
| High | The legacy direct-trip booking route is a second booking code path with duplicated logic | `internal/handler/routes/booking_routes.go`, `internal/usecase/booking_usecase.go` | Booking rules, pricing, and notifications can drift between the trip and slot paths | Consolidate booking creation behind one payment-aware flow |
| Medium | Review eligibility is weaker than the desired completed-booking policy | `internal/usecase/review_usecase.go` | A user can review after the trip ends if they have any non-cancelled booking; there is no explicit completed-booking gate | Add a clear completion rule and enforce it in the use case |
| Medium | Complaint data model only supports booking complaints | `internal/entity/complaint.go`, `internal/dto/complaint_dto.go`, `internal/usecase/complaint_usecase.go` | Platform issues and trip-specific issues cannot be represented cleanly | Add `complaint_type`, make `booking_id` nullable, and add nullable `trip_id` |
| Medium | List endpoints have no pagination and only minimal filtering | `internal/repository/*`, `internal/handler/*` | Users, trips, bookings, reviews, complaints, offers, and notifications can all grow without bounds | Add page/limit and filtering consistently across list endpoints |
| Medium | Slot assignment columns exist without matching guide/driver domain entities | `internal/entity/trip_slot.go`, `internal/usecase/trip_slot_usecase.go` | `guide_id` and `driver_id` are raw IDs with no FK-backed entities or management routes | Add guide/driver models or remove the fields until they are fully modeled |
| Medium | Several endpoints are best-effort on notifications instead of transactional on the whole write | `internal/usecase/booking_usecase.go`, `internal/usecase/ai_trip_request_usecase.go` | Primary writes succeed even if notification fan-out fails, which can hide operational issues | Decide whether this is intentional; if not, move notification creation into the same transaction or add retries |
| Low | Admin notification listing is filtered to booking notifications only | `internal/repository/notification_repository.go` | Other admin-facing notifications are not visible in `/admin/notifications` | Rename or broaden the endpoint if it is meant to be a true admin inbox |
| Low | The booking request body includes `booking_type` for both booking routes, but the legacy trip route ignores it | `internal/handler/booking_handler.go`, `internal/usecase/booking_usecase.go` | The direct-trip path accepts a field that has no effect, which is confusing for clients | Remove the unused field from the legacy route or route both paths through one booking DTO |

## Missing Pagination And Filtering

- `/admin/users`
- `/trips`
- `/admin/trips`
- `/admin/slots`
- `/bookings/my-orders`
- `/admin/orders/all`
- `/reviews/me`
- `/reviews/trip/:trip_id`
- `/admin/reviews`
- `/complaints/me`
- `/admin/complaints`
- `/notifications`
- `/admin/notifications`
- `/offers/active`
- `/admin/offers`
- `/tracking/booking/:booking_id/history`
- `/admin/tracking`

## Foreign Key And Schema Gaps

- `internal/entity/complaint.go` requires `booking_id`, so platform complaints are not representable.
- `internal/entity/trip_slot.go` has `guide_id` and `driver_id` columns, but there are no corresponding entities or FK-backed repositories.
- `internal/entity/booking.go` links to `trip_id`, `slot_id`, `vehicle_id`, and `offer_id`, but payment summary fields are still missing.
- `internal/entity/booking.go` has no payment history table backing the `pending_payment` state.

## Duplicate Or Transitional Paths

- `POST /bookings/trip/:id` is a legacy compatibility route that duplicates much of the slot booking flow.
- `internal/repository/booking_repo.go` still contains `CreateBooking`, while the transactional path uses `CreateBookingTx`.
- `internal/usecase/jwt.go` still contains `RefreshAccessToken`, while the live auth flow uses `RotateRefreshToken`.
- `internal/usecase/notification_usecase.go` still contains `CreateAIReviewNotification`, while AI review notifications are written inline in the AI use case.

## Review Summary

- The codebase is structurally consistent and already uses transactions for the most important multi-step writes.
- The main product gaps are the missing payment workflow, the weaker-than-intended review policy, and the complaint model that cannot yet represent platform or trip-level issues.
- The main security concern is the auth secret mismatch combined with cookie-based auth that lacks CSRF hardening.
