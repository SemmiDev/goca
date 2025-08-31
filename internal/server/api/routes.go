package api

import (
	"github.com/sammidev/goca/internal/server/api/middleware"
)

func (s *Server) registerAPIRoutes() {
	api := s.app.Group("/api/v1")

	// User routes
	auth := api.Group("/auth")
	auth.Post("/register", s.userHandler.Register)
	auth.Post("/login", s.userHandler.Login)
	auth.Post("/verify-otp", s.userHandler.VerifyOTP)
	auth.Post("/resend-otp", s.userHandler.ResendOTP)
	auth.Post("/refresh-token", s.userHandler.RefreshToken)
	auth.Post("/forgot-password", s.userHandler.ForgotPassword)
	auth.Post("/reset-password", s.userHandler.ResetPassword)
	auth.Post("/setup-2fa", middleware.AuthMiddleware(s.token), s.userHandler.Setup2FA)
	auth.Post("/verify-2fa", middleware.AuthMiddleware(s.token), s.userHandler.Verify2FA)
	auth.Post("/disable-2fa", middleware.AuthMiddleware(s.token), s.userHandler.Disable2FA)

	// Protected routes
	protected := api.Use(middleware.AuthMiddleware(s.token))

	// Note routes
	notes := protected.Group("/notes")
	notes.Post("/", s.noteHandler.CreateNote)
	notes.Get("/", s.noteHandler.GetNotes)
	notes.Get("/:id", s.noteHandler.GetNote)
	notes.Put("/:id", s.noteHandler.UpdateNote)
	notes.Delete("/:id", s.noteHandler.DeleteNote)
}
