package models

// Member model adalah model untuk member yang akan disimpan di database
type Member struct {
	ID               string `gorm:"type:varchar(15);primaryKey" json:"id" validate:"required"`
	Name             string `gorm:"type:varchar(100);not null" json:"name" validate:"required"`
	Phone            string `gorm:"type:varchar(100);" json:"phone"`
	Address          string `gorm:"type:text;" json:"address"`
	MemberCategoryId uint   `gorm:"not null" json:"member_category_id" validate:"required"`
	Points           int    `gorm:"type:int;not null;default:0" json:"points" validate:"required"` // Ubah ini
	BranchID         string `gorm:"type:varchar(15);not null" json:"branch_id" validate:"required"`
}

// MemberDetail model adalah model untuk member yang akan ditampilkan di data detail
type MemberDetail struct {
	ID             string `gorm:"type:varchar(15);primaryKey" json:"id" validate:"required"`
	Name           string `gorm:"type:varchar(100);not null" json:"name" validate:"required"`
	Phone          string `gorm:"type:varchar(100);" json:"phone"`
	Address        string `gorm:"type:text;" json:"address"`
	MemberCategory string `gorm:"type:varchar(100);not null" json:"member_category" validate:"required"`
	Points         int    `gorm:"type:int;not null;default:0" json:"points" validate:"required"` // Ubah ini
}

// ComboboxMembers adalah model untuk combo box member
type ComboboxMembers struct {
	MemberID   string `gorm:"type:varchar(15);primaryKey" json:"member_id" validate:"required"`
	MemberName string `gorm:"type:varchar(255);not null" json:"member_name" validate:"required"`
}
