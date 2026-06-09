# Changelog From Code

This changelog is inferred from the current source tree and the implemented feature set.

## New Features

- Slot-first booking flow added through `POST /bookings/slot/:slot_id`
- `TripSlot` entity, repository, usecase, handler, and routes
- Slot seat accounting and overlap validation
- Booking records now capture `SlotID`, `BookingType`, `BaseAmount`, `FinalAmount`, `CouponCode`, and `DiscountPercent`
- `pending_payment` booking state for slot bookings
- Admin and public slot endpoints
- Transactional AI trip request review that can create a trip on approval
- Notification generation for booking and AI request events
- RBAC seeding for canonical roles and permissions
- Role and permission admin endpoints
- Core security-flow test coverage in `internal/usecase/security_flows_test.go`

## Modified Features

- Booking is no longer purely trip-template driven; it now has a slot-first path and still keeps the legacy trip path for compatibility
- Refresh token handling is single-session per user
- Admin user blocking now revokes refresh tokens
- Role assignment syncs both `users.role` and `user_roles`
- Vehicle assignment is transactional and locks rows to prevent double assignment
- AI trip generation supports a Gemini client or a local fallback generator
- OTP generation and verification now use hashed storage with row invalidation for prior codes

## Deprecated or Transitional Features

| Feature | Status |
| --- | --- |
| `POST /bookings/trip/:id` | Legacy compatibility route, still active but no longer the primary path |
| `CreateBooking` repository method | Transitional; current booking usecases use `CreateBookingTx` |
| `BookingResponse` entity DTO | Defined but not used as the main response type |
| `CreateAIReviewNotification` helper | Transitional; AI review notifications are currently written inline in the approval transaction |
| `ChatResponse` persistence path | Auto-migrated, but no repository or handler consumes it yet |

## Notes

- The new slot-first booking path is the clearest change in the current code snapshot.
- The presence of `pending_payment` indicates a payment workflow is expected but not yet implemented.
- Existing docs such as `docs/api-endpoints.md` should be treated as stale compared with the code-backed docs generated here.

## Source References

- `internal/usecase/booking_usecase.go`
- `internal/entity/trip_slot.go`
- `internal/entity/booking.go`
- `internal/usecase/ai_trip_request_usecase.go`
- `internal/usecase/otp_verification.go`
- `internal/usecase/jwt.go`
- `internal/usecase/role_usecase.go`
