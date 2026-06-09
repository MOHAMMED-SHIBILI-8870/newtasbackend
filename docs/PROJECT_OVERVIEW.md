# Project Overview

## Purpose
NewTas is a travel management backend built in Go. It serves public trip discovery, admin trip management, bookings, notifications, RBAC, AI trip planning, complaints, reviews, tracking, and vehicle administration.

Current implementation files:
- Bootstrap and dependency wiring: [cmd/server/main.go](../cmd/server/main.go)
- HTTP routes: [internal/handler/routes/](../internal/handler/routes/)
- Business logic: [internal/usecase/](../internal/usecase/)
- Persistence: [internal/repository/](../internal/repository/)
- Data models: [internal/entity/](../internal/entity/)

## Features Present In Code
- Authentication with register, login, OTP verification, password reset, logout, and refresh token rotation.
- Role-based access control with roles, permissions, and user-role assignments.
- Public and admin trip management.
- Trip plans and trip slots.
- Slot-based booking plus a legacy trip-based booking route kept for compatibility.
- Vehicle management and trip assignment.
- Offers and coupon validation.
- AI trip generation through Gemini with a fallback generator.
- AI trip request submission and admin review.
- Reviews, complaints, notifications, and live tracking.
- Admin user management.

## Features Not Found In Code
- Payment module
- Booking cancellation and refund flows
- Waitlist module
- Guide module
- Driver module
- WebSocket layer
- Cron jobs or background workers
- Docker or CI/CD manifests

## Architecture
The code follows a Clean Architecture style:

- Handlers translate HTTP requests into DTOs, call usecases, and return a shared response envelope.
- Usecases hold business rules, validation, transactional orchestration, and cross-module coordination.
- Repositories encapsulate GORM queries and persistence operations.
- Entities define database-backed models and relations.
- Middleware performs authentication, role checks, and permission checks.

Primary bootstrap path:
- `cmd/server/main.go` loads env vars, connects Postgres, runs migrations, seeds RBAC/users, builds repositories, builds usecases, builds handlers, and mounts routes.

## Technology Stack
- Go 1.25.4
- Fiber v2
- GORM
- PostgreSQL
- JWT (`github.com/golang-jwt/jwt/v5`)
- bcrypt (`golang.org/x/crypto/bcrypt`)
- Google GenAI (`google.golang.org/genai`)
- SMTP for OTP email delivery
- CORS and recovery middleware

## Folder Structure
| Path | Purpose |
| --- | --- |
| `cmd/server` | Application entrypoint and bootstrap |
| `internal/config` | Env loading and database connection |
| `internal/entity` | Database entities and request/response models |
| `internal/dto` | API request/response DTOs |
| `internal/repository` | GORM repositories |
| `internal/usecase` | Business logic and transaction orchestration |
| `internal/handler` | HTTP handlers |
| `internal/handler/routes` | Route registration |
| `internal/middleware` | Auth and permission middleware |
| `internal/response` | Shared response envelope and Fiber error handler |
| `internal/seed` | Startup seed data for users and RBAC |
| `migrations` | Auto-migration helpers |
| `docs` | Generated and legacy documentation |

## Design Patterns
- Repository pattern for database access.
- Use case / service layer for domain rules.
- Dependency injection through constructors in `main.go`.
- Transactional coordination with GORM transactions and row locking.
- DTO mapping in handlers to keep request/response shapes separate from entities.
- Shared envelope response format for all endpoints.
- Middleware-based auth and permission enforcement.

