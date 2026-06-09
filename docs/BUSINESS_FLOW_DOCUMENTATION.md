# Business Flow Documentation

The diagrams below are based on the actual handlers and use cases in the repository.

## Trip Creation

```mermaid
sequenceDiagram
    autonumber
    participant Admin
    participant H as TripHandler
    participant U as TripUsecase
    participant R as TripRepository
    participant DB as PostgreSQL

    Admin->>H: POST /admin/trips/
    H->>H: parse entity.Trip
    H->>U: CreateTrip(trip)
    U->>U: validate from/to, duration, price
    U->>R: Create(ctx, trip)
    R->>DB: INSERT trips
    DB-->>R: created row
    R-->>U: trip persisted
    U-->>H: success
    H-->>Admin: Envelope{trip}
```

Notes:

- File locations: `internal/handler/trip_handler.go`, `internal/usecase/trip_usecase.go`, `internal/repository/trip_repository.go`
- Tables touched: `trips`

## Slot Creation

```mermaid
sequenceDiagram
    autonumber
    participant Admin
    participant H as TripSlotHandler
    participant U as TripSlotUsecase
    participant SR as TripSlotRepository
    participant TR as TripRepository
    participant VR as VehicleRepository
    participant DB as PostgreSQL

    Admin->>H: POST /admin/slots/
    H->>H: parse entity.TripSlot
    H->>U: CreateSlot(ctx, slot)
    U->>U: validate dates, seats, price, status
    U->>DB: transaction
    U->>TR: load trip
    U->>VR: validate vehicle assignment
    U->>SR: CreateTx(tx, slot)
    SR->>DB: INSERT trip_slots
    DB-->>SR: created slot
    SR-->>U: slot persisted
    U-->>H: success
    H-->>Admin: Envelope{slot}
```

Notes:

- Slot creation also enforces overlap checks for vehicle, guide, and driver assignment IDs.
- File locations: `internal/handler/trip_slot_handler.go`, `internal/usecase/trip_slot_usecase.go`, `internal/repository/trip_slot_repository.go`
- Tables touched: `trip_slots`, `trips`, `vehicles`

## Booking

### Slot booking path

```mermaid
sequenceDiagram
    autonumber
    participant User
    participant H as BookingHandler
    participant U as BookingUsecase
    participant SR as TripSlotRepository
    participant BR as BookingRepository
    participant OR as OfferRepository
    participant NR as NotificationUsecase
    participant DB as PostgreSQL

    User->>H: POST /bookings/slot/:slot_id
    H->>H: parse seats/coupon_code/booking_type
    H->>U: BookSlot(slot_id, user_id, seats, coupon, type)
    U->>OR: validate coupon (optional)
    U->>DB: transaction
    U->>SR: load and lock slot + trip + plans
    U->>U: validate slot status, seats, duplicate booking
    U->>SR: update seat counts and status
    U->>BR: CreateBookingTx(tx, booking)
    BR->>DB: INSERT bookings + booking_plans
    DB-->>BR: created booking
    BR-->>U: booking persisted
    U-->>NR: create user/admin notifications
    U-->>H: success
    H-->>User: Envelope{booking}
```

### Legacy direct-trip booking path

```mermaid
sequenceDiagram
    autonumber
    participant User
    participant H as BookingHandler
    participant U as BookingUsecase
    participant VR as VehicleRepository
    participant BR as BookingRepository
    participant NR as NotificationUsecase
    participant DB as PostgreSQL

    User->>H: POST /bookings/trip/:id
    H->>U: BookTrip(trip_id, user_id, seats, coupon)
    U->>DB: transaction
    U->>VR: lock assigned vehicle (if any)
    U->>U: validate offer and seat availability
    U->>BR: CreateBookingTx(tx, booking)
    BR->>DB: INSERT bookings + booking_plans
    DB-->>BR: created booking
    BR-->>U: booking persisted
    U-->>NR: create user/admin notifications
    U-->>H: success
    H-->>User: Envelope{booking}
```

Notes:

- File locations: `internal/handler/booking_handler.go`, `internal/usecase/booking_usecase.go`, `internal/repository/booking_repo.go`
- Tables touched: `bookings`, `booking_plans`, `trip_slots`, `trips`, `vehicles`, `offers`, `notifications`

## Notifications

### Booking notifications

```mermaid
sequenceDiagram
    autonumber
    participant U as BookingUsecase
    participant N as NotificationUsecase
    participant R as NotificationRepository
    participant DB as PostgreSQL

    U->>N: CreateBookingNotification(user_id, booking_id, message)
    N->>N: build notification payload
    N->>R: Create(ctx, notification)
    R->>DB: INSERT notifications
    DB-->>R: created row
    R-->>N: saved
    N-->>U: ok
```

### AI request notifications

```mermaid
sequenceDiagram
    autonumber
    participant U as AITripRequestUsecase
    participant N as NotificationUsecase
    participant R as NotificationRepository
    participant DB as PostgreSQL
    participant Admin

    U->>U: create AI trip request
    U->>N: notify admins about the request
    N->>R: Create(ctx, notification)
    R->>DB: INSERT notifications
    DB-->>R: created row
    R-->>N: saved
    N-->>U: ok
    Admin->>DB: later reads admin notifications
```

Notes:

- File locations: `internal/usecase/notification_usecase.go`, `internal/repository/notification_repository.go`, `internal/usecase/booking_usecase.go`, `internal/usecase/ai_trip_request_usecase.go`
- Tables touched: `notifications`

## Reviews

```mermaid
sequenceDiagram
    autonumber
    participant User
    participant H as ReviewHandler
    participant U as ReviewUsecase
    participant BR as BookingRepository
    participant TR as TripRepository
    participant RR as ReviewRepository
    participant DB as PostgreSQL

    User->>H: POST /reviews/
    H->>U: CreateReview(user_id, trip_id, rating, comment)
    U->>DB: transaction
    U->>TR: lock trip and check end date
    U->>BR: verify user has non-cancelled booking
    U->>RR: check for duplicate review
    U->>RR: Create(ctx, review)
    RR->>DB: INSERT reviews
    DB-->>RR: created row
    RR-->>U: saved review
    U-->>H: success
    H-->>User: Envelope{review}
```

Notes:

- Review eligibility is based on trip completion and the existence of a non-cancelled booking.
- File locations: `internal/handler/review_handler.go`, `internal/usecase/review_usecase.go`, `internal/repository/review_repository.go`
- Tables touched: `reviews`, `bookings`, `trips`

## Complaints

```mermaid
sequenceDiagram
    autonumber
    participant User
    participant H as ComplaintHandler
    participant U as ComplaintUsecase
    participant BR as BookingRepository
    participant CR as ComplaintRepository
    participant DB as PostgreSQL

    User->>H: POST /complaints/
    H->>U: CreateComplaint(user_id, booking_id, title, description)
    U->>BR: load booking
    U->>U: verify booking ownership
    U->>CR: Create(ctx, complaint)
    CR->>DB: INSERT complaints
    DB-->>CR: created row
    CR-->>U: saved complaint
    U-->>H: success
    H-->>User: Envelope{complaint}
```

Notes:

- Admin review uses the same complaint record and updates only the status field.
- File locations: `internal/handler/complaint_handler.go`, `internal/usecase/complaint_usecase.go`, `internal/repository/complaint_repository.go`
- Tables touched: `complaints`, `bookings`

## Tracking

```mermaid
sequenceDiagram
    autonumber
    participant Driver as Driver or Operator
    participant H as TrackingHandler
    participant U as TrackingUsecase
    participant BR as BookingRepository
    participant VR as VehicleRepository
    participant TR as TrackingRepository
    participant DB as PostgreSQL

    Driver->>H: POST /tracking/
    H->>U: UpdateLocation(booking_id, vehicle_id, lat, lng)
    U->>BR: load booking
    U->>VR: load vehicle and verify trip match
    U->>TR: Create(ctx, tracking)
    TR->>DB: INSERT trackings
    DB-->>TR: created row
    TR-->>U: saved tracking
    U-->>H: success
    H-->>Driver: Envelope{tracking}
```

Notes:

- Admin and driver roles can read tracking for any booking; regular users can only read their own booking tracking.
- File locations: `internal/handler/tracking_handler.go`, `internal/usecase/tracking_usecase.go`, `internal/repository/tracking_repository.go`
- Tables touched: `trackings`, `bookings`, `vehicles`

## AI Requests

```mermaid
sequenceDiagram
    autonumber
    participant User
    participant H as AIHandler
    participant U as AITripRequestUsecase
    participant TR as TripRepository
    participant NR as NotificationUsecase
    participant DB as PostgreSQL
    participant Admin

    User->>H: POST /ai/requests
    H->>U: CreateRequest(user_id, request)
    U->>DB: INSERT ai_trip_requests
    DB-->>U: request created
    U->>NR: notify admins about pending review
    NR->>DB: INSERT notifications
    DB-->>NR: saved
    NR-->>U: ok
    U-->>H: success
    H-->>User: Envelope{ai request}

    Admin->>H: PATCH /admin/ai-requests/:id/approve
    H->>U: ReviewRequest(admin_id, request_id, approve=true)
    U->>DB: transaction
    U->>TR: create trip from generated plan
    U->>DB: update request status and notification
    DB-->>U: committed
    U-->>H: updated request
    H-->>Admin: Envelope{ai request}
```

Notes:

- File locations: `internal/handler/ai_handler.go`, `internal/usecase/ai_trip_request_usecase.go`, `internal/repository/ai_trip_request_repository.go`
- Tables touched: `ai_trip_requests`, `trips`, `notifications`

## Source Files

- `internal/handler/*`
- `internal/usecase/*`
- `internal/repository/*`
