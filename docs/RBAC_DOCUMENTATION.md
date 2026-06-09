# RBAC Documentation

Source files:
- Seed data: [internal/seed/rbac_seed.go](../internal/seed/rbac_seed.go)
- Role use case: [internal/usecase/role_usecase.go](../internal/usecase/role_usecase.go)
- Permission use case: [internal/usecase/permission_usecase.go](../internal/usecase/permission_usecase.go)
- Middleware: [internal/middleware/auth_middleware.go](../internal/middleware/auth_middleware.go), [internal/middleware/permission_middleware.go](../internal/middleware/permission_middleware.go)
- Routes: [internal/handler/routes/rbac_routes.go](../internal/handler/routes/rbac_routes.go)

## Roles
Seeded roles in code:
- `admin`
- `agency`
- `guide`
- `driver`
- `support`
- `user`

The canonical role normalizer is implemented in `NormalizeRole` inside [internal/usecase/jwt.go](../internal/usecase/jwt.go).

## Permissions
Seeded permissions in code:
- `manage_users`
- `manage_bookings`
- `manage_tracking`
- `manage_chat`
- `manage_offers`
- `manage_fleet`
- `manage_reviews`
- `manage_complaints`

## Role-Permission Mapping
| Role | Permissions |
| --- | --- |
| `admin` | `manage_users`, `manage_bookings`, `manage_tracking`, `manage_chat`, `manage_offers`, `manage_fleet`, `manage_reviews`, `manage_complaints` |
| `agency` | `manage_bookings`, `manage_offers`, `manage_fleet` |
| `guide` | `manage_bookings`, `manage_tracking`, `manage_chat` |
| `driver` | `manage_tracking` |
| `support` | `manage_chat`, `manage_complaints` |
| `user` | None by default |

## Protected Routes
### Admin role only
- `/admin/users`
- `/admin/trips`
- `/admin/slots`
- `/admin/plans/trip-plans`
- `/admin/orders/all`
- `/admin/notifications`
- `/admin/ai-requests`

### Permission-guarded admin routes
- `manage_users`: `/admin/rbac/*`
- `manage_fleet`: `/admin/vehicles/*`
- `manage_offers`: `/admin/offers/*`
- `manage_reviews`: `/admin/reviews/*`
- `manage_complaints`: `/admin/complaints/*`
- `manage_tracking`: `/tracking` update and `/admin/tracking/*`

### Authenticated routes without extra permission checks
- `/bookings/*`
- `/notifications/*`
- `/reviews/*`
- `/complaints/*`
- `/tracking/booking/*`
- `/ai/*` except admin review endpoints

## How Authorization Works
1. `AuthMiddleware` validates the JWT and populates auth locals.
2. `RoleMiddleware("admin")` restricts routes to admins.
3. `PermissionMiddleware` loads the caller permissions from the role mapping tables.
4. Admins bypass permission checks in `PermissionMiddleware`.

## Persistence Tables
- `roles`
- `permissions`
- `user_roles`
- `role_permissions`

