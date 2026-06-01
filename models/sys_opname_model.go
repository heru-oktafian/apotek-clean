package models

import time "time"

// Opnames model merepresentasikan tabel opnames
type Opnames struct {
	ID           string        `gorm:"type:varchar(15);primaryKey" json:"id" validate:"required"`
	Description  string        `gorm:"type:text;" json:"description"`
	OpnameDate   time.Time     `gorm:"not null" json:"opname_date" validate:"required"`
	BranchID     string        `gorm:"type:varchar(15);not null" json:"branch_id" validate:"required"`
	TotalOpname  int           `gorm:"type:int;not null;default:0" json:"total_opname" validate:"required"`
	Payment      PaymentStatus `gorm:"type:payment_status;not null;default:'opname'" json:"payment" validate:"required"`
	UserID       string        `gorm:"type:varchar(15);not null" json:"user_id" validate:"required"`
	CreatedAt    time.Time     `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt    time.Time     `gorm:"autoUpdateTime" json:"updated_at"`
	OpnameStatus DataStatus    `gorm:"type:data_status;not null;default:'active'" json:"opname_status" validate:"required"`
}

// All Opnames model merepresentasikan tabel opnames dengan join ke tabel branches
type AllOpnames struct {
	ID          string `gorm:"type:varchar(15);primaryKey" json:"id" validate:"required"`
	Description string `gorm:"type:text;" json:"description"`
	OpnameDate  string `gorm:"type:text;" json:"opname_date"`
	TotalOpname int    `gorm:"type:int;not null;default:0" json:"total_opname" validate:"required"`
}

// All Opname mobiles model merepresentasikan tabel opnames dengan join ke tabel branches
type AllOpnameMobiles struct {
	ID          string `gorm:"type:varchar(15);primaryKey" json:"id" validate:"required"`
	Description string `gorm:"type:text;" json:"description"`
	OpnameDate  string `gorm:"type:text;" json:"opname_date"`
	TotalOpname string `gorm:"type:text;" json:"total_opname"`
}

// Opname Items model merepresentasikan tabel opname_items
type OpnameItems struct {
	ID            string    `gorm:"type:varchar(15);primaryKey" json:"id" validate:"required"`
	OpnameId      string    `gorm:"type:varchar(15);not null" json:"opname_id" validate:"required"`
	ProductId     string    `gorm:"type:varchar(15);not null" json:"product_id" validate:"required"`
	Price         int       `gorm:"type:int;not null;default:0" json:"price" validate:"required"`
	Qty           int       `gorm:"type:int;not null;default:0" json:"qty" validate:"required"`
	QtyExist      int       `gorm:"type:int;not null;default:0" json:"qty_exist" validate:"required"`
	ExpiredDate   time.Time `gorm:"not null;default:(NOW() + interval '2 year')" json:"expired_date" validate:"required"`
	SubTotal      int       `gorm:"type:int;not null;default:0" json:"sub_total" validate:"required"`
	SubTotalExist int       `gorm:"type:int;not null;default:0" json:"sub_total_exist" validate:"required"`
}

// All Opname Items model merepresentasikan tabel opname_items dengan join ke tabel products
type AllOpnameItems struct {
	ID            string `gorm:"type:varchar(15);primaryKey" json:"id" validate:"required"`
	OpnameId      string `gorm:"type:varchar(15);not null" json:"opname_id" validate:"required"`
	ProductId     string `gorm:"type:varchar(15);not null" json:"product_id" validate:"required"`
	ProductName   string `gorm:"type:varchar(255);not null" json:"product_name" validate:"required"`
	Price         int    `gorm:"type:int;not null;default:0" json:"price" validate:"required"`
	Qty           int    `gorm:"type:int;not null;default:0" json:"qty" validate:"required"`
	QtyExist      int    `gorm:"type:int;not null;default:0" json:"qty_exist" validate:"required"`
	SubTotal      int    `gorm:"type:int;not null;default:0" json:"sub_total" validate:"required"`
	SubTotalExist int    `gorm:"type:int;not null;default:0" json:"sub_total_exist" validate:"required"`
}

// All Opname Items mobile model merepresentasikan tabel opname_items dengan join ke tabel products
type AllOpnameItemMobiles struct {
	ID            string `gorm:"type:varchar(15);primaryKey" json:"id"`
	OpnameId      string `gorm:"type:varchar(15);not null" json:"opname_id"`
	ProductId     string `gorm:"type:varchar(15);not null" json:"product_id"`
	ProductName   string `gorm:"type:varchar(255);not null" json:"product_name"`
	Price         string `gorm:"type:varchar(255);not null" json:"price"`
	Qty           int    `gorm:"type:int;not null;default:0" json:"qty"`
	QtyExist      int    `gorm:"type:int;not null;default:0" json:"qty_exist"`
	SubTotal      string `gorm:"type:varchar(255);not null" json:"sub_total"`
	SubTotalExist string `gorm:"type:varchar(255);not null" json:"sub_total_exist"`
	ExpiredDate   string `gorm:"column:expired_date" json:"expired_date"`
}

// All Opname Item details model merepresentasikan tabel opname_items dengan join ke tabel products
type AllOpnameItemDetails struct {
	ID            string `gorm:"type:varchar(15);primaryKey" json:"id" validate:"required"`
	OpnameId      string `gorm:"type:varchar(15);not null" json:"opname_id" validate:"required"`
	ProductId     string `gorm:"type:varchar(15);not null" json:"product_id" validate:"required"`
	ProductName   string `gorm:"type:varchar(255);not null" json:"product_name" validate:"required"`
	Price         int    `gorm:"type:int;not null;default:0" json:"price" validate:"required"`
	QtyAdjustment int    `gorm:"type:int;not null;default:0" json:"qty_adjustment" validate:"required"`
	SubAdjustment int    `gorm:"type:int;not null;default:0" json:"sub_adjustment" validate:"required"`
	ExpiredDate   string `gorm:"type:varchar(255);not null" json:"expired_date" validate:"required"`
}

// OpnameQueryResult model merepresentasikan hasil query opname
type OpnameQueryResult struct {
	ID          string    `gorm:"column:id"`
	Description string    `gorm:"column:description"`
	OpnameDate  time.Time `gorm:"column:opname_date"`  // <--- Ini akan menerima tipe time.Time
	TotalOpname string    `gorm:"column:total_opname"` // Ini tetap string karena sudah diformat dari DB
}

// OpnameInput model merepresentasikan input opname
type OpnameInput struct {
	OpnameDate  string `json:"opname_date" validate:"required"`
	Description string `gorm:"type:text;" json:"description"`
	TotalOpname int    `gorm:"type:int;not null;default:0" json:"total_opname" validate:"required"`
	Payment     string `json:"payment"`
}

// Input struct untuk request body
type CreateOpnameItemInput struct {
	OpnameId    string `json:"opname_id" validate:"required"`
	ProductId   string `json:"product_id" validate:"required"`
	Qty         int    `json:"qty" validate:"required"`
	Price       int    `json:"price" validate:"required"`
	ExpiredDate string `json:"expired_date" validate:"required"`
}

// Input struct untuk request body
type CreateOpnameItemUpdate struct {
	OpnameId    string `json:"opname_id" validate:"required"`
	ProductId   string `json:"product_id" validate:"required"`
	Qty         int    `json:"qty" validate:"required"`
	Price       int    `json:"price" validate:"required"`
	ExpiredDate string `json:"expired_date" validate:"required"`
}

// Response menampilkan satu opname beserta semua item-nya
type ResponseOpnameWithItemsResponse struct {
	Status      string      `json:"status"`
	Message     string      `json:"message"`
	OpnameId    string      `json:"opname_id"`
	Description string      `json:"description"`
	OpnameDate  string      `json:"opname_date"`
	TotalOpname int         `json:"total_opname"`
	Payment     string      `json:"payment"`
	Items       interface{} `json:"items"`
}
