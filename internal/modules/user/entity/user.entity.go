package entity

import (
	"time"

	"github.com/google/uuid"
)

type UserStatus string

const (
	UserStatusActive    UserStatus = "active"
	UserStatusInactive  UserStatus = "inactive"
	UserStatusPending   UserStatus = "pending"
	UserStatusSuspended UserStatus = "suspended"
)

type User struct {
	ID               uuid.UUID  `db:"id"`
	Email            string     `db:"email"`
	EmailVerifiedAt  *time.Time `db:"email_verified_at"`
	FirstName        string     `db:"first_name"`
	LastName         *string    `db:"last_name"`
	FullName         string     `db:"full_name"`
	Password         string     `db:"password"`
	Status           UserStatus `db:"status"`
	TwoFactorSecret  *string    `db:"two_factor_secret"`
	TwoFactorEnabled bool       `db:"two_factor_enabled"`
	CreatedAt        time.Time  `db:"created_at"`
	UpdatedAt        time.Time  `db:"updated_at"`
}

func (u *User) GenerateFullName() {
	if u.LastName == nil {
		u.FullName = u.FirstName
	}
	u.FullName = u.FirstName + " " + *u.LastName
}

func (u *User) IsActive() bool {
	return u.Status == UserStatusActive
}

func (u *User) IsSuspended() bool {
	return u.Status == UserStatusSuspended
}

func (u *User) IsPending() bool {
	return u.Status == UserStatusPending
}

func (u *User) IsInactive() bool {
	return u.Status == UserStatusInactive
}

func (u *User) Suspend() {
	u.Status = UserStatusSuspended
}

func (u *User) Activate() {
	u.SetStatus(UserStatusActive)
}

func (u *User) IsEmailVerified() bool {
	return u.EmailVerifiedAt != nil
}

func (u *User) IsTwoFactorEnabled() bool {
	return u.TwoFactorEnabled && u.TwoFactorSecret != nil
}

func (u *User) SetStatus(status UserStatus) {
	u.Status = status
}

func (u *User) SetTwoFactorSecret(secret string) {
	u.TwoFactorSecret = &secret
}

func (u *User) GetTwoFactorSecret() string {
	if u.TwoFactorSecret == nil {
		return ""
	}
	return *u.TwoFactorSecret
}

func (u *User) HasTwoFactorSecret() bool {
	return u.TwoFactorSecret != nil
}

func (u *User) DisableTwoFactor() {
	u.TwoFactorEnabled = false
	u.TwoFactorSecret = nil
}

func (u *User) EnableTwoFactor() {
	u.TwoFactorEnabled = true
}

func (u *User) VerifyEmail(t time.Time) {
	u.EmailVerifiedAt = &t
}
