# Environment Variables

Source files:
- Env loading: [internal/config/env.go](../internal/config/env.go)
- Required-variable validation: [cmd/server/main.go](../cmd/server/main.go)
- Database DSN: [internal/config/databse.go](../internal/config/databse.go)
- OTP email sending: [internal/usecase/send_opt.go](../internal/usecase/send_opt.go)
- JWT generation and validation: [internal/usecase/jwt.go](../internal/usecase/jwt.go), [internal/usecase/validate_jwt.go](../internal/usecase/validate_jwt.go)

| Variable | Purpose | Required | Default / Notes |
| --- | --- | --- | --- |
| `DB_HOST` | PostgreSQL host | Yes | No default in code |
| `DB_USER` | PostgreSQL user | Yes | No default in code |
| `DB_PASSWORD` | PostgreSQL password | Yes | No default in code |
| `DB_NAME` | PostgreSQL database name | Yes | No default in code |
| `DB_PORT` | PostgreSQL port | Yes | No default in code |
| `DB_SSLMODE` | PostgreSQL SSL mode | Optional | Defaults to `disable` |
| `SMTP_HOST` | SMTP host for OTP email delivery | Yes | No default in code |
| `SMTP_PORT` | SMTP port for OTP email delivery | Yes | No default in code |
| `EMAIL_FROM` | Sender mailbox for OTP email delivery | Yes | No default in code |
| `EMAIL_PASSWORD` | SMTP password for OTP email delivery | Yes | No default in code |
| `PORT` | HTTP server port | Yes | Runtime fallback is `8997`, but startup validation requires it to be set |
| `JWT_SECRET` | JWT signing secret used by token generation | One of `JWT_SECRET` or `JWT_SECRETKEY` is required | `GenerateAccessToken` prefers this variable |
| `JWT_SECRETKEY` | JWT secret used by token validation | One of `JWT_SECRET` or `JWT_SECRETKEY` is required | `ValidateJwt` currently reads only this variable |
| `CORS_ORIGINS` | Allowed frontend origins | Optional | Defaults to local dev origins on ports `5173` and `5174` |
| `GEMINI_API_KEY` | Google GenAI API key | Optional | If missing, the code falls back to the local trip-plan generator |

## Important Runtime Note
The code accepts either `JWT_SECRET` or `JWT_SECRETKEY` at startup, but validation currently reads only `JWT_SECRETKEY`. If only `JWT_SECRET` is set, access-token validation fails even though token generation succeeds.

