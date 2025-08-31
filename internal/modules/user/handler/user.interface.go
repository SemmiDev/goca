package handler

import (
	"context"

	"github.com/sammidev/goca/internal/modules/user/dto"
)

type UserService interface {
	Register(ctx context.Context, req *dto.RegisterRequest) (*dto.RegisterResponse, error)
	Login(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error)
	VerifyOTP(ctx context.Context, req *dto.VerifyOTPRequest) (*dto.VerifyOTPResponse, error)
	ResendOTP(ctx context.Context, req *dto.ResendOTPRequest) error
	RefreshToken(ctx context.Context, req *dto.RefreshTokenRequest) (*dto.RefreshTokenResponse, error)
	ForgotPassword(ctx context.Context, req *dto.ForgotPasswordRequest) error
	ResetPassword(ctx context.Context, req *dto.ResetPasswordRequest) error
	Setup2FA(ctx context.Context, req *dto.Setup2FARequest) (*dto.Setup2FAResponse, error)
	Verify2FA(ctx context.Context, req *dto.Verify2FARequest) (*dto.Verify2FAResponse, error)
	Disable2FA(ctx context.Context, req *dto.Disable2FARequest) error
}
