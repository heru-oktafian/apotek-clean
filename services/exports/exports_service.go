package services

import "gorm.io/gorm"

type ExportServices struct {
	db *gorm.DB
}

func NewExcelServices(db *gorm.DB) *ExportServices {
	return &ExportServices{db: db}
}

func NewPDFService(db *gorm.DB) *ExportServices {
	return &ExportServices{db: db}
}
