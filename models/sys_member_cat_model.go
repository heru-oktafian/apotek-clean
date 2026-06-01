package models

// MemberCategory model adalah model untuk kategori member yang akan disimpan di database
type MemberCategory struct {
	ID                   uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	Name                 string `gorm:"type:varchar(100);not null" json:"name" validate:"required"`
	PointsConversionRate int    `gorm:"type:int;not null;default:0" json:"points_conversion_rate" validate:"required"`
	BranchID             string `gorm:"type:varchar(15);not null" json:"branch_id" validate:"required"`
}

// ComboMemberCategory model adalah model untuk kategori member yang akan ditampilkan di data combo
type ComboMemberCategory struct {
	MemberCategoryId   uint   `gorm:"primaryKey;autoIncrement" json:"member_category_id"`
	MemberCategoryName string `gorm:"type:varchar(100);not null" json:"member_category_name" validate:"required"`
}

// ComboboxMemberCategories adalah model untuk combo box kategori member
type ComboboxMemberCategories struct {
	MCID     uint   `gorm:"primaryKey;autoIncrement" json:"mc_id" validate:"required"`
	Category string `gorm:"type:varchar(255);not null" json:"category" validate:"required"`
}
