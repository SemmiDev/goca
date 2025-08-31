package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/sammidev/goca/internal/modules/user/entity"
)

type UserResponse struct {
	ID               uuid.UUID         `json:"id" example:"0198f10c-98c7-71ab-bc9a-7e148b5ece17"`
	Email            string            `json:"email" example:"sammi@example.com"`
	FirstName        string            `json:"first_name" example:"Sammi"`
	LastName         *string           `json:"last_name,omitempty" validate:"optional" example:"Aldhi Yanto"`
	FullName         string            `json:"full_name" example:"Sammi Aldhi Yanto"`
	Status           entity.UserStatus `json:"status" example:"active"`
	TwoFactorEnabled bool              `json:"two_factor_enabled" example:"false"`
	CreatedAt        time.Time         `json:"created_at" example:"2025-06-01T20:50:35.388851+07:00"`
	UpdatedAt        time.Time         `json:"updated_at" example:"2025-06-01T20:50:35.388851+07:00"`
}

type (
	RegisterRequest struct {
		Email     string  `json:"email" validate:"required,email,max=255" example:"sammi@example.com"`
		FirstName string  `json:"first_name" validate:"required,min=2,max=100" example:"Sammi"`
		LastName  *string `json:"last_name" example:"Aldhi Yanto"`
		Password  string  `json:"password" validate:"required,password" example:"Password123@"`
	}

	RegisterResponse struct {
		*UserResponse
	}
)

type (
	LoginRequest struct {
		Email    string `json:"email" validate:"required,email" example:"sammi@example.com"`
		Password string `json:"password" validate:"required,password" example:"Password123@"`
		Remember bool   `json:"remember" example:"false"`
	}

	LoginResponse struct {
		*UserResponse
		TwoFactorEnabled      bool      `json:"two_factor_enabled" example:"false"`
		AccessToken           string    `json:"access_token" validate:"required" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"`
		RefreshToken          string    `json:"refresh_token" validate:"required" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"`
		AccessTokenExpiresAt  time.Time `json:"access_token_expires_at" example:"2025-06-01T20:50:35.388851+07:00"`
		RefreshTokenExpiresAt time.Time `json:"refresh_token_expires_at" example:"2025-06-01T20:50:35.388851+07:00"`
	}
)

type (
	VerifyOTPRequest struct {
		Email string `json:"email" validate:"required,email" example:"sammi@example.com"`
		OTP   string `json:"otp" validate:"required,len=6,numeric" example:"123456"`
	}

	VerifyOTPResponse struct {
		*UserResponse
	}
)

type (
	ResendOTPRequest struct {
		Email string `json:"email" validate:"required,email" example:"sammi@example.com"`
	}
)

type (
	RefreshTokenRequest struct {
		RefreshToken string `json:"refresh_token" validate:"required" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"`
	}

	RefreshTokenResponse struct {
		AccessToken           string    `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"`
		RefreshToken          string    `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"`
		AccessTokenExpiresAt  time.Time `json:"access_token_expires_at" example:"2025-06-01T20:50:35.388851+07:00"`
		RefreshTokenExpiresAt time.Time `json:"refresh_token_expires_at" example:"2025-06-01T20:50:35.388851+07:00"`
	}
)

type (
	ForgotPasswordRequest struct {
		Email string `json:"email" validate:"required,email" example:"sammi@example.com"`
	}
)

type (
	ResetPasswordRequest struct {
		Email       string `json:"email" validate:"required,email" example:"sammi@example.com"`
		OTP         string `json:"otp" validate:"required,len=6,numeric" example:"123456"`
		NewPassword string `json:"new_password" validate:"required,password" example:"password123"`
	}
)

type (
	Setup2FARequest struct {
		UserID uuid.UUID `json:"user_id" validate:"required" example:"123e4567-e89b-12d3-a456-426655440000"`
	}

	Setup2FAResponse struct {
		Secret string `json:"secret" example:"12345678901234567890"`
		QRCode string `json:"qr_code" example:"data:image/png;base64,..."` // URL atau data URI untuk QR code
	}
)

type (
	Verify2FARequest struct {
		UserID uuid.UUID `json:"user_id" validate:"required" example:"123e4567-e89b-12d3-a456-426655440000"`
		Code   string    `json:"code" validate:"required,len=6,numeric" example:"123456"`
	}

	Verify2FAResponse struct {
		Verified              bool      `json:"verified" example:"true"`
		AccessToken           string    `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"`
		RefreshToken          string    `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"`
		AccessTokenExpiresAt  time.Time `json:"access_token_expires_at" example:"2025-06-01T20:50:35.388851+07:00"`
		RefreshTokenExpiresAt time.Time `json:"refresh_token_expires_at" example:"2025-06-01T20:50:35.388851+07:00"`
	}
)

type (
	Disable2FARequest struct {
		UserID uuid.UUID `json:"user_id" validate:"required" example:"123e4567-e89b-12d3-a456-426655440000"`
	}
)

func RegisterRequestToUserEntity(payload *RegisterRequest) *entity.User {
	user := &entity.User{
		ID:        uuid.Must(uuid.NewV7()),
		Email:     payload.Email,
		FirstName: payload.FirstName,
		LastName:  payload.LastName,
		Password:  payload.Password,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	user.SetStatus(entity.UserStatusPending)
	user.GenerateFullName()

	return user
}

func UserEntityToUserResponse(user *entity.User) *UserResponse {
	return &UserResponse{
		ID:               user.ID,
		Email:            user.Email,
		FirstName:        user.FirstName,
		LastName:         user.LastName,
		FullName:         user.FullName,
		Status:           user.Status,
		TwoFactorEnabled: user.IsTwoFactorEnabled(),
		CreatedAt:        user.CreatedAt,
		UpdatedAt:        user.UpdatedAt,
	}
}
