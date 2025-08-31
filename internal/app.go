package internal

import (
	"context"
	"errors"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/hibiken/asynq"
	"github.com/ulule/limiter/v3"
	"golang.org/x/sync/errgroup"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	noteHdl "github.com/sammidev/goca/internal/modules/note/handler"
	noteRepo "github.com/sammidev/goca/internal/modules/note/repository"
	noteSvc "github.com/sammidev/goca/internal/modules/note/service"
	userHdl "github.com/sammidev/goca/internal/modules/user/handler"
	userRepo "github.com/sammidev/goca/internal/modules/user/repository"
	userSvc "github.com/sammidev/goca/internal/modules/user/service"

	"github.com/sammidev/goca/internal/config"
	"github.com/sammidev/goca/internal/pkg/cache"
	"github.com/sammidev/goca/internal/pkg/database"
	"github.com/sammidev/goca/internal/pkg/email"
	"github.com/sammidev/goca/internal/pkg/logger"
	"github.com/sammidev/goca/internal/pkg/observability"
	"github.com/sammidev/goca/internal/pkg/ratelimit"
	"github.com/sammidev/goca/internal/pkg/scheduler"
	"github.com/sammidev/goca/internal/pkg/token"
	"github.com/sammidev/goca/internal/pkg/validator"
	"github.com/sammidev/goca/internal/pkg/worker"
	apiServer "github.com/sammidev/goca/internal/server/api"
)

type Application struct {
	cfg           *config.Config
	server        *apiServer.Server
	logger        logger.Logger
	cronScheduler scheduler.Scheduler
	taskProcessor worker.TaskProcessor
	db            database.Database
	cache         cache.Cache
	observability func(context.Context) error
}

// NewApplication creates and initializes the application with all dependencies
func NewApplication(cfg *config.Config) (*Application, error) {
	// Initialize core dependencies
	zapLogger, err := logger.NewZapLogger(cfg)
	if err != nil {
		return nil, err
	}

	postgresDB, err := database.NewPostgreSQLDatabase(cfg, zapLogger)
	if err != nil {
		return nil, err
	}

	redisClient, err := cache.NewRedisClient(cfg)
	if err != nil {
		return nil, err
	}

	// Run database migrations
	if err := runDatabaseMigration(cfg); err != nil {
		return nil, err
	}

	// Setup observability
	shutdown, err := observability.SetupOTelProvider(context.Background(), cfg, zapLogger)
	if err != nil {
		return nil, errors.New("failed to setup OTel meter provider: " + err.Error())
	}

	// Initialize remaining dependencies
	jwtToken, err := token.NewJWT(cfg)
	if err != nil {
		return nil, err
	}

	smtpClient, err := email.NewSMTPClient(cfg, zapLogger)
	if err != nil {
		return nil, err
	}

	authRateLimiter, err := ratelimit.NewRedisRateLimiter(
		redisClient.Client,
		config.AuthRateLimiterKey,
		limiter.Rate{Period: time.Minute, Limit: 100},
		zapLogger,
	)
	if err != nil {
		return nil, err
	}

	asynqRedisOpt := asynq.RedisClientOpt{
		Addr:     cfg.RedisDSN(),
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	}

	taskProcessor := worker.NewRedisTaskProcessor(postgresDB, zapLogger, asynqRedisOpt, smtpClient)
	taskDistributor := worker.NewRedisTaskDistributor(asynqRedisOpt, cfg)
	goPlaygroundValidator := validator.NewGoPlaygroundValidatorWithLocale(validator.Locale(cfg.AppLocale))

	cronScheduler, err := scheduler.New(zapLogger)
	if err != nil {
		return nil, err
	}

	// Initialize server with handlers
	server, err := initializeServer(cfg, zapLogger, postgresDB, redisClient, jwtToken, authRateLimiter, goPlaygroundValidator, taskDistributor)
	if err != nil {
		return nil, err
	}

	return &Application{
		cfg:           cfg,
		server:        server,
		observability: shutdown,
		logger:        zapLogger,
		cronScheduler: cronScheduler,
		taskProcessor: taskProcessor,
		db:            postgresDB,
		cache:         redisClient,
	}, nil
}

// Run starts the application and handles graceful shutdown
func (a *Application) Run(ctx context.Context) error {
	a.logger.Info("Starting application")

	a.cronScheduler.Start()
	defer a.cronScheduler.Stop()

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		a.logger.Info("Starting API server")
		return a.server.Start()
	})

	g.Go(func() error {
		a.logger.Info("Starting task processor")
		return a.taskProcessor.Start()
	})

	go a.handleShutdown(ctx)

	if err := g.Wait(); err != nil {
		a.logger.Error("Service failed: %v", err)
		return err
	}

	a.logger.Info("Application exited gracefully")
	return nil
}

// Close performs cleanup of all resources
func (a *Application) Close() {
	if a.observability != nil {
		a.observability(context.Background())
	}
	if a.db != nil {
		a.db.Close()
	}
	if a.cache != nil {
		a.cache.Close()
	}
}

// handleShutdown handles graceful shutdown of all services
func (a *Application) handleShutdown(ctx context.Context) {
	<-ctx.Done()
	a.logger.Info("Shutdown signal received, initiating graceful shutdown")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), a.cfg.ServerShutdownTimeout)
	defer cancel()

	if err := a.server.Shutdown(shutdownCtx); err != nil {
		a.logger.Error("API server shutdown failed: %v", err)
	}

	a.taskProcessor.Shutdown()
}

// initializeServer creates the API server with all handlers
func initializeServer(
	cfg *config.Config,
	logger logger.Logger,
	db database.Database,
	cache cache.Cache,
	jwtToken token.Token,
	authRateLimit ratelimit.RateLimiter,
	validator validator.Validator,
	taskDistributor worker.TaskDistributor,
) (*apiServer.Server, error) {
	// Initialize user module
	userRepo := userRepo.NewUserPostgresRepository(db.(*database.PostgreSQLDatabase))
	userService := userSvc.NewUserService(
		cfg,
		logger,
		validator,
		db,
		jwtToken,
		cache,
		taskDistributor,
		authRateLimit,
		userRepo,
	)
	userHandler := userHdl.NewUserHandler(userService)

	// Initialize note module
	noteRepo := noteRepo.NewNotePostgresRepository(db.(*database.PostgreSQLDatabase))
	noteService := noteSvc.NewNoteService(
		cfg,
		logger,
		validator,
		db,
		noteRepo,
	)
	noteHandler := noteHdl.NewNoteHandler(noteService)

	server, err := apiServer.NewServer(cfg, logger, jwtToken, userHandler, noteHandler)
	if err != nil {
		return nil, err
	}

	return server, nil
}

// runDatabaseMigration runs database migrations
func runDatabaseMigration(cfg *config.Config) error {
	migration, err := migrate.New("file://migrations", cfg.DSN())
	if err != nil {
		return err
	}
	defer migration.Close()

	if err = migration.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}
