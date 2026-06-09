# Feature Status Report

This report categorizes features based on the current source tree.

## Completed

| Feature | Evidence |
| --- | --- |
| Authentication | `internal/handler/auth_handler.go`, `internal/usecase/jwt.go`, `internal/usecase/otp_verification.go` |
| Public trips and admin trip CRUD | `internal/handler/routes/trip_routes.go`, `internal/handler/trip_handler.go` |
| Trip plans | `internal/handler/routes/trip_plans.go`, `internal/handler/trip_plan_handler.go` |
| Slot CRUD and trip-slot listing | `internal/handler/routes/trip_slot_routes.go`, `internal/handler/trip_slot_handler.go` |
| Slot-first booking flow | `internal/handler/routes/booking_routes.go`, `internal/usecase/booking_usecase.go` |
| Legacy trip booking compatibility | `internal/handler/routes/booking_routes.go`, `internal/usecase/booking_usecase.go` |
| Fleet management | `internal/handler/routes/vehicle_routes.go`, `internal/usecase/vehicle_usecase.go` |
| Offers | `internal/handler/routes/offer_routes.go`, `internal/usecase/offer_usecase.go` |
| Reviews | `internal/handler/routes/review_routes.go`, `internal/usecase/review_usecase.go` |
| Complaints | `internal/handler/routes/complaint_routes.go`, `internal/usecase/complaint_usecase.go` |
| Tracking | `internal/handler/routes/tracking_routes.go`, `internal/usecase/tracking_usecase.go` |
| Notifications | `internal/handler/routes/notification_routes.go`, `internal/usecase/notification_usecase.go` |
| RBAC and admin user management | `internal/handler/routes/rbac_routes.go`, `internal/handler/routes/admin_routes.go` |
| AI trip generation and request review | `internal/handler/routes/ai_routes.go`, `internal/usecase/ai_trip_request_usecase.go` |
| Startup migrations and seeds | `migrations/migrate.go`, `internal/seed/seed.go`, `internal/seed/rbac_seed.go` |
| Usecase tests for core security flows | `internal/usecase/security_flows_test.go` |

## Partially Implemented

| Feature | Why it is partial |
| --- | --- |
| Payment flow | Bookings can enter `pending_payment`, but no payment module exists |
| Guide assignment | `TripSlot.GuideID` exists, but there is no guide entity, handler, or repository |
| Driver assignment | `TripSlot.DriverID` exists, but there is no driver entity, handler, or repository |
| AI chat persistence | `ChatResponse` is auto-migrated, but no repository or handler stores chat messages |
| Admin notifications | Notifications exist, but admin notifications are filtered by `type = admin_booking` only |
| Review policy | The code checks for any non-cancelled booking, not a stricter completed-booking rule |

## Missing

| Feature | Evidence of absence |
| --- | --- |
| Payment gateway | No payment packages, routes, or integrations |
| Booking cancellation/refund | No route or usecase path |
| Waitlist | No entity or handler |
| WebSocket live updates | No websocket server or routes |
| Cron jobs / workers | No scheduler or worker code |
| Docker / CI/CD | No Dockerfile, compose file, or CI workflow |
| Guide management module | No guide entity or admin CRUD |
| Driver management module | No driver entity or admin CRUD |

## Planned or Implied by the Code

These are not explicitly planned in a roadmap file, but the code structure suggests they are intended future work:

- Payment confirmation for `pending_payment` bookings
- Full guide and driver domain models for slot assignments
- A richer notification center for more than booking and AI request events
- A retirement path for the legacy direct-trip booking route

## Source References

- `internal/entity/trip_slot.go`
- `internal/entity/booking.go`
- `internal/handler/routes/*`
- `internal/usecase/*`
