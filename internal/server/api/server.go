package api

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/adaptor/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/etag"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sammidev/goca/internal/config"
	"github.com/sammidev/goca/internal/pkg/logger"
	"github.com/sammidev/goca/internal/pkg/observability"
	"github.com/sammidev/goca/internal/pkg/token"
	"github.com/sammidev/goca/internal/server/api/middleware"
	fiberSwagger "github.com/swaggo/fiber-swagger"
)

type Server struct {
	app         *fiber.App
	cfg         *config.Config
	logger      logger.Logger
	token       token.Token
	userHandler UserHandler
	noteHandler NoteHandler
}

// NewServer creates a new HTTP server with all middleware and routes configured
func NewServer(
	cfg *config.Config,
	logger logger.Logger,
	token token.Token,
	userHandler UserHandler,
	noteHandler NoteHandler,
) (*Server, error) {
	app := fiber.New(fiber.Config{
		AppName:       cfg.AppName,
		ReadTimeout:   cfg.ServerReadTimeout,
		WriteTimeout:  cfg.ServerWriteTimeout,
		IdleTimeout:   cfg.ServerIdleTimeout,
		StrictRouting: false,
	})

	s := &Server{
		app:         app,
		cfg:         cfg,
		logger:      logger,
		token:       token,
		userHandler: userHandler,
		noteHandler: noteHandler,
	}

	if err := s.setupMiddleware(); err != nil {
		return nil, fmt.Errorf("failed to setup middleware: %w", err)
	}

	s.setupRoutes()
	return s, nil
}

func (s *Server) setupMiddleware() error {
	obs, err := observability.NewObservability(s.cfg.AppName)
	if err != nil {
		return err
	}

	// Order matters for middleware
	s.app.Use(observability.TracingMiddleware(s.cfg.AppName))
	s.app.Use(recover.New())
	s.app.Use(obs.MetricsMiddleware())
	s.app.Use(middleware.CORSMiddleware())
	s.app.Use(compress.New())
	s.app.Use(etag.New())
	s.app.Use(middleware.RequestIDMiddleware())
	s.app.Use(middleware.LoggerMiddleware(s.logger))

	return nil
}

func (s *Server) setupRoutes() {
	// Health check
	s.app.Get("/health", s.healthCheck)

	// Documentation and monitoring
	s.app.Get("/swagger/*", fiberSwagger.WrapHandler)
	s.app.Get("/metrics", adaptor.HTTPHandler(promhttp.Handler()))

	// API routes
	s.registerAPIRoutes()
}

func (s *Server) healthCheck(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status":    "ok",
		"service":   s.cfg.AppName,
		"version":   "1.0.0",
		"timestamp": time.Now(),
	})
}

func (s *Server) Start() error {
	addr := fmt.Sprintf("%s:%d", s.cfg.ServerHost, s.cfg.ServerPort)
	s.logger.Info("Server starting on %s", addr)
	return s.app.Listen(addr)
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("Shutting down server gracefully...")

	shutdownCtx, cancel := context.WithTimeout(ctx, s.cfg.ServerShutdownTimeout)
	defer cancel()

	if err := s.app.ShutdownWithContext(shutdownCtx); err != nil {
		return fmt.Errorf("failed to shutdown server: %w", err)
	}

	s.logger.Info("Server shutdown completed")
	return nil
}

func (s *Server) GetFiberApp() *fiber.App {
	return s.app
}
