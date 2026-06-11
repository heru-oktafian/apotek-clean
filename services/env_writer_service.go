package services

import (
	fmt "fmt"
	os "os"
	filepath "path/filepath"
	strings "strings"
)

func ResolveEnvFilePath() (string, error) {
	candidates := []string{".env"}

	if wd, err := os.Getwd(); err == nil {
		candidates = append(candidates, filepath.Join(wd, ".env"))
	}

	if exePath, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exePath)
		candidates = append(candidates,
			filepath.Join(exeDir, ".env"),
			filepath.Join(exeDir, "..", ".env"),
		)
	}

	seen := make(map[string]struct{})
	checked := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		absPath, err := filepath.Abs(candidate)
		if err != nil {
			continue
		}
		if _, ok := seen[absPath]; ok {
			continue
		}
		seen[absPath] = struct{}{}
		checked = append(checked, absPath)

		info, err := os.Stat(absPath)
		if err == nil && !info.IsDir() {
			return absPath, nil
		}
	}

	if exePath, err := os.Executable(); err == nil {
		fallback := filepath.Join(filepath.Dir(exePath), "..", ".env")
		absPath, absErr := filepath.Abs(fallback)
		if absErr == nil {
			return absPath, nil
		}
	}

	return "", fmt.Errorf("file .env tidak ditemukan; checked: %s", strings.Join(checked, ", "))
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
