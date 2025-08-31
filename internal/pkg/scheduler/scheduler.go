package scheduler

// Scheduler mendefinisikan interfaces untuk penjadwal tugas.
// Ini memungkinkan kita membuat implementasi palsu (mock) untuk testing.
type Scheduler interface {
	RegisterJob(schedule string, jobFunc func()) error
	Start()
	Stop()
}
