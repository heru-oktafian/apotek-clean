package models

import time "time"

// FirstStock Items model
type FirstStockItems struct {
	ID           string    `gorm:"type:varchar(15);primaryKey" json:"id" validate:"required"`
	FirstStockId string    `gorm:"type:varchar(15);not null" json:"first_stock_id" validate:"required"`
	ProductId    string    `gorm:"type:varchar(15);not null" json:"product_id" validate:"required"`
	Price        int       `gorm:"type:int;not null;default:0" json:"price" validate:"required"`
	Qty          int       `gorm:"type:int;not null;default:0" json:"qty" validate:"required"`
	SubTotal     int       `gorm:"type:int;not null;default:0" json:"sub_total" validate:"required"`
	ExpiredDate  time.Time `gorm:"not null;default:(NOW() + interval '2 year')" json:"expired_date" validate:"required"`
}

// All FirstStock Items model
type AllFirstStockItems struct {
	ID           string `gorm:"type:varchar(15);primaryKey" json:"id" validate:"required"`
	FirstStockId string `gorm:"type:varchar(15);not null" json:"first_stock_id" validate:"required"`
	ProductId    string `gorm:"type:varchar(15);not null" json:"product_id" validate:"required"`
	ProductName  string `gorm:"type:varchar(255);not null" json:"product_name" validate:"required"`
	Price        int    `gorm:"type:int;not null;default:0" json:"price" validate:"required"`
	Qty          int    `gorm:"type:int;not null;default:0" json:"qty" validate:"required"`
	UnitName     string `gorm:"type:varchar(255);not null" json:"unit_name" validate:"required"`
	SubTotal     int    `gorm:"type:int;not null;default:0" json:"sub_total" validate:"required"`
}

// Struct untuk input FirstStockItem
// Struct untuk input FirstStockItem
type FirstStockItemInput struct {
	ProductId   string `json:"product_id"`
	UnitId      string `json:"unit_id"`
	Qty         int    `json:"qty"`
	ExpiredDate string `json:"expired_date"`
}

// Struct untuk respons FirstStockItem
type FirstStockItemResponse struct {
	ID          string `json:"id"`
	ProductID   string `json:"product_id"`
	ProductName string `json:"product_name"` // Nama produk untuk respons
	UnitID      string `json:"unit_id"`
	UnitName    string `json:"unit_name"`    // Nama unit untuk respons
	Price       int    `json:"price"`        // Harga beli per unit dasar
	Qty         int    `json:"qty"`          // Qty dalam unit input
	SubTotal    int    `json:"sub_total"`    // SubTotal berdasarkan Price * Qty (nilai stok)
	ExpiredDate string `json:"expired_date"` // Format tanggal kedaluwarsa
}

// Struct untuk input FirstStock
// type FirstStockInput struct {
// 	Description    string `json:"description"`
// 	FirstStockDate string `json:"first_stock_date"` // String untuk parsing dari request
// }

// Response menampilkan satu first_stock beserta semua item-nya
type ResponseFirstStockWithItemsResponse struct {
	Status          string      `json:"status"`
	Message         string      `json:"message"`
	FirstStockId    string      `json:"first_stock_id"`
	Description     string      `json:"description"`
	FirstStockDate  string      `json:"first_stock_date"`
	TotalFirstStock int         `json:"total_first_stock"`
	Payment         string      `json:"payment"`
	Items           interface{} `json:"items"`
}
