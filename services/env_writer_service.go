package services

import (
	fmt "fmt"
	os "os"
	strings "strings"
)

func ResolveEnvFilePath() (string, error) {
	return ResolveProjectFilePath(".env")
}

// WriteRawEnvFile menulis string mentah ke file .env
func WriteRawEnvFile(filePath string, content string) error {
	resolvedPath := filePath
	if strings.TrimSpace(filePath) == "" || filePath == ".env" {
		var err error
		resolvedPath, err = ResolveEnvFilePath()
		if err != nil {
			return err
		}
	}

	err := os.WriteFile(resolvedPath, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("gagal menulis file .env: %v", err)
	}
	return nil
}
