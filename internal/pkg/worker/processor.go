package worker

import (
	"context"
	"log/slog"
	"time"

	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
	"github.com/sammidev/goca/internal/pkg/database"
	"github.com/sammidev/goca/internal/pkg/email"
	"github.com/sammidev/goca/internal/pkg/logger"
)

const (
	Low      = "low"
	Default  = "default"
	Critical = "critical"
)

type TaskProcessor interface {
	Start() error
	Shutdown()
	ProcessTaskHello(ctx context.Context, task *asynq.Task) error
	ProcessTaskSendVerifyEmail(ctx context.Context, task *asynq.Task) error
	ProcessTaskSendForgotPasswordEmail(ctx context.Context, task *asynq.Task) error
}

type RedisTaskProcessor struct {
	server *asynq.Server
	logger logger.Logger
	db     database.Database
	email  email.Email
}

var _ TaskProcessor = (*RedisTaskProcessor)(nil)

func NewRedisTaskProcessor(
	db database.Database,
	logger logger.Logger,
	redisOpt asynq.RedisClientOpt,
	email email.Email,
) *RedisTaskProcessor {

	asynqLogger := NewLogger(logger)
	redis.SetLogger(asynqLogger)

	server := asynq.NewServer(
		redisOpt,
		asynq.Config{
			// priority value. Keys are the names of the queues and values are associated priority value.
			Queues: map[string]int{
				Critical: 6,
				Default:  3,
				Low:      1,
			},
			ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
				slog.Error("process task failed", "error", err, "type", task.Type(), "payload", string(task.Payload()))
			}),
			// maximum number of concurrent processing of tasks.
			Concurrency: 100,
			Logger:      asynqLogger,
			// Graceful shutdown timeout
			ShutdownTimeout: 30 * time.Second,
		},
	)

	return &RedisTaskProcessor{
		server: server,
		db:     db,
		logger: logger,
		email:  email,
	}
}

func (p *RedisTaskProcessor) Start() error {
	mux := asynq.NewServeMux()

	mux.HandleFunc(TaskHello, p.ProcessTaskHello)
	mux.HandleFunc(TaskSendVerifyEmail, p.ProcessTaskSendVerifyEmail)
	mux.HandleFunc(TaskSendForgotPasswordEmail, p.ProcessTaskSendForgotPasswordEmail)

	return p.server.Start(mux)
}

func (p *RedisTaskProcessor) Shutdown() {
	p.server.Shutdown()
}
