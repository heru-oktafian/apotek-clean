package models

import (
	time "time"

	gorm "gorm.io/gorm"
)

type UserBranch struct {
	UserID    string `gorm:"type:varchar(15);primaryKey" json:"user_id" validate:"required"`
	BranchID  string `gorm:"type:varchar(15);primaryKey" json:"branch_id" validate:"required"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type UserBranchDetail struct {
	UserID     string `json:"user_id"`
	UserName   string `json:"user_name"`
	BranchID   string `json:"branch_id"`
	BranchName string `json:"branch_name"`
	SiaName    string `json:"sia_name"`
	SipaName   string `json:"sipa_name"`
	Phone      string `json:"phone"`
}
