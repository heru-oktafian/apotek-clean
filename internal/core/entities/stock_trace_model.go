package models

import "time"

type StockTrace struct {
	ID            uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	ProductID     string    `gorm:"type:varchar(15);not null;index" json:"product_id"`
	Location      string    `gorm:"type:varchar(20);not null" json:"location"` // "showcase" or "warehouse" or "total"
	MutationType  string    `gorm:"type:varchar(30);not null" json:"mutation_type"`
	QtyChange     int       `gorm:"type:int;not null" json:"qty_change"`
	BalanceAfter  int       `gorm:"type:int;not null" json:"balance_after"`
	ReferenceID   string    `gorm:"type:varchar(50)" json:"reference_id"`
	ReferenceType string    `gorm:"type:varchar(30)" json:"reference_type"`
	Note          string    `gorm:"type:text" json:"note"`
	UserID        string    `gorm:"type:varchar(15);not null" json:"user_id"`
	BranchID      string    `gorm:"type:varchar(15);not null;index" json:"branch_id"`
	CreatedAt     time.Time `json:"created_at"`
}
