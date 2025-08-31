package worker

import (
	"context"

	"github.com/hibiken/asynq"
)

type TaskDistributor interface {
	DistributeTaskHello(
		ctx context.Context,
		payload *PayloadHello,
		opts ...asynq.Option,
	) error

	DistributeTaskSendVerifyEmail(
		ctx context.Context,
		payload *PayloadSendVerifyEmail,
		opts ...asynq.Option,
	) error

	DistributeTaskSendForgotPasswordEmail(
		ctx context.Context,
		payload *PayloadSendForgotPasswordEmail,
		opts ...asynq.Option,
	) error
}
