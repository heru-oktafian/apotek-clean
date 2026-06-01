package crons

import (
	fmt "fmt"
	log "log"
	time "time"

	cron "github.com/robfig/cron/v3"
	gorm "gorm.io/gorm"
)

func SchedulerJobs(db *gorm.DB) (*cron.Cron, error) {
	// Inisialisasi cron dengan lokasi waktu WIB (Asia/Jakarta)
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		return nil, fmt.Errorf("failed to load location: %w", err)
	}

	// Inisialisasi cron
	c := cron.New(cron.WithLocation(loc))

	// Tambahkan job
	c.AddFunc("0 7 * * *", func() {
		// 1. Backup database
		if err := DBDump(); err != nil {
			log.Printf("[SCHEDULER] Error running db dump: %v", err)
		}

		// 2. Hitung asset
		if err := AssetCounter(db); err != nil {
			log.Printf("[SCHEDULER] Error running asset counter: %v", err)
		}
	})

	c.Start()
	log.Println("[SCHEDULER] Semua job terjadwal aktif!")
	return c, nil
}
