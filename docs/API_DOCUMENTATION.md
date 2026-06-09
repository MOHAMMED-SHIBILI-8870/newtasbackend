# API Documentation

This reference is derived from the actual route registration, handlers, use cases, and entities in the repository.

## Common Conventions

- Response envelope: `internal/response/response.go`
- Standard shape: `{"success": bool, "message": string, "data": any, "error": string}`
- Most endpoints return their payload in `data`
- Public auth endpoints do not require a token
- Protected routes use either `internal/middleware/auth_middleware.go` or `internal/middleware/permission_middleware.go`

## Authentication APIs

| Method | Route | Description | Auth | Roles / Permissions | Request / Params / Validation | Response | Errors | DB Tables | Source |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| POST | `/auth/register` | Register a new customer account and send a signup OTP | Public | None | Body: `full_name`, `email`, `password`; all required, password >= 6 chars | `data.user` (`AuthUserResponse`) | 400 invalid/missing fields, 409 duplicate email, 500 hash/create/OTP/email failure | `users`, `otps` | `internal/handler/auth_handler.go`, `internal/usecase/otp_verification.go`, `internal/usecase/send_opt.go` |
| POST | `/auth/login` | Authenticate user and issue access/refresh tokens | Public | None | Body: `email`, `password`; account must be verified and not blocked | `data` (`AuthResponse`) + `access_token`/`refresh_token` cookies | 400 invalid body, 404 email not found, 401 wrong password/unverified, 403 blocked, 500 token/save failure | `users`, `refresh_tokens` | `internal/handler/auth_handler.go`, `internal/usecase/jwt.go` |
| POST | `/auth/verify-otp` | Verify signup or reset OTP | Public | None | Body: `email`, `otp`, `purpose`; all required | Success envelope, `data = null` | 400 invalid body or OTP, 404 user not found | `users`, `otps` | `internal/handler/auth_handler.go`, `internal/usecase/otp_verification.go` |
| POST | `/auth/forgot-password` | Send reset-password OTP | Public | None | Body: `email`; required | Success envelope, `data = null` | 400 missing email, 404 user not found, 500 OTP/email failure | `users`, `otps` | `internal/handler/auth_handler.go`, `internal/usecase/otp_verification.go`, `internal/usecase/send_opt.go` |
| POST | `/auth/reset-password` | Reset the password after OTP verification | Public | None | Body: `email`, `new_password`, `otp`; password >= 6 chars | Success envelope, `data = null` | 400 invalid body/OTP, 404 user not found, 500 hash/update/revoke failure | `users`, `otps`, `refresh_tokens` | `internal/handler/auth_handler.go`, `internal/usecase/otp_verification.go`, `internal/usecase/jwt.go` |
| POST | `/auth/resend-otp` | Send a new OTP for signup or reset flows | Public | None | Body: `email`, `purpose`; signup resend blocked if user already verified | Success envelope, `data = null` | 400 missing fields, 404 user not found, 500 OTP/email failure | `users`, `otps` | `internal/handler/auth_handler.go`, `internal/usecase/otp_verification.go`, `internal/usecase/send_opt.go` |
| POST | `/auth/logout` | Revoke refresh token and clear cookies | Public | None | No body required; refresh cookie is optional | Success envelope, `data = null` | 500 only if token deletion fails silently in the background | `refresh_tokens` | `internal/handler/auth_handler.go`, `internal/usecase/jwt.go` |
| POST | `/auth/refresh` | Rotate refresh token and mint a new access token | Public | None | Cookie: `refresh_token` required | `data` (`AuthResponse`) + refreshed cookies | 401 missing/invalid token, 403 blocked/unverified, 500 token generation failure | `refresh_tokens`, `users` | `internal/handler/auth_handler.go`, `internal/usecase/jwt.go` |

## Trip, Plan, and Slot APIs

| Method | Route | Description | Auth | Roles / Permissions | Request / Params / Validation | Response | Errors | DB Tables | Source |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| GET | `/trips/` | List public trips | Public | None | No body; no query params | `data` array of `Trip` | 500 load failure | `trips`, `trip_plans` | `internal/handler/trip_handler.go`, `internal/repository/trip_repository.go` |
| GET | `/trips/:name` | Search trips by decoded name fragment | Public | None | Path param `name`; URL-decoded and required | `data` single trip | 400 invalid name, 404 not found | `trips`, `trip_plans` | `internal/handler/trip_handler.go`, `internal/usecase/trip_usecase.go` |
| POST | `/admin/trips/` | Create a trip template | Authenticated admin | `admin` | Body: `entity.Trip`; `from`, `to`, `duration`, `price` required by usecase | `data` trip | 400 invalid body/required fields, 422 validation failures | `trips` | `internal/handler/trip_handler.go`, `internal/usecase/trip_usecase.go` |
| GET | `/admin/trips/` | List all trips for admin | Authenticated admin | `admin` | No body | `data` array of trips | 500 load failure | `trips`, `trip_plans` | `internal/handler/trip_handler.go`, `internal/usecase/trip_usecase.go` |
| PATCH | `/admin/trips/:id` | Partially update a trip template | Authenticated admin | `admin` | Path `id`; body `entity.UpdateTripInput` with pointer fields | Success envelope, `data = null` | 400 invalid body/id, 404 not found, 422 validation failures | `trips`, `trip_plans` | `internal/handler/trip_handler.go`, `internal/entity/updateTrip_Input.go`, `internal/usecase/trip_usecase.go` |
| DELETE | `/admin/trips/:id` | Delete a trip template | Authenticated admin | `admin` | Path `id` required | Success envelope, `data = null` | 400 invalid id, 404 not found | `trips` | `internal/handler/trip_handler.go`, `internal/usecase/trip_usecase.go` |
| GET | `/trips/:trip_id/slots` | List upcoming slots for a trip | Public | None | Path `trip_id` required | `data` array of `TripSlot` | 400 invalid trip id, 404 not found | `trip_slots`, `trips`, `vehicles` | `internal/handler/trip_slot_handler.go`, `internal/usecase/trip_slot_usecase.go` |
| POST | `/admin/slots/` | Create a trip slot | Authenticated admin | `admin` | Body `entity.TripSlot`; trip, start/end required; end >= start; total seats > 0; price override >= 0 | `data` created slot | 400 invalid body/validation, 404 trip/vehicle not found, 409 overlap/seat conflicts | `trip_slots`, `trips`, `vehicles` | `internal/handler/trip_slot_handler.go`, `internal/usecase/trip_slot_usecase.go` |
| GET | `/admin/slots/` | List all slots | Authenticated admin | `admin` | No body | `data` array of slots | 500 load failure | `trip_slots`, `trips`, `vehicles` | `internal/handler/trip_slot_handler.go`, `internal/repository/trip_slot_repository.go` |
| GET | `/admin/slots/:id` | Load a single slot | Authenticated admin | `admin` | Path `id` required | `data` slot | 400 invalid id, 404 not found | `trip_slots`, `trips`, `vehicles` | `internal/handler/trip_slot_handler.go`, `internal/usecase/trip_slot_usecase.go` |
| PUT | `/admin/slots/:id` | Update a slot | Authenticated admin | `admin` | Path `id`; body `entity.UpdateTripSlotInput`; same seat/date/overlap rules as create | `data` updated slot | 400 invalid body/id, 404 not found, 409 overlap/seat conflicts | `trip_slots`, `trips`, `vehicles` | `internal/handler/trip_slot_handler.go`, `internal/usecase/trip_slot_usecase.go` |
| DELETE | `/admin/slots/:id` | Delete a slot | Authenticated admin | `admin` | Path `id` required | Success envelope, `data = null` | 400 invalid id, 404 not found | `trip_slots` | `internal/handler/trip_slot_handler.go`, `internal/usecase/trip_slot_usecase.go` |
| POST | `/admin/plans/trip-plans` | Batch create trip plans | Authenticated admin | `admin` | Body: array of `TripPlan`; each item needs `trip_id`, `day_number`, `title` | `data` array of created plans | 400 invalid body/validation, 422 persistence failure | `trip_plans`, `trips` | `internal/handler/trip_plan_handler.go`, `internal/usecase/trip_plan_usecase.go` |
| GET | `/admin/plans/trip-plans/:trip_id` | Load trip plans by trip | Authenticated admin | `admin` | Path `trip_id` required | `data` array of plans | 400 invalid trip id, 404 not found | `trip_plans`, `trips` | `internal/handler/trip_plan_handler.go`, `internal/usecase/trip_plan_usecase.go` |
| DELETE | `/admin/plans/trip-plans/:id` | Delete one trip plan | Authenticated admin | `admin` | Path `id` required | Success envelope, `data = null` | 400 invalid id, 404 not found | `trip_plans` | `internal/handler/trip_plan_handler.go`, `internal/usecase/trip_plan_usecase.go` |

## Admin User APIs

| Method | Route | Description | Auth | Roles / Permissions | Request / Params / Validation | Response | Errors | DB Tables | Source |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| GET | `/admin/users` | List/search users | Authenticated admin | `admin` | Query: `role`, `search` | `data` array of users | 500 load failure | `users`, `user_roles` | `internal/handler/admin_handler.go`, `internal/usecase/admin_usecase.go` |
| PATCH | `/admin/users/:id/block` | Toggle block/unblock | Authenticated admin | `admin` | Path `id` required | `data.name`, `data.is_blocked` | 400 invalid id, 403 cannot block admin, 404 not found | `users`, `refresh_tokens` | `internal/handler/admin_handler.go`, `internal/usecase/admin_usecase.go` |
| PATCH | `/admin/users/:id/role` | Change a user role | Authenticated admin | `admin` | Path `id`; body `{ "role": "..." }`; role required | `data.role` | 400 invalid body/role, 403/404 for denied or missing records | `users`, `roles`, `user_roles` | `internal/handler/admin_handler.go`, `internal/usecase/admin_usecase.go` |
| POST | `/admin/users` | Create a verified user as admin | Authenticated admin | `admin` | Body `AdminCreateUserRequest`; `full_name`, `email`, `password`, `role` required | `data` created user summary | 400 invalid body, 409 email exists, 400 invalid role, 500 hash/create failure | `users`, `roles`, `user_roles` | `internal/handler/admin_handler.go`, `internal/entity/AdminCreateUserRequest.go`, `internal/usecase/admin_usecase.go` |
| DELETE | `/admin/users/:id` | Delete a user account | Authenticated admin | `admin` | Path `id` required; cannot delete self or admin user | Success envelope, `data = null` | 400 invalid id, 403 security risk, 404 not found | `users`, `refresh_tokens`, `user_roles` | `internal/handler/admin_handler.go`, `internal/usecase/admin_usecase.go` |

## RBAC APIs

| Method | Route | Description | Auth | Roles / Permissions | Request / Params / Validation | Response | Errors | DB Tables | Source |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| GET | `/rbac/me` | Return current user roles and permissions | Authenticated user | Auth only | No body | `data` `UserAccessResponse` | 401 unauthorized, 500 load failure | `users`, `user_roles`, `roles`, `role_permissions`, `permissions` | `internal/handler/role_handler.go`, `internal/usecase/role_usecase.go`, `internal/usecase/permission_usecase.go` |
| GET | `/admin/rbac/roles` | List roles | Authenticated admin | `manage_users` | No body | `data` array of roles | 500 load failure | `roles` | `internal/handler/role_handler.go`, `internal/usecase/role_usecase.go` |
| POST | `/admin/rbac/roles` | Create role | Authenticated admin | `manage_users` | Body `RoleRequest`; role must be canonical | `data` role | 400 invalid body, 409 already exists, 400 invalid role | `roles` | `internal/handler/role_handler.go`, `internal/usecase/role_usecase.go` |
| PATCH | `/admin/rbac/roles/:id` | Update role | Authenticated admin | `manage_users` | Path `id`; body `RoleRequest`; canonical name cannot change to another canonical role | `data = null` | 400 invalid body/id, 404 not found, 400 invalid role | `roles` | `internal/handler/role_handler.go`, `internal/usecase/role_usecase.go` |
| DELETE | `/admin/rbac/roles/:id` | Delete role | Authenticated admin | `manage_users` | Path `id` required; default roles and assigned roles cannot be deleted | `data = null` | 400 invalid id, 404 not found, 403 default/assigned role | `roles`, `user_roles` | `internal/handler/role_handler.go`, `internal/usecase/role_usecase.go` |
| PATCH | `/admin/rbac/users/:id/role` | Assign a role to a user | Authenticated admin | `manage_users` | Path `id`; body `AssignRoleRequest` with `role_id` | `data = null` | 400 invalid body/id, 404 not found | `users`, `roles`, `user_roles` | `internal/handler/role_handler.go`, `internal/usecase/role_usecase.go` |
| GET | `/admin/rbac/permissions` | List permissions | Authenticated admin | `manage_users` | No body | `data` array of permissions | 500 load failure | `permissions` | `internal/handler/permission_handler.go`, `internal/usecase/permission_usecase.go` |
| POST | `/admin/rbac/permissions` | Create permission | Authenticated admin | `manage_users` | Body `PermissionRequest`; `key` and `name` required; key normalized lowercase | `data` permission | 400 invalid body, 409 already exists | `permissions` | `internal/handler/permission_handler.go`, `internal/usecase/permission_usecase.go` |
| PATCH | `/admin/rbac/permissions/:id` | Update permission | Authenticated admin | `manage_users` | Path `id`; body `PermissionRequest` | `data = null` | 400 invalid body/id, 404 not found | `permissions` | `internal/handler/permission_handler.go`, `internal/usecase/permission_usecase.go` |
| DELETE | `/admin/rbac/permissions/:id` | Delete permission | Authenticated admin | `manage_users` | Path `id` required | `data = null` | 400 invalid id, 404 not found | `permissions`, `role_permissions` | `internal/handler/permission_handler.go`, `internal/usecase/permission_usecase.go` |
| GET | `/admin/rbac/roles/:id/permissions` | List permissions for a role | Authenticated admin | `manage_users` | Path `id` required | `data` array of permissions | 400 invalid id, 500 load failure | `role_permissions`, `permissions`, `roles` | `internal/handler/permission_handler.go`, `internal/usecase/permission_usecase.go` |
| POST | `/admin/rbac/roles/:id/permissions` | Assign permission to role | Authenticated admin | `manage_users` | Path `id`; body `AssignPermissionRequest` with `permission_id` | `data = null` | 400 invalid body/id, 404 role or permission not found | `role_permissions`, `roles`, `permissions` | `internal/handler/permission_handler.go`, `internal/usecase/permission_usecase.go` |
| DELETE | `/admin/rbac/roles/:id/permissions/:permission_id` | Remove permission from role | Authenticated admin | `manage_users` | Path `id` and `permission_id` required | `data = null` | 400 invalid ids, 404 not found | `role_permissions` | `internal/handler/permission_handler.go`, `internal/usecase/permission_usecase.go` |

## AI APIs

| Method | Route | Description | Auth | Roles / Permissions | Request / Params / Validation | Response | Errors | DB Tables | Source |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| POST | `/ai/chat` | Generate a trip plan from a prompt | Authenticated user | Auth only | Body: `from`, `to`, `days`, `trip_type`, `budget_level`, `members`, `children`, `hotel_type`, `transport`, `created_by`; `from` and `to` required | `data.created_by`, `data.prompt`, `data.result` | 400 invalid body/missing locations, 429 rate limit from Gemini, 500 AI failure | None persisted | `internal/handler/ai_handler.go` |
| POST | `/ai/requests` | Submit an AI trip request for admin review | Authenticated user | Auth only | Body `AITripRequestInput`; `from`, `to`, `generated_plan` required; days/members default to 1 | `data` `AITripRequestResponse` | 400 invalid body/validation, 401 unauthorized, 500 persistence/notification failure | `ai_trip_requests`, `users`, `notifications` | `internal/handler/ai_handler.go`, `internal/usecase/ai_trip_request_usecase.go` |
| GET | `/ai/requests` | List current user's AI requests, or all for admin | Authenticated user | Auth only | No body | `data` array of AI requests | 500 load failure | `ai_trip_requests`, `users`, `trips` | `internal/handler/ai_handler.go`, `internal/usecase/ai_trip_request_usecase.go` |
| GET | `/admin/ai-requests/` | List all AI requests | Authenticated admin | `admin` | No body | `data` array of AI requests | 500 load failure | `ai_trip_requests`, `users`, `trips` | `internal/handler/ai_handler.go`, `internal/usecase/ai_trip_request_usecase.go` |
| PATCH | `/admin/ai-requests/:id/approve` | Approve an AI request and create a trip | Authenticated admin | `admin` | Path `id`; body may include optional `admin_note` | `data` updated AI request with `trip_id` | 400 invalid body/id, 409 already reviewed, 404 not found | `ai_trip_requests`, `trips`, `notifications` | `internal/handler/ai_handler.go`, `internal/usecase/ai_trip_request_usecase.go` |
| PATCH | `/admin/ai-requests/:id/reject` | Reject an AI request | Authenticated admin | `admin` | Path `id`; body may include optional `admin_note` | `data` updated AI request | 400 invalid body/id, 409 already reviewed, 404 not found | `ai_trip_requests`, `notifications` | `internal/handler/ai_handler.go`, `internal/usecase/ai_trip_request_usecase.go` |

## Booking and Order APIs

| Method | Route | Description | Auth | Roles / Permissions | Request / Params / Validation | Response | Errors | DB Tables | Source |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| POST | `/bookings/slot/:slot_id` | Create a booking against a specific slot | Authenticated user | Auth only | Path `slot_id`; body `seats`, `coupon_code`, `booking_type`; `booking_type` is `shared` or `private`; private requires an empty slot | `data` booking with `pending_payment` status | 400 invalid body/id, 401 unauthorized, 409 slot full/already booked/overlap, 404 slot or offer not found | `bookings`, `booking_plans`, `trip_slots`, `trips`, `offers`, `notifications` | `internal/handler/booking_handler.go`, `internal/usecase/booking_usecase.go` |
| POST | `/bookings/trip/:id` | Legacy direct-trip booking compatibility route | Authenticated user | Auth only | Path `id`; body `seats`, `coupon_code` | `data` booking with `confirmed` status | 400 invalid body/id, 401 unauthorized, 409 insufficient seats or offer/availability issues | `bookings`, `booking_plans`, `trips`, `vehicles`, `offers`, `notifications` | `internal/handler/booking_handler.go`, `internal/usecase/booking_usecase.go` |
| GET | `/bookings/my-orders` | List current user's bookings | Authenticated user | Auth only | No body | `data` array of bookings | 401 unauthorized, 500 load failure | `bookings`, `booking_plans`, `trips`, `slots`, `vehicles` | `internal/handler/booking_handler.go`, `internal/repository/booking_repo.go` |
| PATCH | `/bookings/custom-plan/:id` | Replace the custom plan copy for a user's booking | Authenticated user | Auth only | Path `id`; body `plans` array of booking-plan objects; caller must own booking | `data = null` | 400 invalid body/id, 403 access denied, 404 booking not found | `bookings`, `booking_plans` | `internal/handler/booking_handler.go`, `internal/usecase/booking_usecase.go` |
| GET | `/admin/orders/all` | List all bookings for admins | Authenticated admin | `admin` | No body | `data` array of bookings with trip/slot preloads | 500 load failure | `bookings`, `booking_plans`, `trips`, `trip_slots`, `vehicles` | `internal/handler/booking_handler.go`, `internal/repository/booking_repo.go` |

## Notification APIs

| Method | Route | Description | Auth | Roles / Permissions | Request / Params / Validation | Response | Errors | DB Tables | Source |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| GET | `/notifications/` | List notifications for the current user | Authenticated user | Auth only | No body | `data` array of notifications | 500 load failure | `notifications` | `internal/handler/notification_handler.go`, `internal/usecase/notification_usecase.go` |
| PATCH | `/notifications/:id/read` | Mark a user notification as read | Authenticated user | Auth only | Path `id` required; user must own notification unless admin | `data = null` | 400 invalid id, 403 access denied, 404 not found | `notifications` | `internal/handler/notification_handler.go`, `internal/usecase/notification_usecase.go` |
| GET | `/admin/notifications/` | List admin notifications | Authenticated admin | `admin` | No body | `data` array of admin notifications only | 500 load failure | `notifications` | `internal/handler/notification_handler.go`, `internal/usecase/notification_usecase.go` |
| PATCH | `/admin/notifications/:id/read` | Mark admin notification as read | Authenticated admin | `admin` | Path `id` required | `data = null` | 400 invalid id, 403 access denied, 404 not found | `notifications` | `internal/handler/notification_handler.go`, `internal/usecase/notification_usecase.go` |

## Vehicle APIs

| Method | Route | Description | Auth | Roles / Permissions | Request / Params / Validation | Response | Errors | DB Tables | Source |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| GET | `/vehicles/` | List vehicles with role-aware filtering | Authenticated user | Auth only | No body | `data` array of vehicles | 500 load failure | `vehicles` | `internal/handler/vehicle_handler.go`, `internal/usecase/vehicle_usecase.go` |
| GET | `/vehicles/trip/:trip_id` | Return the vehicle assigned to a trip, if any | Authenticated user | Auth only | Path `trip_id` required | `data` vehicle or `null` | 400 invalid id, 404 not found | `vehicles`, `trips` | `internal/handler/vehicle_handler.go`, `internal/usecase/vehicle_usecase.go` |
| GET | `/vehicles/:id` | Load a vehicle by id | Authenticated user | Auth only | Path `id` required | `data` vehicle | 400 invalid id, 404 not found | `vehicles` | `internal/handler/vehicle_handler.go`, `internal/usecase/vehicle_usecase.go` |
| POST | `/admin/vehicles/` | Create a vehicle | Authenticated user | `manage_fleet` | Body `VehicleRequest`; name/type/total seats required; non-admin creators are forced to their own agency id | `data` vehicle | 400 invalid body/validation, 403 access denied, 500 create failure | `vehicles`, `users` | `internal/handler/vehicle_handler.go`, `internal/usecase/vehicle_usecase.go` |
| PUT | `/admin/vehicles/:id` | Update a vehicle | Authenticated user | `manage_fleet` | Path `id`; body `VehicleRequest` | `data = null` | 400 invalid body/id, 403 access denied, 404 not found | `vehicles`, `trips` | `internal/handler/vehicle_handler.go`, `internal/usecase/vehicle_usecase.go` |
| DELETE | `/admin/vehicles/:id` | Delete a vehicle | Authenticated user | `manage_fleet` | Path `id` required | `data = null` | 400 invalid id, 403 access denied, 404 not found | `vehicles` | `internal/handler/vehicle_handler.go`, `internal/usecase/vehicle_usecase.go` |
| PATCH | `/admin/vehicles/:id/assign-trip` | Assign a vehicle to a trip | Authenticated user | `manage_fleet` | Path `id`; body `AssignVehicleRequest` with `trip_id` | `data = null` | 400 invalid body/id, 403 access denied, 404 not found, 409 assignment conflict | `vehicles`, `trips` | `internal/handler/vehicle_handler.go`, `internal/usecase/vehicle_usecase.go` |

## Offer APIs

| Method | Route | Description | Auth | Roles / Permissions | Request / Params / Validation | Response | Errors | DB Tables | Source |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| GET | `/offers/active` | List active offers | Authenticated user | Auth only | No body | `data` array of active offers | 500 load failure | `offers` | `internal/handler/offer_handler.go`, `internal/usecase/offer_usecase.go` |
| POST | `/offers/validate` | Validate a coupon code | Authenticated user | Auth only | Body `code` required | `data` offer if valid | 400 invalid body/code, 404 not found, 400 inactive/expired | `offers` | `internal/handler/offer_handler.go`, `internal/usecase/offer_usecase.go` |
| GET | `/admin/offers/` | List all offers | Authenticated user | `manage_offers` | No body | `data` array of offers | 500 load failure | `offers` | `internal/handler/offer_handler.go`, `internal/usecase/offer_usecase.go` |
| POST | `/admin/offers/` | Create an offer | Authenticated user | `manage_offers` | Body `OfferRequest`; code/title required; discount 0-100; expiry must be future | `data` offer | 400 invalid body/validation, 409 duplicate code | `offers` | `internal/handler/offer_handler.go`, `internal/usecase/offer_usecase.go` |
| PUT | `/admin/offers/:id` | Update an offer | Authenticated user | `manage_offers` | Path `id`; body `OfferRequest` | `data = null` | 400 invalid body/id, 404 not found, 400 invalid discount/expiry | `offers` | `internal/handler/offer_handler.go`, `internal/usecase/offer_usecase.go` |
| DELETE | `/admin/offers/:id` | Delete an offer | Authenticated user | `manage_offers` | Path `id` required | `data = null` | 400 invalid id, 404 not found | `offers` | `internal/handler/offer_handler.go`, `internal/usecase/offer_usecase.go` |

## Review APIs

| Method | Route | Description | Auth | Roles / Permissions | Request / Params / Validation | Response | Errors | DB Tables | Source |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| POST | `/reviews/` | Create a trip review | Authenticated user | Auth only | Body `ReviewRequest`; `trip_id`, `rating` required; rating 1-5; trip must be completed and user must have a non-cancelled booking | `data` review | 400 invalid body/validation, 403 access denied, 404 not found, 409 duplicate review | `reviews`, `bookings`, `trips` | `internal/handler/review_handler.go`, `internal/usecase/review_usecase.go` |
| GET | `/reviews/me` | List current user's reviews | Authenticated user | Auth only | No body | `data` array of reviews | 500 load failure | `reviews` | `internal/handler/review_handler.go`, `internal/usecase/review_usecase.go` |
| GET | `/reviews/trip/:trip_id` | List reviews for a trip | Authenticated user | Auth only | Path `trip_id` required | `data` array of reviews | 400 invalid id, 404 not found | `reviews`, `trips` | `internal/handler/review_handler.go`, `internal/usecase/review_usecase.go` |
| GET | `/reviews/trip/:trip_id/summary` | Return average rating and count | Authenticated user | Auth only | Path `trip_id` required | `data` `ReviewSummaryResponse` | 400 invalid id, 500 summary failure | `reviews` | `internal/handler/review_handler.go`, `internal/usecase/review_usecase.go` |
| GET | `/admin/reviews/` | List all reviews | Authenticated user | `manage_reviews` | No body | `data` array of reviews | 500 load failure | `reviews` | `internal/handler/review_handler.go`, `internal/usecase/review_usecase.go` |
| DELETE | `/admin/reviews/:id` | Delete a review | Authenticated user | `manage_reviews` | Path `id` required | `data = null` | 400 invalid id, 404 not found | `reviews` | `internal/handler/review_handler.go`, `internal/usecase/review_usecase.go` |

## Complaint APIs

| Method | Route | Description | Auth | Roles / Permissions | Request / Params / Validation | Response | Errors | DB Tables | Source |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| POST | `/complaints/` | Create a booking complaint | Authenticated user | Auth only | Body `ComplaintRequest`; `booking_id`, `title`, `description` required; caller must own booking | `data` complaint | 400 invalid body/validation, 403 access denied, 404 booking not found | `complaints`, `bookings`, `users` | `internal/handler/complaint_handler.go`, `internal/usecase/complaint_usecase.go` |
| GET | `/complaints/me` | List current user's complaints | Authenticated user | Auth only | No body | `data` array of complaints | 500 load failure | `complaints` | `internal/handler/complaint_handler.go`, `internal/usecase/complaint_usecase.go` |
| GET | `/complaints/:id` | Load a complaint by id | Authenticated user | Auth only | Path `id` required; owner or admin only | `data` complaint | 400 invalid id, 403 access denied, 404 not found | `complaints`, `users`, `bookings` | `internal/handler/complaint_handler.go`, `internal/usecase/complaint_usecase.go` |
| GET | `/admin/complaints/` | List all complaints | Authenticated user | `manage_complaints` | No body | `data` array of complaints | 500 load failure | `complaints` | `internal/handler/complaint_handler.go`, `internal/usecase/complaint_usecase.go` |
| PATCH | `/admin/complaints/:id/status` | Update complaint status | Authenticated user | `manage_complaints` | Path `id`; body `ComplaintStatusRequest`; status normalized to pending/in_progress/resolved | `data = null` | 400 invalid body/id, 404 not found | `complaints` | `internal/handler/complaint_handler.go`, `internal/usecase/complaint_usecase.go` |
| DELETE | `/admin/complaints/:id` | Delete a complaint | Authenticated user | `manage_complaints` | Path `id` required | `data = null` | 400 invalid id, 404 not found | `complaints` | `internal/handler/complaint_handler.go`, `internal/usecase/complaint_usecase.go` |

## Tracking APIs

| Method | Route | Description | Auth | Roles / Permissions | Request / Params / Validation | Response | Errors | DB Tables | Source |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| POST | `/tracking/` | Create a tracking point | Authenticated user | `manage_tracking` | Body `TrackingUpdateRequest`; booking and vehicle ids required; latitude -90..90; longitude -180..180; vehicle must match booking's trip | `data` tracking point | 400 invalid body/validation, 403 access denied, 404 not found | `trackings`, `bookings`, `vehicles` | `internal/handler/tracking_handler.go`, `internal/usecase/tracking_usecase.go` |
| GET | `/tracking/booking/:booking_id` | Return latest tracking point for a booking | Authenticated user | Auth only | Path `booking_id` required; admin/driver can read any booking, user can only their own booking | `data` tracking point or `null` | 400 invalid id, 403 access denied, 404 not found | `trackings`, `bookings` | `internal/handler/tracking_handler.go`, `internal/usecase/tracking_usecase.go` |
| GET | `/tracking/booking/:booking_id/history` | Return tracking history for a booking | Authenticated user | Auth only | Path `booking_id` required; same access rules as latest endpoint | `data` array of tracking points | 400 invalid id, 403 access denied, 404 not found | `trackings`, `bookings` | `internal/handler/tracking_handler.go`, `internal/usecase/tracking_usecase.go` |
| GET | `/admin/tracking/` | List all tracking points | Authenticated user | `manage_tracking` | No body | `data` array of tracking points | 500 load failure | `trackings` | `internal/handler/tracking_handler.go`, `internal/usecase/tracking_usecase.go` |

## Notes

- All handlers use the shared response envelope from `internal/response/response.go`.
- Most validation is implemented manually in handlers and use cases rather than through a central validation framework.
- When a route mentions a permission, the actual check happens through `internal/middleware/permission_middleware.go`.
- Public doc consumers should treat `docs/api-endpoints.md` as stale relative to the code-backed endpoints documented here.
