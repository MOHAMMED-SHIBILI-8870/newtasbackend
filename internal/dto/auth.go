package dto

type RegisterRequest struct {
	FullName string `json:"full_name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type OTPRequest struct {
	Email   string `json:"email"`
	OTP     string `json:"otp"`
	Purpose string `json:"purpose"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

type ResetPasswordRequest struct {
	Email       string `json:"email"`
	NewPassword string `json:"new_password"`
	OTP         string `json:"otp"`
}

type ResendOTPRequest struct {
	Email   string `json:"email"`
	Purpose string `json:"purpose"`
}

type AuthUserResponse struct {
	ID         uint   `json:"id"`
	FullName   string `json:"full_name"`
	Email      string `json:"email"`
	Role       string `json:"role"`
	IsBlocked  bool   `json:"is_blocked,omitempty"`
	IsVerified bool   `json:"is_verified,omitempty"`
}

type AuthResponse struct {
	AccessToken string           `json:"access_token"`
	User        AuthUserResponse `json:"user"`
}
