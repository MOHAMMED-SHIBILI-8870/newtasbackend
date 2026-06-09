# Codebase Analysis

## Dead Code and Unused Paths

| Item | File | Observation |
| --- | --- | --- |
| `optionalEnv` | `internal/config/validation.go` | Defined but not used |
| `CreateBooking` | `internal/repository/booking_repo.go` | `CreateBookingTx` is the path used by current booking flows |
| `AdjustAvailableSeats` | `internal/repository/vehicle_repository.go` | Not referenced by the current usecases |
| `isSlotAssignmentColumn` | `internal/repository/trip_slot_repository.go` | Helper appears unused |
| `CreateAIReviewNotification` | `internal/usecase/notification_usecase.go` | AI approval creates notifications inline instead |
| `RefreshAccessToken` | `internal/usecase/jwt.go` | Current auth flow uses `RotateRefreshToken` |
| `BookingResponse` | `internal/entity/booking.go` | Defined but not used as a primary response type |
| `MarkNotificationReadRequest` | `internal/dto/notification.go` | Not used by the current handler code |
| `AITripReviewInput` | `internal/entity/ai_trip_request.go` | DTO exists but handlers use `AITripReviewRequest` instead |
| `Message`, `ChatRequest`, `ChatResponse` | `internal/entity/ai.go` | No repository or handler path currently persists or reads them |

## Duplicate Logic

- Booking money math is duplicated in `BookTrip` and `BookSlot`.
- Offer lookup and expiry validation appear in both booking paths.
- Admin and user notification creation is repeated for both booking methods.
- Several handlers convert string-based errors to status codes with similar logic.

## Refactoring Opportunities

1. Extract a shared booking pricing helper so trip booking and slot booking do not drift.
2. Replace string matching in handler error mappers with typed errors.
3. Add a centralized validation layer for DTOs instead of scattered manual checks.
4. Add guide and driver aggregates if `TripSlot.GuideID` and `TripSlot.DriverID` are meant to be real relations.
5. Consolidate booking notification creation into one helper.
6. Normalize the JWT secret lookup so generation and validation use the same logic.

## Performance Notes

- `VehicleUsecase.ListVehicles` loads all vehicles for non-admin/non-agency users and filters in memory.
- `BookingUsecase.UpdateUserBookingPlans` deletes and recreates plans instead of diffing updates.
- `ReviewUsecase.GetTripAverageRating` loads all reviews and computes the average in memory.
- `PermissionUsecase.GetUserPermissions` resolves and deduplicates permissions in application code.
- Several endpoints fetch user, booking, or trip records repeatedly during auth and authorization checks.

## Maintainability Notes

- `cmd/server/main.go` acts as the main composition root and is already the correct place for dependency wiring.
- The repository layer is consistently separated from business logic, which makes the codebase testable.
- The new slot-first booking path is integrated cleanly, but the legacy booking route still introduces a second code path that will need to be retired later.

## Source References

- `internal/config/validation.go`
- `internal/repository/*`
- `internal/usecase/*`
- `internal/entity/*`
- `internal/handler/*`
