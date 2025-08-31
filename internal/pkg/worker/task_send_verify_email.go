package worker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/sammidev/goca/internal/pkg/assets"
)

const (
	TaskSendVerifyEmailMaxRetry = 3
	TaskSendVerifyEmail         = "task:send_verify_email"
	TaskSendVerifyEmailSubject  = "Verifikasi Email"
)

type PayloadSendVerifyEmail struct {
	UserID                     uuid.UUID `json:"user_id"`
	Name                       string    `json:"name"`
	Email                      string    `json:"email"`
	VerificationCode           string    `json:"verification_code"`
	VerificationCodeExpiration int       `json:"verification_code_expiration"` // in minutes

	// fill by distributor
	From    string `json:"from"`
	Subject string `json:"subject"`
}

func (d *RedisTaskDistributor) DistributeTaskSendVerifyEmail(
	ctx context.Context,
	payload *PayloadSendVerifyEmail,
	opts ...asynq.Option,
) error {
	payload.Subject = TaskSendVerifyEmailSubject
	payload.From = d.cfg.AppName

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal task payload: %w", err)
	}

	task := asynq.NewTask(TaskSendVerifyEmail, jsonPayload, opts...)

	_, err = d.client.EnqueueContext(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to enqueue task: %w", err)
	}

	return nil
}

func (p *RedisTaskProcessor) ProcessTaskSendVerifyEmail(ctx context.Context, task *asynq.Task) error {
	var payload PayloadSendVerifyEmail
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		p.logger.Error("failed to unmarshal payload", "error", err)
		return fmt.Errorf("failed to unmarshal payload: %w", asynq.SkipRetry)
	}

	var tpl *template.Template
	tpl, err := template.ParseFS(assets.EmbeddedFiles, assets.EmailVerificationTemplatePath)
	if err != nil {
		p.logger.Error("failed to parse email template", "error", err)
		return fmt.Errorf("failed to parse email template: %w", err)
	}

	var body bytes.Buffer
	if err := tpl.ExecuteTemplate(&body, "htmlBody", payload); err != nil {
		p.logger.Error("failed to execute email template", "error", err)
		return fmt.Errorf("failed to execute email template: %w", err)
	}

	htmlContent := body.String()

	subject := payload.Subject

	err = p.email.Send(payload.Email, subject, htmlContent, payload)
	if err != nil {
		p.logger.Error("failed to send verify email", "error", err)
		return fmt.Errorf("failed to send verify email: %w", err)
	}

	p.logger.Info("verify email sent", "email", payload.Email)

	return nil
}
