# API Endpoint Documentation

> Legacy note: this file is retained for historical context only. The source-of-truth API reference is [API_DOCUMENTATION.md](./API_DOCUMENTATION.md).

All endpoints return the shared envelope:

```json
{
  "success": true,
  "message": "string",
  "data": {}
}
```

Failures return the same envelope with `success: false` and a sanitized `error` field.

## Auth

### POST /auth/register
- Description: Register a user and send a signup OTP.
- Auth: No.
- Roles: None.
- Request body: `dto.RegisterRequest` (`full_name`, `email`, `password`).
- Query params: None.
- Path params: None.
- Success response: `201 Envelope[data: { user: dto.AuthUserResponse }]`.
- Error responses: `400`, `409`, `500`.
- Validation: `full_name`, `email`, and `password` are required; password min length is 6.
- DB tables: `users`, `otps`.

### POST /auth/login
- Description: Authenticate a user and issue access/refresh tokens.
- Auth: No.
- Roles: None.
- Request body: `dto.LoginRequest` (`email`, `password`).
- Query params: None.
- Path params: None.
- Success response: `200 Envelope[data: dto.AuthResponse]` plus `access_token` and `refresh_token` cookies.
- Error responses: `400`, `401`, `403`, `404`, `500`.
- Validation: Email and password required; user must be verified and not blocked.
- DB tables: `users`, `refresh_tokens`.

### POST /auth/verify-otp
- Description: Verify an OTP for signup or password reset.
- Auth: No.
- Roles: None.
- Request body: `dto.OTPRequest` (`email`, `otp`, `purpose`).
- Query params: None.
- Path params: None.
- Success response: `200 Envelope[data: null]`.
- Error responses: `400`, `404`, `500`.
- Validation: `email`, `otp`, and `purpose` are required.
- DB tables: `otps`, `users`.

### POST /auth/forgot-password
- Description: Send a password-reset OTP.
- Auth: No.
- Roles: None.
- Request body: `dto.ForgotPasswordRequest` (`email`).
- Query params: None.
- Path params: None.
- Success response: `200 Envelope[data: null]`.
- Error responses: `400`, `404`, `500`.
- Validation: Email required.
- DB tables: `users`, `otps`.

### POST /auth/reset-password
- Description: Verify reset OTP, update password, and revoke refresh tokens.
- Auth: No.
- Roles: None.
- Request body: `dto.ResetPasswordRequest` (`email`, `new_password`, `otp`).
- Query params: None.
- Path params: None.
- Success response: `200 Envelope[data: null]`.
- Error responses: `400`, `404`, `500`.
- Validation: `email`, `new_password`, and `otp` are required; new password min length is 6.
- DB tables: `users`, `otps`, `refresh_tokens`.

### POST /auth/resend-otp
- Description: Resend a signup or reset OTP.
- Auth: No.
- Roles: None.
- Request body: `dto.ResendOTPRequest` (`email`, `purpose`).
- Query params: None.
- Path params: None.
- Success response: `200 Envelope[data: null]`.
- Error responses: `400`, `404`, `500`.
- Validation: `email` and `purpose` are required; signup OTP is blocked for already verified users.
- DB tables: `users`, `otps`.

### POST /auth/logout
- Description: Clear auth cookies and delete the refresh token if present.
- Auth: No.
- Roles: None.
- Request body: None.
- Query params: None.
- Path params: None.
- Success response: `200 Envelope[data: null]`.
- Error responses: `200` only; token deletion errors are ignored by design.
- Validation: None.
- DB tables: `refresh_tokens`.

### POST /auth/refresh
- Description: Rotate the refresh token and issue a new access token.
- Auth: No.
- Roles: None.
- Request body: None.
- Query params: None.
- Path params: None.
- Success response: `200 Envelope[data: dto.AuthResponse]` plus refreshed cookies.
- Error responses: `401`, `403`, `500`.
- Validation: Refresh cookie required; token must be valid, unexpired, and owned by an unblocked, verified user.
- DB tables: `refresh_tokens`, `users`.

## Trips

### GET /trips/
- Description: List all public trip templates.
- Auth: No.
- Roles: None.
- Request body: None.
- Query params: None.
- Path params: None.
- Success response: `200 Envelope[data: []entity.Trip]`.
- Error responses: `500`.
- Validation: None.
- DB tables: `trips`, `trip_plans`.

### GET /trips/:name
- Description: Search a trip template by name fragment.
- Auth: No.
- Roles: None.
- Request body: None.
- Query params: None.
- Path params: `name`.
- Success response: `200 Envelope[data: entity.Trip]`.
- Error responses: `400`, `404`, `500`.
- Validation: `name` must be non-empty and URL-decoded.
- DB tables: `trips`, `trip_plans`.

### POST /admin/trips/
- Description: Create a master trip template.
- Auth: Yes.
- Roles: `admin`.
- Request body: `entity.Trip`.
- Query params: None.
- Path params: None.
- Success response: `201 Envelope[data: entity.Trip]`.
- Error responses: `400`, `422`, `500`.
- Validation: `from`, `to`, `start_date`, `end_date`, and `duration` are required; price cannot be negative.
- DB tables: `trips`, `trip_plans`.

### GET /admin/trips/
- Description: List all trip templates for administrators.
- Auth: Yes.
- Roles: `admin`.
- Request body: None.
- Query params: None.
- Path params: None.
- Success response: `200 Envelope[data: []entity.Trip]`.
- Error responses: `500`.
- Validation: None.
- DB tables: `trips`, `trip_plans`.

### PATCH /admin/trips/:id
- Description: Update a trip template and optionally replace plans.
- Auth: Yes.
- Roles: `admin`.
- Request body: `entity.UpdateTripInput`.
- Query params: None.
- Path params: `id`.
- Success response: `200 Envelope[data: null]`.
- Error responses: `400`, `404`, `422`, `500`.
- Validation: Partial fields are allowed; `duration` must be at least 1 day and `price` cannot be negative.
- DB tables: `trips`, `trip_plans`.

### DELETE /admin/trips/:id
- Description: Delete a trip template.
- Auth: Yes.
- Roles: `admin`.
- Request body: None.
- Query params: None.
- Path params: `id`.
- Success response: `200 Envelope[data: null]`.
- Error responses: `400`, `404`, `500`.
- Validation: `id` must be a valid positive integer.
- DB tables: `trips`, `trip_plans`.

## Trip Slots

### POST /admin/slots/
- Description: Create a trip slot for a specific departure window.
- Auth: Yes.
- Roles: `admin`.
- Request body: `entity.TripSlot`.
- Query params: None.
- Path params: None.
- Success response: `201 Envelope[data: entity.TripSlot]`.
- Error responses: `400`, `404`, `409`, `422`, `500`.
- Validation: `trip_id`, `start_date`, `end_date`, and `total_seats` are required; the end date must be after the start date; vehicle, guide, and driver assignments cannot overlap.
- DB tables: `trip_slots`, `trips`, `vehicles`.

### GET /admin/slots/
- Description: List all trip slots for administrators.
- Auth: Yes.
- Roles: `admin`.
- Request body: None.
- Query params: None.
- Path params: None.
- Success response: `200 Envelope[data: []entity.TripSlot]`.
- Error responses: `500`.
- Validation: None.
- DB tables: `trip_slots`, `trips`, `vehicles`.

### GET /admin/slots/:id
- Description: Load one trip slot by ID.
- Auth: Yes.
- Roles: `admin`.
- Request body: None.
- Query params: None.
- Path params: `id`.
- Success response: `200 Envelope[data: entity.TripSlot]`.
- Error responses: `400`, `404`, `500`.
- Validation: `id` must be positive.
- DB tables: `trip_slots`, `trips`, `vehicles`.

### PUT /admin/slots/:id
- Description: Update a trip slot.
- Auth: Yes.
- Roles: `admin`.
- Request body: `entity.UpdateTripSlotInput`.
- Query params: None.
- Path params: `id`.
- Success response: `200 Envelope[data: entity.TripSlot]`.
- Error responses: `400`, `404`, `409`, `422`, `500`.
- Validation: Partial updates are allowed; date ranges must remain valid; overlapping vehicle, guide, and driver assignments are rejected.
- DB tables: `trip_slots`, `trips`, `vehicles`.

### DELETE /admin/slots/:id
- Description: Delete a trip slot.
- Auth: Yes.
- Roles: `admin`.
- Request body: None.
- Query params: None.
- Path params: `id`.
- Success response: `200 Envelope[data: null]`.
- Error responses: `400`, `404`, `500`.
- Validation: `id` must be positive.
- DB tables: `trip_slots`.

### GET /trips/:trip_id/slots
- Description: List upcoming slots for a public trip template.
- Auth: No.
- Roles: None.
- Request body: None.
- Query params: None.
- Path params: `trip_id`.
- Success response: `200 Envelope[data: []entity.TripSlot]`.
- Error responses: `400`, `500`.
- Validation: `trip_id` must be positive; cancelled and completed slots are excluded.
- DB tables: `trip_slots`, `trips`, `vehicles`.

## Trip Plans

### POST /admin/plans/trip-plans
- Description: Batch-create trip plans atomically.
- Auth: Yes.
- Roles: `admin`.
- Request body: `[]entity.TripPlan`.
- Query params: None.
- Path params: None.
- Success response: `201 Envelope[data: []entity.TripPlan]`.
- Error responses: `400`, `422`, `500`.
- Validation: Each plan must have `trip_id`, `day_number > 0`, and `title`.
- DB tables: `trip_plans`.

### GET /admin/plans/trip-plans/:trip_id
- Description: Load all plans for a trip.
- Auth: Yes.
- Roles: `admin`.
- Request body: None.
- Query params: None.
- Path params: `trip_id`.
- Success response: `200 Envelope[data: []entity.TripPlan]`.
- Error responses: `400`, `404`, `500`.
- Validation: `trip_id` must be positive.
- DB tables: `trip_plans`.

### DELETE /admin/plans/trip-plans/:id
- Description: Delete a single trip plan.
- Auth: Yes.
- Roles: `admin`.
- Request body: None.
- Query params: None.
- Path params: `id`.
- Success response: `200 Envelope[data: null]`.
- Error responses: `400`, `404`, `500`.
- Validation: `id` must be positive.
- DB tables: `trip_plans`.

## Admin Users

### GET /admin/users
- Description: List and search users.
- Auth: Yes.
- Roles: `admin`.
- Request body: None.
- Query params: `role`, `search`.
- Path params: None.
- Success response: `200 Envelope[data: []entity.User]`.
- Error responses: `500`.
- Validation: `role` is normalized; `search` is a substring filter.
- DB tables: `users`.

### PATCH /admin/users/:id/block
- Description: Toggle block status and revoke refresh tokens when blocking.
- Auth: Yes.
- Roles: `admin`.
- Request body: None.
- Query params: None.
- Path params: `id`.
- Success response: `200 Envelope[data: { name, is_blocked }]`.
- Error responses: `400`, `403`, `404`, `500`.
- Validation: Cannot block admin users.
- DB tables: `users`, `refresh_tokens`.

### PATCH /admin/users/:id/role
- Description: Change the primary role for a user.
- Auth: Yes.
- Roles: `admin`.
- Request body: `{ "role": string }`.
- Query params: None.
- Path params: `id`.
- Success response: `200 Envelope[data: { role }]`.
- Error responses: `400`, `404`, `500`.
- Validation: Role must be one of `admin`, `agency`, `guide`, `driver`, `support`, or `user`.
- DB tables: `users`, `roles`, `user_roles`.

### POST /admin/users
- Description: Create a user directly from the admin panel.
- Auth: Yes.
- Roles: `admin`.
- Request body: `entity.AdminCreateUserRequest`.
- Query params: None.
- Path params: None.
- Success response: `201 Envelope[data: { id, full_name, email, role, is_blocked, is_verified }]`.
- Error responses: `400`, `409`, `500`.
- Validation: `full_name`, `email`, `password`, and `role` are required; password min length is 6.
- DB tables: `users`, `roles`, `user_roles`.

### DELETE /admin/users/:id
- Description: Delete a user and all refresh tokens.
- Auth: Yes.
- Roles: `admin`.
- Request body: None.
- Query params: None.
- Path params: `id`.
- Success response: `200 Envelope[data: null]`.
- Error responses: `400`, `403`, `404`, `500`.
- Validation: Admin cannot delete their own account or another admin.
- DB tables: `users`, `refresh_tokens`, `user_roles`.

## AI

### POST /ai/chat
- Description: Generate a trip plan text with Gemini or the fallback generator.
- Auth: Yes.
- Roles: Any authenticated user.
- Request body: `handler.TripRequest` (`from`, `to`, `days`, `trip_type`, `budget_level`, `members`, `children`, `hotel_type`, `transport`, `created_by`).
- Query params: None.
- Path params: None.
- Success response: `200 Envelope[data: { created_by, prompt, result }]`.
- Error responses: `400`, `429`, `500`.
- Validation: `from` and `to` required; `days` and `members` default to 1; `created_by` defaults to `user`.
- DB tables: None.

### POST /ai/requests
- Description: Store an AI trip request for later review.
- Auth: Yes.
- Roles: Any authenticated user.
- Request body: `dto.AITripRequestInput`.
- Query params: None.
- Path params: None.
- Success response: `201 Envelope[data: dto.AITripRequestResponse]`.
- Error responses: `400`, `401`, `500`.
- Validation: `from`, `to`, and `generated_plan` are required; days/members default to 1.
- DB tables: `ai_trip_requests`, `notifications`.

### GET /ai/requests
- Description: List the current user’s AI trip requests.
- Auth: Yes.
- Roles: Any authenticated user.
- Request body: None.
- Query params: None.
- Path params: None.
- Success response: `200 Envelope[data: []dto.AITripRequestResponse]`.
- Error responses: `500`.
- Validation: Admins see all requests; users must have a valid auth context.
- DB tables: `ai_trip_requests`.

### GET /admin/ai-requests/
- Description: List all AI trip requests for review.
- Auth: Yes.
- Roles: `admin`.
- Request body: None.
- Query params: None.
- Path params: None.
- Success response: `200 Envelope[data: []dto.AITripRequestResponse]`.
- Error responses: `500`.
- Validation: None.
- DB tables: `ai_trip_requests`.

### PATCH /admin/ai-requests/:id/approve
- Description: Approve a request, create a trip, and notify the requester.
- Auth: Yes.
- Roles: `admin`.
- Request body: `dto.AITripReviewRequest` (`admin_note` optional).
- Query params: None.
- Path params: `id`.
- Success response: `200 Envelope[data: dto.AITripRequestResponse]`.
- Error responses: `400`, `404`, `409`, `500`.
- Validation: Request must be pending; approval is transactional.
- DB tables: `ai_trip_requests`, `trips`, `notifications`.

### PATCH /admin/ai-requests/:id/reject
- Description: Reject a request and notify the requester.
- Auth: Yes.
- Roles: `admin`.
- Request body: `dto.AITripReviewRequest` (`admin_note` optional).
- Query params: None.
- Path params: `id`.
- Success response: `200 Envelope[data: dto.AITripRequestResponse]`.
- Error responses: `400`, `404`, `409`, `500`.
- Validation: Request must be pending; rejection is transactional.
- DB tables: `ai_trip_requests`, `notifications`.

## Bookings

### POST /bookings/slot/:slot_id
- Description: Create a booking against a specific trip slot.
- Auth: Yes.
- Roles: Any authenticated user.
- Request body: Inline JSON `{ "seats": int, "coupon_code": string, "booking_type": "shared|private" }`.
- Query params: None.
- Path params: `slot_id`.
- Success response: `201 Envelope[data: entity.Booking]`.
- Error responses: `400`, `401`, `403`, `404`, `409`, `500`.
- Validation: Seats default to 1; slot must exist and be `scheduled` or `active`; duplicate bookings for the same slot are rejected; private bookings reserve the full slot; offer code must be valid; seat inventory must be sufficient.
- DB tables: `bookings`, `booking_plans`, `trip_slots`, `vehicles`, `offers`, `notifications`, `users`, `trips`.

### POST /bookings/trip/:id
- Description: Legacy compatibility route for creating a booking from a trip template.
- Auth: Yes.
- Roles: Any authenticated user.
- Request body: Inline JSON `{ "seats": int, "coupon_code": string }`.
- Query params: None.
- Path params: `id`.
- Success response: `201 Envelope[data: entity.Booking]`.
- Error responses: `400`, `401`, `403`, `404`, `409`, `500`.
- Validation: Seats default to 1; trip must be active; offer code must be valid; seat inventory must be sufficient.
- DB tables: `bookings`, `booking_plans`, `vehicles`, `offers`, `notifications`, `users`, `trips`.

### GET /bookings/my-orders
- Description: List bookings for the current user.
- Auth: Yes.
- Roles: Any authenticated user.
- Request body: None.
- Query params: None.
- Path params: None.
- Success response: `200 Envelope[data: []entity.Booking]`.
- Error responses: `401`, `500`.
- Validation: Auth context must include a user ID.
- DB tables: `bookings`, `booking_plans`, `trip_slots`, `trips`.

### PATCH /bookings/custom-plan/:id
- Description: Replace the custom plans attached to a booking owned by the caller.
- Auth: Yes.
- Roles: Any authenticated user.
- Request body: `entity.UpdateBookingPlanInput` (`plans` array).
- Query params: None.
- Path params: `id`.
- Success response: `200 Envelope[data: null]`.
- Error responses: `400`, `401`, `403`, `404`, `500`.
- Validation: Booking must belong to the caller.
- DB tables: `bookings`, `booking_plans`.

### GET /admin/orders/all
- Description: List all bookings for the admin dashboard.
- Auth: Yes.
- Roles: `admin`.
- Request body: None.
- Query params: None.
- Path params: None.
- Success response: `200 Envelope[data: []entity.Booking]`.
- Error responses: `500`.
- Validation: None.
- DB tables: `bookings`, `booking_plans`, `trip_slots`, `trips`.

## Notifications

### GET /notifications/
- Description: List notifications for the current user.
- Auth: Yes.
- Roles: Any authenticated user.
- Request body: None.
- Query params: None.
- Path params: None.
- Success response: `200 Envelope[data: []dto.NotificationResponse]`.
- Error responses: `500`.
- Validation: Auth context must include a user ID.
- DB tables: `notifications`.

### PATCH /notifications/:id/read
- Description: Mark a user notification as read.
- Auth: Yes.
- Roles: Any authenticated user.
- Request body: None.
- Query params: None.
- Path params: `id`.
- Success response: `200 Envelope[data: null]`.
- Error responses: `400`, `401`, `403`, `404`, `500`.
- Validation: Non-admin users can only mark their own notifications.
- DB tables: `notifications`.

### GET /admin/notifications/
- Description: List admin-facing booking notifications.
- Auth: Yes.
- Roles: `admin`.
- Request body: None.
- Query params: None.
- Path params: None.
- Success response: `200 Envelope[data: []dto.NotificationResponse]`.
- Error responses: `500`.
- Validation: None.
- DB tables: `notifications`.

### PATCH /admin/notifications/:id/read
- Description: Mark an admin notification as read.
- Auth: Yes.
- Roles: `admin`.
- Request body: None.
- Query params: None.
- Path params: `id`.
- Success response: `200 Envelope[data: null]`.
- Error responses: `400`, `403`, `404`, `500`.
- Validation: Admin role required.
- DB tables: `notifications`.

## RBAC

### GET /rbac/me
- Description: Return the caller’s roles and permissions.
- Auth: Yes.
- Roles: Any authenticated user.
- Request body: None.
- Query params: None.
- Path params: None.
- Success response: `200 Envelope[data: dto.UserAccessResponse]`.
- Error responses: `401`, `500`.
- Validation: Auth context must include a user ID.
- DB tables: `users`, `roles`, `user_roles`, `permissions`, `role_permissions`.

### GET /admin/rbac/roles
- Description: List all roles.
- Auth: Yes.
- Roles: `admin` plus `manage_users` permission.
- Request body: None.
- Query params: None.
- Path params: None.
- Success response: `200 Envelope[data: []dto.RoleResponse]`.
- Error responses: `403`, `500`.
- Validation: Permission `manage_users` is required.
- DB tables: `roles`.

### POST /admin/rbac/roles
- Description: Create a role.
- Auth: Yes.
- Roles: `admin` plus `manage_users` permission.
- Request body: `dto.RoleRequest`.
- Query params: None.
- Path params: None.
- Success response: `201 Envelope[data: dto.RoleResponse]`.
- Error responses: `400`, `403`, `409`, `500`.
- Validation: Role name is required and must be one of the supported canonical roles.
- DB tables: `roles`.

### PATCH /admin/rbac/roles/:id
- Description: Update a role description.
- Auth: Yes.
- Roles: `admin` plus `manage_users` permission.
- Request body: `dto.RoleRequest`.
- Query params: None.
- Path params: `id`.
- Success response: `200 Envelope[data: null]`.
- Error responses: `400`, `403`, `404`, `500`.
- Validation: Role name cannot be changed to a different canonical role.
- DB tables: `roles`.

### DELETE /admin/rbac/roles/:id
- Description: Delete a custom role.
- Auth: Yes.
- Roles: `admin` plus `manage_users` permission.
- Request body: None.
- Query params: None.
- Path params: `id`.
- Success response: `200 Envelope[data: null]`.
- Error responses: `400`, `403`, `404`, `409`, `500`.
- Validation: Default roles and assigned roles cannot be deleted.
- DB tables: `roles`, `user_roles`.

### PATCH /admin/rbac/users/:id/role
- Description: Assign one primary role to a user.
- Auth: Yes.
- Roles: `admin` plus `manage_users` permission.
- Request body: `dto.AssignRoleRequest` (`role_id`).
- Query params: None.
- Path params: `id`.
- Success response: `200 Envelope[data: null]`.
- Error responses: `400`, `403`, `404`, `500`.
- Validation: User and role must exist; write is transactional.
- DB tables: `users`, `roles`, `user_roles`.

### GET /admin/rbac/permissions
- Description: List permissions.
- Auth: Yes.
- Roles: `admin` plus `manage_users` permission.
- Request body: None.
- Query params: None.
- Path params: None.
- Success response: `200 Envelope[data: []dto.PermissionResponse]`.
- Error responses: `403`, `500`.
- Validation: Permission `manage_users` is required.
- DB tables: `permissions`.

### POST /admin/rbac/permissions
- Description: Create a permission.
- Auth: Yes.
- Roles: `admin` plus `manage_users` permission.
- Request body: `dto.PermissionRequest`.
- Query params: None.
- Path params: None.
- Success response: `201 Envelope[data: dto.PermissionResponse]`.
- Error responses: `400`, `403`, `409`, `500`.
- Validation: `key` and `name` are required; keys are normalized to lowercase.
- DB tables: `permissions`.

### PATCH /admin/rbac/permissions/:id
- Description: Update a permission.
- Auth: Yes.
- Roles: `admin` plus `manage_users` permission.
- Request body: `dto.PermissionRequest`.
- Query params: None.
- Path params: `id`.
- Success response: `200 Envelope[data: null]`.
- Error responses: `400`, `403`, `404`, `500`.
- Validation: Partial updates allowed.
- DB tables: `permissions`.

### DELETE /admin/rbac/permissions/:id
- Description: Delete a permission.
- Auth: Yes.
- Roles: `admin` plus `manage_users` permission.
- Request body: None.
- Query params: None.
- Path params: `id`.
- Success response: `200 Envelope[data: null]`.
- Error responses: `400`, `403`, `404`, `500`.
- Validation: Permission must exist.
- DB tables: `permissions`, `role_permissions`.

### GET /admin/rbac/roles/:id/permissions
- Description: List permissions assigned to a role.
- Auth: Yes.
- Roles: `admin` plus `manage_users` permission.
- Request body: None.
- Query params: None.
- Path params: `id`.
- Success response: `200 Envelope[data: []dto.PermissionResponse]`.
- Error responses: `403`, `500`.
- Validation: Role must exist.
- DB tables: `permissions`, `role_permissions`.

### POST /admin/rbac/roles/:id/permissions
- Description: Assign a permission to a role.
- Auth: Yes.
- Roles: `admin` plus `manage_users` permission.
- Request body: `dto.AssignPermissionRequest` (`permission_id`).
- Query params: None.
- Path params: `id`.
- Success response: `200 Envelope[data: null]`.
- Error responses: `400`, `403`, `404`, `500`.
- Validation: Role and permission must exist.
- DB tables: `permissions`, `role_permissions`.

### DELETE /admin/rbac/roles/:id/permissions/:permission_id
- Description: Remove a permission from a role.
- Auth: Yes.
- Roles: `admin` plus `manage_users` permission.
- Request body: None.
- Query params: None.
- Path params: `id`, `permission_id`.
- Success response: `200 Envelope[data: null]`.
- Error responses: `400`, `403`, `404`, `500`.
- Validation: Role and permission must exist.
- DB tables: `permissions`, `role_permissions`.

## Vehicles

### GET /vehicles/
- Description: List vehicles visible to the caller.
- Auth: Yes.
- Roles: Any authenticated user, agency, or admin.
- Request body: None.
- Query params: None.
- Path params: None.
- Success response: `200 Envelope[data: []dto.VehicleResponse]`.
- Error responses: `500`.
- Validation: Non-admin users only see active or assigned vehicles.
- DB tables: `vehicles`.

### GET /vehicles/trip/:trip_id
- Description: Load the vehicle assigned to a trip, if any.
- Auth: Yes.
- Roles: Any authenticated user.
- Request body: None.
- Query params: None.
- Path params: `trip_id`.
- Success response: `200 Envelope[data: dto.VehicleResponse | null]`.
- Error responses: `400`, `404`, `500`.
- Validation: `trip_id` must be positive.
- DB tables: `vehicles`, `trips`.

### GET /vehicles/:id
- Description: Load a vehicle by ID.
- Auth: Yes.
- Roles: Any authenticated user.
- Request body: None.
- Query params: None.
- Path params: `id`.
- Success response: `200 Envelope[data: dto.VehicleResponse]`.
- Error responses: `400`, `404`, `500`.
- Validation: `id` must be positive.
- DB tables: `vehicles`.

### POST /admin/vehicles/
- Description: Create a vehicle.
- Auth: Yes.
- Roles: `admin` plus `manage_fleet` permission.
- Request body: `dto.VehicleRequest`.
- Query params: None.
- Path params: None.
- Success response: `201 Envelope[data: dto.VehicleResponse]`.
- Error responses: `400`, `403`, `500`.
- Validation: `name`, `type`, and `total_seats` are required; type must be `Bus`, `Car`, or `Traveler`.
- DB tables: `vehicles`, `users`, `trips`.

### PUT /admin/vehicles/:id
- Description: Update a vehicle.
- Auth: Yes.
- Roles: `admin` plus `manage_fleet` permission.
- Request body: `dto.VehicleRequest`.
- Query params: None.
- Path params: `id`.
- Success response: `200 Envelope[data: null]`.
- Error responses: `400`, `403`, `404`, `500`.
- Validation: Partial updates allowed; agency ownership is enforced for non-admin callers.
- DB tables: `vehicles`, `users`, `trips`.

### DELETE /admin/vehicles/:id
- Description: Delete a vehicle.
- Auth: Yes.
- Roles: `admin` plus `manage_fleet` permission.
- Request body: None.
- Query params: None.
- Path params: `id`.
- Success response: `200 Envelope[data: null]`.
- Error responses: `400`, `403`, `404`, `500`.
- Validation: Agency ownership is enforced for non-admin callers.
- DB tables: `vehicles`.

### PATCH /admin/vehicles/:id/assign-trip
- Description: Assign a vehicle to a trip.
- Auth: Yes.
- Roles: `admin` plus `manage_fleet` permission.
- Request body: `dto.AssignVehicleRequest` (`trip_id`).
- Query params: None.
- Path params: `id`.
- Success response: `200 Envelope[data: null]`.
- Error responses: `400`, `403`, `404`, `409`, `500`.
- Validation: Vehicle and trip must exist; a trip can have only one vehicle.
- DB tables: `vehicles`, `trips`.

## Offers

### GET /offers/active
- Description: List active offers.
- Auth: Yes.
- Roles: Any authenticated user.
- Request body: None.
- Query params: None.
- Path params: None.
- Success response: `200 Envelope[data: []dto.OfferResponse]`.
- Error responses: `500`.
- Validation: None.
- DB tables: `offers`.

### POST /offers/validate
- Description: Validate a coupon code.
- Auth: Yes.
- Roles: Any authenticated user.
- Request body: `dto.ApplyCouponRequest` (`code`).
- Query params: None.
- Path params: None.
- Success response: `200 Envelope[data: dto.OfferResponse]`.
- Error responses: `400`, `404`, `500`.
- Validation: Code required.
- DB tables: `offers`.

### GET /admin/offers/
- Description: List all offers.
- Auth: Yes.
- Roles: `admin` plus `manage_offers` permission.
- Request body: None.
- Query params: None.
- Path params: None.
- Success response: `200 Envelope[data: []dto.OfferResponse]`.
- Error responses: `403`, `500`.
- Validation: None.
- DB tables: `offers`.

### POST /admin/offers/
- Description: Create an offer.
- Auth: Yes.
- Roles: `admin` plus `manage_offers` permission.
- Request body: `dto.OfferRequest`.
- Query params: None.
- Path params: None.
- Success response: `201 Envelope[data: dto.OfferResponse]`.
- Error responses: `400`, `403`, `409`, `500`.
- Validation: `code`, `title`, and `expiry_date` are required.
- DB tables: `offers`.

### PUT /admin/offers/:id
- Description: Update an offer.
- Auth: Yes.
- Roles: `admin` plus `manage_offers` permission.
- Request body: `dto.OfferRequest`.
- Query params: None.
- Path params: `id`.
- Success response: `200 Envelope[data: null]`.
- Error responses: `400`, `403`, `404`, `500`.
- Validation: Partial updates allowed.
- DB tables: `offers`.

### DELETE /admin/offers/:id
- Description: Delete an offer.
- Auth: Yes.
- Roles: `admin` plus `manage_offers` permission.
- Request body: None.
- Query params: None.
- Path params: `id`.
- Success response: `200 Envelope[data: null]`.
- Error responses: `400`, `403`, `404`, `500`.
- Validation: Offer must exist.
- DB tables: `offers`.

## Reviews

### POST /reviews/
- Description: Create a trip review.
- Auth: Yes.
- Roles: Any authenticated user.
- Request body: `dto.ReviewRequest` (`trip_id`, `rating`, `comment`).
- Query params: None.
- Path params: None.
- Success response: `201 Envelope[data: dto.ReviewResponse]`.
- Error responses: `400`, `401`, `403`, `404`, `409`, `500`.
- Validation: `trip_id` required; rating must be 1-5; user must have a completed booking.
- DB tables: `reviews`, `bookings`, `trips`.

### GET /reviews/me
- Description: List the caller’s reviews.
- Auth: Yes.
- Roles: Any authenticated user.
- Request body: None.
- Query params: None.
- Path params: None.
- Success response: `200 Envelope[data: []dto.ReviewResponse]`.
- Error responses: `400`, `500`.
- Validation: Auth context must include a user ID.
- DB tables: `reviews`.

### GET /reviews/trip/:trip_id
- Description: List reviews for a trip.
- Auth: Yes.
- Roles: Any authenticated user.
- Request body: None.
- Query params: None.
- Path params: `trip_id`.
- Success response: `200 Envelope[data: []dto.ReviewResponse]`.
- Error responses: `400`, `404`, `500`.
- Validation: `trip_id` must be positive.
- DB tables: `reviews`.

### GET /reviews/trip/:trip_id/summary
- Description: Return average rating and review count.
- Auth: Yes.
- Roles: Any authenticated user.
- Request body: None.
- Query params: None.
- Path params: `trip_id`.
- Success response: `200 Envelope[data: dto.ReviewSummaryResponse]`.
- Error responses: `400`, `500`.
- Validation: `trip_id` must be positive.
- DB tables: `reviews`.

### GET /admin/reviews/
- Description: List all reviews.
- Auth: Yes.
- Roles: `admin` plus `manage_reviews` permission.
- Request body: None.
- Query params: None.
- Path params: None.
- Success response: `200 Envelope[data: []dto.ReviewResponse]`.
- Error responses: `403`, `500`.
- Validation: None.
- DB tables: `reviews`.

### DELETE /admin/reviews/:id
- Description: Delete a review.
- Auth: Yes.
- Roles: `admin` plus `manage_reviews` permission.
- Request body: None.
- Query params: None.
- Path params: `id`.
- Success response: `200 Envelope[data: null]`.
- Error responses: `400`, `403`, `404`, `500`.
- Validation: Review must exist.
- DB tables: `reviews`.

## Complaints

### POST /complaints/
- Description: Create a complaint against a booking.
- Auth: Yes.
- Roles: Any authenticated user.
- Request body: `dto.ComplaintRequest` (`booking_id`, `title`, `description`).
- Query params: None.
- Path params: None.
- Success response: `201 Envelope[data: dto.ComplaintResponse]`.
- Error responses: `400`, `401`, `403`, `404`, `500`.
- Validation: `booking_id`, `title`, and `description` are required; booking must belong to the caller.
- DB tables: `complaints`, `bookings`.

### GET /complaints/me
- Description: List the caller’s complaints.
- Auth: Yes.
- Roles: Any authenticated user.
- Request body: None.
- Query params: None.
- Path params: None.
- Success response: `200 Envelope[data: []dto.ComplaintResponse]`.
- Error responses: `400`, `500`.
- Validation: Auth context must include a user ID.
- DB tables: `complaints`.

### GET /complaints/:id
- Description: Load one complaint if the caller owns it or is admin.
- Auth: Yes.
- Roles: Any authenticated user.
- Request body: None.
- Query params: None.
- Path params: `id`.
- Success response: `200 Envelope[data: dto.ComplaintResponse]`.
- Error responses: `400`, `401`, `403`, `404`, `500`.
- Validation: Complaint must exist.
- DB tables: `complaints`.

### GET /admin/complaints/
- Description: List all complaints.
- Auth: Yes.
- Roles: `admin` plus `manage_complaints` permission.
- Request body: None.
- Query params: None.
- Path params: None.
- Success response: `200 Envelope[data: []dto.ComplaintResponse]`.
- Error responses: `403`, `500`.
- Validation: None.
- DB tables: `complaints`.

### PATCH /admin/complaints/:id/status
- Description: Update complaint status.
- Auth: Yes.
- Roles: `admin` plus `manage_complaints` permission.
- Request body: `dto.ComplaintStatusRequest` (`status`).
- Query params: None.
- Path params: `id`.
- Success response: `200 Envelope[data: null]`.
- Error responses: `400`, `403`, `404`, `500`.
- Validation: Status is normalized to `pending`, `in_progress`, or `resolved`.
- DB tables: `complaints`.

### DELETE /admin/complaints/:id
- Description: Delete a complaint.
- Auth: Yes.
- Roles: `admin` plus `manage_complaints` permission.
- Request body: None.
- Query params: None.
- Path params: `id`.
- Success response: `200 Envelope[data: null]`.
- Error responses: `400`, `403`, `404`, `500`.
- Validation: Complaint must exist.
- DB tables: `complaints`.

## Tracking

### POST /tracking/
- Description: Record a live tracking point.
- Auth: Yes.
- Roles: Users with `manage_tracking` permission or admin.
- Request body: `dto.TrackingUpdateRequest` (`booking_id`, `vehicle_id`, `latitude`, `longitude`).
- Query params: None.
- Path params: None.
- Success response: `201 Envelope[data: dto.TrackingResponse]`.
- Error responses: `400`, `403`, `404`, `500`.
- Validation: Booking and vehicle IDs are required; latitude must be -90 to 90; longitude must be -180 to 180; vehicle must match the booking’s trip.
- DB tables: `trackings`, `bookings`, `vehicles`.

### GET /tracking/booking/:booking_id
- Description: Get the latest tracking point for a booking.
- Auth: Yes.
- Roles: Admin, driver, or booking owner.
- Request body: None.
- Query params: None.
- Path params: `booking_id`.
- Success response: `200 Envelope[data: dto.TrackingResponse]`.
- Error responses: `400`, `401`, `403`, `404`, `500`.
- Validation: Booking must exist and belong to the caller unless the caller is admin/driver.
- DB tables: `trackings`, `bookings`.

### GET /tracking/booking/:booking_id/history
- Description: Get tracking history for a booking.
- Auth: Yes.
- Roles: Admin, driver, or booking owner.
- Request body: None.
- Query params: None.
- Path params: `booking_id`.
- Success response: `200 Envelope[data: []dto.TrackingResponse]`.
- Error responses: `400`, `401`, `403`, `404`, `500`.
- Validation: Booking must exist and belong to the caller unless the caller is admin/driver.
- DB tables: `trackings`, `bookings`.

### GET /admin/tracking/
- Description: List all tracking rows.
- Auth: Yes.
- Roles: `admin` plus `manage_tracking` permission.
- Request body: None.
- Query params: None.
- Path params: None.
- Success response: `200 Envelope[data: []dto.TrackingResponse]`.
- Error responses: `403`, `500`.
- Validation: None.
- DB tables: `trackings`.
