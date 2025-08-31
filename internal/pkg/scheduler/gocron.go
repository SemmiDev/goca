package scheduler

import (
	"time"

	"github.com/go-co-op/gocron"
	"github.com/sammidev/goca/internal/pkg/logger"
)

// GoCronScheduler adalah implementasi konkret dari interfaces Scheduler.
// Nama struct tidak di-export agar pengguna fokus pada interfaces.
type GoCronScheduler struct {
	scheduler *gocron.Scheduler
	logger    logger.Logger
}

// Memastikan GoCronScheduler mengimplementasikan interfaces Scheduler.
var _ Scheduler = (*GoCronScheduler)(nil)

// New membuat instance baru dari scheduler.
// Perhatikan, fungsi ini tidak lagi otomatis menjalankan scheduler.
func New(logger logger.Logger) (Scheduler, error) {
	asiaJakartaTime, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		return nil, err
	}

	s := gocron.NewScheduler(asiaJakartaTime)

	return &GoCronScheduler{
		scheduler: s,
		logger:    logger.WithComponent("gocron"),
	}, nil
}

// RegisterJob mendaftarkan tugas baru ke scheduler.
// `schedule` bisa berupa cron expression ("* * * * *") atau interval ("1m", "5s").
func (gs *GoCronScheduler) RegisterJob(schedule string, jobFunc func()) error {
	_, err := gs.scheduler.Every(schedule).Do(func() {
		// Kita bisa tambahkan logging atau error handling di sini
		gs.logger.Info("Running scheduled job", "schedule", schedule)
		jobFunc()
	})
	if err != nil {
		gs.logger.Error("Failed to register job", "schedule", schedule, "error", err)
		return err
	}
	gs.logger.Info("Successfully registered job", "schedule", schedule)
	return nil
}

// Start memulai scheduler secara asynchronous.
func (gs *GoCronScheduler) Start() {
	gs.logger.Info("Starting scheduler...")
	gs.scheduler.StartAsync()
}

// Stop menghentikan semua job yang sedang berjalan.
// Penting untuk graceful shutdown.
func (gs *GoCronScheduler) Stop() {
	gs.logger.Info("Stopping scheduler...")
	gs.scheduler.Stop()
	gs.logger.Info("Scheduler stopped.")
}
