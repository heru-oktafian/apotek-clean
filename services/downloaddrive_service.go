package services

import (
	context "context"
	fmt "fmt"
	io "io"
	log "log"
	os "os"
	filepath "path/filepath"

	google "golang.org/x/oauth2/google"
	drive "google.golang.org/api/drive/v3"
	option "google.golang.org/api/option"
)

// DownloadFileFromGoogleDrive mengunduh file dari Google Drive dan menyimpannya ke folder rest/
func DownloadFileFromGoogleDrive(fileID string) (string, error) {
	ctx := context.Background()

	// Baca credentials.json
	b, err := os.ReadFile("credentials.json")
	if err != nil {
		return "", fmt.Errorf("gagal membaca credentials.json: %v", err)
	}

	// Gunakan scope Drive penuh agar bisa akses semua file
	config, err := google.ConfigFromJSON(b, drive.DriveScope)
	if err != nil {
		return "", fmt.Errorf("gagal memproses credentials.json: %v", err)
	}

	client := getClient(config)

	srv, err := drive.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return "", fmt.Errorf("gagal membuat service Google Drive: %v", err)
	}

	// Ambil metadata file
	f, err := srv.Files.Get(fileID).Fields("name").Do()
	if err != nil {
		return "", fmt.Errorf("gagal mengambil metadata file: %v", err)
	}

	// Pastikan folder rest/ ada
	if err := os.MkdirAll("rest", os.ModePerm); err != nil {
		return "", fmt.Errorf("gagal membuat folder rest/: %v", err)
	}

	// Path penyimpanan file
	savePath := filepath.Join("rest", f.Name)

	// Buat file lokal
	outFile, err := os.Create(savePath)
	if err != nil {
		return "", fmt.Errorf("gagal membuat file lokal: %v", err)
	}
	defer outFile.Close()

	// Download dari Drive
	resp, err := srv.Files.Get(fileID).Download()
	if err != nil {
		return "", fmt.Errorf("gagal mengunduh file dari Google Drive: %v", err)
	}
	defer resp.Body.Close()

	// Simpan file ke lokal
	if _, err := io.Copy(outFile, resp.Body); err != nil {
		return "", fmt.Errorf("gagal menyimpan file: %v", err)
	}

	log.Printf("✅ File berhasil diunduh ke: %s\n", savePath)
	return savePath, nil
}
