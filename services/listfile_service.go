package services

import (
	fmt "fmt"
	os "os"
	sort "sort"
	time "time"
)

// FileInfo adalah struct untuk menyimpan informasi file
type FileInfo struct {
	Name    string    `json:"name"`
	Size    int64     `json:"size"`
	ModTime time.Time `json:"mod_time"`
}

// ListFilesFromFolder menampilkan file di folder tertentu
func ListFilesFromFolder(folderPath string) ([]FileInfo, error) {
	var files []FileInfo

	// Pastikan folder ada
	if _, err := os.Stat(folderPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("folder %s tidak ditemukan", folderPath)
	}

	// Baca isi folder
	entries, err := os.ReadDir(folderPath)
	if err != nil {
		return nil, fmt.Errorf("gagal membaca folder %s: %v", folderPath, err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			info, err := entry.Info()
			if err != nil {
				continue
			}
			files = append(files, FileInfo{
				Name:    info.Name(),
				Size:    info.Size(),
				ModTime: info.ModTime(),
			})
		}
	}

	// Urutkan dari terbaru ke terlama
	sort.Slice(files, func(i, j int) bool {
		return files[i].ModTime.After(files[j].ModTime)
	})

	return files, nil
}

// ListDumpFiles menampilkan file dari folder dump/
func ListDumpFiles() ([]FileInfo, error) {
	return ListFilesFromFolder("dump")
}

// ListRestFiles menampilkan file dari folder rest/
func ListRestFiles() ([]FileInfo, error) {
	return ListFilesFromFolder("rest")
}
