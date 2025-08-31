package service

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"github.com/sammidev/goca/internal/config"
	"github.com/sammidev/goca/internal/modules/user/dto"
	"github.com/sammidev/goca/internal/modules/user/entity"
	"github.com/sammidev/goca/internal/pkg/apperror"
	"github.com/sammidev/goca/internal/pkg/cache"
	"github.com/sammidev/goca/internal/pkg/database"
	"github.com/sammidev/goca/internal/pkg/encoding"
	"github.com/sammidev/goca/internal/pkg/logger"
	"github.com/sammidev/goca/internal/pkg/observability"
	"github.com/sammidev/goca/internal/pkg/password"
	"github.com/sammidev/goca/internal/pkg/random"
	"github.com/sammidev/goca/internal/pkg/ratelimit"
	"github.com/sammidev/goca/internal/pkg/token"
	"github.com/sammidev/goca/internal/pkg/validator"
	"github.com/sammidev/goca/internal/pkg/worker"
	"github.com/skip2/go-qrcode"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type UserService struct {
	cfg       *config.Config
	logger    logger.Logger
	validator validator.Validator
	db        database.Database
	token     token.Token
	cache     cache.Cache
	tracer    trace.Tracer
	worker    worker.TaskDistributor
	limiter   ratelimit.RateLimiter
	userRepo  UserRepository
}

func NewUserService(
	cfg *config.Config,
	logger logger.Logger,
	validator validator.Validator,
	db database.Database,
	token token.Token,
	cache cache.Cache,
	worker worker.TaskDistributor,
	limiter ratelimit.RateLimiter,
	userRepo UserRepository,
) *UserService {
	return &UserService{
		cfg:       cfg,
		logger:    logger.WithComponent("user_service"),
		validator: validator,
		db:        db,
		token:     token,
		cache:     cache,
		worker:    worker,
		tracer:    otel.Tracer("user_service"),
		limiter:   limiter,
		userRepo:  userRepo,
	}
}

// =============================================================================
// VALIDATION & RATE LIMITING HELPERS
// =============================================================================

func (s *UserService) hashPassword(ctx context.Context, plainPassword string) (string, error) {
	return observability.TraceOperation(ctx, s.tracer, "hash_password", func(ctx context.Context) (string, error) {
		hashedPassword, err := password.HashPassword(plainPassword)
		if err != nil {
			return "", apperror.NewAppError(apperror.ErrCodeInternalError, "Failed to hash password")
		}
		return hashedPassword, nil
	})
}

func (s *UserService) checkRateLimit(ctx context.Context, operation, identifier string) error {
	_, err := observability.TraceOperation(ctx, s.tracer, "check_rate_limit", func(ctx context.Context) (struct{}, error) {
		limiterKey := fmt.Sprintf("%s:%s:%s", config.AuthRateLimiterKey, operation, identifier)
		limiterCtx, err := s.limiter.Take(ctx, limiterKey)
		if err != nil {
			s.logger.WithContext(ctx).Error("Failed to check rate limit", "error", err)
			return struct{}{}, apperror.NewAppError(apperror.ErrCodeInternalError, "Failed to check rate limit")
		}
		if limiterCtx.IsExceeded {
			return struct{}{}, apperror.NewAppError(apperror.ErrCodeTooManyRequests, "Too many requests")
		}
		return struct{}{}, nil
	})
	return err
}

// =============================================================================
// USER REPOSITORY HELPERS
// =============================================================================

func (s *UserService) getUserByEmail(ctx context.Context, email string) (*entity.User, error) {
	ctx, span := s.tracer.Start(ctx, "repo.GetByEmail")
	defer span.End()

	span.SetAttributes(attribute.String("email", email))

	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}

	return user, err
}

func (s *UserService) getUserByID(ctx context.Context, userID uuid.UUID) (*entity.User, error) {
	ctx, span := s.tracer.Start(ctx, "repo.GetByID")
	defer span.End()

	span.SetAttributes(attribute.String("user_id", userID.String()))

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}

	return user, err
}

func (s *UserService) createUser(ctx context.Context, user *entity.User) error {
	_, err := observability.TraceOperation(ctx, s.tracer, "repo.CreateUser", func(ctx context.Context) (struct{}, error) {
		return struct{}{}, s.userRepo.Create(ctx, user)
	}, attribute.String("user_id", user.ID.String()))
	return err
}

func (s *UserService) updateUser(ctx context.Context, user *entity.User) error {
	_, err := observability.TraceOperation(ctx, s.tracer, "repo.UpdateUser", func(ctx context.Context) (struct{}, error) {
		return struct{}{}, s.userRepo.Update(ctx, user)
	}, attribute.String("user_id", user.ID.String()))
	return err
}

// =============================================================================
// PASSWORD HELPERS
// =============================================================================

func (s *UserService) verifyPassword(ctx context.Context, plainPassword, hashedPassword string) error {
	_, err := observability.TraceOperation(ctx, s.tracer, "verify_password", func(ctx context.Context) (struct{}, error) {
		if !password.CheckPasswordHash(plainPassword, hashedPassword) {
			return struct{}{}, apperror.ErrUserIncorrectPassword
		}
		return struct{}{}, nil
	})
	return err
}

// =============================================================================
// OTP & CACHE HELPERS
// =============================================================================

func (s *UserService) generateAndCacheOTP(ctx context.Context, cacheKeyPrefix, identifier string) (string, error) {
	ctx, span := s.tracer.Start(ctx, "generate_and_cache_otp")
	defer span.End()

	cacheKey := fmt.Sprintf("%s:%s", cacheKeyPrefix, identifier)
	span.SetAttributes(attribute.String("cache.key", cacheKey))

	otpCode, err := random.GenerateNumericOTP(config.OTPCodeLength)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return "", apperror.NewAppError(apperror.ErrCodeInternalError, "Failed to generate OTP")
	}

	if err := s.cache.Set(ctx, cacheKey, otpCode, config.EmailConfirmationExpInMin); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return "", apperror.NewAppError(apperror.ErrCodeInternalError, "Failed to cache OTP")
	}

	return otpCode, nil
}

func (s *UserService) getCachedOTP(ctx context.Context, cacheKeyPrefix, identifier string) (string, error) {
	ctx, span := s.tracer.Start(ctx, "get_cached_otp")
	defer span.End()

	cacheKey := fmt.Sprintf("%s:%s", cacheKeyPrefix, identifier)
	span.SetAttributes(attribute.String("cache.key", cacheKey))

	cachedOTP, err := s.cache.Get(ctx, cacheKey)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return "", apperror.ErrOTPExpired
	}

	return cachedOTP.(string), nil
}

func (s *UserService) deleteCachedOTP(ctx context.Context, cacheKeyPrefix, identifier string) error {
	_, err := observability.TraceOperation(ctx, s.tracer, "delete_cached_otp", func(ctx context.Context) (struct{}, error) {
		cacheKey := fmt.Sprintf("%s:%s", cacheKeyPrefix, identifier)
		if err := s.cache.Delete(ctx, cacheKey); err != nil {
			s.logger.WithContext(ctx).Warn("Failed to delete cache", "cache_key", cacheKey, "error", err)
		}
		return struct{}{}, nil
	})
	return err
}

// =============================================================================
// TOKEN HELPERS
// =============================================================================

type TokenPair struct {
	AccessToken  *token.GenerateTokenResponse
	RefreshToken *token.GenerateTokenResponse
}

func (s *UserService) generateAuthTokens(ctx context.Context, userID uuid.UUID, remember bool) (*TokenPair, error) {
	_, span := s.tracer.Start(ctx, "generate_auth_tokens")
	defer span.End()

	span.SetAttributes(
		attribute.String("user_id", userID.String()),
		attribute.Bool("remember", remember),
	)

	var accessTokenExpiry, refreshTokenExpiry time.Duration

	if remember {
		accessTokenExpiry = s.cfg.AuthAccessTokenExpiryExtended
		refreshTokenExpiry = s.cfg.AuthRefreshTokenExpiryExtended
	} else {
		accessTokenExpiry = s.cfg.AuthAccessTokenExpiry
		refreshTokenExpiry = s.cfg.AuthRefreshTokenExpiry
	}

	accessToken, err := s.token.GenerateToken(userID, accessTokenExpiry)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, apperror.NewAppError(apperror.ErrCodeInternalError, "Failed to generate access token")
	}

	refreshToken, err := s.token.GenerateToken(userID, refreshTokenExpiry)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, apperror.NewAppError(apperror.ErrCodeInternalError, "Failed to generate refresh token")
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *UserService) verifyToken(ctx context.Context, tokenString string) (*token.Payload, error) {
	_, span := s.tracer.Start(ctx, "verify_token")
	defer span.End()

	payload, err := s.token.VerifyToken(tokenString)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, apperror.ErrInvalidToken
	}

	return payload, nil
}

// =============================================================================
// WORKER HELPERS
// =============================================================================

func (s *UserService) queueVerificationEmail(ctx context.Context, user *entity.User, otpCode string) error {
	_, err := observability.TraceOperation(ctx, s.tracer, "queue.SendVerifyEmail", func(ctx context.Context) (struct{}, error) {
		span := trace.SpanFromContext(ctx)
		span.SetAttributes(
			attribute.String("worker.queue", worker.Critical),
			attribute.String("worker.user_id", user.ID.String()),
		)

		emailPayload := worker.PayloadSendVerifyEmail{
			UserID:                     user.ID,
			Name:                       user.FullName,
			Email:                      user.Email,
			VerificationCode:           otpCode,
			VerificationCodeExpiration: int(config.EmailConfirmationExpInMin.Minutes()),
		}

		taskOptions := []asynq.Option{
			asynq.MaxRetry(worker.TaskSendVerifyEmailMaxRetry),
			asynq.Queue(worker.Critical),
		}

		return struct{}{}, s.worker.DistributeTaskSendVerifyEmail(ctx, &emailPayload, taskOptions...)
	})
	return err
}

func (s *UserService) queueForgotPasswordEmail(ctx context.Context, user *entity.User, otpCode string) error {
	_, err := observability.TraceOperation(ctx, s.tracer, "queue.SendForgotPasswordEmail", func(ctx context.Context) (struct{}, error) {
		span := trace.SpanFromContext(ctx)
		span.SetAttributes(
			attribute.String("worker.queue", worker.Critical),
			attribute.String("worker.user_id", user.ID.String()),
		)

		emailPayload := worker.PayloadSendForgotPasswordEmail{
			UserID:                     user.ID,
			Name:                       user.FullName,
			Email:                      user.Email,
			VerificationCode:           otpCode,
			VerificationCodeExpiration: int(config.EmailConfirmationExpInMin.Minutes()),
		}

		taskOptions := []asynq.Option{
			asynq.MaxRetry(worker.TaskSendVerifyEmailMaxRetry),
			asynq.Queue(worker.Critical),
		}

		return struct{}{}, s.worker.DistributeTaskSendForgotPasswordEmail(ctx, &emailPayload, taskOptions...)
	})
	return err
}

// =============================================================================
// BUSINESS LOGIC VALIDATION HELPERS
// =============================================================================

func (s *UserService) validateUserState(user *entity.User, requireEmailVerified, requireActive bool) error {
	if requireEmailVerified && !user.IsEmailVerified() {
		return apperror.ErrUserEmailNotVerified
	}
	if requireActive && !user.IsActive() {
		return apperror.ErrUserInactive
	}
	return nil
}

// =============================================================================
// 2FA HELPERS
// =============================================================================

func (s *UserService) generate2FASecret(ctx context.Context, userEmail string) (*otp.Key, error) {
	_, span := s.tracer.Start(ctx, "generate_2fa_secret")
	defer span.End()

	span.SetAttributes(attribute.String("email", userEmail))

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      s.cfg.AppName,
		AccountName: userEmail,
		SecretSize:  20,
		Digits:      otp.DigitsSix,
		Algorithm:   otp.AlgorithmSHA1,
		Period:      30,
	})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, apperror.NewAppError(apperror.ErrCodeInternalError, "Failed to generate 2FA secret")
	}

	return key, nil
}

func (s *UserService) generateQRCode(ctx context.Context, otpAuthURL string) (string, error) {
	_, span := s.tracer.Start(ctx, "generate_qr_code")
	defer span.End()

	qrCodeData, err := qrcode.Encode(otpAuthURL, qrcode.Medium, 256)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return "", apperror.NewAppError(apperror.ErrCodeInternalError, "Failed to generate QR code")
	}

	return fmt.Sprintf("data:image/png;base64,%s", encoding.Base64Encode(qrCodeData)), nil
}

// =============================================================================
// MAIN SERVICE METHODS
// =============================================================================

func (s *UserService) Register(ctx context.Context, req *dto.RegisterRequest) (*dto.RegisterResponse, error) {
	s.logger.WithContext(ctx).Info("Creating new user", "email", req.Email)

	ctx, span := s.tracer.Start(ctx, "service.Register")
	defer span.End()

	// Pre-transaction validations
	if err := s.checkRateLimit(ctx, "register", req.Email); err != nil {
		return nil, err
	}

	if err := s.validator.ValidateAndGetErrors(req); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, apperror.NewValidationError(err)
	}

	var createdUser *entity.User

	// Database transaction
	err := s.db.WithTransaction(ctx, func(txCtx context.Context) error {
		// Check if user already exists
		existingUser, err := s.getUserByEmail(txCtx, req.Email)
		if err != nil && !errors.Is(err, apperror.ErrUserNotFound) {
			return err
		}
		if existingUser != nil {
			return apperror.ErrUserAlreadyExists
		}

		// Create user entity
		user := dto.RegisterRequestToUserEntity(req)

		// Hash password
		hashedPassword, err := s.hashPassword(txCtx, user.Password)
		if err != nil {
			return err
		}
		user.Password = hashedPassword

		// Create user in database
		if err := s.createUser(txCtx, user); err != nil {
			return err
		}
		createdUser = user

		// Generate and cache OTP
		otpCode, err := s.generateAndCacheOTP(txCtx, "email_verification", user.Email)
		if err != nil {
			return err
		}

		// Queue verification email
		return s.queueVerificationEmail(txCtx, user, otpCode)
	})

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	s.logger.WithContext(ctx).Info("User created successfully", "user_id", createdUser.ID)
	return &dto.RegisterResponse{
		UserResponse: dto.UserEntityToUserResponse(createdUser),
	}, nil
}

func (s *UserService) Login(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error) {
	s.logger.WithContext(ctx).Info("User login attempt", "email", req.Email)

	ctx, span := s.tracer.Start(ctx, "service.Login")
	defer span.End()

	// Pre-authentication checks
	if err := s.checkRateLimit(ctx, "login", req.Email); err != nil {
		return nil, err
	}

	if err := s.validator.ValidateAndGetErrors(req); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, apperror.NewValidationError(err)
	}

	// Get user
	user, err := s.getUserByEmail(ctx, req.Email)
	if err != nil {
		// Hide the actual error (user not found) for security
		return nil, apperror.ErrUserIncorrectPassword
	}

	// Verify password
	if err := s.verifyPassword(ctx, req.Password, user.Password); err != nil {
		return nil, err
	}

	// Business logic validations
	if err := s.validateUserState(user, true, true); err != nil {
		return nil, err
	}

	// Handle 2FA if enabled
	if user.IsTwoFactorEnabled() {
		s.logger.WithContext(ctx).Info("Login requires 2FA", "user_id", user.ID)
		return &dto.LoginResponse{
			UserResponse:     dto.UserEntityToUserResponse(user),
			TwoFactorEnabled: true,
		}, nil
	}

	// Generate authentication tokens
	tokenPair, err := s.generateAuthTokens(ctx, user.ID, req.Remember)
	if err != nil {
		return nil, err
	}

	s.logger.WithContext(ctx).Info("User logged in successfully", "user_id", user.ID)
	return &dto.LoginResponse{
		AccessToken:           tokenPair.AccessToken.Value,
		RefreshToken:          tokenPair.RefreshToken.Value,
		AccessTokenExpiresAt:  tokenPair.AccessToken.ExpiresAt,
		RefreshTokenExpiresAt: tokenPair.RefreshToken.ExpiresAt,
		UserResponse:          dto.UserEntityToUserResponse(user),
	}, nil
}

func (s *UserService) VerifyOTP(ctx context.Context, req *dto.VerifyOTPRequest) (*dto.VerifyOTPResponse, error) {
	s.logger.WithContext(ctx).Info("Verifying OTP", "email", req.Email)

	ctx, span := s.tracer.Start(ctx, "service.VerifyOTP")
	defer span.End()

	if err := s.checkRateLimit(ctx, "verify_otp", req.Email); err != nil {
		return nil, err
	}

	if err := s.validator.ValidateAndGetErrors(req); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, apperror.NewValidationError(err)
	}

	var user *entity.User

	err := s.db.WithTransaction(ctx, func(txCtx context.Context) error {
		var err error
		user, err = s.getUserByEmail(txCtx, req.Email)
		if err != nil {
			if errors.Is(err, apperror.ErrUserNotFound) {
				return apperror.ErrInvalidOTP
			}
			return err
		}

		if user.IsEmailVerified() {
			return apperror.NewAppError(apperror.ErrCodeBadRequest, "Email already verified")
		}

		// Get and verify OTP
		cachedOTP, err := s.getCachedOTP(txCtx, "email_verification", req.Email)
		if err != nil {
			return err
		}

		if cachedOTP != req.OTP {
			return apperror.ErrInvalidOTP
		}

		// Update user
		user.VerifyEmail(time.Now())
		user.Activate()

		if err := s.updateUser(txCtx, user); err != nil {
			return err
		}

		// Clean up cache
		return s.deleteCachedOTP(txCtx, "email_verification", req.Email)
	})

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	s.logger.WithContext(ctx).Info("Email verified successfully", "user_id", user.ID)
	return &dto.VerifyOTPResponse{
		UserResponse: dto.UserEntityToUserResponse(user),
	}, nil
}

func (s *UserService) ResendOTP(ctx context.Context, req *dto.ResendOTPRequest) error {
	s.logger.WithContext(ctx).Info("Resending OTP", "email", req.Email)

	ctx, span := s.tracer.Start(ctx, "service.ResendOTP")
	defer span.End()

	if err := s.checkRateLimit(ctx, "resend_otp", req.Email); err != nil {
		return err
	}

	if err := s.validator.ValidateAndGetErrors(req); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return apperror.NewValidationError(err)
	}

	err := s.db.WithTransaction(ctx, func(txCtx context.Context) error {
		user, err := s.getUserByEmail(txCtx, req.Email)
		if err != nil {
			if errors.Is(err, apperror.ErrUserNotFound) {
				return nil // Silent fail for security
			}
			return err
		}

		if user.IsEmailVerified() {
			return apperror.NewAppError(apperror.ErrCodeBadRequest, "Email already verified")
		}

		// Generate and cache new OTP
		otpCode, err := s.generateAndCacheOTP(txCtx, "email_verification", user.Email)
		if err != nil {
			return err
		}

		// Queue verification email
		return s.queueVerificationEmail(txCtx, user, otpCode)
	})

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}

	return err
}

func (s *UserService) RefreshToken(ctx context.Context, req *dto.RefreshTokenRequest) (*dto.RefreshTokenResponse, error) {
	s.logger.WithContext(ctx).Info("Refreshing token")

	ctx, span := s.tracer.Start(ctx, "service.RefreshToken")
	defer span.End()

	if err := s.validator.ValidateAndGetErrors(req); err != nil {
		span.RecordError(err)
		return nil, apperror.NewValidationError(err)
	}

	// Verify refresh token
	payload, err := s.verifyToken(ctx, req.RefreshToken)
	if err != nil {
		return nil, err
	}

	// Get user
	user, err := s.getUserByID(ctx, payload.UserID)
	if err != nil {
		return nil, apperror.ErrUserNotFound
	}

	// Validate user state
	if err := s.validateUserState(user, true, true); err != nil {
		return nil, err
	}

	// Generate new tokens
	newAccess, err := s.token.GenerateToken(user.ID, s.cfg.AuthAccessTokenExpiry)
	if err != nil {
		span.RecordError(err)
		return nil, apperror.NewAppError(apperror.ErrCodeInternalError, "Failed to generate access token")
	}

	newRefresh, err := s.token.GenerateToken(user.ID, s.cfg.AuthRefreshTokenExpiry)
	if err != nil {
		span.RecordError(err)
		return nil, apperror.NewAppError(apperror.ErrCodeInternalError, "Failed to generate refresh token")
	}

	s.logger.WithContext(ctx).Info("Token refreshed successfully", "user_id", user.ID)
	return &dto.RefreshTokenResponse{
		AccessToken:           newAccess.Value,
		RefreshToken:          newRefresh.Value,
		AccessTokenExpiresAt:  newAccess.ExpiresAt,
		RefreshTokenExpiresAt: newRefresh.ExpiresAt,
	}, nil
}

func (s *UserService) ForgotPassword(ctx context.Context, req *dto.ForgotPasswordRequest) error {
	s.logger.WithContext(ctx).Info("Forgot password request", "email", req.Email)

	ctx, span := s.tracer.Start(ctx, "service.ForgotPassword")
	defer span.End()

	if err := s.checkRateLimit(ctx, "forgot_password", req.Email); err != nil {
		return err
	}

	if err := s.validator.ValidateAndGetErrors(req); err != nil {
		span.RecordError(err)
		return apperror.NewValidationError(err)
	}

	err := s.db.WithTransaction(ctx, func(txCtx context.Context) error {
		user, err := s.getUserByEmail(txCtx, req.Email)
		if err != nil {
			if errors.Is(err, apperror.ErrUserNotFound) {
				return nil // Silent fail for security
			}
			return err
		}

		// Generate and cache OTP
		otpCode, err := s.generateAndCacheOTP(txCtx, "password_reset", user.Email)
		if err != nil {
			return err
		}

		// Queue forgot password email
		return s.queueForgotPasswordEmail(txCtx, user, otpCode)
	})

	if err != nil {
		span.RecordError(err)
	}

	return err
}

func (s *UserService) ResetPassword(ctx context.Context, req *dto.ResetPasswordRequest) error {
	s.logger.WithContext(ctx).Info("Reset password", "email", req.Email)

	ctx, span := s.tracer.Start(ctx, "service.ResetPassword")
	defer span.End()

	if err := s.checkRateLimit(ctx, "reset_password", req.Email); err != nil {
		return err
	}

	if err := s.validator.ValidateAndGetErrors(req); err != nil {
		span.RecordError(err)
		return apperror.NewValidationError(err)
	}

	err := s.db.WithTransaction(ctx, func(txCtx context.Context) error {
		user, err := s.getUserByEmail(txCtx, req.Email)
		if err != nil {
			if errors.Is(err, apperror.ErrUserNotFound) {
				return apperror.ErrInvalidOTP
			}
			return err
		}

		// Get and verify OTP
		cachedOTP, err := s.getCachedOTP(txCtx, "password_reset", req.Email)
		if err != nil {
			return err
		}

		if cachedOTP != req.OTP {
			return apperror.ErrInvalidOTP
		}

		// Check if new password is different from old one
		if password.CheckPasswordHash(req.NewPassword, user.Password) {
			return apperror.NewAppError(apperror.ErrCodeInvalidInput, "New password must be different from the old password")
		}

		// Hash new password
		hashedPassword, err := s.hashPassword(txCtx, req.NewPassword)
		if err != nil {
			return err
		}

		// Update user
		user.Password = hashedPassword
		user.UpdatedAt = time.Now()

		if err := s.updateUser(txCtx, user); err != nil {
			return err
		}

		// Clean up cache
		return s.deleteCachedOTP(txCtx, "password_reset", req.Email)
	})

	if err != nil {
		span.RecordError(err)
		return err
	}

	s.logger.WithContext(ctx).Info("Password reset successfully", "email", req.Email)
	return nil
}

func (s *UserService) Setup2FA(ctx context.Context, req *dto.Setup2FARequest) (*dto.Setup2FAResponse, error) {
	s.logger.WithContext(ctx).Info("Setting up 2FA", "user_id", req.UserID)

	ctx, span := s.tracer.Start(ctx, "service.Setup2FA")
	defer span.End()

	if err := s.validator.ValidateAndGetErrors(req); err != nil {
		span.RecordError(err)
		return nil, apperror.NewValidationError(err)
	}

	var user *entity.User
	err := s.db.WithTransaction(ctx, func(txCtx context.Context) error {
		var err error
		user, err = s.userRepo.GetByID(txCtx, req.UserID)
		if err != nil {
			s.logger.WithContext(txCtx).Error("Failed to get user", "error", err)
			return apperror.ErrUserNotFound
		}

		if !user.IsEmailVerified() {
			return apperror.ErrUserEmailNotVerified
		}

		if !user.IsActive() {
			return apperror.ErrUserInactive
		}

		if user.IsTwoFactorEnabled() {
			return apperror.NewAppError(apperror.ErrCodeBadRequest, "2FA already enabled")
		}

		key, err := s.generate2FASecret(txCtx, user.Email)
		if err != nil {
			s.logger.WithContext(txCtx).Error("Failed to generate 2FA secret", "error", err)
			return apperror.NewAppError(apperror.ErrCodeInternalError, "Failed to generate 2FA secret")
		}

		user.SetTwoFactorSecret(key.Secret())
		user.EnableTwoFactor()

		if err := s.userRepo.Update(txCtx, user); err != nil {
			s.logger.WithContext(txCtx).Error("Failed to update user", "error", err)
			return err
		}

		return nil
	})
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	issuer := url.QueryEscape(s.cfg.AppName)
	accountName := url.QueryEscape(user.Email)
	secret := user.GetTwoFactorSecret()

	otpAuthURL := fmt.Sprintf("otpauth://totp/%s:%s?secret=%s&issuer=%s&algorithm=SHA1&digits=6&period=30",
		issuer, accountName, secret, issuer)

	qrCodeBase64, err := s.generateQRCode(ctx, otpAuthURL)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	s.logger.WithContext(ctx).Info("2FA setup successfully", "user_id", req.UserID)
	return &dto.Setup2FAResponse{
		Secret: user.GetTwoFactorSecret(),
		QRCode: qrCodeBase64,
	}, nil
}

func (s *UserService) Verify2FA(ctx context.Context, req *dto.Verify2FARequest) (*dto.Verify2FAResponse, error) {
	s.logger.WithContext(ctx).Info("Verifying 2FA code", "user_id", req.UserID)

	ctx, span := s.tracer.Start(ctx, "service.Verify2FA")
	defer span.End()

	if err := s.checkRateLimit(ctx, "verify_2fa", fmt.Sprintf("verify_2fa:%s", req.UserID)); err != nil {
		span.RecordError(err)
		return nil, err
	}

	if err := s.validator.ValidateAndGetErrors(req); err != nil {
		span.RecordError(err)
		return nil, apperror.NewValidationError(err)
	}

	var user *entity.User
	var accessToken, refreshToken *token.GenerateTokenResponse
	err := s.db.WithTransaction(ctx, func(txCtx context.Context) error {
		var err error
		user, err = s.userRepo.GetByID(txCtx, req.UserID)
		if err != nil {
			s.logger.WithContext(txCtx).Error("Failed to get user", "error", err)
			return apperror.ErrUserNotFound
		}

		if !user.IsEmailVerified() {
			return apperror.ErrUserEmailNotVerified
		}
		if !user.IsActive() {
			return apperror.ErrUserInactive
		}
		if !user.HasTwoFactorSecret() {
			return apperror.NewAppError(apperror.ErrCodeBadRequest, "2FA not set up")
		}

		if !totp.Validate(req.Code, user.GetTwoFactorSecret()) {
			s.logger.WithContext(txCtx).Warn("Invalid 2FA code", "user_id", req.UserID)
			return apperror.NewAppError(apperror.ErrCodeInvalidInput, "Invalid 2FA code")
		}

		if !user.IsTwoFactorEnabled() {
			user.EnableTwoFactor()
			if err := s.userRepo.Update(txCtx, user); err != nil {
				s.logger.WithContext(txCtx).Error("Failed to update user", "error", err)
				return err
			}
		}

		accessToken, err = s.token.GenerateToken(user.ID, s.cfg.AuthAccessTokenExpiry)
		if err != nil {
			s.logger.WithContext(txCtx).Error("Failed to generate access token", "error", err)
			return apperror.NewAppError(apperror.ErrCodeInternalError, "Failed to generate access token")
		}
		refreshToken, err = s.token.GenerateToken(user.ID, s.cfg.AuthRefreshTokenExpiry)
		if err != nil {
			s.logger.WithContext(txCtx).Error("Failed to generate refresh token", "error", err)
			return apperror.NewAppError(apperror.ErrCodeInternalError, "Failed to generate refresh token")
		}
		return nil
	})
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	s.logger.WithContext(ctx).Info("2FA verified successfully", "user_id", req.UserID)
	return &dto.Verify2FAResponse{
		Verified:              true,
		AccessToken:           accessToken.Value,
		RefreshToken:          refreshToken.Value,
		AccessTokenExpiresAt:  accessToken.ExpiresAt,
		RefreshTokenExpiresAt: refreshToken.ExpiresAt,
	}, nil
}

func (s *UserService) Disable2FA(ctx context.Context, req *dto.Disable2FARequest) error {
	s.logger.WithContext(ctx).Info("Disabling 2FA", "user_id", req.UserID)

	ctx, span := s.tracer.Start(ctx, "service.Disable2FA")
	defer span.End()

	if err := s.validator.ValidateAndGetErrors(req); err != nil {
		span.RecordError(err)
		return apperror.NewValidationError(err)
	}

	err := s.db.WithTransaction(ctx, func(txCtx context.Context) error {
		user, err := s.userRepo.GetByID(txCtx, req.UserID)
		if err != nil {
			s.logger.WithContext(txCtx).Error("Failed to get user", "error", err)
			return apperror.ErrUserNotFound
		}

		if !user.IsEmailVerified() {
			return apperror.ErrUserEmailNotVerified
		}
		if !user.IsActive() {
			return apperror.ErrUserInactive
		}
		if !user.IsTwoFactorEnabled() {
			return apperror.NewAppError(apperror.ErrCodeBadRequest, "2FA not enabled")
		}

		user.DisableTwoFactor()
		if err := s.userRepo.Update(txCtx, user); err != nil {
			s.logger.WithContext(txCtx).Error("Failed to update user", "error", err)
			return err
		}
		return nil
	})
	if err != nil {
		span.RecordError(err)
		return err
	}

	s.logger.WithContext(ctx).Info("2FA disabled successfully", "user_id", req.UserID)
	return nil
}
