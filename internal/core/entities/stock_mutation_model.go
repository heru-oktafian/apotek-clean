package models

import "time"

type StockMutation struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	ProductID string    `gorm:"type:varchar(15);not null;index" json:"product_id"`
	FromLoc   string    `gorm:"type:varchar(20);not null" json:"from_loc"` // "warehouse" or "showcase"
	ToLoc     string    `gorm:"type:varchar(20);not null" json:"to_loc"`   // "showcase" or "warehouse"
	Qty       int       `gorm:"type:int;not null" json:"qty"`
	Note      string    `gorm:"type:text" json:"note"`
	UserID    string    `gorm:"type:varchar(15);not null" json:"user_id"`
	BranchID  string    `gorm:"type:varchar(15);not null;index" json:"branch_id"`
	CreatedAt time.Time `json:"created_at"`
}
