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
	TaskSendForgotPasswordEmailMaxRetry = 3
	TaskSendForgotPasswordEmail         = "task:send_forgot_password_email"
	TaskSendForgotPasswordEmailSubject  = "Permintaan Reset Password"
)

type PayloadSendForgotPasswordEmail struct {
	UserID                     uuid.UUID `json:"user_id"`
	Name                       string    `json:"name"`
	Email                      string    `json:"email"`
	VerificationCode           string    `json:"verification_code"`
	VerificationCodeExpiration int       `json:"verification_code_expiration"` // in minutes

	// fill by distributor
	From      string `json:"from"`
	Subject   string `json:"subject"`
	ResetLink string `json:"reset_link"`
}

func (d *RedisTaskDistributor) DistributeTaskSendForgotPasswordEmail(
	ctx context.Context,
	payload *PayloadSendForgotPasswordEmail,
	opts ...asynq.Option,
) error {
	payload.Subject = TaskSendForgotPasswordEmailSubject
	payload.From = d.cfg.AppName
	payload.ResetLink = fmt.Sprintf("%s/reset-password/%s", d.cfg.AppFrontendURL, payload.VerificationCode)

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal task payload: %w", err)
	}

	task := asynq.NewTask(TaskSendForgotPasswordEmail, jsonPayload, opts...)

	_, err = d.client.EnqueueContext(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to enqueue task: %w", err)
	}

	return nil
}

func (p *RedisTaskProcessor) ProcessTaskSendForgotPasswordEmail(ctx context.Context, task *asynq.Task) error {
	var payload PayloadSendForgotPasswordEmail
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		p.logger.Error("failed to unmarshal payload", "error", err)
		return fmt.Errorf("failed to unmarshal payload: %w", asynq.SkipRetry)
	}

	var tpl *template.Template
	tpl, err := template.ParseFS(assets.EmbeddedFiles, assets.EmailForgotPasswordTemplatePath)
	if err != nil {
		p.logger.Error("failed to parse forgot password email template", "error", err)
		return fmt.Errorf("failed to parse forgot password email template: %w", err)
	}

	var body bytes.Buffer
	if err := tpl.ExecuteTemplate(&body, "htmlBody", payload); err != nil {
		p.logger.Error("failed to execute forgot password email template", "error", err)
		return fmt.Errorf("failed to execute forgot password email template: %w", err)
	}

	htmlContent := body.String()

	subject := payload.Subject

	err = p.email.Send(payload.Email, subject, htmlContent, payload)
	if err != nil {
		p.logger.Error("failed to send forgot password email", "error", err)
		return fmt.Errorf("failed to send forgot password email: %w", err)
	}

	p.logger.Info("forgot password email sent", "email", payload.Email)

	return nil
}
