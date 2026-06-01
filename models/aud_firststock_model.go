package models

import time "time"

// FirstStocks model
type FirstStocks struct {
	ID              string        `gorm:"type:varchar(15);primaryKey" json:"id" validate:"required"`
	Description     string        `gorm:"type:text;" json:"description"`
	FirstStockDate  time.Time     `gorm:"not null" json:"first_stock_date" validate:"required"`
	BranchID        string        `gorm:"type:varchar(15);not null" json:"branch_id" validate:"required"`
	TotalFirstStock int           `gorm:"type:int;not null;default:0" json:"total_first_stock" validate:"required"`
	Payment         PaymentStatus `gorm:"type:payment_status;not null;default:'nocost'" json:"payment" validate:"required"`
	UserID          string        `gorm:"type:varchar(15);not null" json:"user_id" validate:"required"`
	CreatedAt       time.Time     `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt       time.Time     `gorm:"autoUpdateTime" json:"updated_at"`
}

// All FirstStocks model
type AllFirstStocks struct {
	ID              string        `gorm:"type:varchar(15);primaryKey" json:"id" validate:"required"`
	Description     string        `gorm:"type:text;" json:"description"`
	FirstStockDate  time.Time     `gorm:"not null" json:"first_stock_date" validate:"required"`
	TotalFirstStock int           `gorm:"type:int;not null;default:0" json:"total_first_stock" validate:"required"`
	Payment         PaymentStatus `gorm:"type:payment_status;not null;default:'nocost'" json:"payment" validate:"required"`
}

// FirstStockInput is the input struct for creating or updating a first stock
type FirstStockInput struct {
	FirstStockDate  string `json:"first_stock_date"`
	Description     string `json:"description"`
	TotalFirstStock int    `json:"total_first_stock"`
	Payment         string `json:"payment"`
}

// --- Structs Permintaan untuk First Stock ---
// --- Structs Permintaan untuk First Stock ---
type FirstStockTransactionRequest struct {
	FirstStock      FirstStockInput       `json:"first_stock"`
	FirstStockItems []FirstStockItemInput `json:"first_stock_items"`
}

// --- Structs Respons untuk First Stock ---
type FirstStockOutput struct {
	ID              string `json:"id"`
	Description     string `json:"description"`
	FirstStockDate  string `json:"first_stock_date"` // Format YYYY-MM-DD
	BranchID        string `json:"branch_id"`
	TotalFirstStock int    `json:"total_first_stock"` // Ini adalah nilai stok yang ditambahkan
	Payment         string `json:"payment"`           // Akan diisi default "unpaid" atau "no_cost"
	UserID          string `json:"user_id"`
	CreatedAt       string `json:"created_at"` // Format YYYY-MM-DD
	UpdatedAt       string `json:"updated_at"` // Format YYYY-MM-DD
}

// Struct untuk respons FirstStockTransaction
type FirstStockTransactionResponse struct {
	FirstStock      FirstStockOutput         `json:"first_stock"`
	FirstStockItems []FirstStockItemResponse `json:"first_stock_items"`
}
