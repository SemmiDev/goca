package token

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/sammidev/goca/internal/config"
)

type JWT struct {
	secretKey []byte
	issuer    string
}

func NewJWT(cfg *config.Config) (*JWT, error) {
	if len(cfg.AuthJWTSecret) < 32 {
		return nil, fmt.Errorf("secret key JWT harus memiliki panjang minimal 32 karakter")
	}
	return &JWT{
		secretKey: []byte(cfg.AuthJWTSecret),
		issuer:    cfg.AppName,
	}, nil
}

var _ Token = (*JWT)(nil)

// customClaims adalah payload kustom untuk JWT.
// Dibuat unexported karena ini adalah detail implementasi.
type customClaims struct {
	jwt.RegisteredClaims
	UserID uuid.UUID `json:"user_id"`
}

func (j *JWT) GenerateToken(userID uuid.UUID, exp time.Duration) (*GenerateTokenResponse, error) {
	now := time.Now()
	expiration := now.Add(exp)

	claims := &customClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        userID.String(),
			Issuer:    j.issuer,
			ExpiresAt: jwt.NewNumericDate(expiration),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(j.secretKey)
	if err != nil {
		return nil, err
	}

	return &GenerateTokenResponse{
		Value:     signedToken,
		ExpiresAt: expiration,
	}, nil
}

func (j *JWT) VerifyBearerToken(bearerToken string) (*Payload, error) {
	tokenStr, err := j.extractBearerToken(bearerToken)
	if err != nil {
		return nil, err
	}
	return j.VerifyToken(tokenStr)
}

func (j *JWT) VerifyToken(tokenStr string) (*Payload, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &customClaims{}, func(token *jwt.Token) (any, error) {
		// Memastikan algoritma penandatanganan adalah HMAC
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("algoritma signing tidak terduga: %v", token.Header["alg"])
		}
		return j.secretKey, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*customClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return &Payload{
		UserID:    claims.UserID,
		ExpiresAt: claims.ExpiresAt.Time,
	}, nil
}

func (j *JWT) extractBearerToken(authHeader string) (string, error) {
	if authHeader == "" {
		return "", ErrInvalidTokenFormat
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", ErrInvalidTokenFormat
	}
	if parts[1] == "" {
		return "", ErrInvalidTokenFormat
	}
	return parts[1], nil
}
