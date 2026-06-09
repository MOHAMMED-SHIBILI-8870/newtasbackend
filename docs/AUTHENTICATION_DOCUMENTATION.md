# Authentication Documentation

Source files:
- Handler: [internal/handler/auth_handler.go](../internal/handler/auth_handler.go)
- JWT utilities: [internal/usecase/jwt.go](../internal/usecase/jwt.go), [internal/usecase/validate_jwt.go](../internal/usecase/validate_jwt.go)
- OTP utilities: [internal/usecase/otp_verification.go](../internal/usecase/otp_verification.go)
- Password hashing: [internal/usecase/hashing_pass.go](../internal/usecase/hashing_pass.go)
- Auth middleware: [internal/middleware/auth_middleware.go](../internal/middleware/auth_middleware.go)

## Login Flow
1. Client posts `email` and `password` to `POST /auth/login`.
2. Handler loads the user by email.
3. The account must be verified and not blocked.
4. Password is checked with bcrypt.
5. `GenerateAccessToken` creates a JWT with a 1 hour lifetime.
6. `GenerateRefreshToken` creates a random refresh token and hashed DB value.
7. `SaveRefreshToken` stores the hashed token and deletes previous tokens for the same user.
8. Cookies are set:
   - `access_token`
   - `refresh_token`
9. Response body returns `dto.AuthResponse` with the access token and user data.

## Registration Flow
1. Client posts `full_name`, `email`, and `password` to `POST /auth/register`.
2. The handler checks for duplicate email addresses.
3. Password is hashed with bcrypt.
4. A new `entity.User` is created with role `user` and `is_verified = false`.
5. `CreateOTP` stores a signup OTP hash in the `otps` table.
6. `SentOTPEmail` sends the raw OTP through SMTP.
7. The client receives the created user in `dto.AuthUserResponse` format.

## OTP Flow
### Signup verification
1. `POST /auth/register` creates a user and sends an OTP.
2. `POST /auth/verify-otp` verifies the OTP for purpose `signup`.
3. `VerifyOTP` marks the OTP as used and updates the user row to `is_verified = true`.

### Password reset OTP
1. `POST /auth/forgot-password` creates a `reset_password` OTP.
2. `POST /auth/reset-password` verifies the OTP and changes the password.

### Resend OTP
1. `POST /auth/resend-otp` creates a new OTP for the requested purpose.
2. Signup OTPs are refused for already verified users.

## Refresh Token Flow
1. `POST /auth/refresh` reads the `refresh_token` cookie.
2. `RotateRefreshToken` hashes the token, locks the matching row, loads the user, and rejects blocked or unverified users.
3. The old refresh token is deleted.
4. A new refresh token is generated and stored.
5. A fresh access token is generated.
6. New cookies are returned.

## Password Reset Flow
1. `POST /auth/forgot-password` checks that the email exists and sends a `reset_password` OTP.
2. `POST /auth/reset-password` verifies the OTP and updates the password hash.
3. `RevokeRefreshTokensByUserID` removes all live refresh tokens so old sessions cannot survive the reset.

## Session Management
- Access tokens are stored in an HttpOnly cookie and also returned in the response body.
- Refresh tokens are stored as hashes in the database and as plain values in an HttpOnly cookie.
- Refresh tokens are single-session per user because the save and rotate logic deletes older rows.
- `AuthMiddleware` accepts either the `Authorization: Bearer ...` header or the `access_token` cookie.
- The middleware also checks that the user still exists, is verified, and is not blocked.

## Session Settings From Code
- Access cookie expiration: 15 minutes
- Access JWT expiration: 1 hour
- Refresh token cookie expiration: 7 days
- Cookies use `HttpOnly`, `SameSite=Lax`, and `Secure=false` in the current code snapshot

## Security Notes
- The login flow depends on the JWT secret mismatch behavior described in the environment-variable documentation.
- Cookies are not marked secure, so production deployments should sit behind HTTPS and revisit cookie settings.

