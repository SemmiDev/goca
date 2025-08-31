package handler

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/sammidev/goca/internal/modules/user/dto"
	"github.com/sammidev/goca/internal/pkg/response"
	"github.com/sammidev/goca/internal/server/api/middleware"
)

type UserHandler struct {
	userService UserService
}

func NewUserHandler(userService UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// Register godoc
//
//	@Summary		Register a new user
//	@Description	Register a new user with email and password
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		dto.RegisterRequest								true	"User registration data"
//	@Success		201		{object}	response.Response{data=dto.RegisterResponse}	"Register Successfully"
//	@Failure		400		{object}	response.Response
//	@Failure		500		{object}	response.Response
//	@Router			/auth/register [post]
func (h *UserHandler) Register(c *fiber.Ctx) error {
	var req dto.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return response.HandleErrorAPI(c, err)
	}

	res, err := h.userService.Register(c.UserContext(), &req)
	if err != nil {
		return response.HandleErrorAPI(c, err)
	}

	return response.HandleSuccessAPI(c, http.StatusCreated, "Register Successfully", res, nil)
}

// Login godoc
//
//	@Summary		Login a user
//	@Description	Authenticate a user with email and password
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		dto.LoginRequest							true	"User login data"
//	@Success		200		{object}	response.Response{data=dto.LoginResponse}	"Login Successfully"
//	@Failure		400		{object}	response.Response
//	@Failure		500		{object}	response.Response
//	@Router			/auth/login [post]
func (h *UserHandler) Login(c *fiber.Ctx) error {
	var req dto.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return response.HandleErrorAPI(c, err)
	}

	res, err := h.userService.Login(c.UserContext(), &req)
	if err != nil {
		return response.HandleErrorAPI(c, err)
	}

	return response.HandleSuccessAPI(c, http.StatusOK, "Login Successfully", res, nil)
}

// VerifyOTP godoc
//
//	@Summary		Verify OTP
//	@Description	Verify the OTP sent to the user
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		dto.VerifyOTPRequest							true	"OTP verification data"
//	@Success		200		{object}	response.Response{data=dto.VerifyOTPResponse}	"OTP verified successfully"
//	@Failure		400		{object}	response.Response
//	@Failure		500		{object}	response.Response
//	@Router			/auth/verify-otp [post]
func (h *UserHandler) VerifyOTP(c *fiber.Ctx) error {
	var req dto.VerifyOTPRequest
	if err := c.BodyParser(&req); err != nil {
		return response.HandleErrorAPI(c, err)
	}

	res, err := h.userService.VerifyOTP(c.UserContext(), &req)
	if err != nil {
		return response.HandleErrorAPI(c, err)
	}

	return response.HandleSuccessAPI(c, http.StatusOK, "OTP verified successfully", res, nil)
}

// ResendOTP godoc
//
//	@Summary		Resend OTP
//	@Description	Resend the OTP sent to the user
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		dto.ResendOTPRequest	true	"Resend OTP data"
//	@Success		200		{object}	response.Response		"OTP resent successfully"
//	@Failure		400		{object}	response.Response
//	@Failure		500		{object}	response.Response
//	@Router			/auth/resend-otp [post]
func (h *UserHandler) ResendOTP(c *fiber.Ctx) error {
	var req dto.ResendOTPRequest
	if err := c.BodyParser(&req); err != nil {
		return response.HandleErrorAPI(c, err)
	}

	err := h.userService.ResendOTP(c.UserContext(), &req)
	if err != nil {
		return response.HandleErrorAPI(c, err)
	}

	return response.HandleSuccessAPI(c, http.StatusOK, "OTP resent successfully", nil, nil)
}

// RefreshToken godoc
//
//	@Summary		Refresh Token
//	@Description	Refresh the authentication token
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		dto.RefreshTokenRequest								true	"Refresh token data"
//	@Success		200		{object}	response.Response{data=dto.RefreshTokenResponse}	"Token refreshed successfully"
//	@Failure		400		{object}	response.Response
//	@Failure		500		{object}	response.Response
//	@Router			/auth/refresh-token [post]
func (h *UserHandler) RefreshToken(c *fiber.Ctx) error {
	var req dto.RefreshTokenRequest
	if err := c.BodyParser(&req); err != nil {
		return response.HandleErrorAPI(c, err)
	}

	res, err := h.userService.RefreshToken(c.UserContext(), &req)
	if err != nil {
		return response.HandleErrorAPI(c, err)
	}

	return response.HandleSuccessAPI(c, http.StatusOK, "Token refreshed successfully", res, nil)
}

// ForgotPassword godoc
//
//	@Summary		Forgot Password
//	@Description	Send a reset password email to the user
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		dto.ForgotPasswordRequest	true	"Forgot password data"
//	@Success		200		{object}	response.Response			"Reset password email sent"
//	@Failure		400		{object}	response.Response
//	@Failure		500		{object}	response.Response
//	@Router			/auth/forgot-password [post]
func (h *UserHandler) ForgotPassword(c *fiber.Ctx) error {
	var req dto.ForgotPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return response.HandleErrorAPI(c, err)
	}

	err := h.userService.ForgotPassword(c.UserContext(), &req)
	if err != nil {
		return response.HandleErrorAPI(c, err)
	}

	return response.HandleSuccessAPI(c, http.StatusOK, "Reset password email sent", nil, nil)
}

// ResetPassword godoc
//
//	@Summary		Reset Password
//	@Description	Reset the user's password
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		dto.ResetPasswordRequest	true	"Reset password data"
//	@Failure		400		{object}	response.Response
//	@Failure		500		{object}	response.Response
//	@Router			/auth/reset-password [post]
func (h *UserHandler) ResetPassword(c *fiber.Ctx) error {
	var req dto.ResetPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return response.HandleErrorAPI(c, err)
	}

	err := h.userService.ResetPassword(c.UserContext(), &req)
	if err != nil {
		return response.HandleErrorAPI(c, err)
	}

	return response.HandleSuccessAPI(c, http.StatusOK, "Password reset successfully", nil, nil)
}

// Setup2FA godoc
//
//	@Summary		Setup Two-Factor Authentication
//	@Description	Initiate 2FA setup for the authenticated user
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	response.Response{data=dto.Setup2FAResponse}	"2FA setup initiated"
//	@Failure		400	{object}	response.Response
//	@Failure		500	{object}	response.Response
//	@Router			/auth/setup-2fa [post]
func (h *UserHandler) Setup2FA(c *fiber.Ctx) error {
	req := dto.Setup2FARequest{
		UserID: middleware.GetUser(c).UserID,
	}

	res, err := h.userService.Setup2FA(c.UserContext(), &req)
	if err != nil {
		return response.HandleErrorAPI(c, err)
	}

	return response.HandleSuccessAPI(c, http.StatusOK, "2FA setup initiated", res, nil)
}

// Verify2FA godoc
//
//	@Summary		Verify Two-Factor Authentication
//	@Description	Verify 2FA for the authenticated user
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		dto.Verify2FARequest							true	"Verify 2FA data"
//	@Success		200		{object}	response.Response{data=dto.Verify2FAResponse}	"2FA verified successfully"
//	@Failure		400		{object}	response.Response
//	@Failure		500		{object}	response.Response
//	@Router			/auth/verify-2fa [post]
func (h *UserHandler) Verify2FA(c *fiber.Ctx) error {
	var req dto.Verify2FARequest
	if err := c.BodyParser(&req); err != nil {
		return response.HandleErrorAPI(c, err)
	}

	req.UserID = middleware.GetUser(c).UserID

	res, err := h.userService.Verify2FA(c.UserContext(), &req)
	if err != nil {
		return response.HandleErrorAPI(c, err)
	}

	return response.HandleSuccessAPI(c, http.StatusOK, "2FA verified successfully", res, nil)
}

// Disable2FA godoc
//
//	@Summary		Disable Two-Factor Authentication
//	@Description	Disable 2FA for the authenticated user
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	response.Response	"2FA disabled successfully"
//	@Failure		400	{object}	response.Response
//	@Failure		500	{object}	response.Response
//	@Router			/auth/disable-2fa [post]
func (h *UserHandler) Disable2FA(c *fiber.Ctx) error {
	req := dto.Disable2FARequest{
		UserID: middleware.GetUser(c).UserID,
	}

	err := h.userService.Disable2FA(c.UserContext(), &req)
	if err != nil {
		return response.HandleErrorAPI(c, err)
	}

	return response.HandleSuccessAPI(c, http.StatusOK, "2FA disabled successfully", nil, nil)
}
