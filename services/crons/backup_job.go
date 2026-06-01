package crons

import (
	fmt "fmt"
	log "log"
	os "os"
	exec "os/exec"
	filepath "path/filepath"
	time "time"
)

func DBDump() error {
	filename := os.Getenv("PROJECT_NAME") + "_" + time.Now().Format("02-01-2006") + ".sql"
	outputDir := ".backup_db"
	outputPath := filepath.Join(outputDir, filename)

	// 2. Buat folder dump jika belum ada
	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create dump directory: %w", err)
	}

	// 3. Ambil config DB dari environment
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASS")
	dbName := os.Getenv("DB_NAME")

	// 4. Set password untuk pg_dump
	os.Setenv("PGPASSWORD", dbPassword)

	// 5. Jalankan perintah pg_dump
	cmd := exec.Command("/usr/bin/pg_dump",
		"-h", dbHost,
		"-p", dbPort,
		"-U", dbUser,
		"-F", "p",
		"--inserts",       // solusi agar data yang ada tanda ' di-escape dengan benar
		"--encoding=UTF8", // pastikan encoding benar
		"-f", outputPath,
		dbName,
	)
	cmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", dbPassword))

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error running pg_dump: %v - %s", err, string(output))
	}

	log.Printf("✅ Backup berhasil: %s\n", outputPath)
	return nil
}
