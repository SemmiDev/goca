package worker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
)

const (
	TaskHelloMaxRetry = 3
	TaskHello         = "task:hello"
)

type PayloadHello struct {
	Name string `json:"name"`
}

func (d *RedisTaskDistributor) DistributeTaskHello(ctx context.Context, payload *PayloadHello, opts ...asynq.Option) error {
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal task payload: %w", err)
	}

	task := asynq.NewTask(TaskHello, jsonPayload, opts...)

	_, err = d.client.EnqueueContext(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to enqueue task: %w", err)
	}

	return nil
}

func (p *RedisTaskProcessor) ProcessTaskHello(ctx context.Context, task *asynq.Task) error {
	var payload PayloadHello
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", asynq.SkipRetry)
	}

	p.logger.Info("hello", "name", payload.Name)

	return nil
}
