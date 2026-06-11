package services

import (
	fmt "fmt"
	os "os"
	filepath "path/filepath"
	strings "strings"
)

func ResolveProjectRoot() (string, error) {
	candidates := make([]string, 0, 4)

	if wd, err := os.Getwd(); err == nil {
		candidates = append(candidates, wd)
	}

	if exePath, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exePath)
		candidates = append(candidates, exeDir, filepath.Join(exeDir, ".."))
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
		if err != nil || !info.IsDir() {
			continue
		}
		if _, err := os.Stat(filepath.Join(absPath, "go.mod")); err == nil {
			return absPath, nil
		}
	}

	return "", fmt.Errorf("project root tidak ditemukan; checked: %s", strings.Join(checked, ", "))
}

func ResolveProjectFilePath(name string) (string, error) {
	if strings.TrimSpace(name) == "" {
		return "", fmt.Errorf("nama file tidak boleh kosong")
	}

	root, err := ResolveProjectRoot()
	if err != nil {
		return "", err
	}

	return filepath.Join(root, name), nil
}

func ResolveProjectDirPath(name string) (string, error) {
	if strings.TrimSpace(name) == "" {
		return "", fmt.Errorf("nama folder tidak boleh kosong")
	}

	root, err := ResolveProjectRoot()
	if err != nil {
		return "", err
	}

	return filepath.Join(root, name), nil
}
