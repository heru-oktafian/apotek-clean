package services

import (
	fmt "fmt"
	log "log"
	os "os"
	exec "os/exec"
	filepath "path/filepath"
)

// RestoreDatabaseFromFile me-restore database dari file .sql yang ada di folder rest/
func RestoreDatabaseFromFile(fileName string, targetDB string) error {
	filePath := filepath.Join("rest", fileName)

	// Cek apakah file ada
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("file %s tidak ditemukan di folder rest/", fileName)
	}

	// Ambil konfigurasi koneksi DB dari environment
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASS")

	// Set password agar psql/createdb tidak prompt
	os.Setenv("PGPASSWORD", dbPassword)

	// 1. Buat database baru
	createCmd := exec.Command("createdb",
		"-h", dbHost,
		"-p", dbPort,
		"-U", dbUser,
		targetDB,
	)
	createCmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", dbPassword))

	if output, err := createCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("gagal membuat database: %v - %s", err, string(output))
	}
	log.Printf("✅ Database %s berhasil dibuat\n", targetDB)

	// 2. Restore dari file .sql
	restoreCmd := exec.Command("psql",
		"-h", dbHost,
		"-p", dbPort,
		"-U", dbUser,
		"-d", targetDB,
		"-f", filePath,
	)
	restoreCmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", dbPassword))

	if output, err := restoreCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("gagal restore database: %v - %s", err, string(output))
	}

	log.Printf("✅ Database %s berhasil direstore dari file %s\n", targetDB, fileName)
	return nil
}
