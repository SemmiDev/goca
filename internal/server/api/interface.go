package api

import "github.com/gofiber/fiber/v2"

type UserHandler interface {
	Register(c *fiber.Ctx) error
	Login(c *fiber.Ctx) error
	VerifyOTP(c *fiber.Ctx) error
	ResendOTP(c *fiber.Ctx) error
	RefreshToken(c *fiber.Ctx) error
	ForgotPassword(c *fiber.Ctx) error
	ResetPassword(c *fiber.Ctx) error
	Setup2FA(c *fiber.Ctx) error
	Verify2FA(c *fiber.Ctx) error
	Disable2FA(c *fiber.Ctx) error
}

type NoteHandler interface {
	CreateNote(c *fiber.Ctx) error
	GetNote(c *fiber.Ctx) error
	UpdateNote(c *fiber.Ctx) error
	DeleteNote(c *fiber.Ctx) error
	GetNotes(c *fiber.Ctx) error
}
