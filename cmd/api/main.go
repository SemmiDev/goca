package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2/log"
	_ "github.com/sammidev/goca/api/swagger"
	"github.com/sammidev/goca/internal"
	"github.com/sammidev/goca/internal/config"
)

// @title						Notes Taking API
// @version					1.0
// @description				RESTful API for notes taking app
// @termsOfService				http://swagger.io/terms/
// @contact.name				Sammi Aldhi Yanto
// @contact.url				https://lab-sammi.gitbook.io
// @contact.email				sammidev4@gmail.com
// @license.name				Apache 2.0
// @license.url				http://www.apache.org/licenses/LICENSE-2.0.html
// @host						localhost:8080
// @BasePath					/api/v1
// @securityDefinitions.apikey	BearerAuth
// @in							header
// @name						Authorization
// @description				Type "Bearer" followed by a space and JWT token. Example: "Bearer {token}"
func main() {
	if err := run(); err != nil {
		log.Fatalf("Application failed: %v", err)
	}
}

func run() error {
	cfg, err := config.NewConfig()
	if err != nil {
		return err
	}

	// Create application with all dependencies
	application, err := internal.NewApplication(cfg)
	if err != nil {
		return err
	}
	defer application.Close()

	// Set up graceful shutdown
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Start application
	return application.Run(ctx)
}
