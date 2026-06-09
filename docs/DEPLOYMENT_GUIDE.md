# Deployment Guide

This guide is derived from the actual startup path in `cmd/server/main.go` and the configuration helpers in `internal/config/*`.

## Local Setup

### Prerequisites

- Go 1.25.4 or compatible
- PostgreSQL
- SMTP credentials for OTP email delivery
- A JWT signing secret

### Setup steps

1. Clone the repository.
2. Create a `.env` file with the variables documented in `ENVIRONMENT_VARIABLES.md`.
3. Create the PostgreSQL database referenced by `DB_NAME`.
4. Ensure the SMTP account can send OTP messages.
5. Start the server:

```bash
go run ./cmd/server
```

### Recommended verification

Run the current test suite before starting the service:

```bash
go test ./...
```

If Windows build cache issues appear, set `GOCACHE` to a writable path inside the workspace first.

## Docker Setup

The repository does not include a Dockerfile, compose file, or CI build definition.

That means:

- There is no repo-authored container image to build
- There is no repo-authored compose stack to start
- Container deployment must be created externally from the source code

If you containerize this service, the runtime should mirror the environment variables and command used by `cmd/server/main.go`.

## Production Setup

### Required runtime inputs

- Database connection variables
- SMTP credentials
- `PORT`
- One of `JWT_SECRET` or `JWT_SECRETKEY`
- Optional `GEMINI_API_KEY`

### Production considerations

- Put the service behind HTTPS.
- Revisit cookie security because the current code sets `Secure: false`.
- Keep the PostgreSQL schema migrations and seeders in the startup path or run them as a controlled release step.
- Ensure the reverse proxy forwards the `Authorization` header if bearer-token auth is used.

### Typical production launch

```bash
go build ./cmd/server
./server
```

The binary name can vary by platform and build flags.

## Build Commands

```bash
go build ./cmd/server
go test ./...
```

## Run Commands

```bash
go run ./cmd/server
```

If you need a custom port:

```bash
PORT=9000 go run ./cmd/server
```

## Startup Order

The actual startup sequence in `cmd/server/main.go` is:

1. Load `.env` if present.
2. Validate required environment variables.
3. Connect to PostgreSQL.
4. Run migrations.
5. Seed users and RBAC data.
6. Build repositories, use cases, handlers, and middleware.
7. Register routes.
8. Start the Fiber server.

## Source References

- `cmd/server/main.go`
- `internal/config/env.go`
- `internal/config/databse.go`
- `migrations/migrate.go`
- `internal/seed/seed.go`
- `internal/seed/rbac_seed.go`
