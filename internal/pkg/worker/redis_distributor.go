package worker

import (
	"github.com/hibiken/asynq"
	"github.com/sammidev/goca/internal/config"
)

type RedisTaskDistributor struct {
	client *asynq.Client
	cfg    *config.Config
}

func NewRedisTaskDistributor(redisOpt asynq.RedisClientOpt, cfg *config.Config) *RedisTaskDistributor {
	client := asynq.NewClient(redisOpt)

	return &RedisTaskDistributor{
		client: client,
		cfg:    cfg,
	}
}

var _ TaskDistributor = (*RedisTaskDistributor)(nil)
