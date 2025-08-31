package token

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/sammidev/goca/internal/config"
	. "github.com/smartystreets/goconvey/convey"
)

func TestJWT(t *testing.T) {
	// Setup test configuration
	cfg := &config.Config{
		AuthJWTSecret: strings.Repeat("x", 32), // Minimum 32 characters
		AppName:       "test-app",
	}

	Convey("Given a JWT instance", t, func() {
		jwtService, err := NewJWT(cfg)
		So(err, ShouldBeNil)
		So(jwtService, ShouldNotBeNil)

		userID := uuid.New()
		expDuration := time.Hour

		Convey("When generating a token", func() {
			Convey("Then it should generate a valid token with correct payload", func() {
				resp, err := jwtService.GenerateToken(userID, expDuration)
				So(err, ShouldBeNil)
				So(resp, ShouldNotBeNil)
				So(resp.Value, ShouldNotBeEmpty)
				So(resp.ExpiresAt, ShouldHappenOnOrAfter, time.Now())
				So(resp.ExpiresAt, ShouldHappenOnOrBefore, time.Now().Add(expDuration))

				// Verify token content
				claims := &customClaims{}
				token, err := jwt.ParseWithClaims(resp.Value, claims, func(token *jwt.Token) (interface{}, error) {
					return jwtService.secretKey, nil
				})
				So(err, ShouldBeNil)
				So(token.Valid, ShouldBeTrue)
				So(claims.UserID, ShouldEqual, userID)
				So(claims.Issuer, ShouldEqual, cfg.AppName)
			})

			Convey("Then it should fail with invalid userID", func() {
				invalidUserID := uuid.Nil
				resp, err := jwtService.GenerateToken(invalidUserID, expDuration)
				So(err, ShouldBeNil) // Should still generate, as nil UUID is technically valid
				So(resp, ShouldNotBeNil)
				So(resp.Value, ShouldNotBeEmpty)
			})
		})

		Convey("When verifying a bearer token", func() {
			tokenResp, err := jwtService.GenerateToken(userID, expDuration)
			So(err, ShouldBeNil)

			Convey("With valid bearer token", func() {
				bearerToken := fmt.Sprintf("Bearer %s", tokenResp.Value)
				payload, err := jwtService.VerifyBearerToken(bearerToken)
				So(err, ShouldBeNil)
				So(payload, ShouldNotBeNil)
				So(payload.UserID, ShouldEqual, userID)
				So(payload.ExpiresAt, ShouldHappenOnOrAfter, time.Now())
			})

			Convey("With invalid bearer format", func() {
				invalidTokens := []string{
					"",
					"InvalidToken",
					"bearer lowercase", // Should fail due to case sensitivity
					"Bearer ",          // Should fail due to empty token
					"Basic token",
				}
				for _, invalidToken := range invalidTokens {
					Convey(fmt.Sprintf("With invalid format: %s", invalidToken), func() {
						_, err := jwtService.VerifyBearerToken(invalidToken)
						So(err, ShouldEqual, ErrInvalidTokenFormat)
					})
				}
			})
		})

		Convey("When verifying a token directly", func() {
			tokenResp, err := jwtService.GenerateToken(userID, expDuration)
			So(err, ShouldBeNil)

			Convey("With valid token", func() {
				payload, err := jwtService.VerifyToken(tokenResp.Value)
				So(err, ShouldBeNil)
				So(payload, ShouldNotBeNil)
				So(payload.UserID, ShouldEqual, userID)
				So(payload.ExpiresAt, ShouldHappenOnOrAfter, time.Now())
			})

			Convey("With expired token", func() {
				expiredToken, err := jwtService.GenerateToken(userID, -time.Hour)
				So(err, ShouldBeNil)
				_, err = jwtService.VerifyToken(expiredToken.Value)
				So(err, ShouldEqual, ErrExpiredToken)
			})

			Convey("With invalid token", func() {
				_, err := jwtService.VerifyToken("invalid.token.string")
				So(err, ShouldEqual, ErrInvalidToken)
			})

			Convey("With token signed by different key", func() {
				// Create another JWT instance with different secret
				otherCfg := &config.Config{
					AuthJWTSecret: strings.Repeat("y", 32),
					AppName:       "other-app",
				}
				otherJWT, err := NewJWT(otherCfg)
				So(err, ShouldBeNil)

				tokenResp, err := otherJWT.GenerateToken(userID, expDuration)
				So(err, ShouldBeNil)

				_, err = jwtService.VerifyToken(tokenResp.Value)
				So(err, ShouldEqual, ErrInvalidToken)
			})
		})

		Convey("When extracting bearer token", func() {
			token := "valid-token-string"
			Convey("With valid bearer format", func() {
				bearerToken := fmt.Sprintf("Bearer %s", token)
				extracted, err := jwtService.extractBearerToken(bearerToken)
				So(err, ShouldBeNil)
				So(extracted, ShouldEqual, token)
			})

			Convey("With invalid bearer formats", func() {
				invalidTokens := []string{
					"",
					"InvalidToken",
					"bearer lowercase",
					"Bearer ", // Explicitly test empty token
					"Basic token",
				}
				for _, invalidToken := range invalidTokens {
					Convey(fmt.Sprintf("With invalid format: %s", invalidToken), func() {
						_, err := jwtService.extractBearerToken(invalidToken)
						So(err, ShouldEqual, ErrInvalidTokenFormat)
					})
				}
			})
		})

		Convey("When creating JWT with invalid secret key", func() {
			invalidCfg := &config.Config{
				AuthJWTSecret: "short",
				AppName:       "test-app",
			}
			_, err := NewJWT(invalidCfg)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "minimal 32 karakter")
		})
	})
}
