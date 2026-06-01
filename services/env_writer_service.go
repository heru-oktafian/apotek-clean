package services

import (
	fmt "fmt"
	os "os"
)

// WriteRawEnvFile menulis string mentah ke file .env
func WriteRawEnvFile(filePath string, content string) error {
	// Tulis ulang file (overwrite)
	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("gagal menulis file .env: %v", err)
	}
	return nil
}
