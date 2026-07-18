package services

import "gorm.io/gorm"

type ExportServices struct {
	db *gorm.DB
}

// NewExportService creates an ExportServices instance.
// The same struct is used for both Excel and PDF exports; the difference is
// which export method is called on the returned instance (ExportXxxToExcel vs ExportXxxToPDF).
func NewExportService(db *gorm.DB) *ExportServices {
	return &ExportServices{db: db}
}
